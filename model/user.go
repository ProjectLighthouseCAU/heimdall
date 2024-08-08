package model

import "time"

type User struct {
	Model

	Username          string           `gorm:"uniqueIndex;not null" json:"username"`
	Password          string           `gorm:"not null" json:"-"`
	Email             string           `json:"email"`
	LastLogin         *time.Time       `json:"last_login"`
	RegistrationKeyID *uint            `gorm:"constraint:OnDelete:CASCADE" json:"registration_key_id,omitempty"`
	RegistrationKey   *RegistrationKey `gorm:"foreignkey:RegistrationKeyID" json:"registration_key,omitempty"`

	Roles []Role `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"-"`
}
