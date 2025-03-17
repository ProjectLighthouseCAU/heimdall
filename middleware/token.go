package middleware

import (
	"slices"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/repository"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
)

type TokenMiddleware fiber.Handler

func NewTokenMiddleware(userService *service.UserService, tokenRepository *repository.TokenRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		headers := c.GetReqHeaders()
		authHeader := headers["Authorization"]
		if len(authHeader) == 0 || authHeader[0] == "" {
			return fiber.ErrUnauthorized
		}
		// TODO: figure out how joins work with Gorm and query the user by token
		token, err := tokenRepository.FindByToken(authHeader[0])
		if err != nil {
			return fiber.ErrUnauthorized
		}
		user, err := userService.GetByID(token.UserID)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		c.Locals("user", user)
		return c.Next()
	}
}

func (tm *TokenMiddleware) AllowRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*model.User)
		if !ok {
			return fiber.ErrInternalServerError
		}
		if slices.ContainsFunc(user.Roles, func(r model.Role) bool {
			return r.Name == role
		}) {
			return c.Next()
		}
		return fiber.ErrForbidden
	}

}
