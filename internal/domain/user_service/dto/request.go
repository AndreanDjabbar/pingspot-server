package dto

type SaveUserProfileRequest struct {
	FullName 		  string  `json:"fullName" validate:"required"`
	Username		  *string  `json:"username"`
	Bio    	 		 *string `json:"bio" validate:"omitempty,max=255"`
	ProfilePicture   *string `json:"profilePicture" validate:"omitempty,max=255"`
	Gender   		 *string `json:"gender" validate:"omitempty,oneof=male female"`
	Birthday 	   	 *string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
}

type SaveUserSecurityRequest struct {
	CurrentPassword      string `json:"currentPassword" validate:"required,min=6"`
	CurrentPasswordConfirmation string `json:"currentPasswordConfirmation" validate:"required,eqfield=CurrentPassword"`
	NewPassword          string `json:"newPassword" validate:"required,min=6"`
	NewPasswordConfirmation   string `json:"newPasswordConfirmation" validate:"required,eqfield=NewPassword"`
}