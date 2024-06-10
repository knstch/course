package dto

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName      string
	Surname        string
	Verified       bool `gorm:"not null"`
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

func (user *User) SetStatusUnverified() *User {
	user.Verified = false
	return user
}

func (user *User) AddSubscriptionId(id uint) *User {
	user.SubscriptionId = id
	return user
}

type Credentials struct {
	gorm.Model
	Email    string `gorm:"not null"`
	Password string `gorm:"not null"`
}

func CreateNewCredentials() *Credentials {
	return &Credentials{}
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
