package controller

import (
	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

type RoleController interface {
	GetAll(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	GetUsersOfRole(c *fiber.Ctx) error
	AddUserToRole(c *fiber.Ctx) error
	RemoveUserFromRole(c *fiber.Ctx) error
}

type roleController struct {
	roleService service.RoleService
}

var _ RoleController = (*roleController)(nil) // compile-time interface check

func NewRoleController(s service.RoleService) *roleController {
	return &roleController{
		roleService: s,
	}
}

func (rc *roleController) GetAll(c *fiber.Ctx) error {
	roles, err := rc.roleService.GetAll()
	if err != nil {
		unwrapAndSendError(c, err)
	}
	return c.JSON(roles)
}

func (rc *roleController) Get(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	name := c.Query("name", "")
	if id < 0 && name == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	var role *model.Role
	var err error
	if id >= 0 { // id takes precedence over name
		role, err = rc.roleService.GetByID(uint(id))
	} else {
		role, err = rc.roleService.GetByName(name)
	}
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(role)
}

func (rc *roleController) Create(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	payload := struct {
		Name string `json:"name"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	if payload.Name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Empty field")
	}
	role := model.Role{
		Name: payload.Name,
	}
	err := rc.roleService.Create(&role)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (rc *roleController) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rc.roleService.DeleteByID(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (rc *roleController) GetUsersOfRole(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	users, err := rc.roleService.GetUsersOfRole(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(users)
}

func (rc *roleController) AddUserToRole(c *fiber.Ctx) error {
	roleid, _ := c.ParamsInt("roleid", -1)
	userid, _ := c.ParamsInt("userid", -1)
	if roleid < 0 || userid < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rc.roleService.AddUserToRole(uint(roleid), uint(userid))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (rc *roleController) RemoveUserFromRole(c *fiber.Ctx) error {
	roleid, _ := c.ParamsInt("roleid", -1)
	userid, _ := c.ParamsInt("userid", -1)
	if roleid < 0 || userid < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rc.roleService.RemoveUserFromRole(uint(roleid), uint(userid))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
