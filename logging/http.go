package logging

import (
	"errors"
	"net/http"
	"strings"
)

// HTTPErrorLog carries normalized HTTP error logging details.
// Callers (or app platform glue) should populate status/code/message/fields
// using their error package; this package stays independent of errors/.
type HTTPErrorLog struct {
	Method       string
	Path         string
	StatusCode   int
	ErrorCode    string
	ErrorMsg     string
	ErrorFields  map[string]any
	ErrorCause   string
	ErrorContext string
	ErrorSource  string
	ErrorFrames  []string
}

// ErrorFrameConfig is reserved for callers that supply pre-filtered frames.
type ErrorFrameConfig struct {
	IncludePrefixes []string
	StripPrefixes   []string
	MaxDepth        int
}

// HTTPErrorLogInput is metadata used when building an HTTPErrorLog without a
// domain-error mapper. Status/code/message should be provided by the caller.
type HTTPErrorLogInput struct {
	Method      string
	Path        string
	StatusCode  int
	ErrorCode   string
	ErrorMsg    string
	ErrorFields map[string]any
	FrameConfig ErrorFrameConfig
}

// BuildHTTPErrorLog builds an HTTPErrorLog from caller-supplied metadata.
// It derives ErrorCause via errors.Unwrap when err is non-nil; it does not
// import any sibling packages.
func BuildHTTPErrorLog(err error, input HTTPErrorLogInput) HTTPErrorLog {
	statusCode := input.StatusCode
	errorCode := strings.TrimSpace(input.ErrorCode)
	errorMsg := strings.TrimSpace(input.ErrorMsg)
	errorFields := input.ErrorFields

	if errorMsg == "" && err != nil {
		errorMsg = err.Error()
	}

	return HTTPErrorLog{
		Method:       strings.Clone(input.Method),
		Path:         strings.Clone(input.Path),
		StatusCode:   statusCode,
		ErrorCode:    errorCode,
		ErrorMsg:     errorMsg,
		ErrorFields:  errorFields,
		ErrorCause:   rootCauseMessage(err),
		ErrorContext: "",
		ErrorSource:  "",
		ErrorFrames:  nil,
	}
}

func rootCauseMessage(err error) string {
	if err == nil {
		return ""
	}
	for {
		next := errors.Unwrap(err)
		if next == nil {
			return err.Error()
		}
		err = next
	}
}

func resolveHTTPErrorLogMessage(err error, payload HTTPErrorLog) string {
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(payload.ErrorMsg)
}

// LogHTTPError logs an HTTP error with consistent structured fields.
// 5xx responses are logged at error level, and 4xx at warn level.
func LogHTTPError(logger Logger, err error, payload HTTPErrorLog) {
	if logger == nil {
		return
	}

	logMsg := resolveHTTPErrorLogMessage(err, payload)

	fields := Fields{
		"method":                    strings.Clone(payload.Method),
		"path":                      strings.Clone(payload.Path),
		"status_code":               payload.StatusCode,
		"http.response.status_code": payload.StatusCode,
		"error_code":                strings.Clone(payload.ErrorCode),
		"error_msg":                 strings.Clone(payload.ErrorMsg),
		"error": func() any {
			if err == nil {
				return nil
			}
			return err.Error()
		}(),
		"error_fields":  payload.ErrorFields,
		"error_cause":   strings.Clone(payload.ErrorCause),
		"error_context": strings.Clone(payload.ErrorContext),
		"error_source":  strings.Clone(payload.ErrorSource),
		"error_frames":  payload.ErrorFrames,
	}

	if payload.StatusCode >= http.StatusInternalServerError {
		logger.Error(logMsg, err, fields)
		return
	}
	logger.Warn(logMsg, fields)
}
