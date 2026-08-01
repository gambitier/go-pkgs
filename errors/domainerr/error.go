package domainerr

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"

	cerrors "github.com/cockroachdb/errors"
)

// Error is the canonical application-layer error type.
type Error struct {
	Code    Code
	Message string // Constant message - never concatenated
	Err     error  // Internal error with stack trace preserved
	Fields  map[string]any
}

type StackFrame struct {
	File     string
	Line     int
	Function string
}

// Error returns the stable, client-safe Message only.
//
// Do not change this to include the wrapped cause: HTTP/API layers and
// websocket payloads rely on Error()/Message staying opaque. Use RootCause,
// CauseChain, ErrorContext, or LogFields when logging or debugging.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// Is allows errors.Is to match domain errors by stable Code.
func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	te, ok := target.(*Error)
	if !ok || te == nil {
		return false
	}
	return e.Code == te.Code
}

// newError creates a newError Error.
func newError(code Code, message string, err error, fields map[string]any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     wrapErrWithStack(err),
		Fields:  fields,
	}
}

func wrapErrWithStack(err error) error {
	if err == nil {
		return nil
	}
	return cerrors.WithStackDepth(err, 3)
}

// As returns the first *Error in the chain.
func As(err error) (*Error, bool) {
	var ue *Error
	if errors.As(err, &ue) && ue != nil {
		return ue, true
	}
	return nil, false
}

// Wrap preserves an existing *Error's code/message and adds internal context.
func Wrap(err error, safeMessage string) *Error {
	if err == nil {
		return nil
	}
	if ue, ok := As(err); ok {
		return &Error{
			Code:    ue.Code,
			Message: ue.Message,
			Err:     cerrors.WrapWithDepth(1, err, safeMessage),
			Fields:  ue.Fields,
		}
	}
	return &Error{
		Code:    CodeInternal,
		Message: safeMessage,
		Err:     cerrors.WithStackDepth(err, 1),
	}
}

// CauseChain returns the full error chain as "msg: cause: root".
func CauseChain(err error) string {
	if err == nil {
		return ""
	}
	var parts []string
	seen := map[string]struct{}{}
	for err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg != "" {
			if _, ok := seen[msg]; !ok {
				parts = append(parts, msg)
				seen[msg] = struct{}{}
			}
		}
		// Prefer cockroachdb unwrap for stack-preserving errors.
		// Fall back to stdlib errors.Unwrap for fmt.Errorf(...%w...) chains.
		if next := cerrors.Unwrap(err); next != nil {
			err = next
			continue
		}
		err = errors.Unwrap(err)
	}
	return strings.Join(parts, ": ")
}

// ErrorContext returns deduplicated wrap messages, excluding the public domain
// message and the deepest raw cause.
func ErrorContext(err error) string {
	if err == nil {
		return ""
	}
	domainMessage := ""
	if de, ok := As(err); ok && de != nil {
		domainMessage = strings.TrimSpace(de.Message)
	}
	leafMessage := RootCause(err)

	var parts []string
	seen := map[string]struct{}{}
	for err != nil {
		next := cerrors.Unwrap(err)
		if next == nil {
			next = errors.Unwrap(err)
		}
		if next != nil {
			for _, segment := range contextSegments(err.Error(), domainMessage, leafMessage) {
				if _, ok := seen[segment]; ok {
					continue
				}
				parts = append(parts, segment)
				seen[segment] = struct{}{}
			}
		}
		err = next
	}
	return strings.Join(parts, ": ")
}

// RootCause returns the deepest non-empty message in the chain that differs from
// the public domain message.
func RootCause(err error) string {
	domainMessage := ""
	if de, ok := As(err); ok && de != nil {
		domainMessage = strings.TrimSpace(de.Message)
	}
	return deepestLeafMessage(err, domainMessage)
}

