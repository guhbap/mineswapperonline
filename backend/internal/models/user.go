package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Color        string    `json:"color,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
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
	UserID      int       `json:"userId"`
	GamesPlayed int       `json:"gamesPlayed"`
	GamesWon    int       `json:"gamesWon"`
	GamesLost   int       `json:"gamesLost"`
	LastSeen    time.Time `json:"lastSeen"`
	IsOnline    bool      `json:"isOnline"`
}

type UserProfile struct {
	User  User      `json:"user"`
	Stats UserStats `json:"stats"`
}

