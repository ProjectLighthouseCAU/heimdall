package service

// import (
// 	"errors"
// 	"fmt"
// 	"strings"

// 	"github.com/casbin/casbin/v2/log"
// 	"github.com/casbin/casbin/v2/rbac"

// 	"lighthouse.uni-kiel.de/lighthouse-api/config"
// 	"lighthouse.uni-kiel.de/lighthouse-api/repository"
// )

// // Casbin RoleManager implementation to integrate this services managed users and roles with the Casbin RBAC system
// // We only implement HasLink, GetRoles and GetUsers from rbac.RoleManager
// // Adding and removing links is not handled by Casbin and we don't use domains for roles

// var (
// 	userPrefix = config.GetString("CASBIN_USER_PREFIX", "user::")
// 	rolePrefix = config.GetString("CASBIN_ROLE_PREFIX", "role::")
// )

// func addUserPrefix(user string) string {
// 	return userPrefix + user
// }
// func removeUserPrefix(user string) string {
// 	return strings.TrimPrefix(user, userPrefix)
// }
// func addRolePrefix(role string) string {
// 	return rolePrefix + role
// }
// func removeRolePrefix(role string) string {
// 	return strings.TrimPrefix(role, rolePrefix)
// }

// type RoleManager struct {
// 	roleRepository repository.RoleRepository
// 	userRepository repository.UserRepository
// }

// var _ rbac.RoleManager = (*RoleManager)(nil) // RoleManager implements rbac.RoleManager

// func NewRoleManager(rr repository.RoleRepository, ur repository.UserRepository) rbac.RoleManager {
// 	rm := RoleManager{
// 		roleRepository: rr,
// 		userRepository: ur,
// 	}
// 	return &rm
// }

// func (rm *RoleManager) Clear() error {
// 	return nil
// }

// func (rm *RoleManager) AddLink(name1, name2 string, args ...string) error {
// 	return errors.New("AddLink not implemented")
// }

// func (rm *RoleManager) BuildRelationship(name1, name2 string, args ...string) error {
// 	return errors.New("BuildRelationship not implemented")
// }

// func (rm *RoleManager) DeleteLink(name1, name2 string, args ...string) error {
// 	return errors.New("DeleteLink not implemented")
// }

// // HasLink returns true if a user has a role
// // username: name of user with prefix, role: name of role with prefix, args: ignored
// // TODO: transitive relationships are currently not supported
// func (rm *RoleManager) HasLink(username, rolename string, args ...string) (b bool, err error) {
// 	if username == rolename { // user is always a member of himself (needed by casbin, safe because of the prefixes)
// 		fmt.Println("HasLink(" + username + ", " + rolename + ") -> true")
// 		return true, nil
// 	}
// 	roles, err := rm.GetRoles(username) // ensures name1 is a user
// 	if err != nil {
// 		fmt.Println("HasLink(" + username + ", " + rolename + ") -> false")
// 		return false, err
// 	}
// 	_, err = rm.GetUsers(rolename) // ensures name2 is a role
// 	if err != nil {
// 		fmt.Println("HasLink(" + username + ", " + rolename + ") -> false")
// 		return false, err
// 	}
// 	for _, r := range roles {
// 		if r == rolename {
// 			fmt.Println("HasLink(" + username + ", " + rolename + ") -> true")
// 			return true, nil
// 		}
// 	}
// 	fmt.Println("HasLink(" + username + ", " + rolename + ") -> false")
// 	return false, nil
// }

// // GetRoles returns the roles that a user is a member of
// // username: name of user with prefix, args: ignored
// func (rm *RoleManager) GetRoles(username string, args ...string) (roles []string, err error) {

// 	user, err := rm.userRepository.FindByName(removeUserPrefix(username))
// 	if err != nil {
// 		fmt.Println("GetRoles(" + username + ") -> " + fmt.Sprint(roles))
// 		return nil, err
// 	}
// 	// // TODO: search roles recursively
// 	// var allRoles []model.Role
// 	// for _, r := range roles {
// 	// 	allRoles = append(allRoles, rm.getRolesRecursively(&r)...)
// 	// }
// 	for _, r := range user.Roles {
// 		roles = append(roles, addRolePrefix(r.Name))
// 	}
// 	fmt.Println("GetRoles(" + username + ") -> " + fmt.Sprint(roles))
// 	return roles, nil
// }

// // func (rm *RoleManager) getRolesRecursively(role *model.Role) (roles []model.Role) {
// // 	roles = append(roles, *role)
// // 	for _, r := range role.Roles {
// // 		roles = append(roles, rm.getRolesRecursively(&r)...)
// // 	}
// // 	return roles
// // }

// // func dedupSlice[T comparable](s []T) []T {
// // 	seen := make(map[T]bool)
// // 	var result []T
// // 	for _, v := range s {
// // 		if !seen[v] {
// // 			result = append(result, v)
// // 			seen[v] = true
// // 		}
// // 	}
// // 	return result
// // }

// // GetUsers returns the users that have a role
// // rolename: name of role with prefix, args: ignored
// func (rm *RoleManager) GetUsers(rolename string, args ...string) (users []string, err error) {
// 	fmt.Println("GetUsers(" + rolename + ") -> " + fmt.Sprint(users))

// 	role, err := rm.roleRepository.FindByName(removeRolePrefix(rolename))
// 	if err != nil {
// 		fmt.Println("GetUsers(" + rolename + ") -> " + fmt.Sprint(users))
// 		return nil, err
// 	}
// 	for _, u := range role.Users {
// 		users = append(users, addUserPrefix(u.Username))
// 	}
// 	fmt.Println("GetUsers(" + rolename + ") -> " + fmt.Sprint(users))
// 	return users, nil
// }

// func (rm *RoleManager) GetDomains(name string) ([]string, error) {
// 	return nil, errors.New("GetDomains not implemented")
// }

// func (rm *RoleManager) GetAllDomains() ([]string, error) {
// 	return nil, errors.New("GetAllDomains not implemented")
// }

// func (rm *RoleManager) PrintRoles() error {
// 	return errors.New("PrintRoles not implemented")
// }

// func (rm *RoleManager) SetLogger(logger log.Logger) {
// 	// do nothing
// }

// func (*RoleManager) AddDomainMatchingFunc(name string, fn rbac.MatchingFunc) {
// 	panic("unimplemented")
// }

// func (*RoleManager) AddMatchingFunc(name string, fn rbac.MatchingFunc) {
// 	panic("unimplemented")
// }

// func (*RoleManager) Match(str string, pattern string) bool {
// 	panic("unimplemented")
// }
