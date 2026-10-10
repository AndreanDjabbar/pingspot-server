package validation

import (
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

func FormatSaveUserProfileValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "FullName":
			if e.Tag() == "required" {
				errors["fullName"] = "Full name is required"
			}
		case "Bio":
			if e.Tag() == "max" {
				errors["bio"] = "Bio must be at most 255 characters"
			}
		case "ProfilePicture":
			if e.Tag() == "max" {
				errors["avatar"] = "Avatar must be at most 255 characters"
			}
		case "Gender":
			if e.Tag() == "max" {
				errors["gender"] = "Gender must be at most 20 characters"
			}
			if e.Tag() == "oneof" {
				errors["gender"] = "Gender must be one of male or female"
			}
		case "Birthday":
			if e.Tag() == "datetime" {
				errors["birthday"] = "Birthday must be in YYYY-MM-DD format"
			}
		}
	}
	return errors
}

func FormatSaveUserSecurityValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "CurrentPassword":
			if e.Tag() == "required" {
				errors["currentPassword"] = "Current password is required"
			}
			if e.Tag() == "min" {
				errors["currentPassword"] = "Current password must be at least 6 characters"
			}
		case "CurrentPasswordConfirmation":
			if e.Tag() == "required" {
				errors["currentPasswordConfirmation"] = "Current password confirmation is required"
			}
			if e.Tag() == "eqfield" {
				errors["currentPasswordConfirmation"] = "Current password confirmation must match the current password"
			}
		case "NewPassword":
			if e.Tag() == "required" {
				errors["newPassword"] = "New password is required"
			}
			if e.Tag() == "min" {
				errors["newPassword"] = "New password must be at least 6 characters"
			}
		case "NewPasswordConfirmation":
			if e.Tag() == "required" {
				errors["newPasswordConfirmation"] = "New password confirmation is required"
			}
			if e.Tag() == "eqfield" {
				errors["newPasswordConfirmation"] = "New password confirmation must match the new password"
			}
		}
	}
	return errors
}

func FormatFollowValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "FollowingID":
			if e.Tag() == "required" {
				errors["followingID"] = "Following ID is required"
			}
		case "FollowingType":
			if e.Tag() == "required" {
				errors["followingType"] = "Following type is required"
			}
			if e.Tag() == "oneof" {
				errors["followingType"] = "Following type must be one of user or community"
			}
		}
	}
	return errors
}

func FormatGetFollowDataValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "FollowingID":
			if e.Tag() == "required" {
				errors["followingID"] = "Following ID is required"
			}
		case "FollowingType":
			if e.Tag() == "required" {
				errors["followingType"] = "Following type is required"
			}
			if e.Tag() == "oneof" {
				errors["followingType"] = "Following type must be one of user or community"
			}
		}
	}
	return errors
}

func FormatUpdateEmailNotificationPreferenceValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "IsDisableEmailNotification":
			if e.Tag() == "required" {
				errors["isDisableEmailNotification"] = "Email notification preference is required"
			}
		}
	}
	return errors
}
