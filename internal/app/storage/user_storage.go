package storage

import (
	"context"
	"errors"
	"strings"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errBadPassword             = errors.New("старый пароль передан неверно")
	errOldAndNewPasswordsEqual = errors.New("новый и старый пароль не могут совпадать")

	errInvoiceNotFound    = errors.New("инвойс не найден")
	errOrderNotFound      = errors.New("заказ не найден")
	errBadUserCredentials = errors.New("данные не совпадают с инвойсом")
)

func (storage Storage) newUserProfileUpdate(firstName, surname string, phoneNumber int) map[string]interface{} {
	updates := make(map[string]interface{}, 3)

	updates["phone_number"] = phoneNumber
	updates["first_name"] = firstName
	updates["surname"] = surname

	return updates
}

func (storage Storage) FillUserProfile(ctx context.Context, firstName, surname string, phoneNumber int, userId string) *courseError.CourseError {
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

func (storage Storage) ChangePasssword(ctx context.Context, oldPassword, newPassword string, userId uint) *courseError.CourseError {
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

func (storage Storage) ChangeEmail(ctx context.Context, newEmail string, userId uint) *courseError.CourseError {
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

func (storage Storage) SetPhoto(ctx context.Context, path string) *courseError.CourseError {
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

func (storage Storage) RetreiveUserData(ctx context.Context) (*entity.UserData, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userData := entity.CreateNewUserData()

	userId := ctx.Value("userId").(uint)

	user := dto.CreateNewUser()
	if err := tx.Where("id = ?", userId).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errUserNotFound, 11101)
		}
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
	if err := tx.Joins("JOIN orders ON courses.id = orders.course_id").
		Joins("JOIN billings ON billings.order_id = orders.id").
		Where("orders.user_id = ? AND billings.paid = ?", userId, true).Find(&courses).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	var userOrders []dto.Order
	if err := tx.Where("user_id = ?", userId).Find(&userOrders).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	orderIds := dto.ExtractIds(userOrders, func(item interface{}) uint {
		return item.(dto.Order).ID
	})

	var userBilling []dto.Billing
	if err := tx.Where("id IN (?)", orderIds).Find(&userBilling).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	userData.AddCourses(courses, userOrders, userBilling)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return userData, nil
}

func (storage Storage) GetUserCourses(ctx context.Context) ([]dto.Order, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	courses := dto.NewUserCourses()

	if err := tx.Joins("JOIN billings b ON b.order_id = orders.id").
		Joins("JOIN orders o ON o.id = b.order_id").
		Joins("JOIN courses c ON c.id = o.course_id").
		Where("o.user_id = ? AND b.paid = ?", userId, true).
		Find(&courses).Error; err != nil {
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return courses, nil
}

func (storage Storage) GetCourseCost(ctx context.Context, courseId uint) (*uint, *courseError.CourseError) {
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

	if course.Discount != nil {
		costWithDiscount := course.Cost - *course.Discount
		return &costWithDiscount, nil
	}
	return &course.Cost, nil
}

func (storage Storage) DeactivateProfile(ctx context.Context) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	if err := tx.Table("users").Where("id = ?", userId).Update("active", false).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Table("access_tokens").Where("user_id = ?", userId).Update("available", false).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) SetWatchedStatus(ctx context.Context, lessonId uint) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	lesson := dto.CreateNewLesson()
	if err := tx.Where("id = ?", lessonId).First(&lesson).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errLessonNotExists, 13005)
		}
	}

	watchHistory := dto.CreateNewWatchHistory(lessonId, userId)
	if err := tx.Create(&watchHistory).Error; err != nil {
		return courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage Storage) checkWatchedLessons(ctx context.Context, courseId string) ([]dto.WatchHistory, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	userId := ctx.Value("userId").(uint)

	var watched []dto.WatchHistory
	if err := tx.Table("watch_histories").
		Joins("JOIN lessons ON watch_histories.lesson_id = lessons.id").
		Joins("JOIN modules ON lessons.module_id = modules.id").
		Where("modules.course_id = ?", courseId).
		Where("watch_histories.user_id = ?", userId).
		Find(&watched).Error; err != nil {
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return watched, nil
}
