package service

import (
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/crypto"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/ProjectLighthouseCAU/heimdall/repository"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type UserService struct {
	userRepository            repository.UserRepository
	registrationKeyRepository repository.RegistrationKeyRepository
	roleRepository            repository.RoleRepository
	tokenService              TokenService
}

func NewUserService(userRepo repository.UserRepository,
	regKeyRepo repository.RegistrationKeyRepository,
	roleRepo repository.RoleRepository,
	tokenService TokenService) UserService {
	return UserService{userRepo, regKeyRepo, roleRepo, tokenService}
}

func (s *UserService) GetAll() ([]model.User, error) {
	return s.userRepository.FindAll()
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.userRepository.FindByID(id)
}

func (s *UserService) GetByName(name string) (*model.User, error) {
	return s.userRepository.FindByName(name)
}

func (s *UserService) Login(username, password string, session *session.Session) (*model.User, error) {
	userid := session.Get("userid")
	if userid != nil { // already logged in
		uid, ok := userid.(uint)
		if !ok {
			return nil, model.InternalServerError{Message: "Error retrieving userid from session"}
		}
		return s.userRepository.FindByID(uid)
	}
	user, err := s.userRepository.FindByName(username)
	// don't leak if username exists -> both cases return the same response
	if err != nil {
		return nil, model.UnauthorizedError{Message: "Invalid credentials", Err: nil}
	}
	if !crypto.PasswordMatchesHash(password, user.Password) {
		return nil, model.UnauthorizedError{Message: "Invalid credentials", Err: nil}
	}

	session.Set("userid", user.ID)
	err = session.Save()
	if err != nil {
		return nil, model.InternalServerError{Message: "Could not save session", Err: err}
	}

	now := time.Now()
	user.LastLogin = &now
	err = s.userRepository.Save(user)
	if err != nil {
		return nil, model.InternalServerError{Message: "Could not save user", Err: err}
	}
	s.tokenService.GenerateApiTokenIfNotExists(user)
	return user, nil
}

func (s *UserService) Logout(session *session.Session) error {
	if err := session.Destroy(); err != nil {
		return model.InternalServerError{Message: "Could not destroy session", Err: err}
	}
	return nil
}

func validateUser(username, password, email string) error {
	if !isValidName(username) {
		return model.BadRequestError{Message: "Invalid name"}
	}
	if !isValidPassword(password) {
		return model.BadRequestError{Message: "Password does not meet criteria"}
	}
	if !isValidEmail(email) {
		return model.BadRequestError{Message: "Invalid email"}
	}
	return nil
}

func (s *UserService) checkIfUserExists(username string) error {
	exists, err := s.userRepository.ExistsByName(username)
	if err != nil {
		return model.InternalServerError{Err: err}
	}
	if exists {
		return model.ConflictError{Message: "Username already exists"}
	}
	return nil
}

func (s *UserService) Register(username, password, email, registrationKey string, session *session.Session) (*model.User, error) {
	key, err := s.registrationKeyRepository.FindByKey(registrationKey)
	if err != nil {
		switch err.(type) {
		case model.NotFoundError:
			return nil, model.UnauthorizedError{Message: "invalid registration key", Err: err}
		}
		return nil, err
	}
	// check if registration key is expired
	if time.Now().After(key.ExpiresAt) && !key.Permanent {
		return nil, model.UnauthorizedError{Message: "registration key expired"}
	}

	if err := validateUser(username, password, email); err != nil {
		return nil, err
	}
	if err := s.checkIfUserExists(username); err != nil {
		return nil, err
	}
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, model.InternalServerError{Message: "could not hash password", Err: err}
	}
	now := time.Now()
	user := model.User{
		Username:          username,
		Password:          string(hashedPassword),
		Email:             email,
		LastLogin:         &now,
		RegistrationKeyID: &key.ID,
	}
	err = s.userRepository.Save(&user)
	if err != nil {
		return nil, err
	}
	savedUser, err := s.userRepository.FindByName(user.Username)
	if err != nil {
		return nil, err
	}
	if session != nil {
		session.Set("userid", savedUser.ID)
		err = session.Save()
		if err != nil {
			return nil, model.InternalServerError{Message: "could not save session", Err: err}
		}
	}
	s.tokenService.GenerateApiTokenIfNotExists(savedUser)
	return savedUser, nil
}

func (s *UserService) Create(username, password, email string, permanentAPIToken bool) error {
	if err := validateUser(username, password, email); err != nil {
		return err
	}
	if err := s.checkIfUserExists(username); err != nil {
		return err
	}
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return model.InternalServerError{Message: "could not hash password", Err: err}
	}
	user := model.User{
		Username:          username,
		Password:          string(hashedPassword),
		Email:             email,
		LastLogin:         nil,
		PermanentAPIToken: permanentAPIToken,
	}
	return s.userRepository.Save(&user)
}

func (s *UserService) Update(id uint, username, password, email string, permanentAPIToken bool) error {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return err
	}
	if err := validateUser(username, password, email); err != nil {
		return err
	}
	if username != user.Username || user.PermanentAPIToken != permanentAPIToken {
		// renew token if username changed
		existed, _ := s.tokenService.InvalidateApiTokenIfExists(user)
		user.Username = username
		user.PermanentAPIToken = permanentAPIToken
		if existed {
			_, _ = s.tokenService.GenerateApiTokenIfNotExists(user)
		}
		// TODO: maybe keep list of previous names?
	}
	if !crypto.PasswordMatchesHash(password, user.Password) {
		hashedPassword, err := crypto.HashPassword(password)
		if err != nil {
			return model.InternalServerError{Message: "could not hash password", Err: err}
		}
		user.Password = string(hashedPassword)
	}
	user.Email = email
	return s.userRepository.Save(user)
}

func (s *UserService) DeleteByID(id uint) error {
	// invalidate token
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return model.NotFoundError{Err: err}
	}
	s.tokenService.InvalidateApiTokenIfExists(user)
	return s.userRepository.DeleteByID(id)
}

func (s *UserService) GetRolesOfUser(userid uint) ([]model.Role, error) {
	user, err := s.userRepository.FindByID(userid)
	if err != nil {
		return nil, err
	}
	roles, err := s.userRepository.GetRolesOfUser(user)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
