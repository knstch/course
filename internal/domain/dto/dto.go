package dto

import (
	"reflect"
	"strconv"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName     string
	Surname       string
	Credentials   Credentials
	CredentialsId *uint `gorm:"not null"`
	PhoneNumber   *uint
	Active        bool `gorm:"not null;default:true"`
	PhotoId       *uint
	Photo         Photo
}

func CreateNewUser() *User {
	return &User{}
}

func CreateNewUsers() []User {
	return []User{}
}

type Photo struct {
	gorm.Model
	Path string
}

func CreateNewPhoto() *Photo {
	return &Photo{}
}

func (photo *Photo) AddPath(path string) *Photo {
	photo.Path = path
	return photo
}

func (user *User) AddCredentialsId(id *uint) *User {
	user.CredentialsId = id
	return user
}

type Credentials struct {
	gorm.Model
	Email    string `gorm:"not null;unique"`
	Password string `gorm:"not null"`
	Verified bool   `gorm:"not null;default:false"`
}

func CreateNewCredentials() *Credentials {
	return &Credentials{}
}

func (cr *Credentials) SetStatusUnverified() *Credentials {
	cr.Verified = false
	return cr
}

func (cr *Credentials) AddEmail(email string) *Credentials {
	cr.Email = email
	return cr
}

func (cr *Credentials) AddPassword(password string) *Credentials {
	cr.Password = password
	return cr
}

type AccessToken struct {
	gorm.Model
	User      User
	UserId    uint   `gorm:"not null"`
	Token     string `gorm:"not null"`
	Available bool   `gorm:"not null"`
}

func CreateNewAccessToken() *AccessToken {
	return &AccessToken{}
}

func (accessToken *AccessToken) AddUsedId(id *uint) *AccessToken {
	accessToken.UserId = *id
	return accessToken
}

func (accessToken *AccessToken) AddToken(token *string) *AccessToken {
	accessToken.Token = *token
	return accessToken
}

func (accessToken *AccessToken) SetStatusAvailable() *AccessToken {
	accessToken.Available = true
	return accessToken
}

type UsersCourse struct {
	gorm.Model
	UserId   *uint
	User     User
	CourseId *uint
	Course   Course
}

func NewUsersCourse() *UsersCourse {
	return &UsersCourse{}
}

type Course struct {
	gorm.Model
	Name          string `gorm:"not null"`
	Description   string `gorm:"not null"`
	PreviewImgUrl string `gorm:"not null"`
	Cost          uint   `gorm:"not null"`
	Discount      *uint
}

func CreateNewCourse() *Course {
	return &Course{}
}

func (course *Course) AddName(name string) *Course {
	course.Name = name
	return course
}

func (course *Course) AddDescription(description string) *Course {
	course.Description = description
	return course
}

func (course *Course) AddPreviewImg(path string) *Course {
	course.PreviewImgUrl = path
	return course
}

func (course *Course) AddCost(cost string) *Course {
	intCost, _ := strconv.Atoi(cost)
	course.Cost = uint(intCost)
	return course
}

func (course *Course) AddDiscount(discount string) *Course {
	if discount != "" {
		intDiscount, _ := strconv.Atoi(discount)
		uintDiscount := uint(intDiscount)
		course.Discount = &uintDiscount
	}
	return course
}

func CreateNewCourses() []Course {
	return []Course{}
}

type Lesson struct {
	gorm.Model
	ModuleId      uint   `gorm:"not null"`
	Module        Module `gorm:"not null"`
	Name          string `gorm:"not null"`
	Description   string `gorm:"not null"`
	PreviewImgUrl string `gorm:"not null"`
	VideoUrl      string `gorm:"not null"`
	Position      int    `gorm:"not null"`
}

func GetAllLessons() []Lesson {
	var lessons []Lesson
	return lessons
}

func ExtractAllModulesIds(modules []Module) []interface{} {
	ids := make([]interface{}, 0, len(modules))
	checkIfIdsExist := make(map[uint]interface{})
	for _, v := range modules {
		_, ok := checkIfIdsExist[v.ID]
		if !ok {
			checkIfIdsExist[v.ID] = true
			ids = append(ids, v.ID)
		}
	}
	return ids
}

func ExtractAllCoursesIds(courses []Course) []interface{} {
	ids := make([]interface{}, 0, len(courses))
	checkIfIdsExist := make(map[uint]interface{})
	for _, v := range courses {
		_, ok := checkIfIdsExist[v.ID]
		if !ok {
			checkIfIdsExist[v.ID] = true
			ids = append(ids, v.ID)
		}
	}
	return ids
}

func ExtractIds(items interface{}, idExtractor func(interface{}) uint) []interface{} {
	ids := make([]interface{}, 0)
	checkIfIdsExist := make(map[uint]interface{})

	sliceValue := reflect.ValueOf(items)
	for i := 0; i < sliceValue.Len(); i++ {
		item := sliceValue.Index(i).Interface()
		id := idExtractor(item)

		_, ok := checkIfIdsExist[id]
		if !ok {
			checkIfIdsExist[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

func CreateNewLesson() *Lesson {
	return &Lesson{}
}

func (lesson *Lesson) AddModuleId(id uint) *Lesson {
	lesson.ModuleId = id
	return lesson
}

func (lesson *Lesson) AddName(name string) *Lesson {
	lesson.Name = name
	return lesson
}

func (lesson *Lesson) AddDescription(descr string) *Lesson {
	lesson.Description = descr
	return lesson
}

func (lesson *Lesson) AddPreviewImgUrl(url string) *Lesson {
	lesson.PreviewImgUrl = url
	return lesson
}

func (lesson *Lesson) AddVideoUrl(url string) *Lesson {
	lesson.VideoUrl = url
	return lesson
}

func (lesson *Lesson) AddPosition(pos int) *Lesson {
	lesson.Position = pos
	return lesson
}

type Module struct {
	gorm.Model
	CourseId    uint
	Course      Course
	Name        string
	Description string
	Position    uint
}

func GetAllModules() []Module {
	var modules []Module
	return modules
}

func CreateNewModule() *Module {
	return &Module{}
}

func (m *Module) AddCourseId(id uint) *Module {
	m.CourseId = id
	return m
}

func (m *Module) AddName(name string) *Module {
	m.Name = name
	return m
}

func (m *Module) AddDescription(description string) *Module {
	m.Description = description
	return m
}

func (m *Module) AddPosition(pos uint) *Module {
	m.Position = pos
	return m
}
