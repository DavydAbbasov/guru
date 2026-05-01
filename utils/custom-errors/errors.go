package customerrors

import "errors"

type Error struct {
	Code    string
	Message error
}

func NewError(err error) Error {
	code := ""
	if err != nil {
		code = err.Error()
	}
	return Error{Code: code, Message: err}
}

func NewCodedError(code string, err error) Error {
	return Error{Code: code, Message: err}
}

func (e Error) Error() string {
	if e.Message != nil {
		return e.Message.Error()
	}
	return e.Code
}

// Is compares by Code so wrapping via fmt.Errorf("%w", ...) keeps matching.
func (e Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}
	return e.Code != "" && e.Code == t.Code
}

func (e Error) Unwrap() error {
	return e.Message
}

func (e Error) ToCustomError() *CustomErr {
	if customErr, ok := errMapper[e.Code]; ok {
		return customErr
	}
	return InternalErr
}

type CustomErr struct {
	Err        error  `json:"-"`
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *CustomErr) Error() string {
	return e.Message
}

const (
	CodeInternal        = "INTERNAL_ERROR"
	CodeNotFound        = "NOT_FOUND"
	CodeAlreadyExists   = "ALREADY_EXISTS"
	CodeValidation      = "VALIDATION_FAILED"
	CodeDatabaseRequest = "DATABASE_REQUEST_FAILED"
)

var (
	ErrInternal        = Error{Code: CodeInternal, Message: errors.New("internal error")}
	ErrNotFound        = Error{Code: CodeNotFound, Message: errors.New("not found")}
	ErrAlreadyExists   = Error{Code: CodeAlreadyExists, Message: errors.New("already exists")}
	ErrValidation      = Error{Code: CodeValidation, Message: errors.New("validation failed")}
	ErrDatabaseRequest = Error{Code: CodeDatabaseRequest, Message: errors.New("failed database request")}
)
