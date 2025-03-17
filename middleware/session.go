package middleware

import (
	"slices"

	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type SessionMiddleware fiber.Handler

func NewSessionMiddleware(sessionStore *session.Store,
	userService service.UserService,
	tokenService service.TokenService) SessionMiddleware {
	return func(c *fiber.Ctx) error {
		session, err := sessionStore.Get(c)
		if err != nil {
			return err
		}
		userId, ok := session.Get("userid").(uint)
		if !ok {
			return handler.UnwrapAndSendError(c, model.UnauthorizedError{})
		}

		user, err := userService.GetByID(userId)
		if err != nil { // user was deleted
			if err := session.Destroy(); err != nil { // destroy this invalid session
				return handler.UnwrapAndSendError(c, model.InternalServerError{Message: "Could not destroy session", Err: err})
			}
			return handler.UnwrapAndSendError(c, model.UnauthorizedError{})
		}

		// TODO: do we need to destroy this session if the username was changed?
		username, ok := session.Get("username").(string)
		if !ok || username != user.Username {
			if err := session.Destroy(); err != nil { // destroy this session that was authenticated with the old username
				return handler.UnwrapAndSendError(c, model.InternalServerError{Message: "Could not destroy session", Err: err})
			}
			return handler.UnwrapAndSendError(c, model.UnauthorizedError{})
		}

		// check if password was changed
		password, ok := session.Get("password").(string)
		if !ok || password != user.Password {
			if err := session.Destroy(); err != nil { // destroy this session that was authenticated with the old password
				return handler.UnwrapAndSendError(c, model.InternalServerError{Message: "Could not destroy session", Err: err})
			}
			return handler.UnwrapAndSendError(c, model.UnauthorizedError{})
		}

		c.Locals("user", user)
		tokenService.GenerateApiTokenIfNotExists(user)
		return c.Next()
	}
}

func (s *SessionMiddleware) AllowRole(role string) fiber.Handler {
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

func (s *SessionMiddleware) AllowOwnUserId(pathParamUserId string) fiber.Handler {
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

func (s *SessionMiddleware) AllowRoleOrOwnUserId(role, pathParamUserId string) fiber.Handler {
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
