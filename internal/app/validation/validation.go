package validation

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

const (
	passwordPattern = `^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};:"\\|,.<>\/?]*$`
	lettersPattern  = `^\p{L}+$`
	fileNamePattern = `\.(.+)$`
	loginPattern    = `^[a-zA-Z.\-]{4,20}$`

	errEmailIsNil                 = "email обязательно"
	errBadEmail                   = "email передан неправильно"
	errPasswordIsNil              = "пароль обязателен"
	errBadPassword                = "пароль должен содержать как миниум 8 символов и включать в себя как минимум 1 цифру"
	errBadConfirmCode             = "код верификации передан неверно"
	errPasswordContainsBadSymbols = "пароль может содержать только латинские буквы, спец. символы и цифры"
	errFieldAcceptsOnlyLetters    = "допустимы только буквы"
	errBadBool                    = "допустимы значения только true/fasle"

	errPageIsBad     = "номер страницы не может быть меньше 0"
	errPageIsNil     = "номер страницы обязателен"
	errPageNotInt    = "номер страницы передан не как число"
	errLimitIsNil    = "лимит обязателен"
	errLimitIsNotInt = "лимит передан не как число"
	errLimitIsBad    = "значение лимит не может быть меньше 1"

	errBadMinCostValue = "значение не может быть меньше 0"
	errFieldIsNil      = "поле не может быть пустым"
	errIdIsNil         = "ИД обязательно"

	errCourseNameIsTooBig        = "название курса слишком длинное, ограничение в 100 символов"
	errCourseDescriptionIsTooBig = "описание курса слишком длинное, ограничение в 2000 символов"
	errValueTooSmall             = "значение не может быть ниже 1"

	errModuleNameIsTooBig        = "название модуля слишком длинное"
	errModuleDescriptionIsTooBig = "описание модуля слишком длинное"
	errModulePositionIsBad       = "позиция модуля передана неверно"

	errLessonNameIsTooBig        = "название урока слишком длинное"
	errLessonDescriptionIsTooBig = "описание урока слишком длинное"
	errLessonPositionIsBad       = "позиция модуля передана неверно"
	errBadMinValue               = "значение не может быть меньше 1"

	errBadParam  = "параметр передан неправильно"
	errBadLength = "превышена длина параметра"

	errTokenFieldCantBeEmpty = "поле token не может быть пустым"

	errBadLogin = "логин передан неверно"
	errBadRole  = "такой роли не существует"

	errBadPaymentMethodParam = `допустимы значения только "ru-card" и "foreign-card"`
	errBadDate               = "параметр дата передан неверно"

	errDueEarlierThenFrom = "период задан некорректно"
	errBadUrl             = "ссылка передана неверно"
)

var (
	passwordRegex = regexp.MustCompile(passwordPattern)
	lettersRegex  = regexp.MustCompile(lettersPattern)
	loginRegexp   = regexp.MustCompile(loginPattern)

	fileExtRegex = regexp.MustCompile(fileNamePattern)

	bools = []string{
		"true",
		"false",
	}

	allowedImgExtentions = []string{
		"jpeg",
		"jpg",
		"png",
		"JPEG",
		"JPG",
		"PNG",
	}

	allowedVideoExtentions = []string{
		"mp4",
	}

	allowedRoles = []string{
		"super_admin",
		"admin",
		"editor",
		"moderator",
	}

	allowrdPaymentMethods = []string{
		"ru-card",
		"foreign-card",
	}

	boolsInterfaces          = stringSliceTOInterfaceSlice(bools)
	rolesInterfaces          = stringSliceTOInterfaceSlice(allowedRoles)
	paymentMethodsInterfaces = stringSliceTOInterfaceSlice(allowrdPaymentMethods)

	errValueNotInt = errors.New("значение передано не как число")
	errBadFile     = errors.New("загруженный файл имеет неверный формат")

	errDiscountValueNotInt = errors.New("размер скидки передан не как число")
	errBadDiscountPriect   = errors.New("размер скидки не может быть больше стоимости")
)

func stringSliceTOInterfaceSlice(values []string) []interface{} {
	interfaces := make([]interface{}, len(values))
	for i := range values {
		interfaces[i] = values[i]
	}

	return interfaces
}

