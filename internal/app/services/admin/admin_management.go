package admin

import (
	"context"
	"errors"
	"strconv"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

var (
	errBadLogin = errors.New("поле логин не может быть пустым")
)

// EraseAdmin используется для удаления админа и деактивации его токенов.
// В качестве параметра принимает логин, валидирует его и отбирает доступ к платформе.
func (admin AdminService) EraseAdmin(ctx context.Context, login string) *courseError.CourseError {
	if login == "" {
		return courseError.CreateError(errBadLogin, 400)
	}

	if err := admin.adminManager.RemoveAdmin(ctx, login); err != nil {
		return err
	}

	return nil
}

// ManageRole используется для изменения роли администратора. Принимает в качестве
// параметров логин и роль, далее валидирует их и вносит изменения в БД. Возвращает ошибку.
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

// RetreiveAdmins используется для поиска администраторов по фильтрам. Принимает в качестве
// параметра логин, роль, подключенную двойную аутентификацию, страницу и лимит.
// Далее валидирует параметры, подготавливает offset и получает данные по заданым фильтрам.
// Возвращает массив, содержащий список администраторов, данные пагинации и ошибку.
func (admin AdminService) RetreiveAdmins(ctx context.Context, login, role, auth, page, limit string) (*entity.AdminsInfoWithPagination, *courseError.CourseError) {
	if err := validation.CreateNewAdminQueryToValidate(login, role, auth, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := pageInt * limitInt

	admins, err := admin.adminManager.GetAdmins(ctx, login, role, auth, limitInt, offset)
	if err != nil {
		return nil, err
	}

	packedAdmins := make([]entity.Admin, 0, len(admins))
	for _, v := range admins {
		packedAdmins = append(packedAdmins, *entity.CovertDtoAdmin(&v))
	}

	return &entity.AdminsInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(admins),
			PagesCount: len(admins) / limitInt,
		},
		AdminInfo: packedAdmins,
	}, nil
}
