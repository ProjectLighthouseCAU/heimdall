package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

type RegistrationKeyController interface {
	GetAll(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
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

func (rkc *registrationKeyController) GetAll(c *fiber.Ctx) error {
	keys, err := rkc.registrationKeyService.GetAll()
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(keys)
}

func (rkc *registrationKeyController) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	keyStr := c.Query("key", "")
	if id < 0 && keyStr == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	var key *model.RegistrationKey
	var err error
	if id >= 0 {
		key, err = rkc.registrationKeyService.GetByID(uint(id))
	} else {
		key, err = rkc.registrationKeyService.GetByKey(keyStr)
	}
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
		Closed      bool      `json:"closed"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	regKey := model.RegistrationKey{
		Key:         payload.Key,
		Description: payload.Description,
		Permanent:   payload.Permanent,
		Closed:      payload.Closed,
		ExpiresAt:   payload.ExpiresAt,
	}
	err := rkc.registrationKeyService.Create(&regKey)
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
		Closed      bool      `json:"closed"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}

	key := model.RegistrationKey{
		Description: payload.Description,
		Permanent:   payload.Permanent,
		Closed:      payload.Closed,
		ExpiresAt:   payload.ExpiresAt,
	}
	key.ID = uint(id)
	err := rkc.registrationKeyService.Update(&key)
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
