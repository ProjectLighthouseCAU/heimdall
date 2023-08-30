package service

import (
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RoleService interface {
	GetAll() ([]model.Role, error)
	GetByID(id uint) (*model.Role, error)
	GetByName(name string) (*model.Role, error)
	Create(rolename string) error
	Update(id uint, rolename string) error
	DeleteByID(id uint) error
	GetUsersOfRole(id uint) ([]model.User, error)
	AddUserToRole(roleid, userid uint) error
	RemoveUserFromRole(roleid, userid uint) error
}

type roleService struct {
	roleRepository repository.RoleRepository
	userRepository repository.UserRepository
}

var _ RoleService = (*roleService)(nil) // compile-time interface check

func NewRoleService(rr repository.RoleRepository, ur repository.UserRepository) *roleService {
	return &roleService{
		roleRepository: rr,
		userRepository: ur,
	}
}

func (r *roleService) GetAll() ([]model.Role, error) {
	return r.roleRepository.FindAll()
}

func (r *roleService) GetByID(id uint) (*model.Role, error) {
	return r.roleRepository.FindByID(id)
}

func (r *roleService) GetByName(name string) (*model.Role, error) {
	return r.roleRepository.FindByName(name)
}

func validateRole(rolename string) error {
	if !isValidName(rolename) {
		return model.BadRequestError{Message: "Invalid name"}
	}
	return nil
}

func (r *roleService) checkIfRoleExists(rolename string) error {
	_, err := r.roleRepository.FindByName(rolename)
	if err == nil {
		return model.ConflictError{Message: "Role already exists"}
	}
	return nil
}

func (r *roleService) Create(rolename string) error {
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

func (r *roleService) Update(id uint, rolename string) error {
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

func (r *roleService) DeleteByID(id uint) error {
	return r.roleRepository.DeleteByID(id)
}

func (r *roleService) GetUsersOfRole(roleid uint) ([]model.User, error) {
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

func (r *roleService) AddUserToRole(roleid, userid uint) error {
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

func (r *roleService) RemoveUserFromRole(roleid, userid uint) error {
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
