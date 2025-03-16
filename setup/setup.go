package setup

import (
	"context"
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
	"github.com/gofiber/storage/redis"
	"gorm.io/gorm"
)

func Setup() *fiber.App {
	docs.SwaggerInfo.Host = config.ApiHost
	docs.SwaggerInfo.BasePath = config.ApiBasePath

	log.Println("Starting Heimdall")
	app := fiber.New(fiber.Config{
		AppName:                 "Heimdall",
		CaseSensitive:           true,
		StrictRouting:           true,
		ProxyHeader:             config.ProxyHeader,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{config.TrustedProxy},
	})

	// Dependency Injection

	// database
	log.Println("	Connecting to database")
	db, err := connectPostgres()
	panicOnError(err)

	// session store
	sessionStorage := redis.New(redis.Config{
		Host:      config.RedisHost,
		Port:      config.RedisPort,
		Username:  config.RedisUser,
		Password:  config.RedisPassword,
		Database:  0,
		Reset:     false,
		TLSConfig: nil,
		PoolSize:  10 * runtime.GOMAXPROCS(0),
	})

	sessionStore := session.New(session.Config{
		Storage:        sessionStorage,
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:session_id",
		KeyGenerator:   utils.UUIDv4,
		CookieSecure:   true,
		CookieSameSite: "Lax",
		CookieHTTPOnly: true,
	})

	setupApplication(app, db, sessionStore)

	return app
}

func setupApplication(app *fiber.App, db *gorm.DB, store *session.Store) {
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
	tokenService := service.NewTokenService(tokenRepository, userRepository)
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
	tokenMiddleware := middleware.NewTokenMiddleware(&userService, &tokenRepository)

	// router
	routa := router.NewRouter(
		app,
		userHandler,
		registrationKeyHandler,
		roleHandler,
		tokenHandler,
		sessionMiddleware,
		tokenMiddleware,
	)

	// readyness probe
	readynessProbe := func(c *fiber.Ctx) bool {
		// Ping redis
		ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelCtx()
		err := store.Storage.(*redis.Storage).Conn().Ping(ctx).Err()
		if err != nil {
			log.Println("Readyness check failed for Redis:", err)
			return false
		}
		// Ping postgres
		db, err := db.DB()
		if err != nil {
			log.Println("Readyness check failed for PostgreSQL:", err)
			return false
		}
		err = db.Ping()
		if err != nil {
			log.Println("Readyness check failed for PostgreSQL:", err)
			return false
		}
		return true
	}

	routa.Init(store, readynessProbe)
	printRoutes(routa.ListRoutes())

	if config.UseTestDatabase { // TODO: remove in prod - this function deletes the whole database
		SetupTestDatabase(db, store, userService, roleService, registrationKeyService, tokenService)
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
