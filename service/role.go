package service

import (
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RoleService interface {
	GetAll() ([]model.Role, error)
	GetByID(id uint) (*model.Role, error)
	GetByName(name string) (*model.Role, error)
	Create(role *model.Role) error
	Update(role *model.Role) error
	Delete(role *model.Role) error
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

func (r *roleService) Create(role *model.Role) error {
	return r.roleRepository.Save(role)
}

func (r *roleService) Update(role *model.Role) error {
	_, err := r.roleRepository.FindByID(role.ID)
	if err != nil {
		return err
	}
	return r.roleRepository.Save(role)
}

func (r *roleService) Delete(role *model.Role) error {
	return r.roleRepository.Delete(role)
}

func (r *roleService) DeleteByID(id uint) error {
	return r.roleRepository.DeleteByID(id)
}

func (r *roleService) GetUsersOfRole(roleid uint) ([]model.User, error) {
	role, err := r.roleRepository.FindByID(roleid)
	if err != nil {
		return nil, err
	}
	return r.roleRepository.GetUsersOfRole(role)
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
