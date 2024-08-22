package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

type RegistrationKeyRepository struct {
	DB *gorm.DB
}

func NewRegistrationKeyRepository(db *gorm.DB) RegistrationKeyRepository {
	return RegistrationKeyRepository{
		DB: db,
	}
}

func (r *RegistrationKeyRepository) Save(key *model.RegistrationKey) error {
	return wrapError(r.DB.Save(key).Error)
}

func (r *RegistrationKeyRepository) FindAll() ([]model.RegistrationKey, error) {
	var keys []model.RegistrationKey
	err := r.DB.Find(&keys).Error
	return keys, wrapError(err)
}

func (r *RegistrationKeyRepository) FindByID(id uint) (*model.RegistrationKey, error) {
	var key model.RegistrationKey
	err := r.DB.Preload(clause.Associations).First(&key, id).Error
	return &key, wrapError(err)
}

func (r *RegistrationKeyRepository) FindByKey(key string) (*model.RegistrationKey, error) {
	var rkey model.RegistrationKey
	err := r.DB.Preload(clause.Associations).First(&rkey, "key = ?", key).Error
	return &rkey, wrapError(err)
}

func (r *RegistrationKeyRepository) ExistsByID(id uint) (bool, error) {
	var exists bool
	err := r.DB.Model(model.RegistrationKey{}).Select("count(1) > 0").Where("id = ?", id).Find(&exists).Error
	return exists, wrapError(err)
}

func (r *RegistrationKeyRepository) ExistsByKey(key string) (bool, error) {
	var exists bool
	err := r.DB.Model(model.RegistrationKey{}).Select("count(1) > 0").Where("key = ?", key).Find(&exists).Error
	return exists, wrapError(err)
}

func (r *RegistrationKeyRepository) Delete(key *model.RegistrationKey) error {
	return wrapError(r.DB.Unscoped().Delete(key).Error)
}

func (r *RegistrationKeyRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Delete(&model.RegistrationKey{}, id).Error)
}

func (r *RegistrationKeyRepository) Migrate() error {
	return r.DB.AutoMigrate(&model.RegistrationKey{})
}
