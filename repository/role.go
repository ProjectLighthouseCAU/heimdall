package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

type RoleRepository interface {
	Save(role *model.Role) error
	FindAll() ([]model.Role, error)
	FindByID(id uint) (*model.Role, error)
	FindByName(name string) (*model.Role, error)
	FindByNames(names []string) ([]model.Role, error)
	Delete(role *model.Role) error
	DeleteByID(id uint) error
	GetUsersOfRole(role *model.Role) ([]model.User, error)
	AddUserToRole(role *model.Role, user *model.User) error
	RemoveUserFromRole(role *model.Role, user *model.User) error
	Migrate() error
}

type roleRepository struct {
	DB *gorm.DB
}

var _ RoleRepository = (*roleRepository)(nil) // compile-time interface check

func NewRoleRepository(db *gorm.DB) *roleRepository {
	return &roleRepository{
		DB: db,
	}
}

func (r *roleRepository) Save(role *model.Role) error {
	return wrapError(r.DB.Save(role).Error)
}

func (r *roleRepository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.DB.Find(&roles).Error
	return roles, wrapError(err)
}

func (r *roleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.DB.Preload(clause.Associations).First(&role, id).Error
	return &role, wrapError(err)
}

func (r *roleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.DB.Preload(clause.Associations).First(&role, "name = ?", name).Error
	return &role, wrapError(err)
}

func (r *roleRepository) FindByNames(names []string) ([]model.Role, error) {
	var roles []model.Role
	err := r.DB.Preload(clause.Associations).Where("name IN ?", names).Find(&roles).Error
	return roles, wrapError(err)
}

func (r *roleRepository) Delete(role *model.Role) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(role).Error)
}

func (r *roleRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(&model.Role{Model: model.Model{ID: id}}).Error)
}

func (r *roleRepository) GetUsersOfRole(role *model.Role) ([]model.User, error) {
	var users []model.User
	err := r.DB.Model(role).Association("Users").Find(&users)
	return users, wrapError(err)
}

func (r *roleRepository) AddUserToRole(role *model.Role, user *model.User) error {
	return wrapError(r.DB.Model(role).Association("Users").Append(user))
}

func (r *roleRepository) RemoveUserFromRole(role *model.Role, user *model.User) error {
	return wrapError(r.DB.Model(role).Association("Users").Delete(user))
}

func (r *roleRepository) Migrate() error {
	return wrapError(r.DB.AutoMigrate(&model.Role{}))
}
