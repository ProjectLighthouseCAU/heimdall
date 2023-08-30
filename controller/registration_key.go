package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

type RegistrationKeyController interface {
	Get(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type registrationKeyController struct {
	registrationKeyService service.RegistrationKeyService
}

var _ RegistrationKeyController = (*registrationKeyController)(nil) // compile-time interface check

func NewRegistrationKeyController(r service.RegistrationKeyService) *registrationKeyController {
	return &registrationKeyController{
		registrationKeyService: r,
	}
}

func (rkc *registrationKeyController) Get(c *fiber.Ctx) error {
	// query registration keys by key (string value)
	keyStr := c.Query("key", "")
	if keyStr != "" {
		key, err := rkc.registrationKeyService.GetByKey(keyStr)
		if err != nil {
			return unwrapAndSendError(c, err)
		}
		return c.JSON(key)
	}
	// return all keys
	keys, err := rkc.registrationKeyService.GetAll()
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(keys)
}

func (rkc *registrationKeyController) GetByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	key, err := rkc.registrationKeyService.GetByID(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(key)
}

func (rkc *registrationKeyController) Create(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	payload := struct {
		Key         string    `json:"key"`
		Description string    `json:"description"`
		Permanent   bool      `json:"permanent"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rkc.registrationKeyService.Create(payload.Key, payload.Description, payload.Permanent, payload.ExpiresAt)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (rkc *registrationKeyController) Update(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	payload := struct {
		Description string    `json:"description"`
		Permanent   bool      `json:"permanent"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rkc.registrationKeyService.Update(uint(id), payload.Description, payload.Permanent, payload.ExpiresAt)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (rkc *registrationKeyController) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rkc.registrationKeyService.DeleteByID(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
