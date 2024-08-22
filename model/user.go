package model

import "time"

// @Description User account information including username, email, last login date and time, permanent API token flag, registration key (if user registered with a key) and roles
type User struct {
	Model

	Username          string     `gorm:"uniqueIndex;not null" json:"username"` // must be unique
	Password          string     `gorm:"not null" json:"-"`                    // hashed and not serialized
	Email             string     `json:"email"`                                // can be empty
	LastLogin         *time.Time `json:"last_login"`                           // ISO 8601 datetime TODO: redundant with UpdatedAt but only because LastLogin is updated on login
	PermanentAPIToken bool       `json:"permanent_api_token"`                  // if set the users API token never automatically expires

	RegistrationKeyID *uint            `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	RegistrationKey   *RegistrationKey `gorm:"foreignkey:RegistrationKeyID" json:"registration_key,omitempty"` // omitted if null (when user was created and not registered)
	Roles             []Role           `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"-"`     // roles of this user, not serialized
} //@name User
