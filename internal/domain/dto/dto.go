package dto

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName      string
	Surname        string
	Verified       *bool `gorm:"not null"`
	Credentials    Credentials
	CredentialsId  *uint
	Subscription   Subscription
	SubscriptionId Subscription
}

type Credentials struct {
	gorm.Model
	Email    string `gorm:"not null"`
	Password string `gorm:"not null"`
}

type Subscription struct {
	gorm.Model
	SubscriptionType string
	DueTo            time.Time
}

type AccessToken struct {
	gorm.Model
	User      User
	UserId    uint   `gorm:"not null"`
	Token     string `gorm:"not null"`
	Available bool   `gorm:"not null"`
}
