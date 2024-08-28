package middleware

import (
	"slices"

	"github.com/ProjectLighthouseCAU/heimdall/controller"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func NewSessionMiddleware(sessionStore *session.Store,
	userService service.UserService,
	tokenService service.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session, err := sessionStore.Get(c)
		if err != nil {
			return err
		}
		userIdIntf := session.Get("userid")
		if userIdIntf == nil {
			return controller.UnwrapAndSendError(c, model.UnauthorizedError{})
		}
		userId, ok := userIdIntf.(uint)
		if !ok {
			return controller.UnwrapAndSendError(c, model.InternalServerError{})
		}
		user, err := userService.GetByID(userId)
		if err != nil {
			err := session.Destroy()
			if err != nil {
				return controller.UnwrapAndSendError(c, model.InternalServerError{Message: "Could not destroy session", Err: err})
			}
			return controller.UnwrapAndSendError(c, model.UnauthorizedError{})
		}
		c.Locals("user", user)
		tokenService.GenerateApiTokenIfNotExists(user)
		return c.Next()
	}
}

func AllowRole(role string) fiber.Handler {
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

func AllowOwnUserId(pathParamUserId string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt(pathParamUserId)
		if err != nil {
			return fiber.ErrBadRequest
		}
		user, ok := c.Locals("user").(*model.User)
		if !ok {
			return fiber.ErrInternalServerError
		}
		if id < 0 || uint(id) != user.ID {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}

func AllowRoleOrOwnUserId(role, pathParamUserId string) fiber.Handler {
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

		id, err := c.ParamsInt(pathParamUserId)
		if err != nil {
			return fiber.ErrBadRequest
		}

		if id < 0 || uint(id) != user.ID {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}
