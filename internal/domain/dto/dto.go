package dto

import (
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

type Photo struct {
	gorm.Model
	Path string
}

func CreateNewPhoto(path string) *Photo {
	return &Photo{
		Path: path,
	}
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

type Course struct {
	gorm.Model
	User   User
	UserId uint
	Name   string
}

func CreateNewCourse() *Course {
	return &Course{}
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
