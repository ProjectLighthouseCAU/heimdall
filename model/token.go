package model

// type Token struct { // TODO: maybe remove this and use JWT instead
// 	Model

// 	Token string `gorm:"not null" json:"token"`
// 	// IssuedAt  time.Time --> gorm.Model.CreatedAt
// 	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`

// 	UserID uint // foreign key
// }

type Token struct {
	Token string `json:"token"`
}
