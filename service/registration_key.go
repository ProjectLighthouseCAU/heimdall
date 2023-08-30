package service

import (
	"time"

	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/crypto"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RegistrationKeyService interface {
	GetAll() ([]model.RegistrationKey, error)
	GetByID(id uint) (*model.RegistrationKey, error)
	GetByKey(key string) (*model.RegistrationKey, error)
	Create(key, description string, permanent bool, expiresAt time.Time) error
	Update(id uint, description string, permanent bool, expiresAt time.Time) error
	DeleteByID(id uint) error
}

type registrationKeyService struct {
	registrationKeyRepository repository.RegistrationKeyRepository
}

var _ RegistrationKeyService = (*registrationKeyService)(nil) // compile-time interface check

func NewRegistrationKeyService(r repository.RegistrationKeyRepository) *registrationKeyService {
	return &registrationKeyService{
		registrationKeyRepository: r,
	}
}

func (r *registrationKeyService) GetAll() ([]model.RegistrationKey, error) {
	return r.registrationKeyRepository.FindAll()
}

func (r *registrationKeyService) GetByID(id uint) (*model.RegistrationKey, error) {
	return r.registrationKeyRepository.FindByID(id)
}

func (r *registrationKeyService) GetByKey(key string) (*model.RegistrationKey, error) {
	return r.registrationKeyRepository.FindByKey(key)
}

func (r *registrationKeyService) checkIfKeyExists(key string) error {
	return nil
}

func (r *registrationKeyService) Create(key, description string, permanent bool, expiresAt time.Time) error {
	if key == "" { // special case: let the server generate the key
		key = crypto.NewRandomAlphaNumString(config.GetInt("REGISTRATION_KEY_LENGTH", 20))
	}
	if !isValidRegistrationKey(key) {
		return model.BadRequestError{Message: "Invalid registration key"}
	}
	if err := r.checkIfKeyExists(key); err != nil {
		return err
	}
	// no restrictions on description, expiresAt (can be in the past for deactivated key)
	// and permanent (false by default)
	regKey := model.RegistrationKey{
		Key:         key,
		Description: description,
		Permanent:   permanent,
		ExpiresAt:   expiresAt,
	}
	return r.registrationKeyRepository.Save(&regKey)
}

func (r *registrationKeyService) Update(id uint, description string, permanent bool, expiresAt time.Time) error {
	// no restrictions on description, permanent and expiresAt (see Create)
	key, err := r.registrationKeyRepository.FindByID(id)
	if err != nil {
		return err
	}
	key.Description = description
	key.Permanent = permanent
	key.ExpiresAt = expiresAt
	return r.registrationKeyRepository.Save(key)
}

func (r *registrationKeyService) DeleteByID(id uint) error {
	return r.registrationKeyRepository.DeleteByID(id)
}
