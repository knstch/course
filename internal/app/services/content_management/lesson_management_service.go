package contentmanagement

import (
	"context"
	"mime/multipart"
	"strconv"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/app/validation"
	"github.com/knstch/course/internal/domain/entity"
	"golang.org/x/sync/errgroup"
)

// AddLesson используется для добавления урока, в качестве обязательного параметра принимается
// видео, название, название модуля, описание, позиция, название курса и превью. Далее метод валидирует
// параметры, проверяет, может ли такой урок быть создан и отправляет контент на CDN в асинхронном режиме.
// Метод возвращает ID урока или ошибку.
func (manager ContentManagementServcie) AddLesson(
	ctx context.Context,
	video *multipart.FileHeader,
	name string,
	moduleName string,
	description string,
	position string,
	courseName string,
	preview *multipart.FileHeader,
	previewFile *multipart.File,
) (*uint, *courseError.CourseError) {
	if err := validation.NewLessonToValidate(
		name, description, moduleName, preview.Filename, video.Filename, position, courseName,
	).Validate(ctx); err != nil {
		return nil, err
	}

	if err := manager.contentManager.CheckIfLessonCanBeCreated(ctx, name, moduleName, position, courseName); err != nil {
		return nil, err
	}

	var (
		videoPath    *string
		sendVideoErr *courseError.CourseError

		previewPath  *string
		sendPhotoErr *courseError.CourseError
	)

	g, errGroupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		videoPath, sendVideoErr = manager.sendVideo(errGroupCtx, video)
		if sendVideoErr != nil {
			return sendVideoErr.Error
		}
		return nil
	})

	g.Go(func() error {
		readyPreviewFileName := manager.prepareFileName(preview.Filename)
		previewPath, sendPhotoErr = manager.sendPhoto(previewFile, readyPreviewFileName)
		if sendPhotoErr != nil {
			return sendPhotoErr.Error
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		if sendPhotoErr != nil {
			return nil, sendPhotoErr
		}
		if sendVideoErr != nil {
			return nil, sendVideoErr
		}
	}

	lessonId, err := manager.contentManager.CreateLesson(ctx, name, moduleName, description, position, *videoPath, *previewPath)
	if err != nil {
		return nil, err
	}

	return lessonId, nil
}

// GetLessonsInfo используется для получения уроков по фильтрам. В качестве параметров принимает название урока, описание
// название модуля, название курса, страницу и лимит. Последние 2 параметра являются обязательными. Метод валидирует переданные данные
// и если было передано название курса, то проверяет наличие этого курса у клиента. Если он был приобретен, то метод возвращает расширенный контент.
// Метод возвращает информацию об уроке или ошибку.
func (manager ContentManagementServcie) GetLessonsInfo(ctx context.Context,
	name, description, moduleName, courseName, page, limit string) (
	*entity.LessonsInfoWithPagination, *courseError.CourseError) {
	if err := validation.NewLessonsQueryToValidate(name, description, courseName, moduleName, page, limit).Validate(ctx); err != nil {
		return nil, err
	}

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	offset := pageInt * limitInt

	var isPurchased bool
	if ctx.Value("UserId") != nil && courseName != "" {
		userCourses, err := manager.contentManager.GetUserCourses(ctx)
		if err != nil {
			return nil, err
		}

		course, err := manager.contentManager.GetCourseByName(ctx, courseName)
		if err != nil {
			return nil, err
		}

		if course != nil {
			for _, v := range userCourses {
				if v.CourseId == course.ID {
					isPurchased = true
					break
				}
			}
		}
	}

	if ctx.Value("adminId") != nil {
		isPurchased = true
	}

	lessons, err := manager.contentManager.GetLessons(ctx, name, description, moduleName, courseName, limitInt, offset, isPurchased)
	if err != nil {
		return nil, err
	}

	return &entity.LessonsInfoWithPagination{
		Pagination: entity.Pagination{
			Page:       pageInt,
			Limit:      limitInt,
			TotalCount: len(lessons),
			PagesCount: len(lessons) / limitInt,
		},
		LessonInfo: lessons,
	}, nil
}

// ManageLesson используется для редактирования уроков. В качестве параметров принимает видео, название, описание,
// порядковый номер, ID урока (обязательный), превью, и 2 булевых параметра, указывающие на наличие переданного видео или превью.
// Метод валидирует параметры и вносит изменения. Возвращает ошибку.
func (manager ContentManagementServcie) ManageLesson(ctx context.Context,
	video *multipart.FileHeader,
	name string,
	description string,
	position string,
	lessonId string,
	preview *multipart.FileHeader,
	previewFile *multipart.File,
	videoNotExists bool,
	previewNotExists bool,
) *courseError.CourseError {
	if err := validation.NewEditLessonToValidate(name, description, position, lessonId).Validate(ctx); err != nil {
		return err
	}

	var (
		previewPath *string
		videoPath   *string
		err         *courseError.CourseError
	)

	if !previewNotExists {
		if err := validation.NewPreviewFileNameToValidate(preview.Filename).Validate(ctx); err != nil {
			return err
		}

		readyName := manager.prepareFileName(preview.Filename)

		previewPath, err = manager.sendPhoto(previewFile, readyName)
		if err != nil {
			return err
		}
	}

	if !videoNotExists {
		if err := validation.NewVideoFileNameToValidate(video.Filename).Validate(ctx); err != nil {
			return err
		}

		videoPath, err = manager.sendVideo(ctx, video)
		if err != nil {
			return err
		}
	}

	if err := manager.contentManager.EditLesson(ctx, name, description, position, lessonId, videoPath, previewPath); err != nil {
		return err
	}

	return nil
}

// RemoveLesson используется для удаления урока. В качестве обятательного параметра принимает ID урока, валидирует его, и удаляет.
// Возвращает ошибку.
func (manager ContentManagementServcie) RemoveLesson(ctx context.Context, lessonId string) *courseError.CourseError {
	if err := validation.NewStringIdToValidate(lessonId).Validate(ctx); err != nil {
		return err
	}

	if err := manager.contentManager.DeleteLesson(ctx, lessonId); err != nil {
		return err
	}

	return nil
}
