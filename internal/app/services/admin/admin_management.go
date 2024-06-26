package admin

import (
	"context"
	"errors"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
)

var (
	errBadLogin = errors.New("поле логин не может быть пустым")
)

func (admin AdminService) EraseAdmin(ctx context.Context, login string) *courseError.CourseError {
	if login == "" {
		return courseError.CreateError(errBadLogin, 400)
	}

	if err := admin.adminManager.RemoveAdmin(ctx, login); err != nil {
		return err
	}

	return nil
}

func (admin AdminService) ManageRole(ctx context.Context, login, role string) *courseError.CourseError {
	if login == "" {
		return courseError.CreateError(errBadLogin, 400)
	}

	if err := validation.CreateNewRoleToValidate(role).Validate(ctx); err != nil {
		return err
	}

	if err := admin.adminManager.ChangeRole(ctx, login, role); err != nil {
		return err
	}

	return nil
}
