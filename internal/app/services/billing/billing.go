package billing

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/knstch/course/internal/app/config"
	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

type SberBillingService struct {
	apiHost     string
	accessToken string
	banker      Banker
}

type Banker interface {
	CreateOrder(ctx context.Context, courseId, price uint, ruCard bool) (*dto.OrderEssentials, *courseError.CourseError)
	GetCourseCost(ctx context.Context, courseId uint) (*uint, *courseError.CourseError)
	SetInvoiceId(ctx context.Context, invoiceId, orderId uint) *courseError.CourseError
}

func NewSberBillingService(config *config.Config, banker Banker) SberBillingService {
	return SberBillingService{
		apiHost:     config.SberApiHost,
		accessToken: config.SberAccessToken,
		banker:      banker,
	}
}

func (billing SberBillingService) PlaceOrder(ctx context.Context, courseId uint, ruCard bool) (*string, *courseError.CourseError) {
	if err := validation.NewPaymentCredentialsToValidate(courseId, ruCard).Validate(ctx); err != nil {
		return nil, err
	}

	price, err := billing.banker.GetCourseCost(ctx, courseId)
	if err != nil {
		return nil, err
	}

	order, err := billing.banker.CreateOrder(ctx, courseId, *price, ruCard)
	if err != nil {
		return nil, err
	}

	userId := ctx.Value("userId").(uint)

	invoice := entity.CreateOrder(*order, int(courseId), fmt.Sprint(userId), 0)

	invoiceId, err := billing.sendInvoice(invoice)
	if err != nil {
		return nil, err
	}

	if err := billing.banker.SetInvoiceId(ctx, *invoiceId, order.OrderId); err != nil {
		return nil, err
	}

	linkToPay, err := billing.getPayLink(invoiceId)
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

func (billing SberBillingService) getPayLink(invoiceId *uint) (*string, *courseError.CourseError) {
	payLink := fmt.Sprintf("%v/invoices/%v", billing.apiHost, *invoiceId)
	return &payLink, nil
}
