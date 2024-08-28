package repository

import (
	"errors"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"gorm.io/gorm"
)

// wraps an error with a custom error type
func wrapError(err error) error {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.NotFoundError{Err: err}
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.ConflictError{Err: err}
		}
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return model.ConflictError{Err: err}
		}
		return err // model.InternalServerError{Err: err}
	}
	return nil
}
