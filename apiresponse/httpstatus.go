package apiresponse

import (
	"net/http"

	pkgerrors "github.com/gambitier/go-pkgs/errors"
)

// codeByStatus maps HTTP status codes to stable domain error codes.
// 4xx entries never map to INTERNAL so framework/client errors are not masked as 500.
var codeByStatus = map[int]pkgerrors.Code{
	http.StatusBadRequest:                    pkgerrors.CodeInvalidArgument,
	http.StatusUnauthorized:                  pkgerrors.CodeUnauthorized,
	http.StatusPaymentRequired:               pkgerrors.CodePaymentRequired,
	http.StatusForbidden:                     pkgerrors.CodeForbidden,
	http.StatusNotFound:                      pkgerrors.CodeNotFound,
	http.StatusMethodNotAllowed:              pkgerrors.CodeMethodNotAllowed,
	http.StatusNotAcceptable:                 pkgerrors.CodeNotAcceptable,
	http.StatusProxyAuthRequired:             pkgerrors.CodeProxyAuthRequired,
	http.StatusRequestTimeout:                pkgerrors.CodeRequestTimeout,
	http.StatusConflict:                      pkgerrors.CodeConflict,
	http.StatusGone:                          pkgerrors.CodeGone,
	http.StatusLengthRequired:                pkgerrors.CodeLengthRequired,
	http.StatusPreconditionFailed:            pkgerrors.CodePreconditionFailed,
	http.StatusRequestEntityTooLarge:         pkgerrors.CodePayloadTooLarge,
	http.StatusRequestURITooLong:             pkgerrors.CodeURITooLong,
	http.StatusUnsupportedMediaType:          pkgerrors.CodeUnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable:  pkgerrors.CodeRangeNotSatisfiable,
	http.StatusExpectationFailed:             pkgerrors.CodeExpectationFailed,
	http.StatusTeapot:                        pkgerrors.CodeTeapot,
	http.StatusMisdirectedRequest:            pkgerrors.CodeMisdirectedRequest,
	http.StatusUnprocessableEntity:           pkgerrors.CodeUnprocessableEntity,
	http.StatusLocked:                        pkgerrors.CodeLocked,
	http.StatusFailedDependency:              pkgerrors.CodeFailedDependency,
	http.StatusTooEarly:                      pkgerrors.CodeTooEarly,
	http.StatusUpgradeRequired:               pkgerrors.CodeUpgradeRequired,
	http.StatusPreconditionRequired:          pkgerrors.CodePreconditionRequired,
	http.StatusTooManyRequests:               pkgerrors.CodeRateLimited,
	http.StatusRequestHeaderFieldsTooLarge:   pkgerrors.CodeRequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons:    pkgerrors.CodeUnavailableForLegalReasons,
	http.StatusInternalServerError:           pkgerrors.CodeInternal,
	http.StatusNotImplemented:                pkgerrors.CodeNotImplemented,
	http.StatusBadGateway:                    pkgerrors.CodeBadGateway,
	http.StatusServiceUnavailable:            pkgerrors.CodeServiceUnavailable,
	http.StatusGatewayTimeout:                pkgerrors.CodeGatewayTimeout,
	http.StatusHTTPVersionNotSupported:       pkgerrors.CodeHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates:         pkgerrors.CodeVariantAlsoNegotiates,
	http.StatusInsufficientStorage:           pkgerrors.CodeInsufficientStorage,
	http.StatusLoopDetected:                  pkgerrors.CodeLoopDetected,
	http.StatusNotExtended:                   pkgerrors.CodeNotExtended,
	http.StatusNetworkAuthenticationRequired: pkgerrors.CodeNetworkAuthenticationRequired,
}

// statusByCode is the inverse of codeByStatus for response serialization.
var statusByCode map[pkgerrors.Code]int

func init() {
	statusByCode = make(map[pkgerrors.Code]int, len(codeByStatus))
	for status, code := range codeByStatus {
		statusByCode[code] = status
	}
}

// CodeFromHTTPStatus returns a stable domain code for an HTTP status.
// 4xx never returns INTERNAL; unknown 4xx falls back to INVALID_ARGUMENT.
func CodeFromHTTPStatus(status int) pkgerrors.Code {
	if code, ok := codeByStatus[status]; ok {
		return code
	}
	if status >= 400 && status < 500 {
		return pkgerrors.CodeInvalidArgument
	}
	return pkgerrors.CodeInternal
}

// StatusFromCode returns the HTTP status for a domain code.
func StatusFromCode(code pkgerrors.Code) int {
	if status, ok := statusByCode[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// ToDomainError builds a domain error from an HTTP status and message.
func ToDomainError(status int, message string, cause error) *pkgerrors.Error {
	code := CodeFromHTTPStatus(status)
	if message == "" {
		message = http.StatusText(status)
	}
	return pkgerrors.NewFromCode(code, message, cause, nil)
}
