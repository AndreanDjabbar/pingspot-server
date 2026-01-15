package validation

import (
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

func FormatRegisterValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Username":
			if e.Tag() == "required" {
				errors["username"] = "Username wajib diisi"
			}
			if e.Tag() == "min" {
				errors["username"] = "Username minimal 3 karakter"
			}
		case "Email":
			if e.Tag() == "required" {
				errors["email"] = "Email wajib diisi"
			}
			if e.Tag() == "email" {
				errors["email"] = "Email tidak valid"
			}
		case "Password":
			if e.Tag() == "required" {
				errors["password"] = "Password wajib diisi"
			}
			if e.Tag() == "min" {
				errors["password"] = "Password minimal 6 karakter"
			}
		case "FullName":
			if e.Tag() == "required" {
				errors["fullName"] = "Fullname wajib diisi"
			}
		case "Provider":
			if e.Tag() == "required" {
				errors["provider"] = "Provider wajib diisi"
			}
			if e.Tag() == "oneof" {
				errors["provider"] = "Provider harus salah satu dari EMAIL, GOOGLE, atau GITHUB"
			}
		}
	}
	return errors
}

func FormatLoginValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Email":
			if e.Tag() == "required" {
				errors["email"] = "Email wajib diisi"
			}
			if e.Tag() == "email" {
				errors["email"] = "Email tidak valid"
			}
		case "Password":
			if e.Tag() == "required" {
				errors["password"] = "Password wajib diisi"
			}
			if e.Tag() == "min" {
				errors["password"] = "Password minimal 6 karakter"
			}
		}
	}
	return errors
}

func FormatForgotPasswordEmailVerificationValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Email":
			if e.Tag() == "required" {
				errors["email"] = "Email wajib diisi"
			}
			if e.Tag() == "email" {
				errors["email"] = "Format email tidak valid"
			}
		}
	}
	return errors
}

func FormatForgotPasswordResetPasswordValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Password":
			if e.Tag() == "required" {
				errors["password"] = "Password wajib diisi"
			}
			if e.Tag() == "min" {
				errors["password"] = "Password minimal 6 karakter"
			}
		case "PasswordConfirmation":
			if e.Tag() == "required" {
				errors["passwordConfirmation"] = "Konfirmasi password wajib diisi"
			}
			if e.Tag() == "eqfield" {
				errors["passwordConfirmation"] = "Konfirmasi password harus sama dengan password"
			}
		case "Email":
			if e.Tag() == "required" {
				errors["email"] = "Email wajib diisi"
			}
			if e.Tag() == "email" {
				errors["email"] = "Format email tidak valid"
			}
		}
	}
	return errors
}