package model

import "time"

// @Description API token that allows access to the websocket API (beacon) and probably other APIs in the future
type Token struct {
	UserID    uint      `gorm:"primarykey;constraint:OnDelete:SET NULL;not null" json:"-"`
	Token     string    `gorm:"uniqueIndex;not null" json:"api_token"`
	Permanent bool      `json:"permanent"`                  // if permanent is true, expires_at is ignored
	CreatedAt time.Time `json:"created_at"`                 // ISO 8601 datetime
	UpdatedAt time.Time `json:"updated_at"`                 // ISO 8601 datetime
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"` // ISO 8601 datetime
} //@name Token

// @Description Message that is sent to notify subscribers (e.g. Beacon) on changes to one of these authentication related values
type AuthUpdateMessage struct {
	Username  string    `json:"username"`   // unique username associated with this token
	Token     string    `json:"api_token"`  // the actual API token
	ExpiresAt time.Time `json:"expires_at"` // expiration date of this token
	Permanent bool      `json:"permanent"`  // no expiration (ignore ExpiresAt)
	Roles     []string  `json:"roles"`      // roles associated with this token
} //@name AuthUpdateMessage

// @Description Message that is sent to notify subscribers (e.g. Beacon) when a new user is created or a user is removed
type UserUpdateMessage struct {
	Username string `json:"username"`
	Removed  bool   `json:"removed"`
} //@name UserUpdateMessage
