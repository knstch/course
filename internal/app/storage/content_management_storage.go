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

var (
	errModulePosAlreadyExists  = errors.New("модуль с такой позицией уже существует")
	errModuleNameAlreadyExists = errors.New("модуль с таким названием уже существует")
	errLessonNameAlreadyExists = errors.New("урок с таким назавнием уже существует")
	errLessonPosAlreadyExists  = errors.New("урок с такой позицией уже существует")

	errCourseNotExists = errors.New("курса с таким названием не существует")
	errModuleNotExists = errors.New("модуль с таким названием не существует")

	errCourseAlreadyExists = errors.New("курс с таким названием уже существует")
)

func (storage *Storage) CreateCourse(ctx context.Context, name, description, cost, discount, path string) (*uint, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	var isCourseNotExists bool
	course := dto.CreateNewCourse()

	if err := tx.Where("name = ?", name).First(&course).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isCourseNotExists = true
		} else {
			tx.Rollback()
			return nil, courseError.CreateError(err, 10002)
		}
	}

	if !isCourseNotExists {
		tx.Rollback()
		return nil, courseError.CreateError(errCourseAlreadyExists, 13001)
	}

	course.AddName(name).
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

func (storage *Storage) CheckIfLessonCanBeCreated(ctx context.Context, name, moduleName, position, courseName string) *courseError.CourseError {
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

	course := dto.CreateNewCourse()
	if err := tx.Where("name = ?", courseName).First(&course).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return courseError.CreateError(errCourseNotExists, 13003)
		}
		return courseError.CreateError(err, 10002)
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

func (storage *Storage) GetCourse(
	ctx context.Context, id, name, descr, cost, discount string,
	limit, offset int,
	isPurchased bool) (
	[]entity.CourseInfo,
	*courseError.CourseError,
) {
	tx := storage.db.WithContext(ctx).Begin()

	courses := dto.CreateNewCourses()

	query := tx.Model(&dto.Course{})

	if id != "" {
		query = query.Where("id = ?", id)
	} else {
		if name != "" {
			query = query.Where("LOWER(name) LIKE ?", fmt.Sprint("%"+strings.ToLower(name)+"%"))
		}

		if descr != "" {
			query = query.Where("LOWER(description) LIKE ?", fmt.Sprint("%"+strings.ToLower(descr)+"%"))
		}

		if cost != "" {
			query = query.Where("cost = ?", cost)
		}

		if discount != "" {
			query = query.Where("discount = ?", discount)
		}
	}

	if err := query.Limit(limit).Offset(offset).Find(&courses).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, courseError.CreateError(errCourseNotExists, 13003)
		}
		return nil, courseError.CreateError(err, 10002)
	}

	coursesIds := dto.ExtractIds(courses, func(item interface{}) uint {
		return item.(dto.Course).ID
	})
	modules := dto.GetAllModules()
	if err := tx.Where("course_id IN (?)", coursesIds).Order("position").Find(&modules).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	lessonIds := dto.ExtractIds(modules, func(item interface{}) uint {
		return item.(dto.Module).ID
	})
	lessons := dto.GetAllLessons()
	if err := tx.Where("module_id IN (?)", lessonIds).Order("position").Find(&lessons).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	lessonsInfo := make([]entity.LessonInfo, 0, len(lessons))
	for _, v := range lessons {
		lessonInfo := entity.CreateLessonInfo(&v, isPurchased)
		lessonsInfo = append(lessonsInfo, *lessonInfo)
	}

	modulesInfo := make([]entity.ModuleInfo, 0, len(modules))
	for _, module := range modules {
		retreivedLessons := make([]entity.LessonInfo, 0)
		for _, lesson := range lessonsInfo {
			if lesson.ModuleId == module.ID {
				retreivedLessons = append(retreivedLessons, lesson)
				continue
			}
		}
		moduleInfo := entity.CreateModuleInfo(&module, retreivedLessons)
		modulesInfo = append(modulesInfo, *moduleInfo)
	}

	courseInfo := entity.CreateCoursesInfo(courses, modulesInfo)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return courseInfo, nil
}

func (storage *Storage) GetModules(ctx context.Context,
	name, description, courseName string,
	limit, offset int) ([]entity.ModuleInfo, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	query := tx.Model(&dto.Module{})

	if name != "" {
		query = query.Where("LOWER(name) LIKE ?", fmt.Sprint("%"+strings.ToLower(name)+"%"))
	}

	if description != "" {
		query = query.Where("LOWER(description) LIKE ?", fmt.Sprint("%"+strings.ToLower(description)+"%"))
	}

	if courseName != "" {
		query = query.Where("course_id IN (SELECT id FROM courses WHERE LOWER(courses.name) LIKE ?)",
			fmt.Sprint("%"+strings.ToLower(courseName)+"%"))
	}

	modules := dto.GetAllModules()
	if err := query.Order("position").Offset(offset).Limit(limit).Find(&modules).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	modulesIds := dto.ExtractIds(modules, func(item interface{}) uint {
		return item.(dto.Module).ID
	})
	lessons := dto.GetAllLessons()
	if err := tx.Where("module_id IN (?)", modulesIds).Order("position").Find(&lessons).Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10002)
	}

	lessonsInfo := make([]entity.LessonInfo, 0, len(lessons))
	for _, v := range lessons {
		lessonInfo := entity.CreateLessonInfo(&v, true)
		lessonsInfo = append(lessonsInfo, *lessonInfo)
	}

	modulesInfo := make([]entity.ModuleInfo, 0, len(modules))
	for _, module := range modules {
		retreivedLessons := make([]entity.LessonInfo, 0)
		for _, lesson := range lessonsInfo {
			if lesson.ModuleId == module.ID {
				retreivedLessons = append(retreivedLessons, lesson)
				continue
			}
		}
		moduleInfo := entity.CreateModuleInfo(&module, retreivedLessons)
		modulesInfo = append(modulesInfo, *moduleInfo)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	return modulesInfo, nil
}

func (storage *Storage) GetLessons(ctx context.Context,
	name, description, moduleName, courseName string,
	limit, offset int) ([]entity.LessonInfo, *courseError.CourseError) {
	tx := storage.db.WithContext(ctx).Begin()

	query := tx.Model(&dto.Lesson{})

	if name != "" {
		query = query.Where("LOWER(name) LIKE ?", fmt.Sprint("%"+strings.ToLower(name)+"%"))
	}

	if description != "" {
		query = query.Where("LOWER(description) LIKE ?", fmt.Sprint("%"+strings.ToLower(description)+"%"))
	}

	if courseName != "" {
		query = query.Where("module_id IN (SELECT id FROM modules WHERE course_id IN (SELECT id FROM courses WHERE LOWER(courses.name) LIKE ?))",
			fmt.Sprint("%"+strings.ToLower(courseName)+"%"))
	}

	if moduleName != "" {
		query = query.Where("module_id IN (SELECT id FROM modules WHERE LOWER(modules.name) LIKE ?)",
			fmt.Sprint("%"+strings.ToLower(moduleName)+"%"))
	}

	lessons := dto.GetAllLessons()
	if err := query.Offset(offset).Limit(limit).Find(&lessons).Error; err != nil {
		return nil, courseError.CreateError(err, 10002)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, courseError.CreateError(err, 10010)
	}

	lessonsInfo := make([]entity.LessonInfo, 0, len(lessons))
	for _, v := range lessons {
		lessonInfo := entity.CreateLessonInfo(&v, true)
		lessonsInfo = append(lessonsInfo, *lessonInfo)
	}

	return lessonsInfo, nil
}
