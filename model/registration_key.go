package model

import "time"

// To manually expire: set ExpiresAt to now
// Permanent overrides ExpiresAt
// TODO: maybe add #uses to have limited users per key (e.g. only one)
// TODO: maybe add roles that are assigned on registration
// @Description A registration key that can be permanent or expire at a specified date and time with which new users can register an account
type RegistrationKey struct {
	Model

	Key         string    `gorm:"uniqueIndex;not null" json:"key"` // unique registration key
	Description string    `json:"description"`                     // a description for this registration key
	Permanent   bool      `json:"permanent"`                       // if set, ignores the expires_at field and never expires this key
	ExpiresAt   time.Time `json:"expires_at"`                      // expiration date in ISO 8601 datetime

	Users []User `gorm:"constraint:OnDelete:SET NULL" json:"-"` // users that registered with this key, not serialized
} //@name RegistrationKey
