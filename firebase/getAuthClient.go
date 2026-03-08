package firebase_self

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"

)

func GetAuthClient(defaultApp *firebase.App) *auth.Client {
	defaultClient, err := defaultApp.Auth(context.Background())
		if err != nil {
			log.Fatalf("error getting Auth client: %v\n", err)
		}
	return defaultClient
}
