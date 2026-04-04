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
				errors["fullName"] = "Full name wajib diisi"
			}
		case "Bio":
			if e.Tag() == "max" {
				errors["bio"] = "Bio maksimal 255 karakter"
			}
		case "ProfilePicture":
			if e.Tag() == "max" {
				errors["avatar"] = "Avatar maksimal 255 karakter"
			}
		case "Gender":
			if e.Tag() == "max" {
				errors["gender"] = "Gender maksimal 20 karakter"
			}
			if e.Tag() == "oneof" {
				errors["gender"] = "Gender harus salah satu antara male atau female"
			}
		case "Birthday":
			if e.Tag() == "datetime" {
				errors["birthday"] = "Birthday harus dalam format YYYY-MM-DD"
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
				errors["currentPassword"] = "Kata Sandi saat ini wajib diisi"
			}
			if e.Tag() == "min" {
				errors["currentPassword"] = "Kata Sandi saat ini minimal 6 karakter"
			}
		case "CurrentPasswordConfirmation":
			if e.Tag() == "required" {
				errors["currentPasswordConfirmation"] = "Konfirmasi Kata Sandi saat ini wajib diisi"
			}
			if e.Tag() == "eqfield" {
				errors["currentPasswordConfirmation"] = "Konfirmasi kata sandi saat ini harus sama dengan kata sandi saat ini"
			}
		case "NewPassword":
			if e.Tag() == "required" {
				errors["newPassword"] = "Kata sandi baru wajib diisi"
			}
			if e.Tag() == "min" {
				errors["newPassword"] = "Kata sandi baru minimal 6 karakter"
			}
		case "NewPasswordConfirmation":
			if e.Tag() == "required" {
				errors["newPasswordConfirmation"] = "Konfirmasi kata sandi baru wajib diisi"
			}
			if e.Tag() == "eqfield" {
				errors["newPasswordConfirmation"] = "Konfirmasi kata sandi baru harus sama dengan kata sandi baru"
			}
		}
	}
	return errors
}