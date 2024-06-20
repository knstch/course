package entity

import "github.com/knstch/course/internal/domain/dto"

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewCredentials() *Credentials {
	return &Credentials{}
}

type UserInfo struct {
	FirstName   string `json:"firstName"`
	Surname     string `json:"surname"`
	PhoneNumber string `json:"phoneNumber"`
}

func NewUserInfo() *UserInfo {
	return &UserInfo{}
}

type Passwords struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func CreateNewPasswords() *Passwords {
	return &Passwords{}
}

type NewEmail struct {
	NewEmail string `json:"newEmail"`
}

func CreateNewEmail() *NewEmail {
	return &NewEmail{}
}

type ConfirmCode struct {
	Code int `json:"code"`
}

func NewConfirmCodeEntity() *ConfirmCode {
	return &ConfirmCode{}
}

type SuccessResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func CreateSuccessResponse(message string, status bool) *SuccessResponse {
	return &SuccessResponse{
		Message: message,
		Success: status,
	}
}

type PasswordRecoverCredentials struct {
	Email    string `json:"email"`
	Code     int    `json:"code"`
	Password string `json:"password"`
}

func NewPasswordRecoverCredentials() *PasswordRecoverCredentials {
	return &PasswordRecoverCredentials{}
}

type Email struct {
	Email string `json:"email"`
}

func CreateEmail() *Email {
	return &Email{}
}

type CdnResponse struct {
	Path string  `json:"path"`
	Code int     `json:"code"`
	Err  *string `json:"error,omitempty"`
}

func NewCdnResponse() *CdnResponse {
	return &CdnResponse{}
}

type UserCourses struct {
	Id         uint   `json:"id"`
	Name       string `json:"name"`
	PreviewUrl string `json:"previewUrl"`
}

type UserData struct {
	FirstName       string        `json:"firstName"`
	Surname         string        `json:"surname"`
	PhoneNumber     uint          `json:"phoneNumber"`
	Photo           string        `json:"photo"`
	Email           string        `json:"email"`
	IsEmailVerified bool          `json:"isEmailVerified"`
	Courses         []UserCourses `json:"courses"`
}

func CreateNewUserData() *UserData {
	return &UserData{}
}

func (user *UserData) AddFirstName(name string) *UserData {
	user.FirstName = name
	return user
}

func (user *UserData) AddSurname(surname string) *UserData {
	user.Surname = surname
	return user
}

func (user *UserData) AddPhoneNumber(phoneNumber *uint) *UserData {
	if phoneNumber != nil {
		user.PhoneNumber = *phoneNumber
	}
	return user
}

func (user *UserData) AddPhoto(photo string) *UserData {
	user.Photo = photo
	return user
}

func (user *UserData) AddEmail(email string) *UserData {
	user.Email = email
	return user
}

func (user *UserData) AddEmailVerifiedStatus(verified bool) *UserData {
	user.IsEmailVerified = verified
	return user
}

