package controller

import (
	"fmt"
	"log"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/gofiber/fiber/v2"
)

func UnwrapAndSendError(c *fiber.Ctx, err error) error {
	fmt.Println(err)
	switch t := err.(type) {
	case model.BadRequestError:
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("%d %s", fiber.StatusBadRequest, err.Error()))
	case model.NotFoundError:
		return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("%d %s", fiber.StatusNotFound, err.Error()))
	case model.ConflictError:
		return c.Status(fiber.StatusConflict).SendString(fmt.Sprintf("%d %s", fiber.StatusConflict, err.Error()))
	case model.ForbiddenError:
		return c.Status(fiber.StatusForbidden).SendString(fmt.Sprintf("%d %s", fiber.StatusForbidden, err.Error()))
	case model.UnauthorizedError:
		return c.Status(fiber.StatusUnauthorized).SendString(fmt.Sprintf("%d %s", fiber.StatusUnauthorized, err.Error()))
	case model.InternalServerError:
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("%d %s", fiber.StatusInternalServerError, err.Error()))
	default:
		log.Printf("Could not unwrap error: %v of type: %T\n", err, t)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
}