// LogFields returns structured attributes for logging without changing Error().
//
// Keys:
//   - "error": CauseChain when it adds detail beyond Error(); otherwise Error()
//   - "error_cause": RootCause when it differs from Error()
//   - "error_context": ErrorContext when non-empty
//
// Callers should merge these into log fields. Prefer these helpers over
// err.Error() alone so wrapped infrastructure causes (e.g. S3 AccessDenied)
// are not dropped when the domain message is intentionally opaque.
func LogFields(err error) map[string]any {
	if err == nil {
		return nil
	}

	public := strings.TrimSpace(err.Error())
	chain := strings.TrimSpace(CauseChain(err))
	cause := strings.TrimSpace(RootCause(err))
	context := strings.TrimSpace(ErrorContext(err))

	out := make(map[string]any, 3)
	switch {
	case chain != "" && chain != public:
		out["error"] = chain
	case public != "":
		out["error"] = public
	}

	if cause != "" && cause != public {
		out["error_cause"] = cause
	}
	if context != "" {
		out["error_context"] = context
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// OneLineSource returns file:line function for the captured stack.
func OneLineSource(err error) string {
	frames := StackFrames(err)
	if len(frames) == 0 {
		return ""
	}
	return FormatStackFrame(frames[0])
}

// StackFrames returns concise chain-derived frames, ordered from deepest to
// outermost. The caller decides which frames are relevant for logging.
func StackFrames(err error) []StackFrame {
	if err == nil {
		return nil
	}

	carriers := make([]error, 0, 4)
	for current := err; current != nil; {
		carriers = append(carriers, stackCarrier(current))
		next := cerrors.Unwrap(current)
		if next == nil {
			next = errors.Unwrap(current)
		}
		current = next
	}

	frames := make([]StackFrame, 0, 8)
	seen := map[string]struct{}{}
	for i := len(carriers) - 1; i >= 0; i-- {
		if st := cerrors.GetReportableStackTrace(carriers[i]); st != nil {
			for frameIndex := len(st.Frames) - 1; frameIndex >= 0; frameIndex-- {
				fnName := st.Frames[frameIndex].Function
				if mod := st.Frames[frameIndex].Module; mod != "" && mod != "unknown" {
					fnName = mod + "." + fnName
				}
				frame := StackFrame{
					File:     st.Frames[frameIndex].Filename,
					Line:     st.Frames[frameIndex].Lineno,
					Function: fnName,
				}
				key := FormatStackFrame(frame)
				if _, exists := seen[key]; exists || key == "" {
					continue
				}
				frames = append(frames, frame)
				seen[key] = struct{}{}
			}
			continue
		}

		if file, line, fn, ok := cerrors.GetOneLineSource(carriers[i]); ok {
			frame := StackFrame{
				File:     file,
				Line:     line,
				Function: fn,
			}
			key := FormatStackFrame(frame)
			if _, exists := seen[key]; !exists && key != "" {
				frames = append(frames, frame)
				seen[key] = struct{}{}
			}
		}
	}
	if len(frames) > 0 {
		return frames
	}

	// Fallback to the current goroutine stack for non-cockroachdb errors.
	return trimDebugStack(string(debug.Stack()))
}

func FilterStackFrames(frames []StackFrame, includePrefixes []string) []StackFrame {
	if len(frames) == 0 || len(includePrefixes) == 0 {
		return nil
	}

	filtered := make([]StackFrame, 0, len(frames))
	for _, frame := range frames {
		for _, prefix := range includePrefixes {
			normalizedPrefix := strings.TrimSpace(prefix)
			if normalizedPrefix == "" {
				continue
			}
			if strings.HasPrefix(frame.File, normalizedPrefix) || strings.HasPrefix(frame.Function, normalizedPrefix) {
				filtered = append(filtered, frame)
				break
			}
		}
	}
	return filtered
}

func StripStackFramePrefixes(frames []StackFrame, stripPrefixes []string) []StackFrame {
	if len(frames) == 0 {
		return nil
	}

	normalizedPrefixes := make([]string, 0, len(stripPrefixes))
	for _, prefix := range stripPrefixes {
		normalized := normalizeStripPrefix(prefix)
		if normalized != "" {
			normalizedPrefixes = append(normalizedPrefixes, normalized)
		}
	}

	trimmed := make([]StackFrame, 0, len(frames))
	for _, frame := range frames {
		next := frame
		for _, prefix := range normalizedPrefixes {
			if stripped, ok := trimPathPrefix(frame.File, prefix); ok {
				next.File = stripped
				break
			}
		}
		trimmed = append(trimmed, next)
	}
	return trimmed
}

func FormatStackFrame(frame StackFrame) string {
	if strings.TrimSpace(frame.File) == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d %s", filepath.ToSlash(frame.File), frame.Line, frame.Function)
}

// StackTraceString returns a formatted multiline stack for logging.
func StackTraceString(err error) string {
	frames := StackFrames(err)
	if len(frames) == 0 {
		return ""
	}
	lines := make([]string, 0, len(frames))
	for _, frame := range frames {
		lines = append(lines, FormatStackFrame(frame))
	}
	return strings.Join(lines, "\n")
}

func stackCarrier(err error) error {
	if de, ok := As(err); ok && de != nil && de.Err != nil {
		return de.Err
	}
	return err
}

func normalizeStripPrefix(prefix string) string {
	normalized := filepath.ToSlash(strings.TrimSpace(prefix))
	if normalized == "" {
		return ""
	}
	if !strings.HasSuffix(normalized, "/") {
		normalized += "/"
	}
	return normalized
}

func trimPathPrefix(path string, prefix string) (string, bool) {
	normalizedPath := filepath.ToSlash(path)
	if strings.HasPrefix(normalizedPath, prefix) {
		return strings.TrimPrefix(normalizedPath, prefix), true
	}
	trimmedPrefix := strings.TrimSuffix(prefix, "/")
	if normalizedPath == trimmedPrefix {
		return "", true
	}
	return normalizedPath, false
}

func contextSegments(msg string, domainMessage string, leafMessage string) []string {
	if strings.TrimSpace(msg) == "" {
		return nil
	}

	rawSegments := strings.Split(msg, ": ")
	segments := make([]string, 0, len(rawSegments))
	for _, raw := range rawSegments {
		segment := strings.TrimSpace(raw)
		if segment == "" || segment == domainMessage || segment == leafMessage {
			continue
		}
		segments = append(segments, segment)
	}
	return segments
}

func deepestLeafMessage(err error, domainMessage string) string {
	last := ""
	for err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg != "" && msg != domainMessage {
			last = msg
		}
		if next := cerrors.Unwrap(err); next != nil {
			err = next
			continue
		}
		err = errors.Unwrap(err)
	}
	return last
}

func trimDebugStack(raw string) []StackFrame {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	frames := make([]StackFrame, 0, len(lines)/2)
	seen := map[string]struct{}{}
	for i := 0; i < len(lines)-1; i += 2 {
		fn := strings.TrimSpace(lines[i])
		loc := strings.TrimSpace(lines[i+1])
		if fn == "" || loc == "" || strings.HasPrefix(fn, "goroutine ") {
			continue
		}

		file := loc
		line := 0
		if idx := strings.Index(loc, " +"); idx != -1 {
			file = loc[:idx]
		}
		if idx := strings.LastIndex(file, ":"); idx != -1 {
			fmt.Sscanf(file[idx+1:], "%d", &line)
			file = file[:idx]
		}

		frame := StackFrame{
			File:     file,
			Line:     line,
			Function: fn,
		}
		key := FormatStackFrame(frame)
		if _, ok := seen[key]; ok || key == "" {
			continue
		}
		frames = append(frames, frame)
		seen[key] = struct{}{}
	}
	return frames
}
