package middleware

import (
	"github.com/gofiber/fiber/v3"
	"context"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v3/middleware/session"

)

func VerifyTokenIdMiddleWare(firebaseAuthClient *auth.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := context.Background()
		sess := session.FromContext(c)
		var UUID string
		var idToken string
		if v := sess.Get("UUID"); v == nil {
			header := c.GetReqHeaders()["Authorization"]
			if len(header) > 0 {
				idToken = header[0]
			}
			parts := strings.Split(idToken, " ")
			token, err := firebaseAuthClient.VerifyIDToken(ctx, parts[1])
			if err != nil {
				return err
			}
			UUID = token.UID
			sess.Set("UUID", UUID)
		}

		return c.Next()
	}
}
