package middleware

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

func NewSessionMiddleware(sessionStore *session.Store, userService service.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session, err := sessionStore.Get(c)
		if err != nil {
			return err
		}
		userId, ok := session.Get("userid").(uint)
		if !ok {
			return fiber.ErrUnauthorized
		}
		user, err := userService.GetByID(userId)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		c.Locals("user", user)

		// TODO: check API token in redis and create new if it has expired (doesn't exist), store: username and token
		// tokenKey := "user:" + fmt.Sprintf("%d", userId) + ":api_token"
		// apiToken, err := sessionStore.Storage.Get(tokenKey)
		// if err != nil || apiToken == nil {
		// 	// TODO: generate new apiToken
		// 	apiToken = []byte("API-TOK_1234-5678-9012-3456")
		// 	sessionStore.Storage.Set(tokenKey, apiToken, 3*24*time.Hour)
		// }

		// TODO: invalidate (delete) API token in redis if:
		// - username is changed
		// - user is deleted
		// - user manually generates a new token // TODO: implement

		return c.Next()
	}
}

func AllowRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*model.User)
		if user == nil {
			return fiber.ErrInternalServerError
		}
		if slices.ContainsFunc[[]model.Role](user.Roles, func(r model.Role) bool {
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
			return fiber.ErrInternalServerError
		}

		user := c.Locals("user").(*model.User)
		if id < 0 || uint(id) != user.ID {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}

func AllowRoleOrOwnUserId(role, pathParamUserId string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*model.User)
		if user == nil {
			return fiber.ErrInternalServerError
		}
		if slices.ContainsFunc[[]model.Role](user.Roles, func(r model.Role) bool {
			return r.Name == role
		}) {
			return c.Next()
		}

		id, err := c.ParamsInt(pathParamUserId)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		if id < 0 || uint(id) != user.ID {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}

/* TODO
- Get all users, query filter by name reveals registration key
- allow user to query their own registration key and role (not important since available through the /users route)

- implement api token generation, storage in redis and expiry + renewal
- figure out how to share user, api token and roles with beacon
*/
