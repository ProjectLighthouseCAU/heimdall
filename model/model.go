package model

import (
	"time"
)

// Copy of gorm.Model with json tags
type Model struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // marshal this or filter when using soft delete (currently using hard delete)
}

// User
// type User struct {
// 	Model

// 	Username string `gorm:"uniqueIndex;not null" json:"username" validate:"required"`
// 	Password string `gorm:"not null" json:"-" validate:"required"` // skipped at json encode
// 	// Email

// 	Tokens            []Token    `gorm:"constraint:OnDelete:CASCADE" json:"tokens,omitempty"` // one-to-many
// 	Groups            []Group    `gorm:"many2many:user_groups;constraint:OnDelete:CASCADE" json:"groups,omitempty"`
// 	Resources         []Resource `gorm:"many2many:user_resources;constraint:OnDelete:CASCADE" json:"resources,omitempty"`
// 	RegistrationKeyID *uint      `gorm:"constraint:OnDelete:CASCADE" json:"registration_key_id,omitempty"` // foreign key (pointer in order to be nullable)
// }

// Authentication token for a user
// type Token struct { // TODO: maybe remove this and use JWT instead
// 	Model

// 	Token string `gorm:"not null" json:"token"`
// 	// IssuedAt  time.Time --> gorm.Model.CreatedAt
// 	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`

// 	UserID uint // foreign key
// }

// type RegistrationKey struct {
// 	Model

// 	Key         string    `gorm:"not null" json:"registration_key"`
// 	Description string    `json:"description"`
// 	Permanent   bool      // never expires
// 	Closed      bool      // manually expired
// 	ExpiresAt   time.Time `json:"expires_at"`
// 	// TODO: maybe add #uses to have limited users per key (e.g. only one)
// 	// TODO: maybe add groups that are assigned on registration
// 	// TODO: maybe add permissions that are assigned on registration
// 	Users []User `json:"users"`
// }

// A group of users sharing the same permissions
// type Group struct {
// 	Model
// 	Name      string     `gorm:"uniqueIndex;not null" json:"name"`
// 	Users     []User     `gorm:"many2many:user_groups;constraint:OnDelete:CASCADE" json:"users,omitempty"`
// 	Resources []Resource `gorm:"many2many:group_resources;constraint:OnDelete:CASCADE" json:"resources,omitempty"`
// }

// A resource identified by a path
// type Resource struct {
// 	Model
// 	Path string
// 	// Path pq.StringArray `gorm:"type:text[]"`
// 	// TODO: maybe add bool to distinguish between internal (REST API) and external (BEACON) resources
// }

// Permissions of a user or group on a resource
// type Permission struct {
// 	Create bool
// 	Read   bool
// 	Write  bool
// 	Delete bool
// }

// Join Table for User and Resource
// type UserResource struct {
// 	UserID     uint `gorm:"primarykey"`
// 	ResourceID uint `gorm:"primarykey"`
// 	CreatedAt  time.Time
// 	UpdatedAt  time.Time
// 	DeletedAt  gorm.DeletedAt
// 	Permission
// }

// Join Table for Group and Resource
// type GroupResource struct {
// 	GroupID    uint `gorm:"primarykey"`
// 	ResourceID uint `gorm:"primarykey"`
// 	CreatedAt  time.Time
// 	UpdatedAt  time.Time
// 	DeletedAt  gorm.DeletedAt
// 	Permission
// }

// ORM Setup
// func Setup(db *gorm.DB) {
// log.Println("	Setting up database")
// log.Println("		Setting up join tables")
// db.SetupJoinTable(&User{}, "Resources", &UserResource{})
// db.SetupJoinTable(&Group{}, "Resources", &GroupResource{})
// log.Println("		AutoMigrating")
// db.AutoMigrate(&RegistrationKey{}, &User{}, &Group{}, &Token{}, &Resource{})
// }
