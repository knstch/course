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
	GetAllUsersData(ctx context.Context,
		firstName, surname, phoneNumber, email, active, isVerified, courseName, banned, page, limit string) (
		*entity.UserDataWithPagination, *courseError.CourseError)
	DisableUser(ctx context.Context, userId int) *courseError.CourseError
	GetAllUserDataById(ctx context.Context, id string) (*entity.UserDataAdmin, *courseError.CourseError)
	EnableUser(ctx context.Context, userId int) *courseError.CourseError
	DeleteUserProfilePhoto(ctx context.Context, id string) *courseError.CourseError
}

func NewUserManagementService(userManager UserManager) UserManagementService {
	return UserManagementService{
		manager: userManager,
	}
}

func (user UserManagementService) RetreiveUsersByFilters(ctx context.Context,
	firstName, surname, phoneNumber, email, active, isVerified, courseName, banned, page, limit string) (
	*entity.UserDataWithPagination, *courseError.CourseError) {

	if err := validation.NewUserFiltersToValidate(firstName, surname, phoneNumber, email, active, isVerified, banned, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	userData, err := user.manager.GetAllUsersData(ctx, firstName, surname, phoneNumber, email, active, isVerified, courseName, banned, page, limit)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (user UserManagementService) DeactivateUser(ctx context.Context, userId uint) *courseError.CourseError {
	if err := validation.NewIdToValidate(int(userId)).Validate(ctx); err != nil {
		return err
	}

	if err := user.manager.DisableUser(ctx, int(userId)); err != nil {
		return err
	}

	return nil
}

func (user UserManagementService) ActivateUser(ctx context.Context, userId uint) *courseError.CourseError {
	if err := validation.NewIdToValidate(int(userId)).Validate(ctx); err != nil {
		return err
	}

	if err := user.manager.EnableUser(ctx, int(userId)); err != nil {
		return err
	}

	return nil
}

func (user UserManagementService) RetreiveUserById(ctx context.Context, id string) (*entity.UserDataAdmin, *courseError.CourseError) {
	if err := validation.NewStringIdToValidate(id).Validate(ctx); err != nil {
		return nil, err
	}

	userData, err := user.manager.GetAllUserDataById(ctx, id)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (user UserManagementService) EraseUserProfilePhoto(ctx context.Context, id string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(id).Validate(ctx); err != nil {
		return err
	}

	if err := user.manager.DeleteUserProfilePhoto(ctx, id); err != nil {
		return err
	}

	return nil
}
