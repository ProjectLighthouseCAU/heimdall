package handler

import (
	"bufio"
	"encoding/json"
	"log"
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
// @Success      200  {object}  model.Token
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

type UpdateTokenRequest struct {
	Permanent bool `json:"permanent"`
}

// @Summary      Update a user's API token (set permanent)
// @Description  Given a valid user id and new permanent status, sets the permanent status for the users current token
// @Tags         Users
// @Produce      json
// @Param        id  path  int  true  "User ID"
// @Success      200  {object}  model.Token
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id}/api-token [put]
func (tc *TokenHandler) Update(c *fiber.Ctx) error {
	c.Accepts("application/json")
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	var payload UpdateTokenRequest
	if err := c.BodyParser(&payload); err != nil {
		return UnwrapAndSendError(c, model.BadRequestError{Message: "Could not parse request body", Err: err})
	}
	user, err := tc.userService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	err = tc.tokenService.SetPermanent(user, payload.Permanent)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Renew a user's API token
// @Description  Given a valid user id, invalidates the current API token and generates a new one
// @Tags         Users
// @Produce      plain
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  model.Token
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

// INTERNAL API

type AuthRequest struct {
	Username string `json:"username"`
	Token    string `json:"api_token"`
}

var keepalive = []byte{'\r', '\n'} // keepalive message (just newline without content)

// @Summary      Get and subscribe to updates of a user's api token and roles
// @Description  If the initial request was successful, the connection is kept alive and updates are sent using server sent events (SSE).
// @Tags         Internal
// @Produce      json
// @Success      200  {object} AuthUpdateMessage
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal Server Error"
// @Router       /internal/authenticate/{username} [post]
func (tc *TokenHandler) WatchAuthChanges(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	user, ok := c.Locals("user").(*model.User)
	if !ok {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	username := c.Params("username", "")
	if username != user.Username {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// token is not permanent and expired
	if !user.ApiToken.Permanent && time.Now().After(user.ApiToken.ExpiresAt) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// prepare first response
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	resp := model.AuthUpdateMessage{
		Username:  user.Username,
		Token:     user.ApiToken.Token,
		ExpiresAt: user.ApiToken.ExpiresAt,
		Permanent: user.ApiToken.Permanent,
		Roles:     roles,
	}

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// send first response
		err := writeAndFlushJson(w, resp)
		if err != nil {
			return
		}

		// subscribe to changes for this user (and unsubscribe on return)
		ch := tc.tokenService.SubscribeToChanges(user.Username)
		defer tc.tokenService.UnsubscribeFromChanges(user.Username, ch)
		forwardMessagesToClient(ch, w)
	}))
	return nil
}

// @Summary      Get a list of all usernames
// @Description  Returns a list of all users names
// @Tags         Internal
// @Produce      json
// @Success      200  {object}  []model.UserUpdateMessage
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal Server Error"
// @Router       /internal/users [get]
func (tc *TokenHandler) GetUsernames(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	log.Println("GET USERNAMES")

	users, err := tc.userService.GetAll()
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	var messages []model.UserUpdateMessage
	for _, user := range users {
		messages = append(messages, model.UserUpdateMessage{Username: user.Username, Removed: false})
	}
	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// send first response
		err := writeAndFlushMultipleJson(w, messages)
		if err != nil {
			return
		}
		ch := tc.tokenService.SubscribeToUserCreateDeleteEvents()
		defer tc.tokenService.UnsubscribeFromUserCreateDeleteEvents(ch)
		forwardMessagesToClient(ch, w)
	}))

	return nil
}

func forwardMessagesToClient[T any](ch chan T, w *bufio.Writer) {
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
			err := writeAndFlushBytes(w, keepalive) // send keepalive message (just newline without content)
			if err != nil {
				return
			}
		}
	}
}

func writeAndFlushJson(w *bufio.Writer, value any) error {
	respJson, err := json.Marshal(value)
	if err != nil {
		return err
	}
	respJson = append(respJson, '\r', '\n')
	return writeAndFlushBytes(w, respJson)
}

func writeAndFlushMultipleJson[T any](w *bufio.Writer, values []T) error {
	var output []byte
	for _, v := range values {
		respJson, err := json.Marshal(v)
		if err != nil {
			return err
		}
		respJson = append(respJson, '\r', '\n')
		output = append(output, respJson...)
	}
	return writeAndFlushBytes(w, output)
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
