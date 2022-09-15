package controller

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

func unwrapAndSendError(c *fiber.Ctx, err error) error {
	fmt.Println(err)
	switch t := err.(type) {
	case model.NotFoundError:
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	case model.ConflictError:
		return c.Status(fiber.StatusConflict).SendString(err.Error())
	case model.ForbiddenError:
		return c.Status(fiber.StatusForbidden).SendString(err.Error())
	case model.InternalServerError:
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	default:
		log.Printf("Could not unwrap error: %v of type: %T\n", err, t)
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
}
