package auth

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/umaaamm/contact/internal/users"
	"github.com/umaaamm/contact/pkg/jwt"
)

type contextKey struct {
	name string
}

var userCtxKey = &contextKey{"user"}

// Middleware for Fiber (returns fiber.Handler)
func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")

		if header == "" {
			return c.Next()
		}

		username, err := jwt.ParseToken(header)
		if err != nil {
			return c.Status(fiber.StatusForbidden).SendString("Invalid token")
		}

		id, err := users.GetUserIdByUsername(username)
		if err != nil {
			return c.Next()
		}

		user := &users.User{
			ID:       id,
			Username: username,
		}

		ctx := context.WithValue(c.UserContext(), userCtxKey, user)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// Get user from context in resolvers
func ForContext(ctx context.Context) *users.User {
	raw, _ := ctx.Value(userCtxKey).(*users.User)

	if raw == nil {
		return nil
	}

	return raw
}
