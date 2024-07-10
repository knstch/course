// email содержит методы для работы с почтой.
package email

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
	courseError "github.com/knstch/course/internal/app/course_error"
	"google.golang.org/api/gmail/v1"
)

const (
	me = "me"
)

var (
	emailMessage = "From: kostyacherepanov1@gmail.com\r\nTo: %v\r\nSubject: %v\r\n\r\n%d"
)

// EmailService используется для отправки email.
type EmailService struct {
	redis *redis.Client
	gmail *gmail.Service
}

// NewEmailService - это билдер для EmailService.
func NewEmailService(redis *redis.Client, gmail *gmail.Service) *EmailService {
	return &EmailService{
		redis,
		gmail,
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
func (email EmailService) generateEmailConfirmCode() int {
	return rand.Intn(9000) + 1000
}

// sendConfirmEmail отправляет код подтверждения на почту, принимает в качестве параметров код и почту для отправки.
// Возвращает ошибку.
func (email EmailService) sendConfirmEmail(code int, userEmail *string) *courseError.CourseError {
	readyEmail := fmt.Sprintf(emailMessage, *userEmail, "код подтверждения", code)

	message := []byte(readyEmail)

	messageToSend := base64.URLEncoding.EncodeToString(message)

	resp, err := email.gmail.Users.Messages.Send(me, &gmail.Message{
		Raw: messageToSend,
	}).Do()

	if err != nil {
		return courseError.CreateError(err, 17001)
	}
	fmt.Println(resp.Raw)
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
