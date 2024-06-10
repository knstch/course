package courseerror

type CourseError struct {
	Error   error  `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func CreateError(err error, code int) *CourseError {
	return &CourseError{
		Error:   err,
		Code:    code,
		Message: err.Error(),
	}
}
