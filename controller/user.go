package controller

import (
	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

type UserController interface {
	GetAll(c *fiber.Ctx) error
	GetByName(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	Register(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	GetRolesOfUser(c *fiber.Ctx) error
	AddRoleToUser(c *fiber.Ctx) error
	RemoveRoleFromUser(c *fiber.Ctx) error
}

type userController struct {
	userService service.UserService
}

var _ UserController = (*userController)(nil) // compile-time interface check

func NewUserController(s service.UserService) *userController {
	return &userController{
		userService: s,
	}
}

// @description	Returns a list of all users
// @produce		json
// @success		200	{array}	model.User
// @router			/users [get]
func (uc *userController) GetAll(c *fiber.Ctx) error {
	// query users by name
	name := c.Query("name", "")
	if name != "" {
		return c.Next()
	}

	// return all users
	users, err := uc.userService.GetAll()
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(users)
}

func (uc *userController) GetByName(c *fiber.Ctx) error {
	name := c.Query("name", "")
	if name == "" {
		return fiber.ErrBadRequest
	}
	user, err := uc.userService.GetByName(name)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(user)
}

// @description	Returns a user by id
// @produce		json
// @success		200	{object}	model.User
// @failure		400	{object}	string	"Bad Request"
// @failure		404	{object}	string	"Not Found"
// @router			/user/{id} [get]
func (uc *userController) GetByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	user, err := uc.userService.GetByID(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(user)
}

func (uc *userController) Login(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := uc.userService.Login(payload.Username, payload.Password, c)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (uc *userController) Logout(c *fiber.Ctx) error {
	if err := uc.userService.Logout(c); err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @description	Creates a user with a registration key (no extra auth needed)
// @id				RegisterUser
// @accept			json
// @produce		plain
// @success		201	{object}	string	"Created"
// @failure		400	{object}	string	"Bad Request"
// @failure		403	{object}	string	"Forbidden"
// @failure		500	{object}	string	"Internal Server Error"
// @failure		409	{object}	string	"Conflict"
// @router			/register [post]
func (uc *userController) Register(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")

	payload := struct {
		Username         string `json:"username"`
		Password         string `json:"password"`
		Email            string `json:"email"`
		Registration_Key string `json:"registration_key"` // snake case naming for decoding of x-www-form-urlencoded bodies
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := uc.userService.Register(payload.Username, payload.Password, payload.Email, payload.Registration_Key)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// @description	Creates a user without a registration key (permissions needed)
// @accept			json
// @produce		plain
// @success		201	{object}	string	"Created"
// @failure		400	{object}	string	"Bad Request"
// @failure		500	{object}	string	"Internal Server Error"
// @failure		409	{object}	string	"Conflict"
// @router			/user [post]
func (uc *userController) Create(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}

	err := uc.userService.Create(payload.Username, payload.Password, payload.Email)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (uc *userController) Update(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}

	err := uc.userService.Update(uint(id), payload.Username, payload.Password, payload.Email)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @description	Creates a user with a registration key (no extra auth needed)
// @id				RegisterUser
// @produce		plain
// @success		200	{object}	string	"OK"
// @failure		404	{object}	string	"Not Found"
// @router			/user/{id} [delete]
func (uc *userController) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := uc.userService.DeleteByID(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (uc *userController) GetRolesOfUser(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	roles, err := uc.userService.GetRolesOfUser(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(roles)
}

func (uc *userController) AddRoleToUser(c *fiber.Ctx) error {
	userid, _ := c.ParamsInt("userid", -1)
	roleid, _ := c.ParamsInt("roleid", -1)
	if userid < 0 || roleid < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := uc.userService.AddRoleToUser(uint(userid), uint(roleid))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (uc *userController) RemoveRoleFromUser(c *fiber.Ctx) error {
	userid, _ := c.ParamsInt("userid", -1)
	roleid, _ := c.ParamsInt("roleid", -1)
	if userid < 0 || roleid < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := uc.userService.RemoveRoleFromUser(uint(userid), uint(roleid))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
