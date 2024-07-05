// billing содержит методы биллинга.
package billing

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	ErrInvoiceNotFound        = errors.New("инвойс не найден")
	ErrCourseAlreadyPurchased = errors.New("этот курс уже куплен")
)

// SberBillingService содержит данные для работы с API сбербанка, Redis клиент
// и методы для взаимодействия с БД.
type SberBillingService struct {
	apiHost     string
	accessToken string
	banker      Banker
	redis       *redis.Client
}

// Banker объединяет в себе методы для работы с биллингом.
type Banker interface {
	CreateNewOrder(ctx context.Context, courseId, price uint, ruCard bool) (*dto.OrderEssentials, *courseError.CourseError)
	GetCourseCost(ctx context.Context, courseId uint) (*uint, *courseError.CourseError)
	SetInvoiceId(ctx context.Context, invoiceId, orderId uint) *courseError.CourseError
	ApprovePayment(ctx context.Context, invoiceId, hashedUserData string) (*string, *courseError.CourseError)
	DeleteOrder(ctx context.Context, invoiceId string) *courseError.CourseError
	GetUserCourses(ctx context.Context) ([]dto.Order, *courseError.CourseError)
}

// NewSberBillingService - это билдер для сервиса биллинга.
func NewSberBillingService(config *config.Config, banker Banker, redis *redis.Client) SberBillingService {
	return SberBillingService{
		apiHost:     config.SberApiHost,
		accessToken: config.SberAccessToken,
		banker:      banker,
		redis:       redis,
	}
}

// PlaceOrder используется для размещения заказа пользователя. В качестве параметра принимает
// ID курса и страна платежного инструмента. Далее валидирует параметра, проверяет куплен ли уже этот
// курс у пользователя, если он куплен, то возвращаем ошибку. Потом сервис запрашивает цену курса
// формирует новый заказ, подготавливает инвойс и отправялет его в банк. Далее из ID пользователя и
// идентификатора заказа формируется хэш, который записывается в Redis и формируется ссылка на оплату для пользователя.
// Метод возвращает ссылку на оплату для пользователя и ошибку.
func (billing SberBillingService) PlaceOrder(ctx context.Context, buyDetails *entity.BuyDetails) (*string, *courseError.CourseError) {
	if err := validation.NewPaymentCredentialsToValidate(buyDetails).Validate(ctx); err != nil {
		return nil, err
	}

	userCourses, err := billing.banker.GetUserCourses(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range userCourses {
		if v.CourseId == buyDetails.CourseId {
			return nil, courseError.CreateError(ErrCourseAlreadyPurchased, 15004)
		}
	}

	price, err := billing.banker.GetCourseCost(ctx, buyDetails.CourseId)
	if err != nil {
		return nil, err
	}

	order, err := billing.banker.CreateNewOrder(ctx, buyDetails.CourseId, *price, buyDetails.IsRusCard)
	if err != nil {
		return nil, err
	}

	userId := ctx.Value("userId").(uint)

	invoice := entity.CreateOrder(*order, int(buyDetails.CourseId), fmt.Sprint(userId), 0)

	invoiceId, err := billing.sendInvoice(invoice)
	if err != nil {
		return nil, err
	}

	if err := billing.banker.SetInvoiceId(ctx, *invoiceId, order.OrderId); err != nil {
		return nil, err
	}

	userDataHash := md5.New()

	userDataHash.Write([]byte(fmt.Sprintf("%d%v", userId, order.Order)))

	hashedUserData := hex.EncodeToString(userDataHash.Sum(nil))

	linkToPay, err := billing.getPayLink(invoiceId, hashedUserData)
	if err != nil {
		return nil, err
	}

	return linkToPay, nil
}

// UNIMPLEMENTED
// sendInvoice отправляет инвойс в банк. Возвращает номер инвойса или ошибку.
func (billing SberBillingService) sendInvoice(_ entity.InvoiceData) (*uint, *courseError.CourseError) {
	return billing.mockInvoiceNumGenerator(), nil
}

// mockInvoiceNumGenerator генерирует mock номер инвойса и возвращает его.
func (billing SberBillingService) mockInvoiceNumGenerator() *uint {
	rand.NewSource(time.Now().UnixNano())

	randomNumbers := make([]string, 10)

	for i := 0; i < 10; i++ {
		randomNumbers[i] = fmt.Sprint(rand.Intn(100))
	}

	invoice := strings.Join(randomNumbers, "")

	invoiceInt, _ := strconv.Atoi(invoice)

	invoiceUint := uint(invoiceInt)

	return &invoiceUint
}

// UNIMPLEMENTED
// getPayLink подготавливает ссылку на оплату для пользователя. В качестве параметров принимает ID инвойса и
// зашэшированные данные заказа пользователя, записывает их в Redis и отправляет данные в банк,
// указывая в качестве return ссылки GET метод сервиса, принимающим в качестве параметра хэш из данных заказа пользователя.
// Возвращает ссылку на оплату для пользователя или ошибку.
func (billing SberBillingService) getPayLink(invoiceId *uint, hashedUserData string) (*string, *courseError.CourseError) {
	if err := billing.redis.Set(hashedUserData, *invoiceId, 15*time.Minute).Err(); err != nil {
		return nil, courseError.CreateError(err, 10031)
	}

	payLink := fmt.Sprintf("%v/invoices/%v", billing.apiHost, *invoiceId)
	return &payLink, nil
}

// ConfirmPayment используется для подтверждения оплаты пользователем. Принимает в качестве параметра
// захэшированные данные заказа пользователя, проверяет наличие такой записи в Redis. Если она была найдена
// и ID инвойса совпали с данными из БД, то оплата считается завершенной. Метод возвращает название курса или ошибку.
func (billing SberBillingService) ConfirmPayment(ctx context.Context, hashedUserData string) (*string, *courseError.CourseError) {
	invoiceId, err := billing.redis.Get(hashedUserData).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, courseError.CreateError(ErrInvoiceNotFound, 11004)
		}
		return nil, courseError.CreateError(err, 10030)
	}

	courseName, courseErr := billing.banker.ApprovePayment(ctx, invoiceId, hashedUserData)
	if courseErr != nil {
		return nil, courseErr
	}

	if err := billing.redis.Del(hashedUserData).Err(); err != nil {
		return nil, courseError.CreateError(err, 10033)
	}

	return courseName, nil
}

