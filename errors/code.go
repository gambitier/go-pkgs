package errors

// Code is a stable, client-facing error code.
// It must not change once published (treat as API contract).
type Code string

const (
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeConflict        Code = "CONFLICT"
	CodeRateLimited     Code = "RATE_LIMITED"
	CodeInternal        Code = "INTERNAL"

	// Transport-level HTTP error codes (stable client-facing contract).
	CodePaymentRequired              Code = "PAYMENT_REQUIRED"
	CodeMethodNotAllowed             Code = "METHOD_NOT_ALLOWED"
	CodeNotAcceptable                Code = "NOT_ACCEPTABLE"
	CodeProxyAuthRequired            Code = "PROXY_AUTH_REQUIRED"
	CodeRequestTimeout               Code = "REQUEST_TIMEOUT"
	CodeGone                         Code = "GONE"
	CodeLengthRequired               Code = "LENGTH_REQUIRED"
	CodePreconditionFailed           Code = "PRECONDITION_FAILED"
	CodePayloadTooLarge              Code = "PAYLOAD_TOO_LARGE"
	CodeURITooLong                   Code = "URI_TOO_LONG"
	CodeUnsupportedMediaType         Code = "UNSUPPORTED_MEDIA_TYPE"
	CodeRangeNotSatisfiable          Code = "RANGE_NOT_SATISFIABLE"
	CodeExpectationFailed            Code = "EXPECTATION_FAILED"
	CodeTeapot                       Code = "I_AM_A_TEAPOT"
	CodeMisdirectedRequest           Code = "MISDIRECTED_REQUEST"
	CodeUnprocessableEntity          Code = "UNPROCESSABLE_ENTITY"
	CodeLocked                       Code = "LOCKED"
	CodeFailedDependency             Code = "FAILED_DEPENDENCY"
	CodeTooEarly                     Code = "TOO_EARLY"
	CodeUpgradeRequired              Code = "UPGRADE_REQUIRED"
	CodePreconditionRequired         Code = "PRECONDITION_REQUIRED"
	CodeRequestHeaderFieldsTooLarge  Code = "REQUEST_HEADER_FIELDS_TOO_LARGE"
	CodeUnavailableForLegalReasons   Code = "UNAVAILABLE_FOR_LEGAL_REASONS"
	CodeNotImplemented               Code = "NOT_IMPLEMENTED"
	CodeBadGateway                   Code = "BAD_GATEWAY"
	CodeServiceUnavailable           Code = "SERVICE_UNAVAILABLE"
	CodeGatewayTimeout               Code = "GATEWAY_TIMEOUT"
	CodeHTTPVersionNotSupported      Code = "HTTP_VERSION_NOT_SUPPORTED"
	CodeVariantAlsoNegotiates        Code = "VARIANT_ALSO_NEGOTIATES"
	CodeInsufficientStorage          Code = "INSUFFICIENT_STORAGE"
	CodeLoopDetected                 Code = "LOOP_DETECTED"
	CodeNotExtended                  Code = "NOT_EXTENDED"
	CodeNetworkAuthenticationRequired Code = "NETWORK_AUTHENTICATION_REQUIRED"
)
