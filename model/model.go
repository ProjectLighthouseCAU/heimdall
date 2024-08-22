package model

import (
	"time"
)

// Copy of gorm.Model with json tags
type Model struct {
	ID        uint      `gorm:"primarykey" json:"id"` // id (primary key)
	CreatedAt time.Time `json:"created_at"`           // ISO 8601 datetime
	UpdatedAt time.Time `json:"updated_at"`           // ISO 8601 datetime
}
