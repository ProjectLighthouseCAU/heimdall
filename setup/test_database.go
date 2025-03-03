package setup

import (
	"context"
	"log"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func setupTestDatabase(db *gorm.DB, rdb *redis.Client, store *session.Store, userService service.UserService, roleService service.RoleService, registrationKeyService service.RegistrationKeyService) {
	log.Println("	Setting up test database")
	log.Println("		Deleting redis")
	must(store.Storage.Reset())
	must(rdb.FlushDB(context.TODO()).Err())
	log.Println("		Deleting tables")

	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.User{}).Error)
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.Role{}).Error)
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.RegistrationKey{}).Error)
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.Token{}).Error)

	log.Println("		Resetting auto increment sequences")
	must(db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1").Error)
	must(db.Exec("ALTER SEQUENCE roles_id_seq RESTART WITH 1").Error)
	must(db.Exec("ALTER SEQUENCE registration_keys_id_seq RESTART WITH 1").Error)
	must(db.Exec("ALTER SEQUENCE tokens_id_seq RESTART WITH 1").Error)

	log.Println("		Creating test data")
	must(registrationKeyService.Create("test_registration_key", "just for testing", true, time.Now().AddDate(0, 0, 3)))
	must(userService.Create("Admin", "password1234", "admin@example.com", false))
	must(userService.Create("Live", "password1234", "live@example.com", true))
	_, err := userService.Register("User", "password1234", "user@example.com", "test_registration_key", nil)
	must(err)
	must(roleService.Create("admin"))
	must(roleService.Create("deploy"))
	admin, err := userService.GetByName("Admin")
	must(err)
	live, err := userService.GetByName("Live")
	must(err)
	adminRole, err := roleService.GetByName("admin")
	must(err)
	deployRole, err := roleService.GetByName("deploy")
	must(err)
	must(roleService.AddUserToRole(adminRole.ID, admin.ID))
	must(roleService.AddUserToRole(deployRole.ID, live.ID))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
