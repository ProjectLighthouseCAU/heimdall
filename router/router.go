package router

import (
	"fmt"
	"time"

	"slices"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/middleware"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var (
	admin  = config.AdminRoleName
	deploy = config.DeployRoleName
)

type Router struct {
	app                    *fiber.App
	userHandler            handler.UserHandler
	registrationKeyHandler handler.RegistrationKeyHandler
	roleHandler            handler.RoleHandler
	tokenHandler           handler.TokenHandler
	sessionMiddleware      middleware.SessionMiddleware
	tokenMiddleware        middleware.TokenMiddleware
}

func NewRouter(app *fiber.App,
	userHandler handler.UserHandler,
	regKeyHandler handler.RegistrationKeyHandler,
	roleHandler handler.RoleHandler,
	tokenHandler handler.TokenHandler,
	sessionMiddleware middleware.SessionMiddleware,
	tokenMiddleware middleware.TokenMiddleware) Router {
	return Router{app, userHandler, regKeyHandler, roleHandler, tokenHandler, sessionMiddleware, tokenMiddleware}
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

func (r *Router) Init(sessionStore *session.Store, readynessProbe func(*fiber.Ctx) bool) {
	// log requests and responses
	r.app.Use(logger.New())

	// recover from panics in request handlers
	r.app.Use(recover.New())

	// setup helmet middleware for setting HTTP security headers
	r.app.Use(helmet.New())

	// FIXME: csrf prevents everything except GET requests
	// r.app.Use(csrf.New(csrf.Config{
	// 	KeyLookup:         "header:" + csrf.HeaderName,
	// 	CookieName:        "__Host-csrf_",
	// 	CookieSameSite:    "Lax",
	// 	CookieSecure:      true,
	// 	CookieSessionOnly: true,
	// 	CookieHTTPOnly:    true,
	// 	Expiration:        1 * time.Hour,
	// 	KeyGenerator:      utils.UUIDv4,
	// 	ContextKey:        nil,
	// 	Extractor:         csrf.CsrfFromHeader(csrf.HeaderName),
	// 	Session:           sessionStore,
	// 	SessionKey:        "fiber.csrf.token",
	// 	HandlerContextKey: "fiber.csrf.handler",
	// }))

	// setup CORS middleware
	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     config.CorsAllowOrigins,
		AllowCredentials: config.CorsAllowCredentials,
	}))

	r.app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe:     func(c *fiber.Ctx) bool { return true },
		LivenessEndpoint:  "/live",
		ReadinessProbe:    readynessProbe,
		ReadinessEndpoint: "/ready",
	}))

	// allow login and register once every 10 seconds per client
	unauthorizedLimiter := limiter.New(limiter.Config{
		Max:        6,
		Expiration: 1 * time.Minute,
		Next: func(c *fiber.Ctx) bool {
			return config.DisableRateLimiter
		},
	})
	r.app.Post("/register", unauthorizedLimiter, r.userHandler.Register)
	r.app.Post("/login", unauthorizedLimiter, r.userHandler.Login)

	r.initInternalRoutes(r.app.Group("/internal")) // not rate limited and without session middleware

	// allow 5 requests per second per client
	r.app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: 1 * time.Minute,
		Next: func(c *fiber.Ctx) bool {
			return config.DisableRateLimiter
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
	r.app.Use((fiber.Handler)(r.sessionMiddleware))

	// serve fiber monitor
	r.app.Get("/metrics", r.sessionMiddleware.AllowRole(admin), monitor.New())

	// setup pprof monitoring middleware
	r.app.Use("/debug/pprof", r.sessionMiddleware.AllowRole(admin), pprof.New())

	r.app.Post("/logout", r.userHandler.Logout)
	r.initUserRoutes(r.app.Group("/users"))
	r.initRegistrationKeyRoutes(r.app.Group("/registration-keys", r.sessionMiddleware.AllowRole(admin)))
	r.initRoleRoutes(r.app.Group("/roles", r.sessionMiddleware.AllowRole(admin)))

	// catch all requests that could not be handled and send JSON response (instead of fibers plain text)
	r.app.All("*", func(c *fiber.Ctx) error {
		return handler.UnwrapAndSendError(c, model.NotFoundError{Message: fmt.Sprintf("Cannot %s %s", c.Method(), c.Path())})
	})
}

func (r *Router) initInternalRoutes(internal fiber.Router) {
	internal.Use(middleware.AllowLoopbackAndPrivateIPsAnd(config.InternalIPs))
	internal.Use((fiber.Handler)(r.tokenMiddleware))
	internal.Get("/users", r.tokenMiddleware.AllowRole(deploy), r.tokenHandler.GetUsernames)
	internal.Get("/authenticate/:username<string>", r.tokenHandler.WatchAuthChanges)
}

func (r *Router) initUserRoutes(users fiber.Router) {
	users.Get("", r.userHandler.GetAll, r.sessionMiddleware.AllowRole(admin), r.userHandler.GetByName)
	users.Get("/:id<int>", r.sessionMiddleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.GetByID)
	users.Post("", r.sessionMiddleware.AllowRole(admin), r.userHandler.Create)
	users.Put("/:id<int>", r.sessionMiddleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.Update)
	users.Delete("/:id<int>", r.sessionMiddleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.Delete)
	users.Get("/:id<int>/roles", r.sessionMiddleware.AllowRoleOrOwnUserId(admin, "id"), r.userHandler.GetRolesOfUser)
	users.Get("/:id/api-token", r.sessionMiddleware.AllowRoleOrOwnUserId(admin, "id"), r.tokenHandler.Get)
	users.Put("/:id/api-token", r.sessionMiddleware.AllowRole(admin), r.tokenHandler.Update)                     // set permanent
	users.Delete("/:id/api-token", r.sessionMiddleware.AllowRoleOrOwnUserId(admin, "id"), r.tokenHandler.Delete) // invalidate and renew token
}

func (r *Router) initRegistrationKeyRoutes(keys fiber.Router) {
	keys.Get("", r.registrationKeyHandler.Get)
	keys.Get("/:id<int>", r.registrationKeyHandler.GetByID)
	keys.Post("", r.registrationKeyHandler.Create)
	keys.Put("/:id<int>", r.registrationKeyHandler.Update)
	keys.Delete("/:id<int>", r.registrationKeyHandler.Delete)
	keys.Get("/:id<int>/users", r.registrationKeyHandler.GetUsersOfKey)
}

func (r *Router) initRoleRoutes(roles fiber.Router) {
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
			if !slices.Contains(endpoints[endpoint.Path], endpoint.Method) {
				endpoints[endpoint.Path] = append(endpoints[endpoint.Path], endpoint.Method)
			}
		}
	}
	return endpoints
}
