package model

import "time"

// @Description API token that allows access to the websocket API (beacon) and probably other APIs in the future
// Note: API tokens are persisted in Redis and therefore have no gorm annotations and do not include Model
type APIToken struct {
	Token     string    `json:"api_token"`  // the actual API token
	Username  string    `json:"username"`   // unique username associated with this token
	Roles     []string  `json:"roles"`      // roles associated with this token
	ExpiresAt time.Time `json:"expires_at"` // expiration date of this token
} //@name APIToken
