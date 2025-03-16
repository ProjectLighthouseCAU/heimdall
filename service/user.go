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
	uid, uidOk := session.Get("userid").(uint)
	sessionUsername, usernameOk := session.Get("username").(string)
	sessionPassword, passwordOk := session.Get("password").(string)

	if uidOk && usernameOk && passwordOk { // already logged in
		user, err := s.userRepository.FindByID(uid)
		if err == nil { // user exists
			if sessionUsername == user.Username && sessionPassword == user.Password { // username and password weren't changed
				return user, nil // user already logged in and still authenticated
			}
		}
		// user was deleted or changed username or password
		if err := session.Destroy(); err != nil {
			return nil, model.InternalServerError{Message: "Could not destroy session", Err: err}
		}
		// continue with login
		// NOTE: we can use the session after session.Destroy() as a new empty session
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
	session.Set("username", user.Username)
	session.Set("password", user.Password)
	if err = session.Save(); err != nil {
		return nil, model.InternalServerError{Message: "Could not save session", Err: err}
	}

	now := time.Now()
	user.LastLogin = &now
	if err = s.userRepository.Save(user); err != nil {
		return nil, model.InternalServerError{Message: "Could not save user", Err: err}
	}
	tokenWasGenerated, err := s.tokenService.GenerateApiTokenIfNotExists(user)
	if err != nil {
		return nil, err
	}
	if tokenWasGenerated { // query user again to include generated api token
		user, err := s.userRepository.FindByID(user.ID)
		if err != nil {
			return nil, model.NotFoundError{Message: "Could not find user after login - this should not happen", Err: err}
		}
		return user, nil
	}
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
	if session != nil { // session is only nil when Register is called from setupTestDatabase
		if _, ok := session.Get("userid").(uint); ok {
			return nil, model.BadRequestError{Message: "You cannot register when you are logged in!"}
		}
	}
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
		Username:        username,
		Password:        string(hashedPassword),
		Email:           email,
		LastLogin:       &now,
		RegistrationKey: key,
	}
	if err := s.userRepository.Save(&user); err != nil {
		return nil, err
	}
	savedUser, err := s.userRepository.FindByName(user.Username)
	if err != nil {
		return nil, err
	}
	if session != nil { // session is only nil when Register is called from setupTestDatabase
		session.Set("userid", savedUser.ID)
		session.Set("username", savedUser.Username)
		session.Set("password", savedUser.Password)
		if err := session.Save(); err != nil {
			return nil, model.InternalServerError{Message: "could not save session", Err: err}
		}
	}

	s.tokenService.NotifyUserCreated(savedUser)
	if _, err := s.tokenService.GenerateApiTokenIfNotExists(savedUser); err != nil {
		return nil, err
	}
	return savedUser, nil
}

func (s *UserService) Create(username, password, email string) error {
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
		Username:  username,
		Password:  string(hashedPassword),
		Email:     email,
		LastLogin: nil,
	}

	if err = s.userRepository.Save(&user); err != nil {
		return err
	}
	s.tokenService.NotifyUserCreated(&user)
	return nil
}

func (s *UserService) Update(id uint, username, password, email string) error {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return err
	}
	if !isValidName(username) {
		return model.BadRequestError{Message: "Invalid name"}
	}
	if password != "" && !isValidPassword(password) {
		return model.BadRequestError{Message: "Password does not meet criteria"}
	}
	if !isValidEmail(email) {
		return model.BadRequestError{Message: "Invalid email"}
	}
	regenerateApiTokenAfterUpdate := false
	previousUser := *user
	if username != user.Username {
		regenerateApiTokenAfterUpdate = true
		previousUser = *user // copy user before update
		user.Username = username
		// TODO: maybe keep list of previous names?
	}
	if password != "" && !crypto.PasswordMatchesHash(password, user.Password) {
		hashedPassword, err := crypto.HashPassword(password)
		if err != nil {
			return model.InternalServerError{Message: "could not hash password", Err: err}
		}
		regenerateApiTokenAfterUpdate = true
		user.Password = string(hashedPassword)
	}
	user.Email = email
	if err = s.userRepository.Save(user); err != nil {
		return err
	}
	if regenerateApiTokenAfterUpdate {
		s.tokenService.NotifyUsernameInvalid(&previousUser)
		_, _ = s.tokenService.GenerateApiTokenIfNotExists(user)
		// TODO: destroy all (other) sessions of the user
	}
	return nil
}

func (s *UserService) DeleteByID(id uint) error {
	// invalidate token
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return model.NotFoundError{Err: err}
	}

	if err = s.userRepository.DeleteByID(id); err != nil {
		return err
	}
	s.tokenService.NotifyUsernameInvalid(user)
	s.tokenService.NotifyUserDeleted(user)
	// NOTE: We do not need to destroy the deleted user's session
	// since the session middleware checks if the user exists.
	// The session is destroyed when it is used after the user was deleted.
	return nil
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
