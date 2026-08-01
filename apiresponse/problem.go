package apiresponse

import (
	"net/http"

	pkgerrors "github.com/gambitier/go-pkgs/errors"
)

// ContentTypeProblemJSON is the RFC 9457 media type for problem details.
const ContentTypeProblemJSON = "application/problem+json"

// DefaultProblemType is used when no specific problem type URI is registered.
// RFC 9457 allows "about:blank" with title carrying a human-readable summary.
const DefaultProblemType = "about:blank"

const internalDetail = "internal server error"

// Problem is an RFC 9457 problem details object.
// Extension members: Code (domain error code), Fields (structured context).
type Problem struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Code     string         `json:"code,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
}

// BuildOptions configures optional RFC members when building a Problem.
type BuildOptions struct {
	// Instance is a URI reference identifying this occurrence (often the request path).
	Instance string
	// Type overrides DefaultProblemType when non-empty.
	Type string
}

// BuildResult is the HTTP status and problem body to serialize.
type BuildResult struct {
	Status  int
	Problem Problem
}

// BuildProblem maps an error to an RFC 9457 problem details document.
func BuildProblem(err error, opts BuildOptions) BuildResult {
	code := pkgerrors.CodeInternal
	detail := internalDetail
	var fields map[string]any

	if de, ok := pkgerrors.As(err); ok {
		code = de.Code
		if de.Code != pkgerrors.CodeInternal {
			detail = de.Message
		}
		fields = pkgerrors.CollectFields(err)
	}

	status := StatusFromCode(code)
	problemType := DefaultProblemType
	if opts.Type != "" {
		problemType = opts.Type
	}

	return BuildResult{
		Status: status,
		Problem: Problem{
			Type:     problemType,
			Title:    http.StatusText(status),
			Status:   status,
			Detail:   detail,
			Instance: opts.Instance,
			Code:     string(code),
			Fields:   fields,
		},
	}
}
