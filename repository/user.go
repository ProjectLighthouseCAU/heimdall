package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return UserRepository{
		DB: db,
	}
}

func (r *UserRepository) Save(user *model.User) error {
	return wrapError(r.DB.Save(user).Error)
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.DB.Find(&users).Error
	return users, wrapError(err)
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.DB.Preload(clause.Associations).First(&user, id).Error
	return &user, wrapError(err)
}

func (r *UserRepository) FindByName(name string) (*model.User, error) {
	var user model.User
	err := r.DB.Preload(clause.Associations).First(&user, "username = ?", name).Error
	return &user, wrapError(err)
}

func (r *UserRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(&model.User{Model: model.Model{ID: id}}).Error)
}

func (r *UserRepository) GetRolesOfUser(user *model.User) ([]model.Role, error) {
	var roles []model.Role
	err := r.DB.Model(user).Association("Roles").Find(&roles)
	return roles, wrapError(err)
}

func (r *UserRepository) AddRoleToUser(user *model.User, role *model.Role) error {
	return wrapError(r.DB.Model(user).Association("Roles").Append(role))
}

func (r *UserRepository) RemoveRoleFromUser(user *model.User, role *model.Role) error {
	return wrapError(r.DB.Model(user).Association("Roles").Delete(role))
}

func (r *UserRepository) Migrate() error {
	err := r.DB.AutoMigrate(&model.User{})
	if err != nil {
		return model.InternalServerError{Err: err}
	}
	return nil
}
