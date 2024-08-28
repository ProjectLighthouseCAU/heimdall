package router

import (
	"strings"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/controller"
	"github.com/ProjectLighthouseCAU/heimdall/middleware"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	app                       *fiber.App
	userController            controller.UserController
	registrationKeyController controller.RegistrationKeyController
	roleController            controller.RoleController
	tokenController           controller.TokenController
	sessionMiddleware         fiber.Handler
}

func NewRouter(app *fiber.App,
	userContr controller.UserController,
	regKeyContr controller.RegistrationKeyController,
	roleContr controller.RoleController,
	tokenContr controller.TokenController,
	sessionMiddleware fiber.Handler) Router {
	return Router{app, userContr, regKeyContr, roleContr, tokenContr, sessionMiddleware}
}

func (r *Router) Init() {
	r.app.Use(logger.New())
	r.app.Use(recover.New())
	// app.Use(csrf.New()) // FIXME: csrf prevents everything except GET requests
	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{config.GetString("API_HOST", "https://lighthouse.uni-kiel.de"), config.GetString("CORS_ALLOW_ORIGINS", "http://localhost")}, ","), // TODO: remove localhost in production
		AllowCredentials: true,                                                                                                                                                    // TODO: remove in production
	}))
	r.app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: 1 * time.Minute,
	}))
	r.app.Use(pprof.New())

	swag := swagger.New(swagger.Config{
		Title:                    "Heimdall Lighthouse API",
		URL:                      "doc.json",
		DeepLinking:              false,
		DisplayOperationId:       false,
		DefaultModelsExpandDepth: 1,
		DefaultModelExpandDepth:  1,
		DefaultModelRendering:    "example",
		DisplayRequestDuration:   true,
		Filter:                   swagger.FilterConfig{Enabled: true},
		TryItOutEnabled:          true,
		RequestSnippetsEnabled:   true,
		SupportedSubmitMethods:   []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"},
		ValidatorUrl:             "",
		WithCredentials:          false,
	})
	r.app.Get("/swagger", swag)
	r.app.Get("/swagger/*", swag)

	r.app.Post("/register", r.userController.Register)
	r.app.Post("/login", r.userController.Login)

	r.app.Use(r.sessionMiddleware)
	r.app.Get("/metrics", monitor.New())

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
	GET /users
	GET /users/<own-id>
	PUT /users/<own-id>
	DELETE /users/<own-id>
	GET /users/<own-id>/roles

	Authorization not implemented yet (currently admin only):
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
	users.Get("/:id/api-token", middleware.AllowRoleOrOwnUserId(admin, "id"), r.tokenController.Get)       // username, token, roles, expiration
	users.Delete("/:id/api-token", middleware.AllowRoleOrOwnUserId(admin, "id"), r.tokenController.Delete) // invalidate and renew token
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
