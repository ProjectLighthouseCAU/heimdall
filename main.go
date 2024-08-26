package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	fiberRedis "github.com/gofiber/storage/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/controller"
	"lighthouse.uni-kiel.de/lighthouse-api/database"
	"lighthouse.uni-kiel.de/lighthouse-api/docs"
	"lighthouse.uni-kiel.de/lighthouse-api/middleware"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
	"lighthouse.uni-kiel.de/lighthouse-api/router"
	"lighthouse.uni-kiel.de/lighthouse-api/service"
)

// @Title		Heimdall Lighthouse API
// @Version		0.1
// @Description	This is the REST API of Project Lighthouse that manages users, roles, registration keys, API tokens and everything about authentication and authorization.
// @Description NOTE: This API is an early alpha version that still needs a lot of testing (unit tests, end-to-end tests and security tests)
// @Host		https://lighthouse.uni-kiel.de
// @BasePath	/api
func main() {
	docs.SwaggerInfo.Host = config.GetString("API_HOST", "https://lighthouse.uni-kiel.de")
	docs.SwaggerInfo.BasePath = config.GetString("API_BASE_PATH", "/api")

	log.Println("Starting Heimdall")
	app := fiber.New(fiber.Config{
		AppName:       "Heimdall",
		CaseSensitive: true,
		StrictRouting: true,
	})

	// Dependency Injection

	// databases
	log.Println("	Connecting to database")
	db, err := database.ConnectPostgres()
	if err != nil {
		log.Println(err)
	}
	redisdb, err := database.ConnectRedis(0) // use db 0 for api tokens and roles
	if err != nil {
		log.Println(err)
	}

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
		CookieHTTPOnly: true,
	})

	// repositories
	userRepository := repository.NewUserRepository(db)
	registrationKeyRepository := repository.NewRegistrationKeyRepository(db)
	roleRepository := repository.NewRoleRepository(db)

	// services
	tokenService := service.NewTokenService(redisdb)
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

	// controllers
	userController := controller.NewUserController(
		userService,
		roleService,
		store,
	)
	registrationKeyController := controller.NewRegistrationKeyController(
		registrationKeyService,
	)
	roleController := controller.NewRoleController(
		roleService,
	)
	tokenController := controller.NewTokenController(
		tokenService,
		userService,
	)

	// middleware
	sessionMiddleware := middleware.NewSessionMiddleware(store, userService, tokenService)

	// router
	routa := router.NewRouter(
		app,
		userController,
		registrationKeyController,
		roleController,
		tokenController,
		sessionMiddleware,
	)

	// migrate database
	userRepository.Migrate()
	registrationKeyRepository.Migrate()
	roleRepository.Migrate()

	// TODO: remove in prod
	setupTestDatabase(db, redisdb, store, userService, roleService, registrationKeyService)

	routa.Init()
	printRoutes(routa.ListRoutes())

	log.Println("Setup done. Listening until Ragnarök...")
	log.Fatal(app.Listen(":8080"))
}

func setupTestDatabase(db *gorm.DB, rdb *redis.Client, store *session.Store, userService service.UserService, roleService service.RoleService, registrationKeyService service.RegistrationKeyService) {
	log.Println("	Setting up test database")
	log.Println("		Deleting redis")
	must(store.Storage.Reset())
	must(rdb.FlushDB(context.TODO()).Err())
	log.Println("		Deleting tables")

	var users []model.User
	db.Find(&users)
	for _, user := range users {
		must(db.Unscoped().Select(clause.Associations).Delete(user).Error)
	}
	var roles []model.Role
	db.Find(&roles)
	for _, role := range roles {
		must(db.Unscoped().Select(clause.Associations).Delete(role).Error)
	}
	// db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.Token{})
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.RegistrationKey{}).Error)

	log.Println("		Resetting auto increment sequences")
	must(db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1").Error)
	must(db.Exec("ALTER SEQUENCE roles_id_seq RESTART WITH 1").Error)
	// db.Exec("ALTER SEQUENCE tokens_id_seq RESTART WITH 1")
	must(db.Exec("ALTER SEQUENCE registration_keys_id_seq RESTART WITH 1").Error)

	log.Println("		Creating test data")
	must(registrationKeyService.Create("test_registration_key", "just for testing", true, time.Now().AddDate(0, 0, 3)))
	must(userService.Create("User", "password1234", "user@example.com", false))
	must(userService.Create("Admin", "password1234", "admin@example.com", false))
	must(roleService.Create("admin"))
	admin, err := userService.GetByName("Admin")
	must(err)
	adminRole, err := roleService.GetByName("admin")
	must(err)
	must(roleService.AddUserToRole(adminRole.ID, admin.ID))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func printRoutes(routes map[string][]string) {
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
