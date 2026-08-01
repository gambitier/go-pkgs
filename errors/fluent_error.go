package errors

type FluentError interface {
	error
	Message(message string) FluentError
	Err(err error) FluentError
	Fields(fields map[string]any) FluentError
}

// fluentError supports fluent construction of domain errors.
// Example:
//
//	ErrInvalidArgument().Message("invalid request").Err(err).Fields(map[string]any{"k":"v"})
type fluentError struct {
	value *Error
}

func InvalidArgumentError() FluentError { return newErrBuilder(CodeInvalidArgument) }
func NotFoundError() FluentError        { return newErrBuilder(CodeNotFound) }
func ConflictError() FluentError        { return newErrBuilder(CodeConflict) }
func RateLimitedError() FluentError     { return newErrBuilder(CodeRateLimited) }
func InternalError() FluentError        { return newErrBuilder(CodeInternal) }
func ErrUnauthorized() FluentError      { return newErrBuilder(CodeUnauthorized) }
func ErrForbidden() FluentError         { return newErrBuilder(CodeForbidden) }

func newErrBuilder(code Code) *fluentError {
	return &fluentError{
		value: &Error{Code: code},
	}
}

func (b *fluentError) Message(message string) FluentError {
	b.value.Message = message
	return b
}

func (b *fluentError) Error() string {
	return b.value.Error()
}

// Unwrap exposes the inner *Error so errors.As / As can locate it
// and surface the correct Code/Message to HTTP adapters.
func (b *fluentError) Unwrap() error {
	if b == nil {
		return nil
	}
	return b.value
}

func (b *fluentError) Err(err error) FluentError {
	b.value.Err = wrapErrWithStack(err)
	return b
}

func (b *fluentError) withFields(fields map[string]any) {
	b.value.Fields = fields
}

func (b *fluentError) Fields(fields map[string]any) FluentError {
	b.withFields(fields)
	return b
}
