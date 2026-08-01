package httpstatus

import (
	"net/http"

	"github.com/gambitier/go-pkgs/errors/domainerr"
)

// codeByStatus maps HTTP status codes to stable domain error codes.
// 4xx entries never map to INTERNAL so framework/client errors are not masked as 500.
var codeByStatus = map[int]domainerr.Code{
	http.StatusBadRequest:                   domainerr.CodeInvalidArgument,
	http.StatusUnauthorized:                 domainerr.CodeUnauthorized,
	http.StatusPaymentRequired:              domainerr.CodePaymentRequired,
	http.StatusForbidden:                    domainerr.CodeForbidden,
	http.StatusNotFound:                     domainerr.CodeNotFound,
	http.StatusMethodNotAllowed:             domainerr.CodeMethodNotAllowed,
	http.StatusNotAcceptable:                domainerr.CodeNotAcceptable,
	http.StatusProxyAuthRequired:            domainerr.CodeProxyAuthRequired,
	http.StatusRequestTimeout:               domainerr.CodeRequestTimeout,
	http.StatusConflict:                     domainerr.CodeConflict,
	http.StatusGone:                         domainerr.CodeGone,
	http.StatusLengthRequired:               domainerr.CodeLengthRequired,
	http.StatusPreconditionFailed:           domainerr.CodePreconditionFailed,
	http.StatusRequestEntityTooLarge:        domainerr.CodePayloadTooLarge,
	http.StatusRequestURITooLong:            domainerr.CodeURITooLong,
	http.StatusUnsupportedMediaType:         domainerr.CodeUnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable:   domainerr.CodeRangeNotSatisfiable,
	http.StatusExpectationFailed:            domainerr.CodeExpectationFailed,
	http.StatusTeapot:                       domainerr.CodeTeapot,
	http.StatusMisdirectedRequest:           domainerr.CodeMisdirectedRequest,
	http.StatusUnprocessableEntity:          domainerr.CodeUnprocessableEntity,
	http.StatusLocked:                       domainerr.CodeLocked,
	http.StatusFailedDependency:             domainerr.CodeFailedDependency,
	http.StatusTooEarly:                     domainerr.CodeTooEarly,
	http.StatusUpgradeRequired:              domainerr.CodeUpgradeRequired,
	http.StatusPreconditionRequired:         domainerr.CodePreconditionRequired,
	http.StatusTooManyRequests:              domainerr.CodeRateLimited,
	http.StatusRequestHeaderFieldsTooLarge:  domainerr.CodeRequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons:   domainerr.CodeUnavailableForLegalReasons,
	http.StatusInternalServerError:          domainerr.CodeInternal,
	http.StatusNotImplemented:               domainerr.CodeNotImplemented,
	http.StatusBadGateway:                   domainerr.CodeBadGateway,
	http.StatusServiceUnavailable:           domainerr.CodeServiceUnavailable,
	http.StatusGatewayTimeout:               domainerr.CodeGatewayTimeout,
	http.StatusHTTPVersionNotSupported:      domainerr.CodeHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates:        domainerr.CodeVariantAlsoNegotiates,
	http.StatusInsufficientStorage:          domainerr.CodeInsufficientStorage,
	http.StatusLoopDetected:                 domainerr.CodeLoopDetected,
	http.StatusNotExtended:                  domainerr.CodeNotExtended,
	http.StatusNetworkAuthenticationRequired: domainerr.CodeNetworkAuthenticationRequired,
}

// statusByCode is the inverse of codeByStatus for response serialization.
var statusByCode map[domainerr.Code]int

func init() {
	statusByCode = make(map[domainerr.Code]int, len(codeByStatus))
	for status, code := range codeByStatus {
		statusByCode[code] = status
	}
}

// CodeFromHTTPStatus returns a stable domain code for an HTTP status.
// 4xx never returns INTERNAL; unknown 4xx falls back to INVALID_ARGUMENT.
func CodeFromHTTPStatus(status int) domainerr.Code {
	if code, ok := codeByStatus[status]; ok {
		return code
	}
	if status >= 400 && status < 500 {
		return domainerr.CodeInvalidArgument
	}
	return domainerr.CodeInternal
}

// StatusFromCode returns the HTTP status for a domain code.
func StatusFromCode(code domainerr.Code) int {
	if status, ok := statusByCode[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// ToDomainError builds a domain error from an HTTP status and message.
// Used by Fiber ErrorHandler to convert framework errors into the standard envelope.
func ToDomainError(status int, message string, cause error) *domainerr.Error {
	code := CodeFromHTTPStatus(status)
	if message == "" {
		message = http.StatusText(status)
	}
	return domainerr.NewFromCode(code, message, cause, nil)
}
