package entity

import (
	"time"

	"github.com/knstch/course/internal/domain/dto"
)

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
	Id         uint        `json:"id"`
	Name       string      `json:"name"`
	PreviewUrl string      `json:"previewUrl"`
	Billing    UserBilling `json:"billingInfo"`
}

func (courses *UserCourses) AddBilling(id uint, order string, paidStatus bool, paid float64, paymentMethod string, invoiceId uint, time time.Time) *UserCourses {
	courses.Billing.Id = id
	courses.Billing.Order = order
	courses.Billing.PaidStatus = paidStatus
	courses.Billing.Paid = paid
	courses.Billing.PaymentMethod = paymentMethod
	courses.Billing.InvoiceId = invoiceId
	courses.Billing.Timestamp = time

	return courses
}

type UserBilling struct {
	Id            uint      `json:"id"`
	Order         string    `json:"order"`
	PaidStatus    bool      `json:"paidStatus"`
	Paid          float64   `json:"paid"`
	PaymentMethod string    `json:"paymentMethod"`
	InvoiceId     uint      `json:"invoice"`
	Timestamp     time.Time `json:"timestamp"`
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

func (user *UserDataAdmin) AddCourses(courses []dto.Course, orders []dto.Order, billing []dto.Billing) *UserDataAdmin {
	user.Courses = make([]UserCourses, 0, len(courses))
	for _, v := range courses {
		course := UserCourses{
			Id:         v.ID,
			Name:       v.Name,
			PreviewUrl: v.PreviewImgUrl,
		}

	ordersLoop:
		for _, j := range orders {
			if v.ID == j.CourseId {
				for _, k := range billing {
					if j.ID == k.OrderId {
						course.AddBilling(k.ID, j.Order, k.Paid, k.Price, k.PaymentMethod, k.InvoiceId, k.CreatedAt)
						break ordersLoop
					}
				}
			}
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

func NewId(id *uint) *Id {
	if id == nil {
		return &Id{}
	}
	return &Id{
		Id: *id,
	}
}

type Module struct {
	ModuleId    uint   `json:"moduleId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    *uint  `json:"position,omitempty"`
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
			VideoUrl:    &lesson.VideoUrl,
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

type Invoice struct {
	Purchaser struct {
		Email   string `json:"email"`
		Contact string `json:"contact"`
	} `json:"purchaser"`
	Order Order `json:"order"`
}

type Order struct {
	OrderID        string    `json:"order_id"`
	OrderNumber    int       `json:"order_number"`
	OrderDate      time.Time `json:"order_date"`
	ServiceID      int       `json:"service_id"`
	Amount         int       `json:"amount"`
	Currency       string    `json:"currency"`
	Purpose        string    `json:"purpose"`
	Language       string    `json:"language"`
	ExpirationDate time.Time `json:"expiration_date"`
	TaxSystem      int       `json:"tax_system"`
}

type UserID struct {
	PartnerClientID string `json:"partner_client_id"`
}

type InvoiceData struct {
	UserID  UserID  `json:"user_id"`
	PType   int     `json:"ptype"`
	Invoice Invoice `json:"invoice"`
}

func CreateOrder(essentials dto.OrderEssentials, serviceID int, partnerClientID string, ptype int) InvoiceData {
	order := Order{
		OrderID:        essentials.Order,
		OrderNumber:    int(essentials.OrderId),
		OrderDate:      time.Unix(int64(essentials.OrderDate), 0).UTC(),
		ServiceID:      serviceID,
		Amount:         int(essentials.Amount),
		Currency:       essentials.Currency,
		Purpose:        essentials.Purpose,
		Language:       essentials.Language,
		ExpirationDate: time.Unix(int64(essentials.ExpirationDate), 0).UTC(),
		TaxSystem:      int(essentials.TaxSystem),
	}

	invoice := Invoice{
		Order: order,
	}
	invoice.Purchaser.Email = essentials.Purchaser.Email
	invoice.Purchaser.Contact = essentials.Purchaser.Contact

	data := InvoiceData{
		UserID: UserID{
			PartnerClientID: partnerClientID,
		},
		PType:   ptype,
		Invoice: invoice,
	}

	return data
}

type BuyDetails struct {
	CourseId  uint `json:"courseId"`
	IsRusCard bool `json:"isRusCard"`
}

func CreateNewBuyDetails() *BuyDetails {
	return &BuyDetails{}
}

type SuccessPayment struct {
	CourseName string `json:"courseName"`
	Status     bool   `json:"status"`
}

func CreateNewSuccessPayment(courseName string) *SuccessPayment {
	return &SuccessPayment{
		CourseName: courseName,
		Status:     true,
	}
}

type FailedPayment struct {
	Status bool `json:"status"`
}

func CreateNewFailedPayment() *FailedPayment {
	return &FailedPayment{
		Status: true,
	}
}

type BillingHost struct {
	Url string `json:"url"`
}

func CreateBillingHost() *BillingHost {
	return &BillingHost{}
}

type AccessToken struct {
	Token string `json:"token"`
}

func CreateAccessToken() *AccessToken {
	return &AccessToken{}
}

type AdminCredentials struct {
	Login    string
	Password string
	Role     string
	Code     string
}

func CreateNewAdminCredentials() *AdminCredentials {
	return &AdminCredentials{}
}

func (admin *AdminCredentials) AddLogin(login string) *AdminCredentials {
	admin.Login = login
	return admin
}

func (admin *AdminCredentials) AddPassword(password string) *AdminCredentials {
	admin.Password = password
	return admin
}

func (admin *AdminCredentials) AddRole(role string) *AdminCredentials {
	admin.Role = role
	return admin
}

type Admin struct {
	Id         uint   `json:"id"`
	Login      string `json:"login"`
	Role       string `json:"role"`
	AuthStatus bool   `json:"2 steps auth enabled"`
	Key        string `json:"key"`
}

func CovertDtoAdmin(admin *dto.Admin) *Admin {
	return &Admin{
		Id:         admin.ID,
		Login:      admin.Login,
		Role:       admin.Role,
		AuthStatus: admin.TwoStepsAuthEnabled,
		Key:        admin.Key,
	}
}

type AdminsInfoWithPagination struct {
	Pagination Pagination `json:"pagination"`
	AdminInfo  []Admin    `json:"adminInfo"`
}
