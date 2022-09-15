package service

import (
	"time"

	"lighthouse.uni-kiel.de/lighthouse-api/auth"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByID(id uint) (*model.User, error)
	GetByName(name string) (*model.User, error)
	Register(user *model.User, registrationKey string) error
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(user *model.User) error
	DeleteByID(id uint) error
	GetRolesOfUser(userid uint) ([]model.Role, error)
	AddRoleToUser(userid, roleid uint) error
	RemoveRoleFromUser(userid, roleid uint) error
}

type userService struct {
	userRepository            repository.UserRepository
	registrationKeyRepository repository.RegistrationKeyRepository
	roleRepository            repository.RoleRepository
}

var _ UserService = (*userService)(nil) // compile-time interface check

func NewUserService(ur repository.UserRepository, rkr repository.RegistrationKeyRepository, rr repository.RoleRepository) *userService {
	return &userService{
		userRepository:            ur,
		registrationKeyRepository: rkr,
		roleRepository:            rr,
	}
}

func (s *userService) GetAll() ([]model.User, error) {
	return s.userRepository.FindAll()
}

func (s *userService) GetByID(id uint) (*model.User, error) {
	return s.userRepository.FindByID(id)
}

func (s *userService) GetByName(name string) (*model.User, error) {
	return s.userRepository.FindByName(name)
}

func (s *userService) Register(user *model.User, registrationKey string) error {
	// TODO: maybe rewrite to use registration key service instead of repository
	key, err := s.registrationKeyRepository.FindByKey(registrationKey)
	if err != nil {
		switch err.(type) {
		case *model.NotFoundError:
			return model.ForbiddenError{Message: "invalid registration key", Err: err}
		}
		return err
	}

	// check if registration key is expired or closed
	if !key.Permanent && (key.ExpiresAt.Before(time.Now()) || key.Closed) {
		return model.ForbiddenError{Message: "registration key expired"}
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		return model.InternalServerError{Message: "could not hash password", Err: err}
	}
	user.Password = string(hashedPassword)

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	user.RegistrationKeyID = &key.ID
	return s.userRepository.Save(user)
}

func (s *userService) Create(user *model.User) error {
	_, err := s.userRepository.FindByName(user.Username)
	if err == nil {
		return model.ConflictError{Message: "username already exists"}
	}
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		return model.InternalServerError{Message: "could not hash password", Err: err}
	}
	user.Password = string(hashedPassword)
	return s.userRepository.Save(user)
}

func (s *userService) Update(newUser *model.User) error {
	user, err := s.userRepository.FindByID(newUser.ID)
	if err != nil {
		return err
	}
	if newUser.Username != "" {
		user.Username = newUser.Username
	}
	if newUser.Password != "" {
		hashedPassword, err := auth.HashPassword(newUser.Password)
		if err != nil {
			return model.InternalServerError{Message: "could not hash password", Err: err}
		}
		user.Password = string(hashedPassword)
	}
	if newUser.Email != "" {
		user.Email = newUser.Email
	}
	if newUser.DisplayName != "" {
		user.DisplayName = newUser.DisplayName
	}
	return s.userRepository.Save(user)
}

func (s *userService) Delete(user *model.User) error {
	return s.userRepository.Delete(user)
}

func (s *userService) DeleteByID(id uint) error {
	return s.userRepository.DeleteByID(id)
}

func (s *userService) GetRolesOfUser(userid uint) ([]model.Role, error) {
	user, err := s.userRepository.FindByID(userid)
	if err != nil {
		return nil, err
	}
	return s.userRepository.GetRolesOfUser(user)
}

func (s *userService) AddRoleToUser(userid, roleid uint) error {
	user, err := s.userRepository.FindByID(userid)
	if err != nil {
		return err
	}
	role, err := s.roleRepository.FindByID(roleid)
	if err != nil {
		return err
	}
	return s.userRepository.AddRoleToUser(user, role)
}

func (s *userService) RemoveRoleFromUser(userid, roleid uint) error {
	user, err := s.userRepository.FindByID(userid)
	if err != nil {
		return err
	}
	role, err := s.roleRepository.FindByID(roleid)
	if err != nil {
		return err
	}
	return s.userRepository.RemoveRoleFromUser(user, role)
}
