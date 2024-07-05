// courseerror содержит функционал по выводу кастомных ошибок.
package courseerror

// CourseError - кастомная сущность для сбора информации об ошибки и ее вывода при ответе сервиса.
type CourseError struct {
	Error   error  `json:"-"`
	Message string `json:"error"`
	Code    int    `json:"code"`
}

// CreateError - это билдер для ошибки, возвращает готовую сущность CourseError.
func CreateError(err error, code int) *CourseError {
	return &CourseError{
		Error:   err,
		Code:    code,
		Message: err.Error(),
	}
}
