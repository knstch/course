// email содержит методы для работы с почтой.
package email

import (
	"errors"
	"fmt"
	"math/rand"
	"net/smtp"
	"time"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
)

const (
	me           = "me"
	recover      = "recover"
	ConfirmEmail = "confirmEmail"
	confirm      = "confirm"

	recoverPasswordTitle = "Код для восстановления пароля"
	confirmEmailTitle    = "Код для подтверждения почты"

	emailSent = "sent"
)

var (
	emailMessage          = "From: %v\r\nTo: %v\r\nSubject: %v\r\n\r\n%d"
	errDoingAntispamCheck = errors.New("ошибка при проверке антиспам ключа")
	ErrEmailIsAlreadySent = errors.New("письмо уже было отправлено, подождите 1 минуту перед отправкой нового")
)

// EmailService используется для отправки email.
type EmailService struct {
	redis       *redis.Client
	smtpHost    string
	smptPort    string
	auth        smtp.Auth
	senderEmail string
}

// NewEmailService - это билдер для EmailService.
func NewEmailService(redis *redis.Client, config *config.Config) *EmailService {
	auth := smtp.PlainAuth(me, config.ServiceEmail, config.ServiceEmailPassword, config.SmtpHost)
	return &EmailService{
		redis,
		config.SmtpHost,
		config.SmtpPort,
		auth,
		config.ServiceEmail,
	}
}

// SendConfirmCode используется для отправки кода подтверждения. В качестве параметров принимает
// ID пользователя и почту для отправки. Возвращает ошибку.
func (email EmailService) SendConfirmCode(userId *uint, emailToSend *string, source string) *courseError.CourseError {
	antispamKey := fmt.Sprintf("%v%v", confirm, *emailToSend)
	antispamValue, err := email.redis.Get(antispamKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := email.redis.Set(antispamKey, emailSent, time.Minute).Err(); err != nil {
				return courseError.CreateError(err, 10031)
			}
		} else {
			return courseError.CreateError(errDoingAntispamCheck, 11004)
		}
	}

	if antispamValue != "" {
		return courseError.CreateError(ErrEmailIsAlreadySent, 17002)
	}

	confirmCode := email.generateEmailConfirmCode()

	if err := email.sendConfirmEmail(confirmCode, emailToSend, source); err != nil {
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
func (email EmailService) sendConfirmEmail(code int, userEmail *string, sourse string) *courseError.CourseError {
	readyEmail := fmt.Sprintf(emailMessage, email.senderEmail, *userEmail, confirmEmailTitle, code)
	if sourse == recover {
		readyEmail = fmt.Sprintf(emailMessage, email.senderEmail, *userEmail, recoverPasswordTitle, code)
	}

	if err := smtp.SendMail(fmt.Sprintf("%v:%v", email.smtpHost, email.smptPort), email.auth, email.senderEmail, []string{*userEmail}, []byte(readyEmail)); err != nil {
		return courseError.CreateError(err, 17001)
	}

	return nil
}

// SendPasswordRecoverConfirmCode используется для отправки кода подтверждения на почту для изменения пароля.
// Принимает в качестве параметра почту для отправки, возвращает ошибку.
func (email EmailService) SendPasswordRecoverConfirmCode(emailToSend string) *courseError.CourseError {
	antispamKey := fmt.Sprintf("%v%v", recover, emailToSend)
	antispamValue, err := email.redis.Get(antispamKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := email.redis.Set(antispamKey, emailSent, time.Minute).Err(); err != nil {
				return courseError.CreateError(err, 10031)
			}
		} else {
			return courseError.CreateError(errDoingAntispamCheck, 11004)
		}
	}

	if antispamValue != "" {
		return courseError.CreateError(ErrEmailIsAlreadySent, 17002)
	}

	confirmCode := email.generateEmailConfirmCode()

	if err := email.sendConfirmEmail(confirmCode, &emailToSend, recover); err != nil {
		return err
	}

	if err := email.redis.Set(emailToSend, confirmCode, 15*time.Minute).Err(); err != nil {
		return courseError.CreateError(err, 10031)
	}

	return nil
}
