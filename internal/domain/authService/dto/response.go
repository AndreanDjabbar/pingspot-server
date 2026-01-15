package dto

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type VerificationResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

type ForgotPasswordLinkVerificationResponse struct {
	Email string `json:"email"`
}
