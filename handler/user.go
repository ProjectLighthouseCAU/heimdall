package handler

import (
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type UserHandler struct {
	userService  service.UserService
	roleService  service.RoleService
	sessionStore *session.Store
}

func NewUserHandler(userService service.UserService,
	roleService service.RoleService,
	sessionStore *session.Store) UserHandler {
	return UserHandler{userService, roleService, sessionStore}
}

// @Summary      Get all users or query by name
// @Description  Get a list of all users or query a single user by name (returns single object instead of list). NOTE: registration_key is only included when querying a single user
// @Tags         Users
// @Produce      json
// @Param        name  query  string  false  "Username"
// @Success      200  {object}  []model.User
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users [get]
func (uc *UserHandler) GetAll(c *fiber.Ctx) error {
	// query users by name
	name := c.Query("name", "")
	if name != "" {
		return c.Next()
	}

	// return all users
	users, err := uc.userService.GetAll()
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(users)
}

// Documentation included in GetAll
func (uc *UserHandler) GetByName(c *fiber.Ctx) error {
	name := c.Query("name", "")
	if name == "" {
		return fiber.ErrBadRequest
	}
	user, err := uc.userService.GetByName(name)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(user)
}

// @Summary      Get user by id
// @Description  Get a user by its user id
// @Id 			 GetUserByName
// @Tags         Users
// @Produce      json
// @Param        id  path  int  true  "User ID"
// @Success      200  {object}  model.User
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id} [get]
func (uc *UserHandler) GetByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	user, err := uc.userService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(user)
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
} //@name LoginPayload

// @Summary      Login
// @Description  Log in with username and password (sets a cookie with the session id). Returns the full user information if the login was successful or the user is already logged in.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        payload  body  LoginPayload  true  "Username and Password"
// @Success      200  {object}  model.User
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal Server Error"
// @Router       /login [post]
func (uc *UserHandler) Login(c *fiber.Ctx) error {
	c.Accepts("application/json")
	var payload LoginPayload
	if err := c.BodyParser(&payload); err != nil {
		return UnwrapAndSendError(c, model.BadRequestError{Message: "Could not parse request body", Err: err})
	}
	session, err := uc.sessionStore.Get(c)
	if err != nil {
		return UnwrapAndSendError(c, model.InternalServerError{Message: "Could not get session", Err: err})
	}
	user, err := uc.userService.Login(payload.Username, payload.Password, session)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(user)
}

// @Summary      Logout
// @Description  Log out of the current session
// @Tags         Users
// @Accept       json
// @Produce      plain
// @Success      200  "OK"
// @Failure      401  "Unauthorized"
// @Failure      500  "Internal Server Error"
// @Router       /logout [post]
func (uc *UserHandler) Logout(c *fiber.Ctx) error {
	c.Accepts("application/json")
	session, err := uc.sessionStore.Get(c)
	if err != nil {
		return UnwrapAndSendError(c, model.InternalServerError{Message: "Could not get session", Err: err})
	}
	if err := uc.userService.Logout(session); err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

type RegisterPayload struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	Email           string `json:"email"`
	RegistrationKey string `json:"registration_key"` // snake case naming for decoding of x-www-form-urlencoded bodies
} //@name RegisterPayload

// @Summary      Register user
// @Description  Registers a new user using a registration key
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        payload  body  RegisterPayload  true  "Username, Password, Email, RegistrationKey"
// @Success      201  {object}  model.User
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      409  "Conflict"
// @Failure      500  "Internal Server Error"
// @Router       /register [post]
func (uc *UserHandler) Register(c *fiber.Ctx) error {
	c.Accepts("application/json")
	var payload RegisterPayload
	if err := c.BodyParser(&payload); err != nil {
		return UnwrapAndSendError(c, model.BadRequestError{Message: "Could not parse request body", Err: err})
	}
	session, err := uc.sessionStore.Get(c)
	if err != nil {
		return UnwrapAndSendError(c, model.InternalServerError{Message: "Could not get session", Err: err})
	}
	user, err := uc.userService.Register(payload.Username, payload.Password, payload.Email, payload.RegistrationKey, session)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

type CreateOrUpdateUserPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
} //@name CreateOrUpdateUserPayload

// @Summary      Create user
// @Description  Creates a new user
// @Tags         Users
// @Accept       json
// @Produce      plain
// @Param        payload  body  CreateOrUpdateUserPayload  true  "Username, Password, Email"
// @Success      201  "Created"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      409  "Conflict"
// @Failure      500  "Internal Server Error"
// @Router       /users [post]
func (uc *UserHandler) Create(c *fiber.Ctx) error {
	c.Accepts("application/json")
	var payload CreateOrUpdateUserPayload
	if err := c.BodyParser(&payload); err != nil {
		return UnwrapAndSendError(c, model.BadRequestError{Message: "Could not parse request body", Err: err})
	}

	err := uc.userService.Create(payload.Username, payload.Password, payload.Email)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// @Summary      Update user
// @Description  Updates a user (always updates all fields, partial updates currently not supported)
// @Tags         Users
// @Accept       json
// @Produce      plain
// @Param        id  path  int  true  "User ID"
// @Param        payload  body  CreateOrUpdateUserPayload  true  "Username, Password, Email"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      409  "Conflict"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id} [put]
func (uc *UserHandler) Update(c *fiber.Ctx) error {
	c.Accepts("application/json")
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	var payload CreateOrUpdateUserPayload
	if err := c.BodyParser(&payload); err != nil {
		return UnwrapAndSendError(c, model.BadRequestError{Message: "Could not parse request body", Err: err})
	}

	err := uc.userService.Update(uint(id), payload.Username, payload.Password, payload.Email)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Delete user
// @Description  Deletes a user given a user id
// @Tags         Users
// @Produce      plain
// @Param        id  path  int  true  "User ID"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id} [delete]
func (uc *UserHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := uc.userService.DeleteByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Get roles of user
// @Description  Get a list of roles that a user posesses
// @Tags         Users
// @Produce      json
// @Param        id  path  int  true  "User ID"
// @Success      200  {object}  []model.Role
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /users/{id}/roles [get]
func (uc *UserHandler) GetRolesOfUser(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	roles, err := uc.userService.GetRolesOfUser(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(roles)
}
