package config

import (
	"fmt"
	"net/http"
	"pingspot/pkg/utils/env"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func InitGoogleAuth() error {
	googleClientId := env.GoogleClientID()
	googleClientSecret := env.GoogleClientSecret()
	googleCallbackURL := env.GoogleCallbackURL()
	googleSecretSessionKey := env.GoogleSecretSessionKey()

	if googleClientId == "" || googleClientSecret == "" || googleCallbackURL == "" || googleSecretSessionKey == "" {
		return fmt.Errorf("missing required Google Auth environment variables")
	}

	googleProvider := google.New(
		googleClientId,
		googleClientSecret,
		googleCallbackURL,
		"email", "profile",
	)
	googleProvider.SetPrompt("select_account")

	goth.UseProviders(googleProvider)

	maxAge := 60 * 60 * 24 * 30
	isProduction := env.IsProduction()
	httpOnly := env.IsHTTPOnly()

	store := sessions.NewCookieStore([]byte(googleSecretSessionKey))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = httpOnly
	store.Options.Secure = isProduction
	store.Options.SameSite = http.SameSiteNoneMode

	gothic.Store = store

	return nil
}
