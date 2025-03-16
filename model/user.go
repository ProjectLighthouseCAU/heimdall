package model

import "time"

// @Description User account information including username, email, last login date and time, permanent API token flag, registration key (if user registered with a key) and roles
type User struct {
	Model

	Username  string     `gorm:"uniqueIndex;not null" json:"username"` // must be unique
	Password  string     `gorm:"not null" json:"-"`                    // hashed and not serialized
	Email     string     `json:"email"`                                // can be empty
	LastLogin *time.Time `json:"last_login"`                           // ISO 8601 datetime

	RegistrationKeyID *uint            `gorm:"constraint:OnDelete:SET NULL" json:"-"`
	RegistrationKey   *RegistrationKey `gorm:"constraint:OnDelete:SET NULL" json:"registration_key,omitempty"` // omitted if null (when user was created and not registered or when list of users is queried to not leak other users keys)
	Roles             []Role           `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"roles"`
	ApiToken          *Token           `gorm:"constraint:OnDelete:CASCADE;not null" json:"api_token,omitempty"` // omitted if null (user doesn't have an API token)
} //@name User
