package service

import (
	"strings"

	"lighthouse.uni-kiel.de/lighthouse-api/auth"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
	"lighthouse.uni-kiel.de/lighthouse-api/repository"
)

type RegistrationKeyService interface {
	GetAll() ([]model.RegistrationKey, error)
	GetByID(id uint) (*model.RegistrationKey, error)
	GetByKey(key string) (*model.RegistrationKey, error)
	Create(key *model.RegistrationKey) error
	Update(key *model.RegistrationKey) error
	Delete(key *model.RegistrationKey) error
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

func (r *registrationKeyService) Create(key *model.RegistrationKey) error {
	if strings.TrimSpace(key.Key) == "" {
		// auto generate key
		key.Key = auth.NewRandomString(config.GetInt("REGISTRATION_KEY_LENGTH", 20))
	}
	return r.registrationKeyRepository.Save(key)
}

func (r *registrationKeyService) Update(newKey *model.RegistrationKey) error {
	key, err := r.registrationKeyRepository.FindByID(newKey.ID)
	if err != nil {
		return err
	}
	// TODO: figure out how to update only parts of the updatable fields
	key.Description = newKey.Description
	key.Permanent = newKey.Permanent
	key.Closed = newKey.Closed

	return r.registrationKeyRepository.Save(key)
}

func (r *registrationKeyService) Delete(key *model.RegistrationKey) error {
	return r.registrationKeyRepository.Delete(key)
}

func (r *registrationKeyService) DeleteByID(id uint) error {
	return r.registrationKeyRepository.DeleteByID(id)
}
