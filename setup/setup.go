package setup

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/docs"
	"github.com/ProjectLighthouseCAU/heimdall/handler"
	"github.com/ProjectLighthouseCAU/heimdall/middleware"
	"github.com/ProjectLighthouseCAU/heimdall/repository"
	"github.com/ProjectLighthouseCAU/heimdall/router"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	fiberRedis "github.com/gofiber/storage/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Setup() *fiber.App {
	docs.SwaggerInfo.Host = config.GetString("API_HOST", "https://lighthouse.uni-kiel.de")
	docs.SwaggerInfo.BasePath = config.GetString("API_BASE_PATH", "/api")

	log.Println("Starting Heimdall")
	app := fiber.New(fiber.Config{
		AppName:       "Heimdall",
		CaseSensitive: true,
		StrictRouting: true,
		ProxyHeader:   "X-Real-Ip",
	})

	// Dependency Injection

	// databases
	log.Println("	Connecting to database")
	db, err := connectPostgres()
	panicOnError(err)
	redisdb, err := connectRedis(0) // use db 0 for api tokens and roles
	panicOnError(err)

	// session store
	storage := fiberRedis.New(fiberRedis.Config{
		Host:      config.GetString("REDIS_HOST", "127.0.0.1"),
		Port:      config.GetInt("REDIS_PORT", 6379),
		Username:  config.GetString("REDIS_USER", ""),
		Password:  config.GetString("REDIS_PASSWORD", ""),
		Database:  1, // use db 1 for sessions
		Reset:     false,
		TLSConfig: nil,
		PoolSize:  10 * runtime.GOMAXPROCS(0),
	})

	store := session.New(session.Config{
		Storage:        storage,
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:session_id",
		KeyGenerator:   utils.UUIDv4,
		CookieSecure:   false,  // TODO: change to true in production
		CookieSameSite: "None", // TODO: change to Lax or Strict in production
		CookieHTTPOnly: false,  // TODO: change to true in production
	})

	setupApplication(app, db, redisdb, store)

	return app
}

func setupApplication(app *fiber.App, db *gorm.DB, redisdb *redis.Client, store *session.Store) {
	// repositories
	userRepository := repository.NewUserRepository(db)
	registrationKeyRepository := repository.NewRegistrationKeyRepository(db)
	roleRepository := repository.NewRoleRepository(db)
	tokenRepository := repository.NewTokenRepository(db)

	// migrate database
	panicOnError(userRepository.Migrate())
	panicOnError(registrationKeyRepository.Migrate())
	panicOnError(roleRepository.Migrate())
	panicOnError(tokenRepository.Migrate())

	// services
	tokenService := service.NewTokenService(redisdb, tokenRepository)
	userService := service.NewUserService(
		userRepository,
		registrationKeyRepository,
		roleRepository,
		tokenService,
	)
	registrationKeyService := service.NewRegistrationKeyService(
		registrationKeyRepository,
	)
	roleService := service.NewRoleService(
		roleRepository,
		userRepository,
		tokenService,
	)

	// handlers
	userHandler := handler.NewUserHandler(
		userService,
		roleService,
		store,
	)
	registrationKeyHandler := handler.NewRegistrationKeyHandler(
		registrationKeyService,
	)
	roleHandler := handler.NewRoleHandler(
		roleService,
	)
	tokenHandler := handler.NewTokenHandler(
		tokenService,
		userService,
	)

	// middleware
	sessionMiddleware := middleware.NewSessionMiddleware(store, userService, tokenService)

	// router
	routa := router.NewRouter(
		app,
		userHandler,
		registrationKeyHandler,
		roleHandler,
		tokenHandler,
		sessionMiddleware,
	)

	routa.Init()
	printRoutes(routa.ListRoutes())

	if config.GetBool("USE_TEST_DATABASE", false) { // TODO: remove in prod - this function deletes the whole database
		setupTestDatabase(db, redisdb, store, userService, roleService, registrationKeyService)
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func printRoutes(routes map[string][]string) {
	log.Println("Routes:")
	max := 0
	for k := range routes {
		if len(k) > max {
			max = len(k)
		}
	}
	for k, v := range routes {
		pad := strings.Repeat(" ", max-len(k)+1)
		fmt.Printf("%s%s%s\n", k, pad, v)
	}
}
