package storage

import (
	"context"
	"strconv"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"github.com/knstch/course/internal/domain/entity"
)

func (storage *Storage) GetAllUserData(ctx context.Context, firstName, surname, phoneNumber, email, active, isVerified, courseName, page, limit string) ([]entity.UserDataAdmin, *courseError.CourseError) {
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
		query.Where("first_name = ?", firstName)
	}

	if surname != "" {
		query.Where("surname = ?", surname)
	}

	if phoneNumber != "" {
		query.Where("phone_number = ?", phoneNumber)
	}

	if email != "" {
		query.Joins("JOIN credentials ON credentials.email = ?", email).
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
		query.Joins("JOIN courses ON courses.name = ?", courseName).
			Joins("JOIN users_courses ON users_courses.course_id = courses.id").
			Where("users.id = users_courses.user_id")
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
		if err := tx.Joins("JOIN users_courses ON courses.id = users_courses.course_id").
			Where("users_courses.user_id = ?", v.ID).Find(&userCourses).Error; err != nil {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
		user.AddCourses(userCourses)

		usersEntity = append(usersEntity, *user)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return usersEntity, nil
}
