package handler

import (
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
)

type TokenHandler struct {
	tokenService service.TokenService
	userService  service.UserService
}

func NewTokenHandler(tokenService service.TokenService,
	userService service.UserService) TokenHandler {
	return TokenHandler{tokenService, userService}
}

// @Summary      Get a user's API token
// @Description  Given a valid user id, returns the username, API token, associated roles and expiration date
// @Tags         Users
// @Produce      json
// @Param        id  path  int  true  "User ID"
// @Success      200  {object}  APIToken
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id}/api-token [get]
func (tc *TokenHandler) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	user, err := tc.userService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	token, err := tc.tokenService.GetToken(user)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(token)
}

// @Summary      Renew a user's API token
// @Description  Given a valid user id, invalidates the current API token and generates a new one
// @Tags         Users
// @Produce      plain
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  APIToken
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id}/api-token [delete]
func (tc *TokenHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	user, err := tc.userService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	_, err = tc.tokenService.RegenerateApiToken(user)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
