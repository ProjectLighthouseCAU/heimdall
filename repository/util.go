package repository

import (
	"errors"
	"log"

	"github.com/ProjectLighthouseCAU/heimdall/model"
	"gorm.io/gorm"
)

// wraps an error with a custom error type
func wrapError(err error) error {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.NotFoundError{Message: "Record not found"}
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.ConflictError{Message: "Duplicated unique key"}
		}
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return model.ConflictError{Message: "Foreign key constraint violated"}
		}
		log.Println(err)
		return model.InternalServerError{Message: "Database error, see logs"}
	}
	return nil
}
