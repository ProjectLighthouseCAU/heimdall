package handler

import (
	"bufio"
	"encoding/json"
	"log"

	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
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
	// TODO: maybe remove?
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

// @Summary      Get and subscribe to updates of a user's api token and roles
// @Description  If the initial request was successful, the connection is kept alive and updates are sent using server sent events (SSE).
// @Tags         Auth (internal)
// @Produce      json
// @Success      200  {object}  APIToken
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /authenticate [get]
// TODO: think about path
func (tc *TokenHandler) WatchAuthChanges(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	username := c.Query("username", "")
	if username == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	token := c.Query("token", "")
	if token == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	user, err := tc.userService.GetByName(username)
	if err != nil || token != user.ApiToken.Token {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	ch := tc.tokenService.SubscribeToChanges(user.Username)

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for m := range ch {
			json, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				tc.tokenService.UnsubscribeFromChanges(user.Username, ch)
				return
			}
			n, err := w.Write(json)
			if err != nil || n != len(json) {
				log.Println(err)
				tc.tokenService.UnsubscribeFromChanges(user.Username, ch)
				return
			}
			err = w.Flush()
			if err != nil {
				log.Println("Connection closed:", err)
				tc.tokenService.UnsubscribeFromChanges(user.Username, ch)
				return
			}
		}
	}))
	return nil
}
