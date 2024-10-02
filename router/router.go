package router

import (
	"fmt"
	"strings"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/middleware"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	app                    *fiber.App
	userHandler            handler.UserHandler
	registrationKeyHandler handler.RegistrationKeyHandler
	roleHandler            handler.RoleHandler
	tokenHandler           handler.TokenHandler
	sessionMiddleware      fiber.Handler
}

func NewRouter(app *fiber.App,
	userHandler handler.UserHandler,
	regKeyHandler handler.RegistrationKeyHandler,
	roleHandler handler.RoleHandler,
	tokenHandler handler.TokenHandler,
	sessionMiddleware fiber.Handler) Router {
	return Router{app, userHandler, regKeyHandler, roleHandler, tokenHandler, sessionMiddleware}
}

func (r *Router) Init() {
	// log requests and responses
	r.app.Use(logger.New())

	// recover from panics in request handlers
	r.app.Use(recover.New())

	// setup helmet middleware for setting HTTP security headers
	r.app.Use(helmet.New())

	// FIXME: csrf prevents everything except GET requests
	// app.Use(csrf.New())

	// setup CORS middleware
	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{config.GetString("API_HOST", "https://lighthouse.uni-kiel.de"), config.GetString("CORS_ALLOW_ORIGINS", "http://localhost")}, ","), // TODO: remove localhost in production
		AllowCredentials: true,                                                                                                                                                    // TODO: remove in production
	}))

	// TODO: add healthcheck middleware for liveness and readyness endpoints
	// r.app.Use(healthcheck.New())

	// allow login and register once every 10 seconds per client
	unauthorizedLimiter := limiter.New(limiter.Config{
		Max:        6,
		Expiration: 1 * time.Minute,
		Next: func(c *fiber.Ctx) bool {
			return config.GetBool("DISABLE_RATE_LIMITER", false)
		},
	})
	r.app.Post("/register", unauthorizedLimiter, r.userHandler.Register)
	r.app.Post("/login", unauthorizedLimiter, r.userHandler.Login)

	// allow 5 requests per second per client
	r.app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: 1 * time.Minute,
		Next: func(c *fiber.Ctx) bool {
			return config.GetBool("DISABLE_RATE_LIMITER", false)
		},
	}))

	// setup and serve swagger API documentation
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

	// all requests to routes after this point have to be authenticated
	r.app.Use(r.sessionMiddleware)

	// serve fiber monitor
	r.app.Get("/metrics", monitor.New())

	// setup pprof monitoring middleware
	r.app.Use(pprof.New(pprof.Config{Prefix: config.GetString("API_BASE_PATH", "/api")}))

	r.app.Post("/logout", r.userHandler.Logout)
	r.initUserRoutes()
	r.initRegistrationKeyRoutes()
	r.initRoleRoutes()

	// catch all requests that could not be handled and send JSON response (instead of fibers plain text)
	r.app.All("*", func(c *fiber.Ctx) error {
		return handler.UnwrapAndSendError(c, model.NotFoundError{Message: fmt.Sprintf("Cannot %s %s", c.Method(), c.Path())})
	})
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
	users.Get("", r.userHandler.GetAll, middleware.AllowRole(admin), r.userHandler.GetByName)
	users.Get("/:id<int>", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.GetByID)
	users.Post("", middleware.AllowRole(admin), r.userHandler.Create)
	users.Put("/:id<int>", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.Update)
	users.Delete("/:id<int>", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.Delete)
	users.Get("/:id<int>/roles", middleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.GetRolesOfUser)
	users.Get("/:id/api-token", middleware.AllowRoleOrOwnUserId(admin, "id"), r.tokenHandler.Get)       // username, token, roles, expiration
	users.Delete("/:id/api-token", middleware.AllowRoleOrOwnUserId(admin, "id"), r.tokenHandler.Delete) // invalidate and renew token
}

func (r *Router) initRegistrationKeyRoutes() {
	keys := r.app.Group("/registration-keys", middleware.AllowRole(admin))
	keys.Get("", r.registrationKeyHandler.Get)
	keys.Get("/:id<int>", r.registrationKeyHandler.GetByID)
	keys.Post("", r.registrationKeyHandler.Create)
	keys.Put("/:id<int>", r.registrationKeyHandler.Update)
	keys.Delete("/:id<int>", r.registrationKeyHandler.Delete)
	keys.Get("/:id<int>/users", r.registrationKeyHandler.GetUsersOfKey)
}

func (r *Router) initRoleRoutes() {
	roles := r.app.Group("/roles", middleware.AllowRole(admin))
	roles.Get("", r.roleHandler.Get)
	roles.Get("/:id<int>", r.roleHandler.GetByID)
	roles.Post("", r.roleHandler.Create)
	roles.Put("/:id<int>", r.roleHandler.Update)
	roles.Delete("/:id<int>", r.roleHandler.Delete)
	roles.Get("/:id<int>/users", r.roleHandler.GetUsersOfRole)
	roles.Put("/:roleid<int>/users/:userid<int>", r.roleHandler.AddUserToRole)
	roles.Delete("/:roleid<int>/users/:userid<int>", r.roleHandler.RemoveUserFromRole)
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
