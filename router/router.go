package router

import (
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"lighthouse.uni-kiel.de/lighthouse-api/auth"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/controller"
)

type Router interface {
	Init()
}

type router struct {
	app                       *fiber.App
	userController            controller.UserController
	registrationKeyController controller.RegistrationKeyController
	roleController            controller.RoleController
	casbinMiddleware          fiber.Handler
}

var (
	_ Router = (*router)(nil) // compile-time interface check
)

func NewRouter(app *fiber.App, uc controller.UserController, rkc controller.RegistrationKeyController, rc controller.RoleController, cm fiber.Handler) *router {
	return &router{
		app:                       app,
		userController:            uc,
		registrationKeyController: rkc,
		roleController:            rc,
		casbinMiddleware:          cm,
	}
}

func (r *router) Init() {
	r.app.Post("/register", r.userController.Register)

	// r.initJWTMiddleware()
	r.app.Use(r.casbinMiddleware)
	r.initUserRoutes()
	r.initRegistrationKeyRoutes()
	r.initRoleRoutes()
}

func (r *router) initJWTMiddleware() {
	signingKey := []byte(config.GetString("JWT_PRIVATE_KEY", auth.NewRandomString(32)))
	r.app.Use(jwtware.New(jwtware.Config{
		SigningKey: signingKey,
	}))
}

func (r *router) initUserRoutes() {
	r.app.Get("/users", r.userController.GetAll)
	user := r.app.Group("/user")
	user.Get("/:id", r.userController.Get)
	user.Get("/", r.userController.Get)
	user.Post("/", r.userController.Create)
	user.Put("/:id", r.userController.Update)
	user.Delete("/:id", r.userController.Delete)
	user.Get("/:id/roles", r.userController.GetRolesOfUser)
	user.Put("/:userid/role/:roleid", r.userController.AddRoleToUser)
	user.Delete("/:userid/role/:roleid", r.userController.RemoveRoleFromUser)
}

func (r *router) initRegistrationKeyRoutes() {
	r.app.Get("/registration-keys", r.registrationKeyController.GetAll)
	keys := r.app.Group("/registration-key")
	keys.Get("/:id", r.registrationKeyController.Get)
	keys.Get("/", r.registrationKeyController.Get)
	keys.Post("/", r.registrationKeyController.Create)
	keys.Put("/:id", r.registrationKeyController.Update)
	keys.Delete("/:id", r.registrationKeyController.Delete)
}

func (r *router) initRoleRoutes() {
	r.app.Get("/roles", r.roleController.GetAll)
	role := r.app.Group("/role")
	role.Get("/:id", r.roleController.Get)
	role.Get("/", r.roleController.Get)
	role.Post("/", r.roleController.Create)
	role.Delete("/:id", r.roleController.Delete)
	role.Get("/:id/users", r.roleController.GetUsersOfRole)
	role.Put("/:roleid/user/:userid", r.roleController.AddUserToRole)
	role.Delete("/:roleid/user/:userid", r.roleController.RemoveUserFromRole)
}

func (r *router) ListRoutes() map[string][]string {
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
