package dto

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName      string
	Surname        string
	Credentials    Credentials
	CredentialsId  *uint `gorm:"not null"`
	Subscription   Subscription
	SubscriptionId uint
	PhoneNumber    uint
}

func CreateNewUser() *User {
	return &User{}
}

func (user *User) AddCredentialsId(id *uint) *User {
	user.CredentialsId = id
	return user
}

func (user *User) AddSubscriptionId(id *uint) *User {
	user.SubscriptionId = *id
	return user
}

type Credentials struct {
	gorm.Model
	Email    string `gorm:"not null"`
	Password string `gorm:"not null"`
	Verified bool   `gorm:"not null, default:false"`
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

type Subscription struct {
	gorm.Model
	SubscriptionType string
	DueTo            *time.Time
}

func CreateNewSubscription() *Subscription {
	return &Subscription{}
}

func (subscription *Subscription) AddSubscriptionType(subscriptionType string) *Subscription {
	subscription.SubscriptionType = subscriptionType
	return subscription
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
