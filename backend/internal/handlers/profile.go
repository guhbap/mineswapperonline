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

func (h *ProfileHandler) GetProfileByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		utils.JSONError(w, http.StatusBadRequest, "Username parameter is required")
		return
	}

	// Получаем информацию о пользователе по username
	user, err := h.findUserByUsername(username)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Получаем статистику пользователя
	stats, err := h.getUserStats(user.ID)
	if err != nil {
		// Если статистики нет, создаем запись
		stats, err = h.createUserStats(user.ID)
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

func (h *ProfileHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	
	if err := h.UpdateLastSeen(userID); err != nil {
		log.Printf("Error updating last seen for user %d: %v", userID, err)
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update activity")
		return
	}
	
	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProfileHandler) UpdateColor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	
	var req struct {
		Color string `json:"color"`
	}
	
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Валидация цвета (hex формат)
	if req.Color != "" && !isValidHexColor(req.Color) {
		utils.JSONError(w, http.StatusBadRequest, "Invalid color format. Expected hex color (e.g., #FF5733)")
		return
	}
	
	_, err := h.db.Exec(
		"UPDATE users SET color = $1 WHERE id = $2",
		req.Color, userID,
	)
	if err != nil {
		log.Printf("Error updating color for user %d: %v", userID, err)
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update color")
		return
	}
	
	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProfileHandler) FindUserColor(id int) (string, error) {
	var color sql.NullString
	err := h.db.QueryRow(
		"SELECT color FROM users WHERE id = $1",
		id,
	).Scan(&color)

	if err == sql.ErrNoRows {
		return "", err
	}
	if color.Valid {
		return color.String, nil
	}
	return "", nil
}

func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
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
		"SELECT id, username, email, color, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Color, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return models.User{}, err
	}
	return user, err
}

func (h *ProfileHandler) findUserByUsername(username string) (models.User, error) {
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, email, color, created_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Color, &user.CreatedAt)

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

