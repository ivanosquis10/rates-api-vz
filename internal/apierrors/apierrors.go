package apierrors

type Code string

const (
	UNAUTHORIZED   Code = "UNAUTHORIZED"
	RATE_LIMITED   Code = "RATE_LIMITED"
	NOT_FOUND      Code = "NOT_FOUND"
	BAD_REQUEST    Code = "BAD_REQUEST"
	INTERNAL_ERROR Code = "INTERNAL_ERROR"
	PROVIDER_ERROR Code = "PROVIDER_ERROR"
)

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func NewProviderError(err error) *Error {
	return &Error{
		Code:    PROVIDER_ERROR,
		Message: err.Error(),
		Err:     err,
	}
}

var (
	ErrUnauthorized      = New(UNAUTHORIZED, "unauthorized")
	ErrRateLimitExceeded = New(RATE_LIMITED, "too many requests")
)
