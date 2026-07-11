package apierrors

type Code string

const (
	UNAUTHORIZED   Code = "UNAUTHORIZED"
	RATE_LIMITED   Code = "RATE_LIMITED"
	NOT_FOUND      Code = "NOT_FOUND"
	BAD_REQUEST    Code = "BAD_REQUEST"
	INTERNAL_ERROR Code = "INTERNAL_ERROR"
)

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

var (
	ErrUnauthorized      = New(UNAUTHORIZED, "unauthorized")
	ErrRateLimitExceeded = New(RATE_LIMITED, "too many requests")
)
