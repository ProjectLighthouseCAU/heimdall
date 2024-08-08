package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/controller"
	"lighthouse.uni-kiel.de/lighthouse-api/database"
	"lighthouse.uni-kiel.de/lighthouse-api/middleware"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
	"lighthouse.uni-kiel.de/lighthouse-api/router"
	"lighthouse.uni-kiel.de/lighthouse-api/service"

	swagger "github.com/arsmn/fiber-swagger/v2"
	_ "lighthouse.uni-kiel.de/lighthouse-api/docs"
)

// @title			Lighthouse API
// @version		1.0
// @description	This is the REST API of Project Lighthouse
// @host			localhost:8080
// @BasePath		/
func main() {

	log.Println("Starting fiber")
	app := fiber.New(fiber.Config{
		Prefork:       false, // this makes everything complicated and we don't need it behind a reverse proxy
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Lighthouse API",
	})
	log.Println("	Setting up middleware")
	app.Use(logger.New())
	app.Use(recover.New())
	// app.Use(csrf.New()) // FIXME: csrf prevents everything except GET requests
	app.Use(cors.New())
	app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: 1 * time.Minute,
	}))
	// TODO: secure monitoring routes
	app.Use(pprof.New())
	app.Get("/metrics", monitor.New())

	// dependency injection
	log.Println("	Connecting to database")
	db := database.Connect()
	log.Println("	Connected to database")

	storage := redis.New(redis.Config{
		Host:      config.GetString("REDIS_HOST", "127.0.0.1"),
		Port:      config.GetInt("REDIS_PORT", 6379),
		Username:  config.GetString("REDIS_USER", ""),
		Password:  config.GetString("REDIS_PASSWORD", ""),
		Database:  0,
		Reset:     false,
		TLSConfig: nil,
		PoolSize:  10 * runtime.GOMAXPROCS(0),
	})

	store := session.New(session.Config{
		Storage:      storage,
		Expiration:   24 * time.Hour,
		KeyLookup:    "cookie:session_id",
		KeyGenerator: utils.UUIDv4,
	})

	userRepository := repository.NewUserRepository(db)
	registrationKeyRepository := repository.NewRegistrationKeyRepository(db)
	roleRepository := repository.NewRoleRepository(db)

	// roleManager := service.NewRoleManager(
	// 	roleRepository,
	// 	userRepository,
	// )
	// accessControlService := service.NewAccessControlService(
	// 	db,
	// 	userRepository,
	// 	roleRepository,
	// 	roleManager,
	// )
	userService := service.NewUserService(
		userRepository,
		registrationKeyRepository,
		roleRepository,
		store,
	)
	registrationKeyService := service.NewRegistrationKeyService(
		registrationKeyRepository,
	)
	roleService := service.NewRoleService(
		roleRepository,
		userRepository,
	)

	userController := controller.NewUserController(
		userService,
	)
	registrationKeyController := controller.NewRegistrationKeyController(
		registrationKeyService,
	)
	roleController := controller.NewRoleController(
		roleService,
	)

	// casbinMiddleware := middleware.NewCasbinMiddleware(
	// 	accessControlService,
	// )
	sessionMiddleware := middleware.NewSessionMiddleware(store, userService)

	routa := router.NewRouter(
		app,
		userController,
		registrationKeyController,
		roleController,
		// casbinMiddleware,
		sessionMiddleware,
	)

	userRepository.Migrate()
	registrationKeyRepository.Migrate()
	roleRepository.Migrate()

	SetupTestDatabase(db)

	log.Println("	Setting up routes")
	routa.Init()
	printRoutes(routa.ListRoutes())

	// TODO: fix and finish swagger documentation
	app.Get("/swagger", swagger.HandlerDefault)
	app.Get("/swagger/*", swagger.HandlerDefault)
	log.Println("Setup done. Listening...")
	log.Fatal(app.Listen(":8080"))
}

func SetupTestDatabase(db *gorm.DB) {
	log.Println("	Setting up test database")
	log.Println("		Deleting tables")

	var users []model.User
	db.Find(&users)
	for _, user := range users {
		db.Unscoped().Select(clause.Associations).Delete(user)
	}
	var roles []model.Role
	db.Find(&roles)
	for _, role := range roles {
		db.Unscoped().Select(clause.Associations).Delete(role)
	}
	// db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.Token{})
	db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.RegistrationKey{})

	log.Println("		Resetting auto increment sequences")
	db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE roles_id_seq RESTART WITH 1")
	// db.Exec("ALTER SEQUENCE tokens_id_seq RESTART WITH 1")
	db.Exec("ALTER SEQUENCE registration_keys_id_seq RESTART WITH 1")

	log.Println("		Creating test data")
	db.Create(&model.RegistrationKey{Key: "test_registration_key", ExpiresAt: time.Now().AddDate(0, 0, 3)})
	db.Create(&model.User{Username: "Testuser", Password: "hashedPWplaceholder"})
	db.Create(&model.Role{Name: "Testrole"})
	db.Create(&model.Role{Name: "admin"})
	db.Exec("INSERT INTO user_roles (user_id, role_id) values (1, 1)")
	db.Exec("INSERT INTO user_roles (user_id, role_id) values (1, 2)")
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
