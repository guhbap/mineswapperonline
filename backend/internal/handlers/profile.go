package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/rating"
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

// RecordGameResult records game result and updates player rating
// width, height, mines - dimensions of the game field
// gameTime - time taken to complete the game in seconds
// won - whether the player won
// Rating is only updated if:
// 1. Player won AND
// 2. Either this is a new field (no previous best result for this exact field) OR
//    the field is more complex than any previous field the player has played OR
//    the player improved their time on this exact field
// This prevents farming rating on easy fields by requiring either:
// - Playing new fields
// - Playing progressively more complex fields
// - Improving performance on already played fields
func (h *ProfileHandler) RecordGameResult(userID int, width, height, mines int, gameTime float64, won bool) error {
	// Get current player rating
	var currentRating float64
	err := h.db.QueryRow(
		"SELECT COALESCE(rating, 1500.0) FROM users WHERE id = $1",
		userID,
	).Scan(&currentRating)
	if err != nil {
		log.Printf("Error getting player rating: %v", err)
		currentRating = 1500.0 // Default rating
	}

	// Compute complexity of current field
	currentComplexity := rating.ComputeComplexity(float64(width), float64(height), float64(mines))

	// Check if we should update rating
	shouldUpdateRating := false
	var newRating float64 = currentRating

	if won {
		// Get best result for this exact field
		var bestTime sql.NullFloat64
		var bestComplexity sql.NullFloat64
		err = h.db.QueryRow(
			"SELECT best_time, complexity FROM user_best_results WHERE user_id = $1 AND width = $2 AND height = $3 AND mines = $4",
			userID, width, height, mines,
		).Scan(&bestTime, &bestComplexity)

		if err == sql.ErrNoRows {
			// No previous result for this field - give rating
			shouldUpdateRating = true
			log.Printf("First time playing field %dx%d with %d mines - giving rating", width, height, mines)
		} else if err != nil {
			log.Printf("Error getting best result: %v", err)
			// On error, don't update rating to be safe
		} else {
			// We have a previous result for this field
			if bestTime.Valid && bestComplexity.Valid {
				// Check if this is an improvement (better time)
				if gameTime < bestTime.Float64 {
					shouldUpdateRating = true
					log.Printf("Improved time on field %dx%d with %d mines: %.2f -> %.2f", width, height, mines, bestTime.Float64, gameTime)
				} else {
					log.Printf("No improvement on field %dx%d with %d mines: %.2f >= %.2f (no rating)", width, height, mines, gameTime, bestTime.Float64)
				}
			}
		}

		// Also check if this is the most complex field the player has played
		if !shouldUpdateRating {
			var maxComplexity sql.NullFloat64
			err = h.db.QueryRow(
				"SELECT MAX(complexity) FROM user_best_results WHERE user_id = $1",
				userID,
			).Scan(&maxComplexity)

			if err != nil && err != sql.ErrNoRows {
				log.Printf("Error getting max complexity: %v", err)
			} else if err == sql.ErrNoRows || !maxComplexity.Valid {
				// No previous results at all - give rating
				shouldUpdateRating = true
				log.Printf("First game ever - giving rating")
			} else if currentComplexity > maxComplexity.Float64 {
				// This is a more complex field than any previous - give rating
				shouldUpdateRating = true
				log.Printf("Playing more complex field: %.2f > %.2f - giving rating", currentComplexity, maxComplexity.Float64)
			}
		}

		// Update rating if conditions are met
		if shouldUpdateRating {
			Dref := rating.ComputeDref()
			newRating, _ = rating.UpdatePlayerRating(
				float64(width), float64(height), float64(mines),
				gameTime, currentRating, Dref,
			)

			// Update user rating in database
			_, err = h.db.Exec(
				"UPDATE users SET rating = $1 WHERE id = $2",
				newRating, userID,
			)
			if err != nil {
				log.Printf("Error updating player rating: %v", err)
			}

			// Update or insert best result
			_, err = h.db.Exec(
				`INSERT INTO user_best_results (user_id, width, height, mines, best_time, complexity, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
				 ON CONFLICT (user_id, width, height, mines)
				 DO UPDATE SET 
					best_time = LEAST(user_best_results.best_time, $5),
					complexity = $6,
					updated_at = CURRENT_TIMESTAMP`,
				userID, width, height, mines, gameTime, currentComplexity,
			)
			if err != nil {
				log.Printf("Error updating best result: %v", err)
			}
		} else {
			// Still update best result even if rating doesn't change (for tracking)
			_, err = h.db.Exec(
				`INSERT INTO user_best_results (user_id, width, height, mines, best_time, complexity, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
				 ON CONFLICT (user_id, width, height, mines)
				 DO UPDATE SET 
					best_time = LEAST(user_best_results.best_time, $5),
					updated_at = CURRENT_TIMESTAMP`,
				userID, width, height, mines, gameTime, currentComplexity,
			)
			if err != nil {
				log.Printf("Error updating best result: %v", err)
			}
		}
	} else {
		// For lost games, don't update rating or best results
		log.Printf("Game lost - no rating update")
	}

	// Update game statistics
	if won {
		_, err = h.db.Exec(
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
		_, err = h.db.Exec(
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
		"SELECT id, username, email, color, COALESCE(rating, 1500.0), created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Color, &user.Rating, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return models.User{}, err
	}
	return user, err
}

func (h *ProfileHandler) findUserByUsername(username string) (models.User, error) {
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, email, color, COALESCE(rating, 1500.0), created_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Color, &user.Rating, &user.CreatedAt)

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

