package handler

import (
	"bufio"
	"encoding/json"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/model"
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
	return c.JSON(user.ApiToken)
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
	if err = tc.tokenService.RegenerateApiToken(user); err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Get a list of all usernames
// @Description  Returns a list of all users names
// @Tags         Auth (internal)
// @Produce      json
// @Success      200  {object}  []string
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal Server Error"
// @Router       /internal/users [get]
func (tc *TokenHandler) GetUsernames(c *fiber.Ctx) error {
	users, err := tc.userService.GetAll()
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	var usernames []string
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}
	return c.JSON(usernames)
}

type AuthRequest struct {
	Username string `json:"username"`
	Token    string `json:"api_token"`
}

// @Summary      Get and subscribe to updates of a user's api token and roles
// @Description  If the initial request was successful, the connection is kept alive and updates are sent using server sent events (SSE).
// @Tags         Auth (internal)
// @Produce      json
// @Success      200  {object} AuthUpdateMessage
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal Server Error"
// @Router       /internal/authenticate [post]
func (tc *TokenHandler) WatchAuthChanges(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	var request AuthRequest
	if err := c.BodyParser(&request); err != nil {
		return UnwrapAndSendError(c, model.BadRequestError{Message: "Could not parse request body", Err: err})
	}

	if request.Username == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if request.Token == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	user, err := tc.userService.GetByName(request.Username)
	// user not found, has no token, tokens do not match or token is expired
	if err != nil || user.ApiToken == nil || request.Token != user.ApiToken.Token || time.Now().After(user.ApiToken.ExpiresAt) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// prepare first response
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	resp := model.AuthUpdateMessage{
		Username:        user.Username,
		Token:           user.ApiToken.Token,
		ExpiresAt:       user.ApiToken.ExpiresAt,
		Roles:           roles,
		UsernameInvalid: false,
	}

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// send first response
		writeAndFlushJson(w, resp)

		// subscribe to changes for this user (and unsubscribe on return)
		ch := tc.tokenService.SubscribeToChanges(user.Username)
		defer tc.tokenService.UnsubscribeFromChanges(user.Username, ch)
		for {
			select {
			case m, ok := <-ch:
				if !ok {
					return
				}
				err := writeAndFlushJson(w, m) // send update
				if err != nil {
					return
				}
			case <-time.After(time.Second): // detect closed connection
				err := writeAndFlushBytes(w, []byte{'\r', '\n'}) // send keepalive message (just newline without content)
				if err != nil {
					return
				}
			}
		}
	}))
	return nil
}

func writeAndFlushJson(w *bufio.Writer, value any) error {
	respJson, err := json.Marshal(value)
	if err != nil {
		return err
	}
	respJson = append(respJson, '\r', '\n')
	return writeAndFlushBytes(w, respJson)
}

func writeAndFlushBytes(w *bufio.Writer, bs []byte) error {
	n, err := w.Write(bs)
	if err != nil || n != len(bs) {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}
