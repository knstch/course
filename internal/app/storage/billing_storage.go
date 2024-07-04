package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/gorm"
)

func (storage Storage) CreateNewOrder(ctx context.Context, courseId, price uint, ruCard bool) (*dto.OrderEssentials, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	orderHash := md5.New()

	orderHash.Write([]byte(fmt.Sprintf("%d%d%d", userId, courseId, time.Now().Unix())))

	orderNum := hex.EncodeToString(orderHash.Sum(nil))

	order := dto.CreateNewOrder().AddCourseId(courseId).AddUserId(userId).AddOrder(orderNum)

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10001)
	}
	invoice := dto.NewPayment().AddOrderId(order.ID).AddRusCard().AddPrice(float64(price))
	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10001)
	}

	course := dto.CreateNewCourse()
	if err := tx.Where("id = ?", courseId).First(&course).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errCourseNotExists, 13003)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	credentials := dto.CreateNewCredentials()
	if err := tx.Joins("JOIN users ON users.id = ?", userId).
		Where("credentials.id = users.credentials_id").First(&credentials).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errUserNotFound, 11002)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	placedOrder := dto.NewOrderEssentials().
		AddOrderId(order.ID).
		AddOrder(orderNum).
		AddOrderDate(uint(order.CreatedAt.Unix())).
		AddExpDate(uint(order.CreatedAt.Add(15 * time.Minute).Unix())).
		AddAmountToPay(price).
		AddRusLang().
		AddPurpose(fmt.Sprintf("Покупка: %v", course.Name)).
		AddDefaultTaxSystem().
		AddEmail(credentials.Email).
		AddContactEmail()

	if ruCard {
		placedOrder.AddCurrencyRub()
	}

	return placedOrder, nil
}

func (storage Storage) SetInvoiceId(ctx context.Context, invoiceId, orderId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	if err := tx.Model(&dto.Billing{}).Where("order_id = ?", orderId).Update("invoice_id", invoiceId).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) ApprovePayment(ctx context.Context, invoiceId, hashedUserData string) (*string, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	bill := dto.NewPayment()
	if err := tx.Where("invoice_id = ?", invoiceId).First(&bill).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errInvoiceNotFound, 15001)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	order := dto.CreateNewOrder()
	if err := tx.Where("id = ?", bill.OrderId).First(&order).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errOrderNotFound, 15002)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	userDataHash := md5.New()

	userDataHash.Write([]byte(fmt.Sprintf("%d%v", order.UserId, order.Order)))

	checkHashedUserData := hex.EncodeToString(userDataHash.Sum(nil))

	if checkHashedUserData != hashedUserData {
		tx.Rollback()
		return nil, courseError.CreateError(errBadUserCredentials, 15003)
	}

	if err := tx.Model(&dto.Billing{}).Where("id = ?", bill.ID).Update("paid", true).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(errBadUserCredentials, 10003)
	}

	course := dto.CreateNewCourse()
	if err := tx.Where("id = ?", order.CourseId).First(&course).Error; err != nil {
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &course.Name, nil
}

func (storage Storage) DeleteOrder(ctx context.Context, invoiceId string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	bill := dto.NewPayment()
	if err := tx.Where("invoice_id = ?", invoiceId).First(&bill).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errInvoiceNotFound, 15001)
		}
		return courseError.CreateError(err, 10002)
	}

	if err := tx.Model(&dto.Billing{}).Where("order_id = ?", bill.OrderId).Delete(nil).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10004)
	}

	if err := tx.Model(&dto.Order{}).Where("id = ?", bill.OrderId).Delete(nil).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10004)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
