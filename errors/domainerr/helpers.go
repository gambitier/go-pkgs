package domainerr

import "errors"

// Use this when you have an error to wrap and fields to add
func InvalidArgument(message string, err error, fields map[string]any) *Error {
	return newError(CodeInvalidArgument, message, err, fields)
}

// Use this when you don't have an error to wrap
func InvalidArgumentWithFields(message string, fields map[string]any) *Error {
	return newError(CodeInvalidArgument, message, nil, fields)
}

// Use this when you have an error to wrap
func InvalidArgumentWithErr(message string, err error) *Error {
	return newError(CodeInvalidArgument, message, err, nil)
}

func NotFound(message string, err error, fields map[string]any) *Error {
	return newError(CodeNotFound, message, err, fields)
}

func Unauthorized(message string) *Error {
	return newError(CodeUnauthorized, message, nil, nil)
}

func Forbidden(message string) *Error {
	return newError(CodeForbidden, message, nil, nil)
}

func ForbiddenWithFields(message string, fields map[string]any) *Error {
	return newError(CodeForbidden, message, nil, fields)
}

func Conflict(message string, err error, fields map[string]any) *Error {
	return newError(CodeConflict, message, err, fields)
}

func RateLimited(message string, err error, fields map[string]any) *Error {
	return newError(CodeRateLimited, message, err, fields)
}

func Internal(message string, err error, fields map[string]any) *Error {
	return newError(CodeInternal, message, err, fields)
}

// NewFromCode creates a domain error for the given stable code.
func NewFromCode(code Code, message string, err error, fields map[string]any) *Error {
	return newError(code, message, err, fields)
}

func MethodNotAllowed(message string) *Error {
	return newError(CodeMethodNotAllowed, message, nil, nil)
}

func PayloadTooLarge(message string) *Error {
	return newError(CodePayloadTooLarge, message, nil, nil)
}

func UnsupportedMediaType(message string) *Error {
	return newError(CodeUnsupportedMediaType, message, nil, nil)
}

func UnprocessableEntity(message string, fields map[string]any) *Error {
	return newError(CodeUnprocessableEntity, message, nil, fields)
}

func NotImplemented(message string) *Error {
	return newError(CodeNotImplemented, message, nil, nil)
}

func ServiceUnavailable(message string) *Error {
	return newError(CodeServiceUnavailable, message, nil, nil)
}

// Is checks if an error in the chain is a domain error with the given code.
func Is(err error, code Code) bool {
	if err == nil {
		return false
	}
	if ue, ok := As(err); ok {
		return ue.Code == code
	}
	return false
}

// IsCode checks if an error in the chain is a domain error with the given code.
// This also supports errors.Is when target is a *domainerr.Error carrying only Code.
func IsCode(err error, code Code) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, &Error{Code: code}) || Is(err, code)
}

func CodeOf(err error) Code {
	if err == nil {
		return CodeInternal
	}
	if ue, ok := As(err); ok {
		return ue.Code
	}
	return CodeInternal
}
