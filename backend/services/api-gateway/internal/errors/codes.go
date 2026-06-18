package gatewayerrors

const (
	CodeValidation         = "VALIDATION_ERROR"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeFailedPrecondition = "FAILED_PRECONDITION"
	CodeRateLimited        = "RATE_LIMITED"
	CodeRequestCancelled   = "REQUEST_CANCELLED"
	CodeTimeout            = "TIMEOUT"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeBadGateway         = "BAD_GATEWAY"
	CodeNotImplemented     = "NOT_IMPLEMENTED"
	CodeInternal           = "INTERNAL_ERROR"
)

const StatusClientClosedRequest = 499
