package data

import (
	"time"
)

type UserImage struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
