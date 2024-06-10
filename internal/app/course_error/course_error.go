package courseerror

type CourseError struct {
	Error error
	Code  int
}

func CreateError(err error, code int) *CourseError {
	return &CourseError{
		Error: err,
		Code:  code,
	}
}