// FailPayment используется для отмены заказа. Принимает в качестве параметра зашэшированные данные заказа пользователя,
// проверяет по этому ключу данные в Redis. Если они были найдены, удаляет заказ и данные из Redis. Возвращает ошибку.
func (billing SberBillingService) FailPayment(ctx context.Context, hashedUserData string) *courseError.CourseError {
	invoiceId, err := billing.redis.Get(hashedUserData).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return courseError.CreateError(ErrInvoiceNotFound, 11004)
		}
		return courseError.CreateError(err, 10030)
	}

	if err := billing.banker.DeleteOrder(ctx, invoiceId); err != nil {
		return err
	}

	if err := billing.redis.Del(hashedUserData).Err(); err != nil {
		return courseError.CreateError(err, 10033)
	}

	return nil
}

// ChangeApiHost используется для изменения API хоста банка. В качестве
// параметра принимает ссылку на новый хост, валидирует ее и изменяет хост на новый. Возращает ошибку.
func (billing *SberBillingService) ChangeApiHost(ctx context.Context, apiHost string) *courseError.CourseError {
	if err := validation.ValidateWebsite(ctx, apiHost); err != nil {
		return err
	}

	billing.apiHost = apiHost

	return nil
}

// ChangeAccessToken используется для изменения API токена банка. В качестве параметра принимает токен,
// валидирует его и изменяет токен на новый. Возвращает ошибку.
func (billing *SberBillingService) ChangeAccessToken(ctx context.Context, token string) *courseError.CourseError {
	if err := validation.ValidateToken(ctx, token); err != nil {
		return err
	}

	billing.accessToken = token

	return nil
}
