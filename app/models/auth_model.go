package models

type SignUp struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Username string `json:"username" validate:"required,lte=255"`
	Phone    string `json:"phone" validate:"required,lte=20"`
	Gender   string `json:"gender" validate:"omitempty,oneof=male female"`
	Password string `json:"password" validate:"required,lte=255"`
	UserRole string `json:"user_role,omitempty"`
}

type SignIn struct {
	Email    string `json:"email" validate:"required,email,lte=255"`
	Password string `json:"password" validate:"required,lte=255"`
}

type VerifyOTP struct {
	Email string `json:"email" validate:"required,email,lte=255"`
	OTP   string `json:"otp" validate:"required,len=4"`
}
