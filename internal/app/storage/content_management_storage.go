package storage

import (
	"context"
	"errors"
	"strconv"

	courseError "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/dto"
	"gorm.io/gorm"
)

var (
	errModulePosAlreadyExists  = errors.New("модуль с такой позицией уже существует")
	errModuleNameAlreadyExists = errors.New("модуль с таким названием уже существует")
	errLessonNameAlreadyExists = errors.New("урок с таким назавнием уже существует")
	errLessonPosAlreadyExists  = errors.New("урок с такой позицией уже существует")

	errCourseNotExists = errors.New("курса с таким названием не существует")
	errModuleNotExists = errors.New("модуль с таким названием не существует")
)

func (storage *Storage) CreateCourse(ctx context.Context, name, description, cost, discount, path string) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	course := dto.CreateNewCourse().
		AddName(name).
		AddDescription(description).
		AddCost(cost).
		AddDiscount(discount).
		AddPreviewImg(path)

	if err := tx.Create(&course).Error; err != nil {
		return nil, courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &course.ID, nil
}

func (storage *Storage) CreateModule(ctx context.Context, name, description, courseName string, position uint) (*uint, *courseError.CourseError) {
	var (
		moduleWithThisPosNotExists  bool
		moduleWithThisNameNotExists bool
	)

	tx := storage.db.WithContext(ctx).Begin()

	module := dto.CreateNewModule()

	if err := storage.db.Joins("JOIN courses ON courses.name = ?", courseName).Where("course_id = courses.id AND position = ?", position).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			moduleWithThisPosNotExists = true
		} else {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
	}

	if !moduleWithThisPosNotExists {
		tx.Rollback()
		return nil, courseError.CreateError(errModulePosAlreadyExists, 13001)
	}

	if err := storage.db.Joins("JOIN courses ON courses.name = ?", courseName).Where("course_id = courses.id AND modules.name = ?", name).First(&module).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			moduleWithThisNameNotExists = true
		} else {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
	}

	if !moduleWithThisNameNotExists {
		tx.Rollback()
		return nil, courseError.CreateError(errModuleNameAlreadyExists, 13001)
	}

	course := dto.CreateNewCourse()
	if err := storage.db.Where("name = ?", courseName).First(&course).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errCourseNotExists, 13003)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	module.AddCourseId(course.ID).
		AddName(name).
		AddDescription(description).
		AddPosition(position)

	if err := storage.db.Create(&module).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &module.ID, nil
}

func (storage *Storage) CheckIfLessonCanBeCreated(ctx context.Context, name, moduleName, position string) *courseError.CourseError {
	var (
		lessonWithThisPosNotExists  bool
		lessonWithThisNameNotExists bool
	)

	tx := storage.db.WithContext(ctx).Begin()

	module := dto.CreateNewModule()
	if err := tx.Where("name = ?", moduleName).First(&module).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errModuleNotExists, 13002)
		}
		return courseError.CreateError(err, 10002)
	}

	lesson := dto.CreateNewLesson()
	if err := tx.Joins("JOIN modules ON modules.name = ?", moduleName).
		Where("lessons.name = ?", name).Where("module_id = modules.id").First(&lesson).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lessonWithThisNameNotExists = true
		} else {
			tx.Rollback()
			return courseError.CreateError(err, 10002)
		}
	}

	if !lessonWithThisNameNotExists {
		tx.Rollback()
		return courseError.CreateError(errLessonNameAlreadyExists, 13001)
	}

	if err := tx.Joins("JOIN modules ON modules.name = ?", moduleName).Where("lessons.position = ?", position).Where("module_id = modules.id").First(&lesson).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lessonWithThisPosNotExists = true
		} else {
			tx.Rollback()
			return courseError.CreateError(err, 10002)
		}
	}

	if !lessonWithThisPosNotExists {
		tx.Rollback()
		return courseError.CreateError(errLessonPosAlreadyExists, 13001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return courseError.CreateError(err, 10010)
	}

	return nil
}

func (storage *Storage) CreateLesson(ctx context.Context, name, moduleName, description, position, videoPath, previewPath string) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	module := dto.CreateNewModule()
	if err := tx.Where("name = ?", moduleName).First(&module).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(err, 13002)
		}
		return nil, courseError.CreateError(err, 11101)
	}

	intPosition, _ := strconv.Atoi(position)

	lesson := dto.CreateNewLesson().
		AddName(name).
		AddDescription(description).
		AddPosition(intPosition).
		AddVideoUrl(videoPath).
		AddPreviewImgUrl(previewPath).
		AddModuleId(module.ID)

	if err := tx.Create(&lesson).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10001)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return &lesson.ID, nil
}
