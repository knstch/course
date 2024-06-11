package entity

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

type Passwords struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type Emails struct {
	OldEmail string `json:"oldEmail"`
	NewEmail string `json:"newEmail"`
}
