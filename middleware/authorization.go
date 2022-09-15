// DEPRECATED
package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// TODO: maybe rework all of this to use the contained permission system
// ctx locals provides subject (user id from jwt claims)
// ctx method provides verb (translate HTTP Verb to CRUD)
// ctx path provides object (resource path)

func Authorize(isAuthorized fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO:
		// user, ok := c.Locals("user").(*jwt.Token)
		// if !ok {
		// 	return errors.New("invalid jwt token in authorization middleware")
		// }
		// method := c.Method()
		// path := c.Path()

		err := isAuthorized(c)
		if err != nil {
			return err
		}
		return c.Next()
	}
}

func httpToCrwd(httpMethod string) (string, error) {
	switch httpMethod {
	case "GET":
		return "READ", nil
	case "POST":
		return "CREATE", nil
	case "PATCH":
		return "WRITE", nil
	case "DELETE":
		return "DELETE", nil
	default:
		return "", errors.New("invalid http method")
	}
}

func AuthorizeGET(isAuthorized fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != "GET" {
			return c.Next()
		}
		err := isAuthorized(c)
		if err != nil {
			return err
		}
		return c.Next()
	}
}

func AuthorizePOST(isAuthorized fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != "POST" {
			return c.Next()
		}
		err := isAuthorized(c)
		if err != nil {
			return err
		}
		return c.Next()
	}
}

func AuthorizePATCH(isAuthorized fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != "PATCH" {
			return c.Next()
		}
		err := isAuthorized(c)
		if err != nil {
			return err
		}
		return c.Next()
	}
}

func AuthorizeDELETE(isAuthorized fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != "DELETE" {
			return c.Next()
		}
		err := isAuthorized(c)
		if err != nil {
			return err
		}
		return c.Next()
	}
}
