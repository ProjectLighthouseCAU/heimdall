package model

// A group of users sharing the same permissions
// @Description A named role that describes a group of users sharing the same permissions
type Role struct {
	Model

	Name string `gorm:"uniqueIndex;not null" json:"name"` // unique name of the role

	Users []User `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"-"` // users that have this role, not serialized
} //@name Role
