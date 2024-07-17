package contentmanagement

import (
	"context"
	"strconv"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

// AddModule используется для добавления модуля, в качестве обязательного параметра принимает
// название, описание, название курса и порядковый номер модуля. Возвращает ID модуля или ошибку.
// Перед добавлением модуля параметры проходят валидацию.
func (manager ContentManagementServcie) AddModule(ctx context.Context, module *entity.Module) (*uint, *courseError.CourseError) {
	if err := validation.NewModuleToValidate(module.Name, module.Description, module.CourseName, *module.Position).
		Validate(ctx); err != nil {
		return nil, err
	}

	id, err := manager.contentManager.CreateModule(ctx, module.Name, module.Description, module.CourseName, *module.Position)
	if err != nil {
		return nil, err
	}

	return id, nil
}

// GetModulesInfo используется для получения модулей. Принимает в качестве параметров название, описание, название курса, страницу и лимит.
// Последние 2 параметра являются обязательными, остальные используются для поиска по фильтрам. Метод валидирует параметры и проверяет наличие
// купленного курса, если было передано название курса. Если этот курс был куплен пользователем, то возвращается расширенный контент. Метод
// возвращает данные модуля вместе с пагинацией или ошибку.
func (manager ContentManagementServcie) GetModulesInfo(ctx context.Context,
	name, description, courseName, page, limit string) (*entity.ModuleInfoWithPagination, *courseError.CourseError) {
	if err := validation.NewModuleQueryToValidate(name, description, courseName, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := pageInt * limitInt

	var isPurchased bool
	var err *courseError.CourseError
	if ctx.Value("UserId") != nil && courseName != "" {
		isPurchased, err = manager.checkCoursePurchase(ctx, courseName)
		if err != nil {
			return nil, err
		}
	}

	if ctx.Value("adminId") != nil {
		isPurchased = true
	}

	modules, err := manager.contentManager.GetModules(ctx, name, description, courseName, limitInt, offset, isPurchased)
	if err != nil {
		return nil, err
	}

	return &entity.ModuleInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(modules),
			PagesCount: len(modules) / limitInt,
		},
		ModuleInfo: modules,
	}, nil
}

// ManageModule используется для редактировании информации о модуле. Принимает в качестве параметров ID модуля (обязательн),
// навзание, описание, порядковый номер и название курса. Метод валидирует параметры и вносит изменения. Возвращает ошибку.
func (manager ContentManagementServcie) ManageModule(ctx context.Context, module *entity.Module) *courseError.CourseError {
	if err := validation.NewEditModuleToValidate(module.Name, module.Description, *module.Position, module.ModuleId).
		Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.EditModule(ctx, module.Name, module.Description, module.Position, module.ModuleId); err != nil {
		return err
	}

	return nil
}

// RemoveModule используется для удаления модуля. В качестве обятательного параметра принимает ID модуля, валидирует его, и удаляет.
// Возвращает ошибку.
func (manager ContentManagementServcie) RemoveModule(ctx context.Context, moduleId string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(moduleId).Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.DeleteModule(ctx, moduleId); err != nil {
		return err
	}

	return nil
}

func (manager ContentManagementServcie) checkCoursePurchase(ctx context.Context, courseName string) (bool, *courseError.CourseError) {
	userCourses, err := manager.contentManager.GetUserCourses(ctx)
	if err != nil {
		return false, err
	}

	course, err := manager.contentManager.GetCourseByName(ctx, courseName)
	if err != nil {
		return false, err
	}

	if course != nil {
		for _, v := range userCourses {
			if v.CourseId == course.ID {
				return true, nil
			}
		}
	}
	return false, nil
}
