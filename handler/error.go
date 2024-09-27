package handler

import (
	"log"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/gofiber/fiber/v2"
)

type APIError struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func UnwrapAndSendError(c *fiber.Ctx, err error) error {
	if httpErr, ok := err.(model.HTTPError); ok {
		return c.Status(httpErr.Status()).JSON(APIError{Status: httpErr.Status(), Error: httpErr.Error()})
	}
	log.Printf("Could not unwrap error into HTTPError: %v\n", err)
	return c.SendStatus(fiber.StatusInternalServerError)
}
