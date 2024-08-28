package repository

import (
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleRepository struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return RoleRepository{
		DB: db,
	}
}

func (r *RoleRepository) Save(role *model.Role) error {
	return wrapError(r.DB.Save(role).Error)
}

func (r *RoleRepository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.DB.Find(&roles).Order("id ASC").Error
	return roles, wrapError(err)
}

func (r *RoleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.DB.Preload(clause.Associations).First(&role, id).Error
	return &role, wrapError(err)
}

func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.DB.Preload(clause.Associations).First(&role, "name = ?", name).Error
	return &role, wrapError(err)
}

func (r *RoleRepository) FindByNames(names []string) ([]model.Role, error) {
	var roles []model.Role
	err := r.DB.Preload(clause.Associations).Where("name IN ?", names).Find(&roles).Error
	return roles, wrapError(err)
}

func (r *RoleRepository) ExistsByID(id uint) (bool, error) {
	var exists bool
	err := r.DB.Model(model.Role{}).Select("count(1) > 0").Where("id = ?", id).Find(&exists).Error
	return exists, wrapError(err)
}

func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	var exists bool
	err := r.DB.Model(model.Role{}).Select("count(1) > 0").Where("name = ?", name).Find(&exists).Error
	return exists, wrapError(err)
}

func (r *RoleRepository) Delete(role *model.Role) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(role).Error)
}

func (r *RoleRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(&model.Role{Model: model.Model{ID: id}}).Error)
}

func (r *RoleRepository) GetUsersOfRole(role *model.Role) ([]model.User, error) {
	var users []model.User
	err := r.DB.Model(role).Association("Users").Find(&users)
	return users, wrapError(err)
}

func (r *RoleRepository) AddUserToRole(role *model.Role, user *model.User) error {
	return wrapError(r.DB.Model(role).Association("Users").Append(user))
}

func (r *RoleRepository) RemoveUserFromRole(role *model.Role, user *model.User) error {
	return wrapError(r.DB.Model(role).Association("Users").Delete(user))
}

func (r *RoleRepository) Migrate() error {
	return wrapError(r.DB.AutoMigrate(&model.Role{}))
}
