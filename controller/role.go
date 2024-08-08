package controller

import (
	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

type RoleController struct {
	roleService service.RoleService
}

func NewRoleController(s service.RoleService) RoleController {
	return RoleController{
		roleService: s,
	}
}

func (rc *RoleController) Get(c *fiber.Ctx) error {
	// query roles by name
	name := c.Query("name", "")
	if name != "" {
		role, err := rc.roleService.GetByName(name)
		if err != nil {
			return unwrapAndSendError(c, err)
		}
		return c.JSON(role)
	}
	// return all roles
	roles, err := rc.roleService.GetAll()
	if err != nil {
		unwrapAndSendError(c, err)
	}
	return c.JSON(roles)
}

func (rc *RoleController) GetByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	role, err := rc.roleService.GetByID(uint(id))
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.JSON(role)
}

func (rc *RoleController) Create(c *fiber.Ctx) error {
	c.Accepts("json", "application/json", "application/x-www-form-urlencoded")
	payload := struct {
		Name string `json:"name"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rc.roleService.Create(payload.Name)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (rc *RoleController) Update(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	payload := struct {
		Name string `json:"name"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rc.roleService.Update(uint(id), payload.Name)
	if err != nil {
		return unwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (rc *RoleController) Delete(c *fiber.Ctx) error {
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

func (rc *RoleController) GetUsersOfRole(c *fiber.Ctx) error {
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

func (rc *RoleController) AddUserToRole(c *fiber.Ctx) error {
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

func (rc *RoleController) RemoveUserFromRole(c *fiber.Ctx) error {
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
