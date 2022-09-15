package service

import (
	"fmt"

	fibercasbin "github.com/arsmn/fiber-casbin/v2"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/rbac"
	"github.com/casbin/casbin/v2/util"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type AccessControlService interface {
	Enforce(matcher, dom, sub, obj, act string) (bool, error)
	// GetRolesForUser(user *model.User) ([]model.Role, error)
	// GetUsersForRole(role *model.Role) ([]model.User, error)
	// AddRoleForUser(user *model.User, role *model.Role) (bool, error)
	// DeleteRoleForUser(user *model.User, role *model.Role) (bool, error)
	// DeleteRole(role *model.Role) (bool, error)
	// DeleteUser(user *model.User) (bool, error)
	GetEnforcer() *casbin.Enforcer
}

type accessControlService struct {
	enforcer       *casbin.Enforcer
	userRepository repository.UserRepository
	roleRepository repository.RoleRepository
}

var (
	_          AccessControlService = (*accessControlService)(nil) // compile-time interface check
	modelFile                       = config.GetString("CASBIN_MODEL_FILE", "./casbin/model.conf")
	policyFile                      = config.GetString("CASBIN_POLICY_FILE", "./casbin/policy.csv")
)

func NewAccessControlService(db *gorm.DB, userRepository repository.UserRepository, roleRepository repository.RoleRepository, roleManager rbac.RoleManager) *accessControlService {
	// a, _ := gormadapter.NewAdapterByDBWithCustomTable(db, &model.CasbinPolicy{})
	e, err := casbin.NewEnforcer(modelFile, policyFile)
	e.LoadPolicy()
	fmt.Println("POLICIES: ", e.GetPolicy())
	if err != nil {
		panic(err)
	}
	// e.EnableAutoBuildRoleLinks(true)
	e.EnableAutoSave(true)
	e.EnableEnforce(config.GetBool("CASBIN_ENABLE_ENFORCE", true))
	e.SetRoleManager(roleManager)

	fmt.Println("DEBUG ROLEMANAGER:")
	fmt.Println(e.BuildRoleLinks())
	b1, s1, err := e.EnforceEx("internal", "user::Testuser", "/test", "GET")
	if err != nil {
		panic(err)
	}
	b2, s2, err := e.EnforceEx("internal", "role::Testrole", "/test2", "GET")
	if err != nil {
		panic(err)
	}
	b3, s3, err := e.EnforceEx("internal", "user::Testuser", "/test2", "GET")
	if err != nil {
		panic(err)
	}
	b4, s4, err := e.EnforceEx("internal", "user::Testuser", "/test2", "POST")
	if err != nil {
		panic(err)
	}
	fmt.Println("internal", "Testuser", "/test", "GET", b1, s1)
	fmt.Println("internal", "Testrole", "/test2", "GET", b2, s2)
	fmt.Println("internal", "Testuser", "/test2", "GET", b3, s3)
	fmt.Println("internal", "Testuser", "/test2", "POST", b4, s4)
	roles, _ := e.GetRolesForUser("Testuser")
	users, _ := e.GetUsersForRole("Testrole")
	fmt.Println(roles)
	fmt.Println(users)

	f := func(sub, obj, pobj, pathVar string) bool {
		util.KeyGet2(obj, pobj, pathVar)
		return false
	}

	e.AddFunction("mymatch", func(args ...interface{}) (interface{}, error) {
		a1 := args[0].(string)
		a2 := args[1].(string)
		a3 := args[2].(string)
		a4 := args[3].(string)
		return (bool)(f(a1, a2, a3, a4)), nil
	})

	// e.SetAdapter(a)
	// e.SavePolicy()
	fmt.Println("POLICIES: ", e.GetPolicy())
	return &accessControlService{
		enforcer:       e,
		userRepository: userRepository,
		roleRepository: roleRepository,
	}
}

func (acs *accessControlService) NewCasbinMiddleware(lookup func(c *fiber.Ctx) string) *fibercasbin.CasbinMiddleware {
	return fibercasbin.New(fibercasbin.Config{
		ModelFilePath: modelFile,
		PolicyAdapter: acs.enforcer.GetAdapter(),
		Lookup:        lookup,
	})
}

func (acs *accessControlService) GetEnforcer() *casbin.Enforcer {
	return acs.enforcer
}

func (acs *accessControlService) Enforce(matcher, dom, sub, obj, act string) (bool, error) {
	return acs.enforcer.Enforce(matcher, dom, sub, obj, act)
}

// func (acs *accessControlService) GetRolesForUser(user *model.User) ([]model.Role, error) {
// 	roles, err := acs.enforcer.GetRolesForUser(addUserPrefix(user.Username))
// 	if err != nil {
// 		return nil, err
// 	}
// 	for i, role := range roles {
// 		roles[i] = removeRolePrefix(role)
// 	}
// 	return acs.roleRepository.FindByNames(roles)
// }

// func (acs *accessControlService) GetUsersForRole(role *model.Role) ([]model.User, error) {
// 	users, err := acs.enforcer.GetUsersForRole(addRolePrefix(role.Name))
// 	if err != nil {
// 		return nil, err
// 	}
// 	for i, user := range users {
// 		users[i] = removeUserPrefix(user)
// 	}
// 	return acs.userRepository.FindByNames(users)
// }

// --> handled by RoleManager now
// func (acs *accessControlService) AddRoleForUser(user *model.User, role *model.Role) (bool, error) {
// 	return acs.enforcer.AddRoleForUser(addUserPrefix(user.Username), addRolePrefix(role.Name))
// }

// func (acs *accessControlService) DeleteRoleForUser(user *model.User, role *model.Role) (bool, error) {
// 	return acs.enforcer.DeleteRoleForUser(addUserPrefix(user.Username), addRolePrefix(role.Name))
// }

// func (acs *accessControlService) DeleteRole(role *model.Role) (bool, error) {
// 	return acs.enforcer.DeleteRole(addRolePrefix(role.Name))
// }

// func (acs *accessControlService) DeleteUser(user *model.User) (bool, error) {
// 	return acs.enforcer.DeleteUser(addUserPrefix(user.Username))
// }

// TODO: policy management
