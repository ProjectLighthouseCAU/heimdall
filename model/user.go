package model

type User struct {
	Model

	Username    string `gorm:"uniqueIndex;not null" json:"username"`
	Password    string `gorm:"not null" json:"-"` // skipped at json encode
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`

	// Tokens            []Token `gorm:"constraint:OnDelete:CASCADE" json:"tokens,omitempty"`              // one-to-many
	RegistrationKeyID *uint            `gorm:"constraint:OnDelete:CASCADE" json:"registration_key_id,omitempty"` // foreign key (pointer in order to be nullable)
	RegistrationKey   *RegistrationKey `gorm:"foreignkey:RegistrationKeyID" json:"registration_key,omitempty"`

	Roles []Role `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"-"`
}
