package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

var (
	userPrefix     = config.GetString("CASBIN_USER_PREFIX", "user::")
	internalDomain = config.GetString("CASBIN_INTERNAL_DOMAIN", "internal")
)

// type CasbinMiddleware func(c *fiber.Ctx) error

func NewCasbinMiddleware(acs service.AccessControlService) fiber.Handler {
	// TODO: adapt to jwt middleware using c.Locals("user")
	return func(c *fiber.Ctx) error {
		c.Locals("user", "Testuser") // TODO: remove
		user, ok := c.Locals("user").(string)
		if !ok {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		ok, err := acs.GetEnforcer().EnforceWithMatcher(
			`r.dom == p.dom
			&& g(r.sub, p.sub)
			&& keyMatch2(r.obj, p.obj)
			&& regexMatch(r.act, p.act)`,
			internalDomain, userPrefix+user, c.Path(), c.Method(),
		)
		if err != nil {
			fmt.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if !ok {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.Next()
	}
}
