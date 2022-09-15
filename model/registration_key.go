package model

import "time"

type RegistrationKey struct {
	Model

	Key         string    `gorm:"uniqueIndex;not null" json:"key"`
	Description string    `json:"description"`
	Permanent   bool      `json:"permanent"` // never expires
	Closed      bool      `json:"closed"`    // manually expired
	ExpiresAt   time.Time `json:"expires_at"`
	// TODO: maybe add #uses to have limited users per key (e.g. only one)
	// TODO: maybe add groups that are assigned on registration
	// TODO: maybe add permissions that are assigned on registration
	Users []User `json:"users"`
}