func (user *UserData) AddCourses(courses []dto.Course) *UserData {
	user.Courses = make([]UserCourses, 0, len(courses))
	for _, v := range courses {
		course := UserCourses{
			Id:         v.ID,
			Name:       v.Name,
			PreviewUrl: v.PreviewImgUrl,
		}
		user.Courses = append(user.Courses, course)
	}

	return user
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"totalCount"`
	PagesCount int `json:"pagesCount"`
}

type UserDataWithPagination struct {
	Pagination Pagination      `json:"pagination"`
	Users      []UserDataAdmin `json:"users"`
}

type UserDataAdmin struct {
	Id          uint          `json:"id"`
	FirstName   string        `json:"firstName"`
	Surname     string        `json:"surname"`
	PhoneNumber uint          `json:"phoneNumber"`
	Active      bool          `json:"active"`
	Email       string        `json:"email"`
	IsVerified  bool          `json:"isVerified"`
	Photo       *string       `json:"photoPath,omitempty"`
	Courses     []UserCourses `json:"courses"`
}

func CreateUserDataAdmin(user dto.User) *UserDataAdmin {
	var userData UserDataAdmin

	if user.PhoneNumber != nil {
		userData.PhoneNumber = *user.PhoneNumber
	}

	userData.Id = user.ID
	userData.FirstName = user.FirstName
	userData.Surname = user.Surname
	userData.Active = user.Active

	return &userData
}

func (user *UserDataAdmin) AddCourses(courses []dto.Course) *UserDataAdmin {
	user.Courses = make([]UserCourses, 0, len(courses))
	for _, v := range courses {
		course := UserCourses{
			Id:   v.ID,
			Name: v.Name,
		}
		user.Courses = append(user.Courses, course)
	}

	return user
}

func (user *UserDataAdmin) AddCredentials(credentials *dto.Credentials) *UserDataAdmin {
	user.IsVerified = credentials.Verified
	user.Email = credentials.Email

	return user
}

func (user *UserDataAdmin) AddPhoto(photo *dto.Photo) *UserDataAdmin {
	user.Photo = &photo.Path

	return user
}

type Id struct {
	Id uint `json:"id"`
}

func NewId() *Id {
	return &Id{}
}

func (id *Id) AddId(Id *uint) *Id {
	id.Id = *Id
	return id
}

type Module struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    uint   `json:"position"`
	CourseName  string `json:"courseName"`
}

func NewModule() *Module {
	return &Module{}
}

type CourseInfo struct {
	Id          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	PreviewUrl  string       `json:"preview"`
	Cost        uint         `json:"cost"`
	Discount    uint         `json:"discount"`
	Modules     []ModuleInfo `json:"modules"`
}

type CourseInfoWithPagination struct {
	Pagination Pagination   `json:"pagination"`
	CourseInfo []CourseInfo `json:"courses"`
}

func CreateCourseInfo(course dto.Course, modules []ModuleInfo) *CourseInfo {
	return &CourseInfo{
		Id:          course.ID,
		Name:        course.Name,
		Description: course.Description,
		PreviewUrl:  course.PreviewImgUrl,
		Cost:        course.Cost,
		Discount:    *course.Discount,
		Modules:     modules,
	}
}

func CreateCoursesInfo(courses []dto.Course, modules []ModuleInfo) []CourseInfo {
	coursesInfo := make([]CourseInfo, 0, len(courses))

	for _, course := range courses {
		modulesInfo := make([]ModuleInfo, 0)
		for _, module := range modules {
			if module.CourseId == course.ID {
				modulesInfo = append(modulesInfo, module)
			}
		}
		coursesInfo = append(coursesInfo, *CreateCourseInfo(course, modulesInfo))
	}

	return coursesInfo
}

type ModuleInfo struct {
	Id          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Position    uint         `json:"position"`
	Lessons     []LessonInfo `json:"lessons"`
	CourseId    uint         `json:"-"`
}

type ModuleInfoWithPagination struct {
	Pagination Pagination   `json:"pagination"`
	ModuleInfo []ModuleInfo `json:"modulesInfo"`
}

func CreateModuleInfo(module *dto.Module, lessons []LessonInfo) *ModuleInfo {
	return &ModuleInfo{
		Id:          module.ID,
		Position:    module.Position,
		Name:        module.Name,
		Description: module.Description,
		Lessons:     lessons,
		CourseId:    module.CourseId,
	}
}

type LessonInfo struct {
	Id          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	PreviewUrl  string  `json:"preview"`
	VideoUrl    *string `json:"video,omitempty"`
	Position    uint    `json:"position"`
	ModuleId    uint    `json:"-"`
}

func CreateLessonInfo(lesson *dto.Lesson, isPurchased bool) *LessonInfo {
	if isPurchased {
		return &LessonInfo{
			Id:          lesson.ID,
			Name:        lesson.Name,
			Description: lesson.Description,
			PreviewUrl:  lesson.PreviewImgUrl,
			VideoUrl:    lesson.VideoUrl,
			Position:    uint(lesson.Position),
			ModuleId:    lesson.ModuleId,
		}
	}

	return &LessonInfo{
		Id:         lesson.ID,
		Name:       lesson.Name,
		PreviewUrl: lesson.PreviewImgUrl,
		Position:   uint(lesson.Position),
		ModuleId:   lesson.ModuleId,
	}
}

type LessonsInfoWithPagination struct {
	Pagination Pagination   `json:"pagination"`
	LessonInfo []LessonInfo `json:"lessonsInfo"`
}
