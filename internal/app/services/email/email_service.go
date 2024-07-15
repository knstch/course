// email содержит методы для работы с почтой.
package email

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"strings"
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
	errInvalidEmail       = errors.New("передана несуществующая почта")
)

// EmailService используется для отправки email.
type EmailService struct {
	redis       *redis.Client
	smtpHost    string
	smptPort    string
	auth        smtp.Auth
	senderEmail string
	isTest      bool
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
		config.IsTest,
	}
}

// SendConfirmCode используется для отправки кода подтверждения. В качестве параметров принимает
// ID пользователя и почту для отправки. Возвращает ошибку.
func (email EmailService) SendConfirmCode(userId *uint, emailToSend *string, source string) *courseError.CourseError {
	antispamKey := fmt.Sprintf("%v:%v", confirm, *emailToSend)
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

	if email.isTest {
		if err := email.redis.Set(fmt.Sprint(*userId), 1111, 15*time.Minute).Err(); err != nil {
			return courseError.CreateError(err, 10031)
		}
		return nil
	}

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
	antispamKey := fmt.Sprintf("%v:%v", recover, emailToSend)
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

	if email.isTest {
		if err := email.redis.Set(emailToSend, 1111, 15*time.Minute).Err(); err != nil {
			return courseError.CreateError(err, 10031)
		}
		return nil
	}

	if err := email.sendConfirmEmail(confirmCode, &emailToSend, recover); err != nil {
		return err
	}

	if err := email.redis.Set(emailToSend, confirmCode, 15*time.Minute).Err(); err != nil {
		return courseError.CreateError(err, 10031)
	}

	return nil
}

func (email EmailService) ValidateEmail(emailToCheck string) *courseError.CourseError {
	if email.isTest {
		return nil
	}

	parts := strings.Split(emailToCheck, "@")

	mxRecords, err := net.LookupMX(parts[1])
	if err != nil {
		return courseError.CreateError(err, 17003)
	}

	client, err := email.smtpDialTimeout(mxRecords[0].Host, 3*time.Second)
	if err != nil {
		return courseError.CreateError(err, 17003)
	}
	defer client.Close()

	if err := client.Hello(parts[1]); err != nil {
		return courseError.CreateError(err, 17003)
	}

	err = client.Mail(emailToCheck)
	if err != nil {
		return courseError.CreateError(err, 17003)
	}

	err = client.Rcpt(emailToCheck)
	if err != nil {
		return courseError.CreateError(errInvalidEmail, 17004)
	}

	return nil
}

func (email EmailService) smtpDialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr+":25", timeout)
	if err != nil {
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}
