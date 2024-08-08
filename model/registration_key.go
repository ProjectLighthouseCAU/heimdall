package model

import "time"

// To manually expire: set ExpiresAt to now
// Permanent overrides ExpiresAt
// TODO: maybe add #uses to have limited users per key (e.g. only one)
// TODO: maybe add roles that are assigned on registration
type RegistrationKey struct {
	Model

	Key         string    `gorm:"uniqueIndex;not null" json:"key"`
	Description string    `json:"description"`
	Permanent   bool      `json:"permanent"`
	ExpiresAt   time.Time `json:"expires_at"`

	Users []User `json:"users"`
}
