package usermanagement

import (
	"context"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type UserManagementService struct {
	manager UserManager
}

type UserManager interface {
	GetAllUserData(ctx context.Context, firstName, surname, phoneNumber, email, active, isVerified, courseName, page, limit string) ([]entity.UserDataAdmin, *courseError.CourseError)
}

func NewUserManagementService(userManager UserManager) UserManagementService {
	return UserManagementService{
		manager: userManager,
	}
}

func (user UserManagementService) GetAllUserData(ctx context.Context,
	firstName, surname, phoneNumber, email, active, isVerified, courseName, page, limit string) ([]entity.UserDataAdmin, *courseError.CourseError) {

	if err := validation.NewUserFiltersToValidate(firstName, surname, phoneNumber, email, active, isVerified, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	userData, err := user.manager.GetAllUserData(ctx, firstName, surname, phoneNumber, email, active, isVerified, courseName, page, limit)
	if err != nil {
		return nil, err
	}

	return userData, nil
}
