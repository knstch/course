package email

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	courseError "github.com/knstch/course/internal/app/course_error"
)

// EmailService используется для отправки email.
type EmailService struct {
	redis *redis.Client
}

// NewEmailService - это билдер для EmailService.
func NewEmailService(redis *redis.Client) *EmailService {
	return &EmailService{
		redis: redis,
	}
}

// SendConfirmCode используется для отправки кода подтверждения. В качестве параметров принимает
// ID пользователя и почту для отправки. Возвращает ошибку.
func (email EmailService) SendConfirmCode(userId *uint, emailToSend *string) *courseError.CourseError {
	confirmCode := email.generateEmailConfirmCode()

	if err := email.sendConfirmEmail(confirmCode, emailToSend); err != nil {
		return err
	}

	if err := email.redis.Set(fmt.Sprint(*userId), confirmCode, 15*time.Minute).Err(); err != nil {
		return courseError.CreateError(err, 10031)
	}

	return nil
}

// generateEmailConfirmCode используется для генерации кода подтверждения.
func (email EmailService) generateEmailConfirmCode() uint {
	return 1111
}

// UNIMPLEMENTED
// sendConfirmEmail отправляет код подтверждения на почту, принимает в качестве параметров код и почту для отправки.
// Возвращает ошибку.
func (email EmailService) sendConfirmEmail(code uint, emailToSend *string) *courseError.CourseError {
	return nil
}

// SendPasswordRecoverConfirmCode используется для отправки кода подтверждения на почту для изменения пароля.
// Принимает в качестве параметра почту для отправки, возвращает ошибку.
func (email EmailService) SendPasswordRecoverConfirmCode(emailToSend string) *courseError.CourseError {
	confirmCode := email.generateEmailConfirmCode()

	if err := email.sendConfirmEmail(confirmCode, &emailToSend); err != nil {
		return err
	}

	if err := email.redis.Set(emailToSend, confirmCode, 15*time.Minute).Err(); err != nil {
		return courseError.CreateError(err, 10031)
	}

	return nil
}
