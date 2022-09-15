package repository

import (
	"errors"

	"gorm.io/gorm"
	"lighthouse.uni-kiel.de/lighthouse-api/model"
)

// wraps an error with a custom error type
func wrapError(err error) error {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.NotFoundError{Err: err}
		}
		return err // model.InternalServerError{Err: err}
	}
	return nil
}
