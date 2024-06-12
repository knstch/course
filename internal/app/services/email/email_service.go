package email

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	courseError "github.com/knstch/course/internal/app/course_error"
)

type EmailService struct {
	redis *redis.Client
}

func NewEmailService(redis *redis.Client) *EmailService {
	return &EmailService{
		redis: redis,
	}
}

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

func (email EmailService) generateEmailConfirmCode() uint {
	return 1111
}

func (email EmailService) sendConfirmEmail(code uint, emailToSend *string) *courseError.CourseError {
	if code == 1111 || emailToSend != nil {
		return nil
	}
	return nil
}
