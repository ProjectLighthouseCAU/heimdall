package router

import (
	"github.com/gofiber/fiber/v2"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/controller"
	"lighthouse.uni-kiel.de/lighthouse-api/middleware"
)

type Router struct {
	app                       *fiber.App
	userController            controller.UserController
	registrationKeyController controller.RegistrationKeyController
	roleController            controller.RoleController
	sessionMiddleware         fiber.Handler
}

func NewRouter(app *fiber.App, uc controller.UserController, rkc controller.RegistrationKeyController, rc controller.RoleController, sessionMiddleware fiber.Handler) Router {
	return Router{
		app:                       app,
		userController:            uc,
		registrationKeyController: rkc,
		roleController:            rc,
		sessionMiddleware:         sessionMiddleware,
	}
}

func (r *Router) Init() {
	r.app.Post("/register", r.userController.Register)
	r.app.Post("/login", r.userController.Login)

	r.app.Use(r.sessionMiddleware)
	r.app.Post("/logout", r.userController.Logout)
	r.initUserRoutes()
	r.initRegistrationKeyRoutes()
	r.initRoleRoutes()
}

/*
Permissions:
unauthorized:
	/register
	/login
admin: /**
user:
	/logout
	(GET /users)
	GET /users/<own-id>
	PUT /users/<own-id>
	DELETE /users/<own-id>
	GET /users/<own-id>/roles

	(GET /roles/<own-roles-id>)
	(GET /registration-keys/<own-reg-key-id>)
*/

var admin = config.GetString("ADMIN_ROLENAME", "admin")

func (r *Router) initUserRoutes() {
	users := r.app.Group("/users")
	users.Get("", r.userController.GetAll, middleware.AllowRole(admin), r.userController.GetByName)
	users.Get("/:id<int>", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userController.GetByID)
	users.Post("", middleware.AllowRole(admin), r.userController.Create)
	users.Put("/:id<int>", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userController.Update)
	users.Delete("/:id<int>", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userController.Delete)
	users.Get("/:id<int>/roles", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userController.GetRolesOfUser)
	users.Put("/:userid<int>/roles/:roleid<int>", middleware.AllowRole(admin), r.userController.AddRoleToUser)
	users.Delete("/:userid<int>/roles/:roleid<int>", middleware.AllowRole(admin), r.userController.RemoveRoleFromUser)
}

func (r *Router) initRegistrationKeyRoutes() {
	keys := r.app.Group("/registration-keys", middleware.AllowRole(admin))
	keys.Get("", r.registrationKeyController.Get)
	keys.Get("/:id<int>", r.registrationKeyController.GetByID)
	keys.Post("", r.registrationKeyController.Create)
	keys.Put("/:id<int>", r.registrationKeyController.Update)
	keys.Delete("/:id<int>", r.registrationKeyController.Delete)
	keys.Get("/:id<int>/users", r.registrationKeyController.GetUsersOfKey)
}

func (r *Router) initRoleRoutes() {
	roles := r.app.Group("/roles", middleware.AllowRole(admin))
	roles.Get("", r.roleController.Get)
	roles.Get("/:id<int>", r.roleController.GetByID)
	roles.Post("", r.roleController.Create)
	roles.Put("/:id<int>", r.roleController.Update)
	roles.Delete("/:id<int>", r.roleController.Delete)
	roles.Get("/:id<int>/users", r.roleController.GetUsersOfRole)
	roles.Put("/:roleid<int>/users/:userid<int>", r.roleController.AddUserToRole)
	roles.Delete("/:roleid<int>/users/:userid<int>", r.roleController.RemoveUserFromRole)
}

func (r *Router) ListRoutes() map[string][]string {
	endpoints := make(map[string][]string)
	for _, group := range r.app.Stack() {
		for _, endpoint := range group {
			// fmt.Printf("%s %s\n", endpoint.Method, endpoint.Path)
			if endpoints[endpoint.Path] == nil {
				endpoints[endpoint.Path] = []string{}
			}
			if !contains(endpoints[endpoint.Path], endpoint.Method) {
				endpoints[endpoint.Path] = append(endpoints[endpoint.Path], endpoint.Method)
			}
		}
	}
	return endpoints
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
