package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
	"gorm.io/gorm"
)

func (storage *Storage) GetAllUsersData(ctx context.Context,
	firstName, surname, phoneNumber, email, active, isVerified, courseName, page,
	limit string) (*entity.UserDataWithPagination, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	var (
		boolIsActive   bool
		boolIsVerified bool
	)

	users := dto.CreateNewUsers()

	if active == "true" {
		boolIsActive = true
	}

	if isVerified == "true" {
		boolIsVerified = true
	}

	query := tx.Model(&dto.User{})

	if firstName != "" {
		query.Where("LOWER(first_name) LIKE ?", fmt.Sprint("%"+strings.ToLower(firstName)+"%"))
	}

	if surname != "" {
		query.Where("LOWER(surname) LIKE ?", fmt.Sprint("%"+strings.ToLower(surname)+"%"))
	}

	if phoneNumber != "" {
		query.Where("LOWER(phone_number) LIKE ?", fmt.Sprint("%"+strings.ToLower(phoneNumber)+"%"))
	}

	if email != "" {
		query.Joins("JOIN credentials ON LOWER(credentials.email) LIKE ?", fmt.Sprint("%"+strings.ToLower(email)+"%")).
			Where("credentials_id = credentials.id")
	}

	if isVerified != "" {
		query.Joins("JOIN credentials ON credentials.verified = ?", boolIsVerified).
			Where("credentials_id = credentials.id")
	}

	if active != "" {
		query.Where("active = ?", boolIsActive)
	}

	if courseName != "" {
		query.Joins("JOIN courses ON LOWER(courses.name) LIKE ?", fmt.Sprint("%"+strings.ToLower(courseName)+"%")).
			Joins("JOIN orders ON orders.course_id = courses.id").
			Where("users.id = orders.user_id")
	}

	limitInt, _ := strconv.Atoi(limit)
	pageInt, _ := strconv.Atoi(page)

	offset := pageInt * limitInt

	if err := query.Limit(limitInt).Offset(offset).Find(&users).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	usersEntity := make([]entity.UserDataAdmin, 0, len(users))
	for _, v := range users {
		user := entity.CreateUserDataAdmin(v)

		credentials := dto.CreateNewCredentials()
		if err := tx.Where("id = ?", v.CredentialsId).First(&credentials).Error; err != nil {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
		user.AddCredentials(credentials)

		if v.PhotoId != nil {
			photo := dto.CreateNewPhoto()
			if err := tx.Where("id = ?", v.PhotoId).First(&photo).Error; err != nil {
				tx.Rollback()
				return nil, courseError.CreateError(err, 10002)
			}
			user.AddPhoto(photo)
		}

		userCourses := dto.CreateNewCourses()
		if err := tx.Joins("JOIN orders ON courses.id = orders.course_id").
			Where("orders.user_id = ?", v.ID).Find(&userCourses).Error; err != nil {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}

		var userOrders []dto.Order
		if err := tx.Where("user_id = ?", v.ID).Find(&userOrders).Error; err != nil {
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

		user.AddCourses(userCourses, userOrders, userBilling)

		usersEntity = append(usersEntity, *user)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &entity.UserDataWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(usersEntity),
			PagesCount: len(usersEntity) / limitInt,
		},
		Users: usersEntity,
	}, nil
}

func (storage *Storage) DisableUser(ctx context.Context, userId int) *courseError.CourseError {
	tx := storage.db.WithContext(ctx).Begin()

	if err := tx.Model(&dto.User{}).Where("id = ?", userId).Update("active", false).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Model(&dto.AccessToken{}).Where("user_id = ?", userId).Update("available", false).Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10003)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) GetAllUserDataById(ctx context.Context, id string) (*entity.UserDataAdmin, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	user := dto.CreateNewUser()

	if err := tx.Where("id = ?", id).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errUserNotFound, 11101)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	userEntity := entity.CreateUserDataAdmin(*user)

	credentials := dto.CreateNewCredentials()
	if err := tx.Where("id = ?", user.CredentialsId).First(&credentials).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}
	userEntity.AddCredentials(credentials)

	if user.PhotoId != nil {
		photo := dto.CreateNewPhoto()
		if err := tx.Where("id = ?", user.PhotoId).First(&photo).Error; err != nil {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
		userEntity.AddPhoto(photo)
	}

	userCourses := dto.CreateNewCourses()
	if err := tx.Joins("JOIN orders ON courses.id = orders.course_id").
		Where("orders.user_id = ?", id).Find(&userCourses).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	var userOrders []dto.Order
	if err := tx.Where("user_id = ?", id).Find(&userOrders).Error; err != nil {
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

	userEntity.AddCourses(userCourses, userOrders, userBilling)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return userEntity, nil
}
