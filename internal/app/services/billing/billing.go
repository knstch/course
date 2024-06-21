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
	ErrInvoiceNotFound = errors.New("инвойс не найден")
)

type SberBillingService struct {
	apiHost     string
	accessToken string
	banker      Banker
	redis       *redis.Client
}

type Banker interface {
	CreateNewOrder(ctx context.Context, courseId, price uint, ruCard bool) (*dto.OrderEssentials, *courseError.CourseError)
	GetCourseCost(ctx context.Context, courseId uint) (*uint, *courseError.CourseError)
	SetInvoiceId(ctx context.Context, invoiceId, orderId uint) *courseError.CourseError
	ApprovePayment(ctx context.Context, invoiceId, hashedUserData string) (*string, *courseError.CourseError)
	DeleteOrder(ctx context.Context, invoiceId string) *courseError.CourseError
}

func NewSberBillingService(config *config.Config, banker Banker, redis *redis.Client) SberBillingService {
	return SberBillingService{
		apiHost:     config.SberApiHost,
		accessToken: config.SberAccessToken,
		banker:      banker,
		redis:       redis,
	}
}

func (billing SberBillingService) PlaceOrder(ctx context.Context, buyDetails *entity.BuyDetails) (*string, *courseError.CourseError) {
	if err := validation.NewPaymentCredentialsToValidate(buyDetails).Validate(ctx); err != nil {
		return nil, err
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

func (billing SberBillingService) sendInvoice(_ entity.InvoiceData) (*uint, *courseError.CourseError) {
	return billing.mockInvoiceNumGenerator(), nil
}

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

func (billing SberBillingService) getPayLink(invoiceId *uint, hashedUserData string) (*string, *courseError.CourseError) {
	if err := billing.redis.Set(hashedUserData, *invoiceId, 15*time.Minute).Err(); err != nil {
		return nil, courseError.CreateError(err, 10031)
	}

	payLink := fmt.Sprintf("%v/invoices/%v", billing.apiHost, *invoiceId)
	return &payLink, nil
}

func (billing SberBillingService) ConfirmPayment(ctx context.Context, hashedUserData string) (*string, *courseError.CourseError) {
	invoiceId, err := billing.redis.Get(hashedUserData).Result()
	if err != nil {
		return nil, courseError.CreateError(ErrInvoiceNotFound, 11004)
	}

	if err := billing.redis.Del(hashedUserData).Err(); err != nil {
		return nil, courseError.CreateError(err, 10033)
	}

	courseName, courseErr := billing.banker.ApprovePayment(ctx, invoiceId, hashedUserData)
	if courseErr != nil {
		return nil, courseErr
	}

	return courseName, nil
}

func (billing SberBillingService) FailPayment(ctx context.Context, hashedUserData string) *courseError.CourseError {
	invoiceId, err := billing.redis.Get(hashedUserData).Result()
	if err != nil {
		return courseError.CreateError(ErrInvoiceNotFound, 11004)
	}

	if err := billing.banker.DeleteOrder(ctx, invoiceId); err != nil {
		return err
	}

	if err := billing.redis.Del(hashedUserData).Err(); err != nil {
		return courseError.CreateError(err, 10033)
	}

	return nil
}
