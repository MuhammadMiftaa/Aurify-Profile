package model

import (
	"github.com/google/uuid"
)

type Profile struct {
	Base
	UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Fullname string    `gorm:"type:varchar(255)"`
	PhotoURL string    `gorm:"type:text"`
}

func (Profile) TableName() string {
	return "profiles"
}
