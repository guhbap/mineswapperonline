package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/utils"
)

type ProfileHandler struct {
	db *database.DB
}

func NewProfileHandler(db *database.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	// Получаем информацию о пользователе
	user, err := h.findUserByID(userID)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Получаем статистику пользователя
	stats, err := h.getUserStats(userID)
	if err != nil {
		// Если статистики нет, создаем запись
		stats, err = h.createUserStats(userID)
		if err != nil {
			log.Printf("Error creating user stats: %v", err)
			utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	// Проверяем онлайн статус (последний раз был онлайн менее 5 минут назад)
	stats.IsOnline = time.Since(stats.LastSeen) < 5*time.Minute

	profile := models.UserProfile{
		User:  user,
		Stats: stats,
	}

	utils.JSONResponse(w, http.StatusOK, profile)
}

func (h *ProfileHandler) UpdateLastSeen(userID int) error {
	_, err := h.db.Exec(
		`INSERT INTO user_stats (user_id, last_seen, updated_at) 
		 VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (user_id) 
		 DO UPDATE SET last_seen = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP`,
		userID,
	)
	return err
}

func (h *ProfileHandler) RecordGameResult(userID int, won bool) error {
	if won {
		_, err := h.db.Exec(
			`INSERT INTO user_stats (user_id, games_played, games_won, updated_at) 
			 VALUES ($1, 1, 1, CURRENT_TIMESTAMP)
			 ON CONFLICT (user_id) 
			 DO UPDATE SET 
				games_played = user_stats.games_played + 1,
				games_won = user_stats.games_won + 1,
				updated_at = CURRENT_TIMESTAMP`,
			userID,
		)
		return err
	} else {
		_, err := h.db.Exec(
			`INSERT INTO user_stats (user_id, games_played, games_lost, updated_at) 
			 VALUES ($1, 1, 1, CURRENT_TIMESTAMP)
			 ON CONFLICT (user_id) 
			 DO UPDATE SET 
				games_played = user_stats.games_played + 1,
				games_lost = user_stats.games_lost + 1,
				updated_at = CURRENT_TIMESTAMP`,
			userID,
		)
		return err
	}
}

func (h *ProfileHandler) findUserByID(id int) (models.User, error) {
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, email, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return models.User{}, err
	}
	return user, err
}

func (h *ProfileHandler) getUserStats(userID int) (models.UserStats, error) {
	var stats models.UserStats
	err := h.db.QueryRow(
		`SELECT user_id, games_played, games_won, games_lost, last_seen 
		 FROM user_stats WHERE user_id = $1`,
		userID,
	).Scan(&stats.UserID, &stats.GamesPlayed, &stats.GamesWon, &stats.GamesLost, &stats.LastSeen)

	if err == sql.ErrNoRows {
		return models.UserStats{}, err
	}
	return stats, err
}

func (h *ProfileHandler) createUserStats(userID int) (models.UserStats, error) {
	_, err := h.db.Exec(
		`INSERT INTO user_stats (user_id, games_played, games_won, games_lost, last_seen, updated_at) 
		 VALUES ($1, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID,
	)
	if err != nil {
		return models.UserStats{}, err
	}

	return h.getUserStats(userID)
}

