package model

// A group of users sharing the same permissions
type Role struct {
	Model
	Name  string `gorm:"uniqueIndex;not null" json:"name"`
	Users []User `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"-"`
	// Roles []Role `gorm:"many2many:role_roles;constraint:OnDelete:CASCADE;" json:"-"` // TODO: hierarchical roles or not?
}
