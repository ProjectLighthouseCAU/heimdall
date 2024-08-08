package service

import (
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RoleService struct {
	roleRepository repository.RoleRepository
	userRepository repository.UserRepository
}

func NewRoleService(rr repository.RoleRepository, ur repository.UserRepository) RoleService {
	return RoleService{
		roleRepository: rr,
		userRepository: ur,
	}
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
	_, err := r.roleRepository.FindByName(rolename)
	if err == nil {
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
	if !isValidName(rolename) {
		return model.BadRequestError{Message: "Invalid name"}
	}
	role.Name = rolename
	return r.roleRepository.Save(role)
}

func (r *RoleService) DeleteByID(id uint) error {
	return r.roleRepository.DeleteByID(id)
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
	return r.roleRepository.AddUserToRole(role, user)
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
	return r.roleRepository.RemoveUserFromRole(role, user)
}
