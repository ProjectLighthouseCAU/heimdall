package service

import (
	"time"

	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/crypto"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RegistrationKeyService struct {
	registrationKeyRepository repository.RegistrationKeyRepository
}

func NewRegistrationKeyService(regKeyRepo repository.RegistrationKeyRepository) RegistrationKeyService {
	return RegistrationKeyService{regKeyRepo}
}

func (r *RegistrationKeyService) GetAll() ([]model.RegistrationKey, error) {
	return r.registrationKeyRepository.FindAll()
}

func (r *RegistrationKeyService) GetByID(id uint) (*model.RegistrationKey, error) {
	return r.registrationKeyRepository.FindByID(id)
}

func (r *RegistrationKeyService) GetByKey(key string) (*model.RegistrationKey, error) {
	return r.registrationKeyRepository.FindByKey(key)
}

func (r *RegistrationKeyService) Create(key, description string, permanent bool, expiresAt time.Time) error {
	if key == "" { // special case: let the server generate the key
		var err error
		key, err = crypto.NewRandomAlphaNumString(config.GetInt("REGISTRATION_KEY_LENGTH", 20))
		if err != nil {
			return model.InternalServerError{Message: "Could not generate new registration key", Err: err}
		}
	}
	if !isValidRegistrationKey(key) {
		return model.BadRequestError{Message: "Invalid registration key"}
	}
	exists, err := r.registrationKeyRepository.ExistsByKey(key)
	if err != nil {
		return model.InternalServerError{Err: err}
	}
	if exists {
		return model.ConflictError{Message: "Registration key already exists"}
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

func (r *RegistrationKeyService) Update(id uint, description string, permanent bool, expiresAt time.Time) error {
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

func (r *RegistrationKeyService) DeleteByID(id uint) error {
	return r.registrationKeyRepository.DeleteByID(id)
}
