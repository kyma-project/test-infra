package apperrors

import "fmt"

const (
	InternalError                 = 1
	NotFoundError                 = 2
	AlreadyExistsError            = 3
	WrongInputError               = 4
	UpstreamServerCallFailedError = 5
	AuthenticationFailedError     = 6
)

type AppError interface {
	Append(string, ...interface{}) AppError
	Code() int
	Desc() string
	Error() string
}

type appError struct {
	code        int
	description string
	message     string
}

func appErrorf(code int, desc, format string, a ...interface{}) AppError {
	return appError{code: code, description: desc, message: fmt.Sprintf(format, a...)}
}

//Internal - used for generating
func Internal(format string, a ...interface{}) AppError {
	return appErrorf(InternalError, "Internal application error", format, a...)
}

func NotFound(format string, a ...interface{}) AppError {
	return appErrorf(NotFoundError, "not found error", format, a...)
}

func AlreadyExists(format string, a ...interface{}) AppError {
	return appErrorf(AlreadyExistsError, "already exists error", format, a...)
}

func WrongInput(format string, a ...interface{}) AppError {
	return appErrorf(WrongInputError, "wrong input error", format, a...)
}

func UpstreamServerCallFailed(format string, a ...interface{}) AppError {
	return appErrorf(UpstreamServerCallFailedError, "upstream server call error", format, a...)
}

func AuthenticationFailed(format string, a ...interface{}) AppError {
	return appErrorf(AuthenticationFailedError, "authentication error", format, a...)
}

func (ae appError) Append(additionalFormat string, a ...interface{}) AppError {
	format := additionalFormat + ", " + ae.message
	return appErrorf(ae.code, ae.description, format, a...)
}

func (ae appError) Code() int {
	return ae.code
}

func (ae appError) Error() string {
	return ae.message
}

func (ae appError) String() string {
	return ae.message
}

func (ae appError) Desc() string {
	return ae.description
}
