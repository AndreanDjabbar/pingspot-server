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
                errors["username"] = "Username is required"
            }
            if e.Tag() == "min" {
                errors["username"] = "Username must be at least 3 characters"
            }
        case "Email":
            if e.Tag() == "required" {
                errors["email"] = "Email is required"
            }
            if e.Tag() == "email" {
                errors["email"] = "Invalid email format"
            }
        case "Password":
            if e.Tag() == "required" {
                errors["password"] = "Password is required"
            }
            if e.Tag() == "min" {
                errors["password"] = "Password must be at least 6 characters"
            }
        case "FullName":
            if e.Tag() == "required" {
                errors["fullName"] = "Full name is required"
            }
        case "Provider":
            if e.Tag() == "required" {
                errors["provider"] = "Provider is required"
            }
            if e.Tag() == "oneof" {
                errors["provider"] = "Provider must be one of EMAIL, GOOGLE, or GITHUB"
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
        case "EmailOrUsername":
            if e.Tag() == "required" {
                errors["emailOrUsername"] = "Email or username is required"
            }
            if e.Tag() == "min" {
                errors["emailOrUsername"] = "Email or username must be at least 3 characters"
            }
        case "Password":
            if e.Tag() == "required" {
                errors["password"] = "Password is required"
            }
            if e.Tag() == "min" {
                errors["password"] = "Password must be at least 6 characters"
            }
        case "Provider":
            if e.Tag() == "required" {
                errors["provider"] = "Provider is required"
            }
            if e.Tag() == "oneof" {
                errors["provider"] = "Provider must be one of EMAIL, GOOGLE, or GITHUB"
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
                errors["email"] = "Email is required"
            }
            if e.Tag() == "email" {
                errors["email"] = "Invalid email format"
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
                errors["password"] = "Password is required"
            }
            if e.Tag() == "min" {
                errors["password"] = "Password must be at least 6 characters"
            }
        case "PasswordConfirmation":
            if e.Tag() == "required" {
                errors["passwordConfirmation"] = "Password confirmation is required"
            }
            if e.Tag() == "eqfield" {
                errors["passwordConfirmation"] = "Password confirmation must match the password"
            }
        case "Email":
            if e.Tag() == "required" {
                errors["email"] = "Email is required"
            }
            if e.Tag() == "email" {
                errors["email"] = "Invalid email format"
            }
        }
    }
    return errors
}