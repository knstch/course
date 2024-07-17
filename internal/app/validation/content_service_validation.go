package validation

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	courseerror "github.com/knstch/course/internal/app/course_error"
)

type ModuleQueryToValidate struct {
	name        string
	description string
	courseName  string
	page        string
	limit       string
}

func NewModuleQueryToValidate(name, description, courseName, page, limit string) *ModuleQueryToValidate {
	return &ModuleQueryToValidate{
		name,
		description,
		courseName,
		page,
		limit,
	}
}

func (module *ModuleQueryToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, module,
		validation.Field(&module.name,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&module.description,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&module.courseName,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&module.page,
			validation.By(validatePage(module.page)),
		),
		validation.Field(&module.limit,
			validation.By(validateLimit(module.limit)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type LessonQueryToValidate struct {
	name        string
	description string
	courseName  string
	moduleName  string
	page        string
	limit       string
}

func NewLessonsQueryToValidate(name, description, courseName, moduleName, page, limit string) *LessonQueryToValidate {
	return &LessonQueryToValidate{
		name,
		description,
		courseName,
		moduleName,
		page,
		limit,
	}
}

func (lesson *LessonQueryToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, lesson,
		validation.Field(&lesson.name,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&lesson.description,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&lesson.courseName,
			validation.RuneLength(1, 200).Error(errBadLength),
		),
		validation.Field(&lesson.moduleName,
			validation.RuneLength(1, 100).Error(errBadLength),
		),
		validation.Field(&lesson.page,
			validation.By(validatePage(lesson.page)),
		),
		validation.Field(&lesson.limit,
			validation.By(validateLimit(lesson.limit)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
