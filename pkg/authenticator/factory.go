package authenticator

import (
	"os"
)

func GetAuthenticator(name string) Authenticator {
	switch name {
	case "google":
		googleAuthenticator := NewGoogleAuthenticator(
			os.Getenv("GOOGLE_AUTHENTICATOR_SECRET_KEY"),
		)
		return &googleAuthenticator
	}

	return nil
}
