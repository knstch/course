package validation

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	courseerror "github.com/knstch/course/internal/app/course_error"
	"github.com/knstch/course/internal/domain/entity"
)

func NewUserInfoToValidate(userInfo *entity.UserInfo) *UserInfoToValidate {
	return (*UserInfoToValidate)(userInfo)
}

func (user *UserInfoToValidate) Validate(ctx context.Context) *courseerror.CourseError {
	if err := validation.ValidateStructWithContext(ctx, user,
		validation.Field(&user.FirstName,
			validation.Required.Error("имя не может быть пустым"),
			validation.RuneLength(1, 20).Error("имя передано в неверном формате"),
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&user.Surname,
			validation.Required.Error("фамилия не может быть пустой"),
			validation.RuneLength(1, 20).Error("фамилия передана в неверном формате"),
			validation.Match(lettersRegex).Error(errFieldAcceptsOnlyLetters),
		),
		validation.Field(&user.PhoneNumber,
			validation.Required.Error("номер телефона не может быть пустым"),
			validation.Match(phoneRegex).Error("номер телефона передан неверно, введите его в фромате 79123456789"),
			validation.RuneLength(1, 20).Error("номер телефона передан в неверном формате"),
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
