package controller

import (
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
)

type RegistrationKeyController struct {
	registrationKeyService service.RegistrationKeyService
}

func NewRegistrationKeyController(regKeyService service.RegistrationKeyService) RegistrationKeyController {
	return RegistrationKeyController{regKeyService}
}

// @Summary      Get all registration keys or query by key
// @Description  Get a list of all registration keys or query a single registration key by key (returns single object instead of list)
// @Tags         RegistrationKeys
// @Accept       json
// @Produce      json
// @Param        key  query  string  false  "Registration Key"
// @Success      200  {object}  []RegistrationKey
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /registration-keys [get]
func (rkc *RegistrationKeyController) Get(c *fiber.Ctx) error {
	// query registration keys by key (string value)
	keyStr := c.Query("key", "")
	if keyStr != "" {
		key, err := rkc.registrationKeyService.GetByKey(keyStr)
		if err != nil {
			return UnwrapAndSendError(c, err)
		}
		return c.JSON(key)
	}
	// return all keys
	keys, err := rkc.registrationKeyService.GetAll()
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(keys)
}

// @Summary      Get registration key by id
// @Description  Get a registration key by its id
// @Tags         RegistrationKeys
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "Registration Key ID"
// @Success      200  {object}  RegistrationKey
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /registration-keys/{id} [get]
func (rkc *RegistrationKeyController) GetByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	key, err := rkc.registrationKeyService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(key)
}

type CreateRegistrationKeyPayload struct {
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Permanent   bool      `json:"permanent"`
	ExpiresAt   time.Time `json:"expires_at"`
} //@name CreateRegistrationKeyPayload

// @Summary		 Create registration key
// @Description  Create a new registration key
// @Tags         RegistrationKeys
// @Accept       json
// @Produce      plain
// @Param        payload  body  CreateRegistrationKeyPayload  true  "key, description, permament, expires_at"
// @Success      201  "Created"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      409  "Conflict"
// @Failure      500  "Internal Server Error"
// @Router       /registration-keys [post]
func (rkc *RegistrationKeyController) Create(c *fiber.Ctx) error {
	c.Accepts("application/json")
	var payload CreateRegistrationKeyPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rkc.registrationKeyService.Create(payload.Key, payload.Description, payload.Permanent, payload.ExpiresAt)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

type UpdateRegistrationKeyPayload struct {
	Description string    `json:"description"`
	Permanent   bool      `json:"permanent"`
	ExpiresAt   time.Time `json:"expires_at"`
} //@name UpdateRegistrationKeyPayload

// @Summary		 Update registration key
// @Description  Upadte a registration key by its id
// @Tags         RegistrationKeys
// @Accept       json
// @Produce      plain
// @Param        id  path  int  true  "Registration Key ID"
// @Param        payload  body  UpdateRegistrationKeyPayload  true  "description, permament, expires_at"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      500  "Internal Server Error"
// @Router       /registration-keys/{id} [put]
func (rkc *RegistrationKeyController) Update(c *fiber.Ctx) error {
	c.Accepts("application/json")
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	var payload UpdateRegistrationKeyPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rkc.registrationKeyService.Update(uint(id), payload.Description, payload.Permanent, payload.ExpiresAt)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Delete registration key
// @Description  Delete a registration key by its id
// @Tags         RegistrationKeys
// @Produce      plain
// @Param        id  path  int  true  "Registration Key ID"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /registration-keys/{id} [delete]
func (rkc *RegistrationKeyController) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rkc.registrationKeyService.DeleteByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Get users of registration key
// @Description  Get a list of users that registered using this registration key by its id. NOTE: registration_key is not included for users
// @Tags         RegistrationKeys
// @Produce      json
// @Param        id  path  int  true  "Registration Key ID"
// @Success      200  {object}  []User
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /registration-keys/{id}/users [get]
func (rkc *RegistrationKeyController) GetUsersOfKey(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	key, err := rkc.registrationKeyService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(key.Users)
}
