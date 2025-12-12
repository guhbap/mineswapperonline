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
// Rating is updated if:
// 1. Player won AND
// 2. One of the following:
//   - First game ever (no previous results)
//   - Playing more complex field than any previous
//   - Improving or worsening time on already played field
//
// Rating increases for:
// - First game ever
// - Playing more complex fields
// - Improving time on a field
// Rating decreases for:
// - Worsening time on a field (compared to best result)
// Rating is NOT given for:
// - Playing less complex fields than previously played (prevents farming easy fields)
// This prevents farming rating on easy fields and penalizes worse performance
func (h *ProfileHandler) RecordGameResult(userID int, width, height, mines int, gameTime float64, won bool) error {
	// Get current player rating
	var currentRating float64
	err := h.db.QueryRow(
		"SELECT COALESCE(rating, 0.0) FROM users WHERE id = $1",
		userID,
	).Scan(&currentRating)
	if err != nil {
		log.Printf("Error getting player rating: %v", err)
		currentRating = 0.0 // Default rating
	}

	// Compute complexity of current field
	currentComplexity := rating.ComputeComplexity(float64(width), float64(height), float64(mines))
	Dref := rating.ComputeDref()

	// Check if field complexity is sufficient for rating (prevents farming on very easy fields)
	if !rating.IsComplexitySufficient(float64(width), float64(height), float64(mines), Dref) {
		log.Printf("Field %dx%d with %d mines (complexity %.2f) is too simple - no rating",
			width, height, mines, currentComplexity)
		// Still update statistics but not rating
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
		}
		return err
	}

	if won {
		// Новая система рейтинга: вычисляем очки попытки P
		P := rating.ComputeAttemptPoints(
			float64(width), float64(height), float64(mines),
			gameTime, currentRating, Dref,
		)

		// Получаем BestP для этого поля (лучший P, когда-то полученный игроком)
		var bestP sql.NullFloat64
		var bestTime sql.NullFloat64
		err = h.db.QueryRow(
			"SELECT COALESCE(best_p, 0.0), best_time FROM user_best_results WHERE user_id = $1 AND width = $2 AND height = $3 AND mines = $4",
			userID, width, height, mines,
		).Scan(&bestP, &bestTime)

		// Если записи нет, bestP = 0 (первая игра на этом поле)
		bestPValue := 0.0
		if err == nil && bestP.Valid {
			bestPValue = bestP.Float64
		}

		// Вычисляем награду: reward = max(0, P - BestP)
		reward := P - bestPValue
		if reward < 0 {
			reward = 0
		}

		// Обновляем рейтинг только если есть прогресс (reward > 0)
		if reward > 0 {
			newRating := currentRating + reward
			// Ensure rating doesn't go below a minimum (e.g., 0)
			if newRating < 0 {
				newRating = 0
			}

			log.Printf("Field %dx%d with %d mines: P=%.2f, BestP=%.2f, reward=%.2f, rating %.2f -> %.2f",
				width, height, mines, P, bestPValue, reward, currentRating, newRating)

			// Update user rating in database
			_, err = h.db.Exec(
				"UPDATE users SET rating = $1 WHERE id = $2",
				newRating, userID,
			)
			if err != nil {
				log.Printf("Error updating player rating: %v", err)
			}

			// Сохраняем игру в историю (только если был начислен рейтинг)
			_, err = h.db.Exec(
				`INSERT INTO user_game_history (user_id, width, height, mines, game_time, rating_gain, rating_before, rating_after, complexity, attempt_points)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				userID, width, height, mines, gameTime, reward, currentRating, newRating, currentComplexity, P,
			)
			if err != nil {
				log.Printf("Error saving game to history: %v", err)
			}

			// Update or insert best result with new BestP
			_, err = h.db.Exec(
				`INSERT INTO user_best_results (user_id, width, height, mines, best_time, complexity, best_p, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
				 ON CONFLICT (user_id, width, height, mines)
				 DO UPDATE SET 
					best_time = LEAST(user_best_results.best_time, $5),
					complexity = $6,
					best_p = GREATEST(user_best_results.best_p, $7),
					updated_at = CURRENT_TIMESTAMP`,
				userID, width, height, mines, gameTime, currentComplexity, P,
			)
			if err != nil {
				log.Printf("Error updating best result: %v", err)
			}
		} else {
			// Нет прогресса (P <= BestP), но обновляем best_time если улучшили время
			log.Printf("Field %dx%d with %d mines: P=%.2f, BestP=%.2f, no reward (no progress)",
				width, height, mines, P, bestPValue)

			// Обновляем best_time если это улучшение (best_p остается прежним)
			_, err = h.db.Exec(
				`INSERT INTO user_best_results (user_id, width, height, mines, best_time, complexity, best_p, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, COALESCE((SELECT best_p FROM user_best_results WHERE user_id = $1 AND width = $2 AND height = $3 AND mines = $4), 0.0), CURRENT_TIMESTAMP)
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
		"SELECT id, username, email, color, COALESCE(rating, 0.0), created_at FROM users WHERE id = $1",
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
		"SELECT id, username, email, color, COALESCE(rating, 0.0), created_at FROM users WHERE username = $1",
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

// GetLeaderboard возвращает список всех игроков, отсортированных по рейтингу
func (h *ProfileHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT 
			u.id,
			u.username,
			u.color,
			COALESCE(u.rating, 0.0) as rating,
			COALESCE(us.games_played, 0) as games_played,
			COALESCE(us.games_won, 0) as games_won,
			COALESCE(us.games_lost, 0) as games_lost
		FROM users u
		LEFT JOIN user_stats us ON u.id = us.user_id
		ORDER BY COALESCE(u.rating, 0.0) DESC, u.username ASC
	`)
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	type LeaderboardEntry struct {
		ID          int     `json:"id"`
		Username    string  `json:"username"`
		Color       string  `json:"color,omitempty"`
		Rating      float64 `json:"rating"`
		GamesPlayed int     `json:"gamesPlayed"`
		GamesWon    int     `json:"gamesWon"`
		GamesLost   int     `json:"gamesLost"`
	}

	var leaderboard []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		var color sql.NullString
		err := rows.Scan(
			&entry.ID,
			&entry.Username,
			&color,
			&entry.Rating,
			&entry.GamesPlayed,
			&entry.GamesWon,
			&entry.GamesLost,
		)
		if err != nil {
			log.Printf("Error scanning leaderboard row: %v", err)
			continue
		}
		if color.Valid {
			entry.Color = color.String
		}
		leaderboard = append(leaderboard, entry)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating leaderboard rows: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.JSONResponse(w, http.StatusOK, leaderboard)
}

// GetTopGames возвращает топ-10 лучших игр пользователя по начисленному рейтингу
func (h *ProfileHandler) GetTopGames(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из параметра или из контекста (для своего профиля)
	var userID int
	var err error

	username := r.URL.Query().Get("username")
	if username != "" {
		// Получаем userID по username
		user, err := h.findUserByUsername(username)
		if err != nil {
			utils.JSONError(w, http.StatusNotFound, "User not found")
			return
		}
		userID = user.ID
	} else {
		// Используем userID из контекста (свой профиль)
		userIDValue := r.Context().Value("userID")
		if userIDValue == nil {
			utils.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		userID = userIDValue.(int)
	}

	// Получаем топ-10 игр по начисленному рейтингу
	rows, err := h.db.Query(
		`SELECT id, width, height, mines, game_time, rating_gain, rating_before, rating_after, complexity, attempt_points, created_at
		 FROM user_game_history
		 WHERE user_id = $1 AND rating_gain > 0
		 ORDER BY rating_gain DESC
		 LIMIT 10`,
		userID,
	)
	if err != nil {
		log.Printf("Error querying top games: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	type GameHistory struct {
		ID            int     `json:"id"`
		Width         int     `json:"width"`
		Height        int     `json:"height"`
		Mines         int     `json:"mines"`
		GameTime      float64 `json:"gameTime"`
		RatingGain    float64 `json:"ratingGain"`
		RatingBefore  float64 `json:"ratingBefore"`
		RatingAfter   float64 `json:"ratingAfter"`
		Complexity    float64 `json:"complexity"`
		AttemptPoints float64 `json:"attemptPoints"`
		CreatedAt     string  `json:"createdAt"`
	}

	var games []GameHistory
	for rows.Next() {
		var game GameHistory
		var createdAt time.Time
		err := rows.Scan(
			&game.ID, &game.Width, &game.Height, &game.Mines,
			&game.GameTime, &game.RatingGain, &game.RatingBefore, &game.RatingAfter,
			&game.Complexity, &game.AttemptPoints, &createdAt,
		)
		if err != nil {
			log.Printf("Error scanning game history: %v", err)
			continue
		}
		game.CreatedAt = createdAt.Format(time.RFC3339)
		games = append(games, game)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating game history rows: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.JSONResponse(w, http.StatusOK, games)
}
