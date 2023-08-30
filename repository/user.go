package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

type UserRepository interface {
	Save(user *model.User) error
	FindAll() ([]model.User, error)
	FindByID(id uint) (*model.User, error)
	FindByName(name string) (*model.User, error)
	DeleteByID(id uint) error
	GetRolesOfUser(user *model.User) ([]model.Role, error)
	AddRoleToUser(user *model.User, role *model.Role) error
	RemoveRoleFromUser(user *model.User, role *model.Role) error
	Migrate() error
}

type userRepository struct {
	DB *gorm.DB
}

var _ UserRepository = (*userRepository)(nil) // compile-time interface check

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{
		DB: db,
	}
}

func (r *userRepository) Save(user *model.User) error {
	return wrapError(r.DB.Save(user).Error)
}

func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.DB.Find(&users).Error
	return users, wrapError(err)
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.DB.Preload(clause.Associations).First(&user, id).Error
	return &user, wrapError(err)
}

func (r *userRepository) FindByName(name string) (*model.User, error) {
	var user model.User
	err := r.DB.Preload(clause.Associations).First(&user, "username = ?", name).Error
	return &user, wrapError(err)
}

func (r *userRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(&model.User{Model: model.Model{ID: id}}).Error)
}

func (r *userRepository) GetRolesOfUser(user *model.User) ([]model.Role, error) {
	var roles []model.Role
	err := r.DB.Model(user).Association("Roles").Find(&roles)
	return roles, wrapError(err)
}

func (r *userRepository) AddRoleToUser(user *model.User, role *model.Role) error {
	return wrapError(r.DB.Model(user).Association("Roles").Append(role))
}

func (r *userRepository) RemoveRoleFromUser(user *model.User, role *model.Role) error {
	return wrapError(r.DB.Model(user).Association("Roles").Delete(role))
}

func (r *userRepository) Migrate() error {
	err := r.DB.AutoMigrate(&model.User{})
	if err != nil {
		return model.InternalServerError{Err: err}
	}
	return nil
}
