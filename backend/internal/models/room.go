package models

import (
	"time"
)

type Room struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"type:varchar(255);not null" json:"name"`
	Password  string     `gorm:"type:varchar(255)" json:"-"` // Пароль не возвращается в JSON
	Rows      int        `gorm:"not null" json:"rows"`
	Cols      int        `gorm:"not null" json:"cols"`
	Mines     int        `gorm:"not null" json:"mines"`
	CreatorID int        `gorm:"default:0" json:"creatorId"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
	StartTime *time.Time `gorm:"type:timestamp;null" json:"-"` // Время начала игры

	// Связь с GameState
	GameStateData []byte `gorm:"type:bytea" json:"-"` // Бинарные данные состояния игры
}

func (Room) TableName() string {
	return "rooms"
}
