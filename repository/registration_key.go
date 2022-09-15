package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

type RegistrationKeyRepository interface {
	Save(key *model.RegistrationKey) error
	FindAll() ([]model.RegistrationKey, error)
	FindByID(id uint) (*model.RegistrationKey, error)
	FindByKey(key string) (*model.RegistrationKey, error)
	Delete(key *model.RegistrationKey) error
	DeleteByID(id uint) error
	Migrate() error
}

type registrationKeyRepository struct {
	DB *gorm.DB
}

var _ RegistrationKeyRepository = (*registrationKeyRepository)(nil) // compile-time interface check

func NewRegistrationKeyRepository(db *gorm.DB) *registrationKeyRepository {
	return &registrationKeyRepository{
		DB: db,
	}
}

func (r *registrationKeyRepository) Save(key *model.RegistrationKey) error {
	return wrapError(r.DB.Save(key).Error)
}

func (r *registrationKeyRepository) FindAll() ([]model.RegistrationKey, error) {
	var keys []model.RegistrationKey
	err := r.DB.Find(&keys).Error
	return keys, wrapError(err)
}

func (r *registrationKeyRepository) FindByID(id uint) (*model.RegistrationKey, error) {
	var key model.RegistrationKey
	err := r.DB.Preload(clause.Associations).First(&key, id).Error
	return &key, wrapError(err)
}

func (r *registrationKeyRepository) FindByKey(key string) (*model.RegistrationKey, error) {
	var rkey model.RegistrationKey
	err := r.DB.Preload(clause.Associations).First(&rkey, "key = ?", key).Error
	return &rkey, wrapError(err)
}

func (r *registrationKeyRepository) Delete(key *model.RegistrationKey) error {
	return wrapError(r.DB.Unscoped().Delete(key).Error)
}

func (r *registrationKeyRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Delete(&model.RegistrationKey{}, id).Error)
}

func (r *registrationKeyRepository) Migrate() error {
	return r.DB.AutoMigrate(&model.RegistrationKey{})
}
