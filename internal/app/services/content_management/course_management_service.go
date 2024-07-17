package contentmanagement

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
)

type CourseQueryParams struct {
	ID          string
	Name        string
	Description string
	Cost        string
	Discount    string
	Page        string
	Limit       string
}

// AddCourse используется для добавления курса и принимает как обязательные парамерты
// название, описание, цену, скидку и превью. Возвращает ошибку или ID курса. Метод
// валидирует параметры и превью, и отправляет его на CDN.
func (manager ContentManagementServcie) AddCourse(ctx context.Context,
	name, description, cost, discount string,
	formFileHeader *multipart.FileHeader, file *multipart.File) (*uint, *courseError.CourseError) {
	if err := validation.NewCourseToValidate(name, description, cost, discount, formFileHeader.Filename).Validate(ctx); err != nil {
		return nil, err
	}

	readyName := manager.prepareFileName(formFileHeader.Filename)

	path, err := manager.sendPhoto(file, readyName)
	if err != nil {
		return nil, err
	}

	id, err := manager.contentManager.CreateCourse(ctx, name, description, cost, discount, *path)
	if err != nil {
		return nil, err
	}

	return id, nil
}

// GetCourseInfo используется для получения курса по фильтрам, в качестве параметров принимает название курса, описание
// стоимость, скидка, они используются для поиска по фильтрам. В качестве обязательного параметра выступают страница и лимит.
// Если был передан ID, то все остальные параметры игнорируются и происходит проверка на наличие курса у клиента. Если он не приобретен,
// то возвращается только базовая информация без доступа к расширенному контенту. Метод возвращает массив курсов с пагинацией или ошибку.
func (manager ContentManagementServcie) GetCourseInfo(ctx context.Context, params *CourseQueryParams) (*entity.CourseInfoWithPagination, *courseError.CourseError) {
	var isCoursePurchased bool
	if params.ID != "" {
		if ctx.Value("UserId") == nil && ctx.Value("AdminId") == nil {
			return nil, courseError.CreateError(ErrUnautharizedAccess, 13004)
		}

		if err := validation.NewStringIdToValidate(params.ID).Validate(ctx); err != nil {
			return nil, err
		}

		courses, err := manager.contentManager.GetUserCourses(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range courses {
			if fmt.Sprint(v.CourseId) == params.ID {
				isCoursePurchased = true
			}
		}
	}

	if ctx.Value("adminId") != nil {
		isCoursePurchased = true
	}

	if err := validation.NewCourseQueryToValidate(
		params.Name,
		params.Description,
		params.Cost,
		params.Discount,
		params.Page,
		params.Limit,
	).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(params.Page)
	limitInt, _ := strconv.Atoi(params.Limit)

	offset := pageInt * limitInt

	courseInfo, err := manager.contentManager.GetCourse(ctx,
		params.ID,
		params.Name,
		params.Description,
		params.Cost,
		params.Discount,
		limitInt,
		offset,
		isCoursePurchased,
	)
	if err != nil {
		return nil, err
	}

	return &entity.CourseInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(courseInfo),
			PagesCount: len(courseInfo) / limitInt,
		},
		CourseInfo: courseInfo,
	}, nil
}

// ManageCourse используется для управления курсами, в качестве параметров принимает ID курса,
// название, описание, цену, скидку и превью. ID является обязательным. Далее метод валидирует параметры
// и вносит изменения в БД по ID.
func (manager ContentManagementServcie) ManageCourse(ctx context.Context,
	courseId, name, description, cost, discount string,
	formFileHeader *multipart.FileHeader,
	file *multipart.File, fileNotExists bool) *courseError.CourseError {
	if err := validation.NewEditCoruseToValidate(name, description, cost, discount, courseId).Validate(ctx); err != nil {
		return err
	}

	var (
		path         *string
		err          *courseError.CourseError
		uintCost     *uint
		uintDiscount *uint
	)

	if !fileNotExists {
		if err := validation.NewPreviewFileNameToValidate(formFileHeader.Filename).Validate(ctx); err != nil {
			return err
		}

		readyName := manager.prepareFileName(formFileHeader.Filename)

		path, err = manager.sendPhoto(file, readyName)
		if err != nil {
			return err
		}
	}

	if cost != "" {
		intCost, _ := strconv.Atoi(cost)
		bufferCost := uint(intCost)
		uintCost = &bufferCost
	}
	if discount != "" {
		intDiscount, _ := strconv.Atoi(discount)
		bufferDiscount := uint(intDiscount)
		uintCost = &bufferDiscount
	}

	if err := manager.contentManager.EditCourse(ctx, courseId, name, description, path, uintCost, uintDiscount); err != nil {
		return err
	}

	return nil
}

// ManageShowStatus используется для скрытия или открытия урока в общем доступе, в качестве параметров
// принимает ID курса, который является обязательным, и валидирует его. Не затрагивает уже купленный матерал. Возвращает ошибку.
func (manager ContentManagementServcie) ManageShowStatus(ctx context.Context, courseId string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(courseId).Validate(ctx); err != nil {
		return err
	}
	if err := manager.contentManager.ToggleHiddenStatus(ctx, courseId); err != nil {
		return err
	}

	return nil
}
