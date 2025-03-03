package repository

import (
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TokenRepository struct {
	DB *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return TokenRepository{
		DB: db,
	}
}

func (r *TokenRepository) Save(token *model.Token) error {
	return wrapError(r.DB.Save(token).Error)
}

func (r *TokenRepository) FindAll() ([]model.Token, error) {
	var tokens []model.Token
	err := r.DB.Preload("Users").Find(&tokens).Order("id ASC").Error
	return tokens, wrapError(err)
}

func (r *TokenRepository) FindByID(id uint) (*model.Token, error) {
	var token model.Token
	err := r.DB.Preload(clause.Associations).First(&token, id).Error
	return &token, wrapError(err)
}

func (r *TokenRepository) FindByToken(token string) (*model.Token, error) {
	var tokenModel model.Token
	err := r.DB.Preload(clause.Associations).First(&tokenModel, "token = ?", token).Error
	return &tokenModel, wrapError(err)
}

func (r *TokenRepository) ExistsByID(id uint) (bool, error) {
	var exists bool
	err := r.DB.Model(model.Token{}).Select("count(1) > 0").Where("id = ?", id).Find(&exists).Error
	return exists, wrapError(err)
}

func (r *TokenRepository) ExistsByToken(token string) (bool, error) {
	var exists bool
	err := r.DB.Model(model.Token{}).Select("count(1) > 0").Where("token = ?", token).Find(&exists).Error
	return exists, wrapError(err)
}

func (r *TokenRepository) DeleteByID(id uint) error {
	return wrapError(r.DB.Unscoped().Select(clause.Associations).Delete(&model.Token{Model: model.Model{ID: id}}).Error)
}

func (r *TokenRepository) Migrate() error {
	err := r.DB.AutoMigrate(&model.Token{})
	if err != nil {
		return model.InternalServerError{Err: err}
	}
	return nil
}
