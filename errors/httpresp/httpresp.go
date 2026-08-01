package httpresp

import (
	"github.com/gambitier/go-pkgs/errors/domainerr"
	"github.com/gambitier/go-pkgs/errors/httpstatus"
)

const InternalServerErrorMessage = "internal server error"

// Response is the standard JSON error envelope returned by HTTP handlers.
type Response struct {
	ErrorCode string         `json:"errorCode"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
	RequestID string         `json:"requestId,omitempty"`
}

// Envelope is the standard API response wrapper.
type Envelope[T any] struct {
	Data      T         `json:"data,omitempty"`
	Error     *Response `json:"error,omitempty"`
	RequestID string    `json:"requestId,omitempty"`
}

// ErrorBuildResult contains all derived values needed by adapters.
type ErrorBuildResult struct {
	Status int
	Env    Envelope[struct{}]
	Code   domainerr.Code
	Msg    string
	Fields map[string]any
}

// BuildErrorEnvelope maps an error to status code + response envelope.
func BuildErrorEnvelope(err error, requestID string) (int, Envelope[struct{}], domainerr.Code, string, map[string]any) {
	res := BuildError(err, requestID)
	return res.Status, res.Env, res.Code, res.Msg, res.Fields
}

// BuildError maps an error to a response payload and metadata.
func BuildError(err error, requestID string) ErrorBuildResult {
	code := domainerr.CodeInternal
	msg := InternalServerErrorMessage
	var fields map[string]any

	if de, ok := domainerr.As(err); ok {
		code = de.Code
		// Never expose internal implementation messages to clients.
		if de.Code != domainerr.CodeInternal {
			msg = de.Message
		}
		fields = domainerr.CollectFields(err)
	}

	env := Envelope[struct{}]{
		Error: &Response{
			ErrorCode: string(code),
			Message:   msg,
			Fields:    fields,
			RequestID: requestID,
		},
		RequestID: requestID,
	}
	return ErrorBuildResult{
		Status: StatusFromCode(code),
		Env:    env,
		Code:   code,
		Msg:    msg,
		Fields: fields,
	}
}

// BuildOKEnvelope builds a 200 success envelope.
func BuildOKEnvelope[T any](data T, requestID string) Envelope[T] {
	return Envelope[T]{
		Data:      data,
		RequestID: requestID,
	}
}

// BuildCreatedEnvelope builds a 201 success envelope.
func BuildCreatedEnvelope[T any](data T, requestID string) Envelope[T] {
	return Envelope[T]{
		Data:      data,
		RequestID: requestID,
	}
}

func StatusFromCode(code domainerr.Code) int {
	return httpstatus.StatusFromCode(code)
}
