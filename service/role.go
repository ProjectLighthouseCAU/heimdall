package service

import (
	"log"

	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RoleService struct {
	roleRepository repository.RoleRepository
	userRepository repository.UserRepository
	tokenService   TokenService
}

func NewRoleService(roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
	tokenService TokenService) RoleService {
	return RoleService{roleRepo, userRepo, tokenService}
}

func (r *RoleService) GetAll() ([]model.Role, error) {
	return r.roleRepository.FindAll()
}

func (r *RoleService) GetByID(id uint) (*model.Role, error) {
	return r.roleRepository.FindByID(id)
}

func (r *RoleService) GetByName(name string) (*model.Role, error) {
	return r.roleRepository.FindByName(name)
}

func validateRole(rolename string) error {
	if !isValidName(rolename) {
		return model.BadRequestError{Message: "Invalid name"}
	}
	return nil
}

func (r *RoleService) checkIfRoleExists(rolename string) error {
	exists, err := r.roleRepository.ExistsByName(rolename)
	if err != nil {
		return model.InternalServerError{Err: err}
	}
	if exists {
		return model.ConflictError{Message: "Role already exists"}
	}
	return nil
}

func (r *RoleService) Create(rolename string) error {
	if err := validateRole(rolename); err != nil {
		return err
	}
	if err := r.checkIfRoleExists(rolename); err != nil {
		return err
	}
	role := model.Role{
		Name: rolename,
	}
	return r.roleRepository.Save(&role)
}

func (r *RoleService) Update(id uint, rolename string) error {
	role, err := r.roleRepository.FindByID(id)
	if err != nil {
		return err
	}
	if rolename == role.Name {
		return nil
	}
	if !isValidName(rolename) {
		return model.BadRequestError{Message: "Invalid name"}
	}

	// save role
	role.Name = rolename
	err = r.roleRepository.Save(role)
	if err != nil {
		return err
	}

	// get users of role for token update
	users, err := r.GetUsersOfRole(id)
	if err != nil {
		return model.InternalServerError{Message: "Could not get users of role for invalidating tokens", Err: err}
	}

	// update roles in redis
	for _, user := range users {
		r.tokenService.UpdateRolesIfExists(&user)
	}
	return nil
}

func (r *RoleService) DeleteByID(id uint) error {
	// get users of role before deletion
	users, err := r.GetUsersOfRole(id)
	if err != nil {
		return model.InternalServerError{Message: "Could not get users of role for invalidating tokens", Err: err}
	}

	// delete role
	err = r.roleRepository.DeleteByID(id)
	if err != nil {
		return err
	}

	// query the users of the deleted role and update the roles in redis
	for _, user := range users {
		updatedUser, err := r.userRepository.FindByID(user.ID)
		if err != nil {
			log.Println(err)
			continue
		}
		r.tokenService.UpdateRolesIfExists(updatedUser)
	}
	return nil
}

func (r *RoleService) GetUsersOfRole(roleid uint) ([]model.User, error) {
	role, err := r.roleRepository.FindByID(roleid)
	if err != nil {
		return nil, err
	}
	users, err := r.roleRepository.GetUsersOfRole(role)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *RoleService) AddUserToRole(roleid, userid uint) error {
	role, err := r.roleRepository.FindByID(roleid)
	if err != nil {
		return err
	}
	user, err := r.userRepository.FindByID(userid)
	if err != nil {
		return err
	}
	err = r.roleRepository.AddUserToRole(role, user)
	if err != nil {
		return err
	}
	// query user again after update
	user, err = r.userRepository.FindByID(userid)
	if err != nil {
		return err
	}
	// update roles in redis
	_, err = r.tokenService.UpdateRolesIfExists(user)
	return err
}

func (r *RoleService) RemoveUserFromRole(roleid, userid uint) error {
	role, err := r.roleRepository.FindByID(roleid)
	if err != nil {
		return err
	}
	user, err := r.userRepository.FindByID(userid)
	if err != nil {
		return err
	}
	err = r.roleRepository.RemoveUserFromRole(role, user)
	if err != nil {
		return err
	}
	// query user again after update
	user, err = r.userRepository.FindByID(userid)
	if err != nil {
		return err
	}
	// update roles in redis
	_, err = r.tokenService.UpdateRolesIfExists(user)
	return err
}
