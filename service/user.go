package service

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"lighthouse.uni-kiel.de/lighthouse-api/crypto"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type UserService struct {
	userRepository            repository.UserRepository
	registrationKeyRepository repository.RegistrationKeyRepository
	roleRepository            repository.RoleRepository
}

func NewUserService(ur repository.UserRepository, rkr repository.RegistrationKeyRepository, rr repository.RoleRepository) UserService {
	return UserService{
		userRepository:            ur,
		registrationKeyRepository: rkr,
		roleRepository:            rr,
	}
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

func (s *UserService) Login(username, password string, session *session.Session) error {
	_, ok := session.Get("userid").(uint)
	if ok {
		return nil // already logged in
	}
	user, err := s.userRepository.FindByName(username)
	// don't leak if username exists -> both cases return the same response
	if err != nil {
		return model.UnauthorizedError{Message: "Invalid credentials", Err: nil}
	}
	if !crypto.PasswordMatchesHash(password, user.Password) {
		return model.UnauthorizedError{Message: "Invalid credentials", Err: nil}
	}
	session.Set("userid", user.ID)
	session.Save()

	now := time.Now()
	user.LastLogin = &now
	s.userRepository.Save(user)

	// JWT
	// using JWT for now
	// TODO: maybe switch to normal session cookies
	// now := time.Now()
	// claims := jwt.RegisteredClaims{
	// 	// Issuer:    "heimdall",
	// 	Subject: username,
	// 	// Audience:  []string{"heimdall", "beacon"},
	// 	ExpiresAt: jwt.NewNumericDate(now.Add(config.GetDuration("JWT_VALID_DURATION", 1*time.Hour))),
	// 	// NotBefore: jwt.NewNumericDate(now),
	// 	// IssuedAt:  jwt.NewNumericDate(now),
	// }
	// only subject and expires_at: 129 characters
	// all claims: 235 characters
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// tokenStr, err := token.SignedString(crypto.JwtPrivateKey)
	// if err != nil {
	// 	return nil, model.InternalServerError{Message: "Could not sign JWT", Err: err}
	// }

	// token, err = jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
	// 	return crypto.JwtPrivateKey, nil
	// })
	// if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
	// 	fmt.Printf("Token valid! Claims: %+v\n", claims)
	// } else {
	// 	fmt.Println(err)
	// }
	return nil
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
	_, err := s.userRepository.FindByName(username)
	if err == nil {
		return model.ConflictError{Message: "Username already exists"}
	}
	return nil
}

func (s *UserService) Register(username, password, email, registrationKey string, session *session.Session) error {
	key, err := s.registrationKeyRepository.FindByKey(registrationKey)
	if err != nil {
		switch err.(type) {
		case *model.NotFoundError:
			return model.ForbiddenError{Message: "invalid registration key", Err: err}
		}
		return err
	}
	// check if registration key is expired
	if time.Now().After(key.ExpiresAt) && !key.Permanent {
		return model.UnauthorizedError{Message: "registration key expired"}
	}

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
	now := time.Now()
	user := model.User{
		Username:          username,
		Password:          string(hashedPassword),
		Email:             email,
		LastLogin:         &now,
		RegistrationKeyID: &key.ID,
	}
	session.Set("userid", user.ID)
	session.Save()
	return s.userRepository.Save(&user)
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
	return s.userRepository.Save(&user)
}

func (s *UserService) Update(id uint, username, password, email string) error {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return err
	}
	if err := validateUser(username, password, email); err != nil {
		return err
	}
	user.Username = username
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return model.InternalServerError{Message: "could not hash password", Err: err}
	}
	user.Password = string(hashedPassword)
	user.Email = email
	return s.userRepository.Save(user)
}

func (s *UserService) DeleteByID(id uint) error {
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

func (s *UserService) AddRoleToUser(userid, roleid uint) error {
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

func (s *UserService) RemoveRoleFromUser(userid, roleid uint) error {
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
