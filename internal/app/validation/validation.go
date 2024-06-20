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

type CredentialsToValidate struct {
	Email    string
	Password string
}

const (
	emailPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	passwordPattern = `^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};:"\\|,.<>\/?]*$`
	lettersPattern  = `^\p{L}+$`
	fileNamePattern = `\.(.+)$`

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
)

var (
	emailRegex    = regexp.MustCompile(emailPattern)
	passwordRegex = regexp.MustCompile(passwordPattern)
	lettersRegex  = regexp.MustCompile(lettersPattern)

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

	boolsInterfaces = stringSliceTOInterfaceSlice(bools)

	errValueNotInt = errors.New("значение передано не как число")
	errBadFile     = errors.New("загруженный файл имеет неверный формат")

	errDiscountValueNotInt = errors.New("размер скидки передан не как число")
	errBadDiscountPriect   = errors.New("размер скидки не может быть больше стоимости")
)

func NewCredentialsToValidate(credentials *entity.Credentials) *CredentialsToValidate {
	return &CredentialsToValidate{
		Email:    credentials.Email,
		Password: credentials.Password,
	}
}

func (cr *CredentialsToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, cr,
		validation.Field(&cr.Email,
			validation.Required.Error(errEmailIsNil),
			validation.Match(emailRegex).Error(errBadEmail),
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
		validation.Field(&cr.Password,
			validation.Required.Error(errPasswordIsNil),
			validation.Match(passwordRegex).Error(errPasswordContainsBadSymbols),
			validation.By(validatePassword(cr.Password)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
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

type UserInfoToValidate struct {
	firstName   string
	surname     string
	phoneNumber string
}

var (
	phoneRegex = regexp.MustCompile(`^\+?\d{1,20}$`)
)

func NewUserInfoToValidate(userInfo *entity.UserInfo) *UserInfoToValidate {
	return &UserInfoToValidate{
		firstName:   userInfo.FirstName,
		surname:     userInfo.Surname,
		phoneNumber: userInfo.PhoneNumber,
	}
}

func (user *UserInfoToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, user,
		validation.Field(&user.firstName,
			validation.Required.Error("имя не может быть пустым"),
			validation.RuneLength(1, 20).Error("имя передано в неверном формате"),
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&user.surname,
			validation.Required.Error("фамилия не может быть пустой"),
			validation.RuneLength(1, 20).Error("фамилия передана в неверном формате"),
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&user.phoneNumber,
			validation.Required.Error("номер телефона не может быть пустым"),
			validation.Match(phoneRegex).Error("номер телефона передан неверно, введите его в фромате 79123456789"),
			validation.RuneLength(1, 20).Error("номер телефона передан в неверном формате"),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type SignInCredentials struct {
	Email    string
	Password string
}

func NewSignInCredentials(credentials *entity.Credentials) *SignInCredentials {
	return &SignInCredentials{
		Email:    credentials.Email,
		Password: credentials.Password,
	}
}

func (cr *SignInCredentials) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, cr,
		validation.Field(&cr.Email,
			validation.Required.Error(errEmailIsNil),
		),
		validation.Field(&cr.Password,
			validation.Required.Error(errPasswordIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type PasswordToValidate struct {
	password string
}

func NewPasswordToValidate(password string) *PasswordToValidate {
	return &PasswordToValidate{
		password: password,
	}
}

func (password *PasswordToValidate) ValidatePassword(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, password,
		validation.Field(&password.password,
			validation.Required.Error(errPasswordIsNil),
			validation.Match(passwordRegex).Error(errPasswordContainsBadSymbols),
			validation.By(validatePassword(password.password)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type EmailToValidate struct {
	email string
}

func NewEmailToValidate(email string) *EmailToValidate {
	return &EmailToValidate{
		email: email,
	}
}

func (email *EmailToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, email,
		validation.Field(&email.email,
			validation.Required.Error(errEmailIsNil),
			validation.Match(emailRegex).Error(errBadEmail),
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type ConfirmCodeToValidate struct {
	code int
}

func NewConfirmCodeToValidate(code int) *ConfirmCodeToValidate {
	return &ConfirmCodeToValidate{
		code: code,
	}
}

func (code *ConfirmCodeToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, code,
		validation.Field(&code.code,
			validation.Min(1000).Error(errBadConfirmCode),
			validation.Max(9999).Error(errBadConfirmCode),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type PasswordRecoverCredentials struct {
	Email    string
	Password string
	Code     int
}

func NewPasswordRecoverCredentialsToValidate(credentials entity.PasswordRecoverCredentials) *PasswordRecoverCredentials {
	return &PasswordRecoverCredentials{
		Email:    credentials.Email,
		Password: credentials.Password,
		Code:     credentials.Code,
	}
}

func (credentials *PasswordRecoverCredentials) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, credentials,
		validation.Field(&credentials.Email,
			validation.Required.Error(errEmailIsNil),
			validation.Match(emailRegex).Error(errBadEmail),
			validation.RuneLength(5, 40).Error(errBadEmail),
		),
		validation.Field(&credentials.Password,
			validation.Required.Error(errPasswordIsNil),
			validation.Match(passwordRegex).Error(errPasswordContainsBadSymbols),
			validation.By(validatePassword(credentials.Password)),
		),
		validation.Field(&credentials.Code,
			validation.Min(1000).Error(errBadConfirmCode),
			validation.Max(9999).Error(errBadConfirmCode),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

type UserFiltersToValidate struct {
	firstName   string
	surname     string
	phoneNumber string
	active      string
	email       string
	isVerified  string
	page        string
	limit       string
}

func stringSliceTOInterfaceSlice(values []string) []interface{} {
	interfaces := make([]interface{}, len(values))
	for i := range values {
		interfaces[i] = values[i]
	}

	return interfaces
}

func NewUserFiltersToValidate(firstName, surname, phoneNumber, email, active, isVerified, page, limit string) *UserFiltersToValidate {
	return &UserFiltersToValidate{
		firstName:   firstName,
		surname:     surname,
		phoneNumber: phoneNumber,
		active:      active,
		email:       email,
		isVerified:  isVerified,
		page:        page,
		limit:       limit,
	}
}

func (userFilters *UserFiltersToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, userFilters,
		validation.Field(&userFilters.firstName,
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&userFilters.surname,
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&userFilters.phoneNumber,
			validation.Match(phoneRegex).Error("номер телефона передан неверно, введите его в фромате 79123456789"),
		),
		validation.Field(&userFilters.isVerified,
			validation.In(boolsInterfaces...).Error(errBadBool),
		),
		validation.Field(&userFilters.active,
			validation.In(boolsInterfaces...).Error(errBadBool),
		),
		validation.Field(&userFilters.email,
			validation.Match(emailRegex).Error(errBadEmail),
		),
		validation.Field(&userFilters.page,
			validation.By(validatePage(userFilters.page)),
		),
		validation.Field(&userFilters.limit,
			validation.By(validateLimit(userFilters.limit)),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}

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

type ImgExtToValidate struct {
	FileName string
}

func NewImgExtToValidate(fileName string) *ImgExtToValidate {
	return &ImgExtToValidate{
		FileName: fileName,
	}
}

func imgExtValidator(fileName string) validation.RuleFunc {
	return func(value interface{}) error {
		var fileExtention string
		matches := fileExtRegex.FindStringSubmatch(fileName)
		if len(matches) > 1 {
			fileExtention = matches[1]
		} else {
			return errBadFile
		}

		var isExtApproved bool
		for _, v := range allowedImgExtentions {
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

func (img *ImgExtToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, img,
		validation.Field(&img.FileName,
			validation.By(imgExtValidator(img.FileName)),
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
}

func NewCourseToValidate(name, description, previewExt, cost, discount string) *CourseToValidate {
	return &CourseToValidate{
		Name:        name,
		Description: description,
		PreviewExt:  previewExt,
		Cost:        cost,
		Discount:    discount,
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

type LessonToValidate struct {
	description string
	position    string
	moduleName  string
	courseName  string
	name        string
	previewImg  string
	lesson      string
}

func NewLessonToValidate(name, description, moduleName, previewImg, lesson, position, courseName string) *LessonToValidate {
	return &LessonToValidate{
		name:        name,
		description: description,
		position:    position,
		moduleName:  moduleName,
		courseName:  courseName,
		previewImg:  previewImg,
		lesson:      lesson,
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
		validation.Field(&lesson.lesson,
			validation.Required.Error(errFieldIsNil),
			validation.By(videoExtValidator(lesson.lesson)),
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
		name:        name,
		description: descr,
		cost:        cost,
		discount:    discount,
		page:        page,
		limit:       limit,
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

type ModuleQueryToValidate struct {
	name        string
	description string
	courseName  string
	page        string
	limit       string
}

func NewModuleQueryToValidate(name, description, courseName, page, limit string) *ModuleQueryToValidate {
	return &ModuleQueryToValidate{
		name:        name,
		description: description,
		courseName:  courseName,
		page:        page,
		limit:       limit,
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
		name:        name,
		description: description,
		courseName:  courseName,
		moduleName:  moduleName,
		page:        page,
		limit:       limit,
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

type PaymentCredentialsToValidate struct {
	courseId uint
	ruCard   bool
}

func NewPaymentCredentialsToValidate(courseId uint, ruCard bool) *PaymentCredentialsToValidate {
	return &PaymentCredentialsToValidate{
		courseId: courseId,
		ruCard:   ruCard,
	}
}

func (credentials *PaymentCredentialsToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, credentials,
		validation.Field(&credentials.courseId,
			validation.Required.Error(errFieldIsNil),
		),
		validation.Field(&credentials.ruCard,
			validation.Required.Error(errFieldIsNil),
		),
	); err != nil {
		return courseerror.CreateError(err, 400)
	}

	return nil
}
