package setup

import (
	"log"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/service"
	"github.com/gofiber/fiber/v2/middleware/session"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SetupTestDatabase(db *gorm.DB, store *session.Store,
	userService service.UserService,
	roleService service.RoleService,
	registrationKeyService service.RegistrationKeyService,
	tokenService service.TokenService) {

	log.Println("	Setting up test database")
	log.Println("		Deleting redis")
	must(store.Storage.Reset())
	log.Println("		Deleting tables")

	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.User{}).Error)
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.Role{}).Error)
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.RegistrationKey{}).Error)
	must(db.Unscoped().Select(clause.Associations).Where("true").Delete(&model.Token{}).Error)

	log.Println("		Resetting auto increment sequences")
	must(db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1").Error)
	must(db.Exec("ALTER SEQUENCE roles_id_seq RESTART WITH 1").Error)
	must(db.Exec("ALTER SEQUENCE registration_keys_id_seq RESTART WITH 1").Error)

	log.Println("		Creating test data")
	must(registrationKeyService.Create("test_registration_key", "just for testing", true, time.Now().AddDate(0, 0, 3)))
	must(userService.Create("Admin", "password1234", "admin@example.com"))
	must(userService.Create("Live", "password1234", "live@example.com"))
	_, err := userService.Register("User", "password1234", "user@example.com", "test_registration_key", nil)
	must(err)
	admin, err := userService.GetByName("Admin")
	must(err)
	live, err := userService.GetByName("Live")
	must(err)
	user, err := userService.GetByName("User")
	must(err)
	var created bool
	created, err = tokenService.GenerateApiTokenIfNotExists(admin)
	must(err)
	if !created {
		panic("token for Admin was not created but did not error, was it correctly deleted?")
	}
	created, err = tokenService.GenerateApiTokenIfNotExists(live)
	must(err)
	if !created {
		panic("token for Live was not created but did not error, was it correctly deleted?")
	}
	_, err = tokenService.GenerateApiTokenIfNotExists(user)
	must(err)
	// User is created with userService.Register, which creates the token

	must(roleService.Create("admin"))
	must(roleService.Create("deploy"))

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