func validatePassword(password string) validation.RuleFunc {
	return func(value interface{}) error {
		if len([]rune(password)) < 8 {
			return fmt.Errorf("пароль должен содержать как миниум 8 символов")
		}

		var hasLetter, hasNumber, hasUpperCase bool

		for _, v := range password {
			switch {
			case unicode.IsLetter(v):
				hasLetter = true
				if unicode.IsUpper(v) {
					hasUpperCase = true
				}
			case unicode.IsNumber(v):
				hasNumber = true
			}
		}

		if !hasLetter || !hasNumber || !hasUpperCase {
			return fmt.Errorf("пароль должен содержать как минимум 1 букву, 1 заглавную букву и 1 цифру")
		}

		return nil
	}
}

type UserInfoToValidate entity.UserInfo

var (
	phoneRegex = regexp.MustCompile(`^\+?\d{1,20}$`)
)

func validateLimit(limit string) validation.RuleFunc {
	return func(value interface{}) error {
		if limit == "" {
			return fmt.Errorf(errLimitIsNil)
		}

		intLimit, err := strconv.Atoi(limit)
		if err != nil {
			return fmt.Errorf(errLimitIsNotInt)
		}

		if intLimit < 1 {
			return fmt.Errorf(errLimitIsBad)
		}

		return nil
	}
}

func validatePage(page string) validation.RuleFunc {
	return func(value interface{}) error {
		if page == "" {
			return fmt.Errorf(errPageIsNil)
		}

		intPage, err := strconv.Atoi(page)
		if err != nil {
			return fmt.Errorf(errPageNotInt)
		}

		if intPage < 0 {
			return fmt.Errorf(errPageIsBad)
		}

		return nil
	}
}

type IdToValidate struct {
	Id int
}

func NewIdToValidate(id int) *IdToValidate {
	return &IdToValidate{
		Id: id,
	}
}

