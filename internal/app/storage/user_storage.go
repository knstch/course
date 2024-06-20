package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errBadPassword             = errors.New("старый пароль передан неверно")
	errOldAndNewPasswordsEqual = errors.New("новый и старый пароль не могут совпадать")
)

func (storage *Storage) newUserProfileUpdate(firstName, surname string, phoneNumber int) map[string]interface{} {
	updates := make(map[string]interface{}, 3)

	updates["phone_number"] = phoneNumber
	updates["first_name"] = firstName
	updates["surname"] = surname

	return updates
}

func (storage *Storage) FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	user := dto.CreateNewUser()

	if err := tx.Where("id = ?", userId).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(err, 11101)
		}
		return courseError.CreateError(err, 10002)
	}

	if err := tx.Model(&dto.User{}).Where("id = ?", userId).Updates(storage.newUserProfileUpdate(firstName, surname, phoneNumber)).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError {
	credentials := dto.CreateNewCredentials()

	tx := storage.db.WithContext(ctx).Begin()

	if err := tx.Joins("JOIN users ON users.id = ?", userId).
		Where("credentials.id = users.credentials_id").
		First(&credentials).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errUserNotFound, 11002)
		}
		return courseError.CreateError(err, 10002)
	}

	if oldPassword == newPassword {
		tx.Rollback()
		return courseError.CreateError(errOldAndNewPasswordsEqual, 11103)
	}

	if !storage.verifyPassword(credentials.Password, oldPassword) {
		tx.Rollback()
		return courseError.CreateError(errBadPassword, 11104)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword+storage.secret), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 11020)
	}

	if err := tx.Model(dto.Credentials{}).Where("id = ?", credentials.ID).Update("password", hashedPassword).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) ChangeEmail(ctx context.Context, newEmail string, userId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	oldCredentials := dto.CreateNewCredentials()

	if err := tx.Joins("JOIN users ON users.id = ?", userId).
		Where("credentials.id = users.credentials_id AND verified = ?", true).
		First(&oldCredentials).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 11002)
	}

	newCredentials := dto.CreateNewCredentials().
		AddPassword(oldCredentials.Password).
		AddEmail(newEmail).SetStatusUnverified()

	if err := tx.Create(&newCredentials).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return courseError.CreateError(errEmailIsBusy, 11001)
		}
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) SetPhoto(ctx context.Context, path string) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	photo := dto.CreateNewPhoto().AddPath(path)
	if err := tx.Create(&photo).Error; err != nil {
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Model(&dto.User{}).Where("id = ?", userId).Update("photo_id", photo.ID).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) RetreiveUserData(ctx context.Context) (*entity.UserData, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userData := entity.CreateNewUserData()

	userId := ctx.Value("userId").(uint)

	user := dto.CreateNewUser()
	if err := tx.Where("id = ?", userId).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}
	userData.AddFirstName(user.FirstName).
		AddSurname(user.Surname).
		AddPhoneNumber(user.PhoneNumber)

	photo := dto.CreateNewPhoto()
	if err := tx.Where("id = ?", user.PhotoId).First(&photo).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
	}
	userData.AddPhoto(photo.Path)

	credentials := dto.CreateNewCredentials()
	if err := tx.Where("id = ?", user.CredentialsId).First(&credentials).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}
	userData.AddEmail(credentials.Email)
	userData.AddEmailVerifiedStatus(credentials.Verified)

	courses := dto.CreateNewCourses()
	if err := tx.Joins("JOIN users_courses ON courses.id = users_courses.course_id").
		Where("users_courses.user_id = ?", userId).Find(&courses).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}
	userData.AddCourses(courses)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return userData, nil
}

func (storage *Storage) GetUserCourses(ctx context.Context) ([]dto.Order, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	courses := dto.NewUserCourses()

	if err := tx.Where("user_id = ? AND paid = true", userId).Find(&courses).Error; err != nil {
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return courses, nil
}

func (storage *Storage) CreateOrder(ctx context.Context, courseId, price uint, ruCard bool) (*dto.OrderEssentials, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	orderHash := md5.New()

	orderHash.Write([]byte(fmt.Sprintf("%d%d%d", userId, courseId, time.Now().Unix())))

	orderNum := hex.EncodeToString(orderHash.Sum(nil))

	order := dto.NewOrder().AddCourseId(courseId).AddUserId(userId).AddOrder(orderNum)

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

func (storage *Storage) GetCourseCost(ctx context.Context, courseId uint) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	course := dto.CreateNewCourse()
	if err := tx.Where("id = ?", courseId).First(&course).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	finalCost := course.Cost - *course.Discount

	return &finalCost, nil
}

// func (storage *Storage) ConfirmPayment(ctx context.Context)
func (storage *Storage) SetInvoiceId(ctx context.Context, invoiceId, orderId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	if err := tx.Where("order_id = ?", orderId).Update("invoice_id", invoiceId).Error; err != nil {
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}
