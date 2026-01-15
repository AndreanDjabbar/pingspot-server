package dto

type RegisterRequest struct {
	Username   string  `json:"username" validate:"required,min=3"`
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required,min=6"`
	FullName   string  `json:"fullName" validate:"required"`
	Provider   string  `json:"provider" validate:"required,oneof=EMAIL GOOGLE GITHUB"`
	ProviderID *string `json:"providerId"`
}

type LoginRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required,min=6"`
	IPAddress  string  `json:"-"`
	UserAgent  string  `json:"-"`
}

type ForgotPasswordEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6"`
	PasswordConfirmation string `json:"passwordConfirmation" validate:"required,eqfield=Password"`
	Email string `json:"email" validate:"required,email"`
}