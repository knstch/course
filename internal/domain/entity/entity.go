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

type Email struct {
	NewEmail string `json:"newEmail"`
}

func CreateNewEmail() *Email {
	return &Email{}
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