func (id *IdToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, id,
		validation.Field(&id.Id,
			validation.Min(1).Error(errBadMinValue),
			validation.Required.Error(errFieldIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type StringIdToValidate struct {
	Id string
}

func NewStringIdToValidate(id string) *StringIdToValidate {
	return &StringIdToValidate{
		Id: id,
	}
}

func idValidator(id string) validation.RuleFunc {
	return func(value interface{}) error {
		if id == "" || id == "0" {
			return fmt.Errorf(errIdIsNil)
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			return errValueNotInt
		}

		if idInt < 0 {
			return fmt.Errorf(errBadMinValue)
		}

		return nil
	}
}

func (i *StringIdToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Id,
			validation.By(idValidator(i.Id)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type CourseToValidate struct {
	Name        string
	Description string
	PreviewExt  string
	Cost        string
	Discount    string
	CourseId    string
}

func NewCourseToValidate(name, description, cost, discount, previewExt string) *CourseToValidate {
	return &CourseToValidate{
		Name:        name,
		Description: description,
		Cost:        cost,
		Discount:    discount,
		PreviewExt:  previewExt,
	}
}

func (course *CourseToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, course,
		validation.Field(&course.Name,
			validation.Required.Error(errFieldIsNil),
			validation.RuneLength(1, 100).Error(errCourseNameIsTooBig),
		),
		validation.Field(&course.Description,
			validation.Required.Error(errFieldIsNil),
			validation.RuneLength(1, 2000).Error(errCourseDescriptionIsTooBig),
		),
		validation.Field(&course.PreviewExt,
			validation.Required.Error(errFieldIsNil),
			validation.By(imgExtValidator(course.PreviewExt)),
		),
		validation.Field(&course.Discount,
			validation.By(costValidator(course.Discount, "", true)),
		),
		validation.Field(&course.Cost,
			validation.By(costValidator(course.Cost, course.Discount, false)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type EditCourseToValidate CourseToValidate

func NewEditCoruseToValidate(name, description, cost, discount, courseId string) *EditCourseToValidate {
	return &EditCourseToValidate{
		Name:        name,
		Description: description,
		Cost:        cost,
		Discount:    discount,
		CourseId:    courseId,
	}
}

func (course *EditCourseToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, course,
		validation.Field(&course.Name,
			validation.RuneLength(1, 100).Error(errCourseNameIsTooBig),
		),
		validation.Field(&course.Description,
			validation.RuneLength(1, 2000).Error(errCourseDescriptionIsTooBig),
		),
		validation.Field(&course.Discount,
			validation.By(costValidator(course.Discount, "", true)),
		),
		validation.Field(&course.Cost,
			validation.By(costValidator(course.Cost, course.Discount, false)),
		),
		validation.Field(&course.CourseId,
			validation.Required.Error(errFieldIsNil),
			validation.By(idValidator(course.CourseId)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type PreviewFileNameToValidate struct {
	PreviewExt string
}

func NewPreviewFileNameToValidate(fileName string) *PreviewFileNameToValidate {
	return &PreviewFileNameToValidate{
		PreviewExt: fileName,
	}
}

func (preview *PreviewFileNameToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, preview,
		validation.Field(&preview.PreviewExt,
			validation.By(imgExtValidator(preview.PreviewExt)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

func costValidator(cost, discount string, isDiscount bool) validation.RuleFunc {
	return func(value interface{}) error {
		if isDiscount && (cost == "" || cost == "0") {
			return nil
		}

		if cost == "" {
			return fmt.Errorf(errFieldIsNil)
		}

		var (
			costInt     int
			discountInt int
			err         error
		)

		if cost != "0" {
			costInt, err = strconv.Atoi(cost)
			if err != nil {
				return errValueNotInt
			}

			if costInt < 0 {
				return fmt.Errorf(errBadMinCostValue)
			}
		}
		if !isDiscount && cost == "" {
			return fmt.Errorf(errFieldIsNil)
		}

		if discount != "" {
			discountInt, err = strconv.Atoi(discount)
			if err != nil {
				return errDiscountValueNotInt
			}

			if costInt < discountInt {
				return errBadDiscountPriect
			}
		}

		return nil
	}
}

type ModuleToValidate struct {
	name        string
	description string
	position    int
	courseName  string
	moduleId    uint
}

func NewModuleToValidate(name, description, courseName string, position uint) *ModuleToValidate {
	return &ModuleToValidate{
		name:        name,
		description: description,
		position:    int(position),
		courseName:  courseName,
	}
}

func (module *ModuleToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, module,
		validation.Field(&module.courseName,
			validation.Required.Error(errFieldIsNil),
		),
		validation.Field(&module.name,
			validation.Required.Error(errFieldIsNil),
			validation.RuneLength(1, 40).Error(errModuleNameIsTooBig),
		),
		validation.Field(&module.description,
			validation.Required.Error(errFieldIsNil),
			validation.RuneLength(1, 200).Error(errModuleDescriptionIsTooBig),
		),
		validation.Field(&module.position,
			validation.Required.Error(errFieldIsNil),
			validation.Min(1).Error(errModulePositionIsBad),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type EditModuleToValidate ModuleToValidate

func NewEditModuleToValidate(name, description string, position, moduleId uint) *EditModuleToValidate {
	return &EditModuleToValidate{
		name:        name,
		description: description,
		position:    int(position),
		moduleId:    moduleId,
	}
}

func (module *EditModuleToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, module,
		validation.Field(&module.name,
			validation.RuneLength(1, 40).Error(errModuleNameIsTooBig),
		),
		validation.Field(&module.description,
			validation.RuneLength(1, 200).Error(errModuleDescriptionIsTooBig),
		),
		validation.Field(&module.position,
			validation.Min(1).Error(errModulePositionIsBad),
		),
		validation.Field(&module.moduleId,
			validation.By(idValidator(fmt.Sprint(module.moduleId))),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type LessonToValidate struct {
	description string
	position    string
	moduleName  string
	courseName  string
	name        string
	previewImg  string
	video       string
	lessonId    string
}

func NewLessonToValidate(name, description, moduleName, previewImg, video, position, courseName string) *LessonToValidate {
	return &LessonToValidate{
		name:        name,
		description: description,
		position:    position,
		moduleName:  moduleName,
		courseName:  courseName,
		previewImg:  previewImg,
		video:       video,
	}
}

func (lesson *LessonToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, lesson,
		validation.Field(&lesson.name,
			validation.Required.Error(errFieldIsNil),
			validation.RuneLength(1, 40).Error(errLessonNameIsTooBig),
		),
		validation.Field(&lesson.description,
			validation.Required.Error(errFieldIsNil),
			validation.RuneLength(1, 200).Error(errLessonDescriptionIsTooBig),
		),
		validation.Field(&lesson.moduleName,
			validation.Required.Error(errFieldIsNil),
		),
		validation.Field(&lesson.previewImg,
			validation.Required.Error(errFieldIsNil),
			validation.By(imgExtValidator(lesson.previewImg)),
		),
		validation.Field(&lesson.video,
			validation.Required.Error(errFieldIsNil),
			validation.By(videoExtValidator(lesson.video)),
		),
		validation.Field(&lesson.position,
			validation.By(posValidator(lesson.position)),
		),
		validation.Field(&lesson.courseName,
			validation.Required.Error(errFieldIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

func videoExtValidator(fileName string) validation.RuleFunc {
	return func(value interface{}) error {
		var fileExtention string
		matches := fileExtRegex.FindStringSubmatch(fileName)
		if len(matches) > 1 {
			fileExtention = matches[1]
		} else {
			return errBadFile
		}

		var isExtApproved bool
		for _, v := range allowedVideoExtentions {
			if v == fileExtention {
				isExtApproved = true
				break
			}
		}

		if !isExtApproved {
			return errBadFile
		}

		return nil
	}
}

func posValidator(position string) validation.RuleFunc {
	return func(value interface{}) error {
		if position == "" {
			return fmt.Errorf(errFieldIsNil)
		}

		posInt, err := strconv.Atoi(position)
		if err != nil {
			return errValueNotInt
		}

		if posInt < 1 {
			return fmt.Errorf(errBadMinValue)
		}

		return nil
	}
}

type EditLessonToValidate LessonToValidate

func NewEditLessonToValidate(name, description, position, lessonId string) *EditLessonToValidate {
	return &EditLessonToValidate{
		name:        name,
		description: description,
		position:    position,
		lessonId:    lessonId,
	}
}

func (lesson *EditLessonToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, lesson,
		validation.Field(&lesson.name,
			validation.RuneLength(1, 40).Error(errLessonNameIsTooBig),
		),
		validation.Field(&lesson.description,
			validation.RuneLength(1, 200).Error(errLessonDescriptionIsTooBig),
		),
		validation.Field(&lesson.position,
			validation.By(posValidator(lesson.position)),
		),
		validation.Field(&lesson.lessonId,
			validation.Required.Error(errFieldIsNil),
			validation.By(idValidator(lesson.lessonId)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type VideoFileNameToValidate struct {
	VideoExt string
}

func NewVideoFileNameToValidate(fileName string) *VideoFileNameToValidate {
	return &VideoFileNameToValidate{
		VideoExt: fileName,
	}
}

func (video *VideoFileNameToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, video,
		validation.Field(&video.VideoExt,
			validation.By(videoExtValidator(video.VideoExt)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type CourseQueryToValidate struct {
	name        string
	description string
	cost        string
	discount    string
	page        string
	limit       string
}

func NewCourseQueryToValidate(name, descr, cost, discount, page, limit string) *CourseQueryToValidate {
	return &CourseQueryToValidate{
		name,
		descr,
		cost,
		discount,
		page,
		limit,
	}
}

func (query *CourseQueryToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, query,
		validation.Field(&query.name,
			validation.RuneLength(1, 100).Error(errBadLength),
		),
		validation.Field(&query.description,
			validation.RuneLength(1, 100).Error(errBadLength),
		),
		validation.Field(&query.cost,
			validation.RuneLength(1, 100).Error(errBadLength),
		),
		validation.Field(&query.discount,
			validation.RuneLength(1, 100).Error(errBadLength),
		),
		validation.Field(&query.page,
			validation.By(validatePage(query.page)),
		),
		validation.Field(&query.limit,
			validation.By(validateLimit(query.limit)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
