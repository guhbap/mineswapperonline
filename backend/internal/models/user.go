package models

import (
	"time"
)

type User struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex:idx_users_username;not null" json:"username"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex:idx_users_email;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null;column:password_hash" json:"-"`
	Color        *string   `gorm:"type:varchar(7)" json:"color,omitempty"`
	Rating       float64   `gorm:"-" json:"rating"` // Вычисляемое поле, не хранится в БД
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
}

func (User) TableName() string {
	return "users"
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UserStats struct {
	UserID      int       `gorm:"primaryKey;column:user_id" json:"userId"`
	GamesPlayed int       `gorm:"default:0" json:"gamesPlayed"`
	GamesWon    int       `gorm:"default:0" json:"gamesWon"`
	GamesLost   int       `gorm:"default:0" json:"gamesLost"`
	LastSeen    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"lastSeen"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
	IsOnline    bool      `gorm:"-" json:"isOnline"` // Вычисляемое поле, не хранится в БД
}

func (UserStats) TableName() string {
	return "user_stats"
}

type UserProfile struct {
	User  User      `json:"user"`
	Stats UserStats `json:"stats"`
}

type UserGameHistory struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int       `gorm:"not null;column:user_id" json:"userId"`
	RoomID        string    `gorm:"type:varchar(255);column:room_id" json:"roomId"`
	Width         int       `gorm:"not null" json:"width"`
	Height        int       `gorm:"not null" json:"height"`
	Mines         int       `gorm:"not null" json:"mines"`
	GameTime      float64   `gorm:"type:double precision;not null;column:game_time" json:"gameTime"`
	Seed          string    `gorm:"type:varchar(36);not null" json:"seed"`
	HasCustomSeed bool      `gorm:"default:false;column:has_custom_seed" json:"hasCustomSeed"`
	CreatorID     int       `gorm:"not null;column:creator_id" json:"creatorId"`
	Won           bool      `gorm:"default:false" json:"won"`
	Chording      bool      `gorm:"default:false" json:"chording"`
	QuickStart    bool      `gorm:"default:false;column:quick_start" json:"quickStart"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
}

func (UserGameHistory) TableName() string {
	return "user_game_history"
}

type GameParticipant struct {
	GameHistoryID int     `gorm:"primaryKey;column:game_history_id;index:idx_game_participants_game_history_id" json:"gameHistoryId"`
	UserID        int     `gorm:"primaryKey;column:user_id" json:"userId"`
	Nickname      string  `gorm:"type:varchar(100);not null" json:"nickname"`
	Color         *string `gorm:"type:varchar(7)" json:"color,omitempty"`
}

func (GameParticipant) TableName() string {
	return "game_participants"
}
