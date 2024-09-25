package handler

import (
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) RoleHandler {
	return RoleHandler{roleService}
}

// @Summary      Get all roles or query by name
// @Description  Get a list of all roles or query a single role by name (returns single object instead of list)
// @Tags         Roles
// @Produce      json
// @Param        name  query  string  false  "Role name"
// @Success      200  {object}  Role
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /roles [get]
func (rc *RoleHandler) Get(c *fiber.Ctx) error {
	// query roles by name
	name := c.Query("name", "")
	if name != "" {
		role, err := rc.roleService.GetByName(name)
		if err != nil {
			return UnwrapAndSendError(c, err)
		}
		return c.JSON(role)
	}
	// return all roles
	roles, err := rc.roleService.GetAll()
	if err != nil {
		UnwrapAndSendError(c, err)
	}
	return c.JSON(roles)
}

// @Summary      Get role by id
// @Description  Get a role by its role id
// @Tags         Roles
// @Produce      json
// @Param        id  path  int  true  "Role ID"
// @Success      200  {object}  Role
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /roles/{id} [get]
func (rc *RoleHandler) GetByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	role, err := rc.roleService.GetByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(role)
}

type CreateOrUpdateRolePayload struct {
	Name string `json:"name"`
} //@name CreateOrUpdateRolePayload

// @Summary      Create role
// @Description  Create a new role
// @Tags         Roles
// @Accept       json
// @Produce      plain
// @Param        payload  body  CreateOrUpdateRolePayload  true  "Name"
// @Success      201  "Created"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      409  "Conflict"
// @Failure      500  "Internal Server Error"
// @Router       /roles [post]
func (rc *RoleHandler) Create(c *fiber.Ctx) error {
	c.Accepts("application/json")
	var payload CreateOrUpdateRolePayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rc.roleService.Create(payload.Name)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// @Summary      Update role
// @Description  Update a new role by its user id
// @Tags         Roles
// @Accept       json
// @Produce      plain
// @Param        id  path  int  true  "Role ID"
// @Param        payload  body  CreateOrUpdateRolePayload  true  "Name"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      409  "Conflict"
// @Failure      500  "Internal Server Error"
// @Router       /roles/{id} [put]
func (rc *RoleHandler) Update(c *fiber.Ctx) error {
	c.Accepts("application/json")
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	var payload CreateOrUpdateRolePayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not parse request body")
	}
	err := rc.roleService.Update(uint(id), payload.Name)
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Delete role
// @Description  Delete a role by its role id
// @Tags         Roles
// @Produce      plain
// @Param        id  path  int  true  "Role ID"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /roles/{id} [delete]
func (rc *RoleHandler) Delete(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rc.roleService.DeleteByID(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Get users of role
// @Description  Get a list of users that have a role by its role id. NOTE: registration_key is not included for users
// @Tags         Roles
// @Produce      json
// @Param        id  path  int  true  "Role ID"
// @Success      200  {object}  []User
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /roles/{id}/users [get]
func (rc *RoleHandler) GetUsersOfRole(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", -1)
	if id < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	users, err := rc.roleService.GetUsersOfRole(uint(id))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.JSON(users)
}

// @Summary      Add user to role
// @Description  Add a user (by its user id) to a role (by its role id)
// @Tags         Roles
// @Produce      plain
// @Param        roleid  path  int  true  "Role ID"
// @Param        userid  path  int  true  "User ID"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /roles/{roleid}/users/{userid} [put]
func (rc *RoleHandler) AddUserToRole(c *fiber.Ctx) error {
	roleid, _ := c.ParamsInt("roleid", -1)
	userid, _ := c.ParamsInt("userid", -1)
	if roleid < 0 || userid < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rc.roleService.AddUserToRole(uint(roleid), uint(userid))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// @Summary      Remove user from role
// @Description  Remove a user (by its user id) from a role (by its role id)
// @Tags         Roles
// @Produce      plain
// @Param        roleid  path  int  true  "Role ID"
// @Param        userid  path  int  true  "User ID"
// @Success      200  "OK"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      403  "Forbidden"
// @Failure      404  "Not Found"
// @Failure      500  "Internal Server Error"
// @Router       /roles/{roleid}/users/{userid} [delete]
func (rc *RoleHandler) RemoveUserFromRole(c *fiber.Ctx) error {
	roleid, _ := c.ParamsInt("roleid", -1)
	userid, _ := c.ParamsInt("userid", -1)
	if roleid < 0 || userid < 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	err := rc.roleService.RemoveUserFromRole(uint(roleid), uint(userid))
	if err != nil {
		return UnwrapAndSendError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
