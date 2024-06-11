package courseerror

type CourseError struct {
	Error   error  `json:"-"`
	Message string `json:"error"`
	Code    int    `json:"code"`
}

func CreateError(err error, code int) *CourseError {
	return &CourseError{
		Error:   err,
		Code:    code,
		Message: err.Error(),
	}
}
