package handlers

import (
	"errors"
	"log"
	"net/http"
	"sort"
	"time"

	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/rating"
	"minesweeperonline/internal/utils"

	"gorm.io/gorm"
)

type ProfileHandler struct {
	db *database.DB
}

func NewProfileHandler(db *database.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

// calculateUserRating рассчитывает рейтинг пользователя как сумму рейтинга лучших игр
// topGamesCount - количество лучших игр для суммирования (по умолчанию 10)
func (h *ProfileHandler) calculateUserRating(userID int, topGamesCount int) float64 {
	if topGamesCount <= 0 {
		topGamesCount = 10 // По умолчанию топ-10
	}

	var historyRecords []models.UserGameHistory
	err := h.db.Where("user_id = ?", userID).
		Find(&historyRecords).Error

	if err != nil {
		log.Printf("Error querying games for rating calculation: %v", err)
		return 0.0
	}

	// Рассчитываем рейтинг для каждой игры
	type GameRating struct {
		Rating float64
	}
	var gameRatings []GameRating

	for _, record := range historyRecords {
		// Пропускаем игры с seed (нерейтинговые)
		if record.Seed != 0 {
			continue
		}

		// Рассчитываем рейтинг для игры
		var gameRating float64
		if rating.IsRatingEligible(float64(record.Width), float64(record.Height), float64(record.Mines), record.GameTime) {
			gameRating = rating.CalculateGameRating(float64(record.Width), float64(record.Height), float64(record.Mines), record.GameTime)
			
			// Применяем модификаторы
			if record.Chording {
				gameRating = gameRating * 0.8
			}
			if record.QuickStart {
				gameRating = gameRating * 0.9
			}
		}

		if gameRating > 0 {
			gameRatings = append(gameRatings, GameRating{Rating: gameRating})
		}
	}

	// Сортируем по рейтингу (по убыванию)
	sort.Slice(gameRatings, func(i, j int) bool {
		return gameRatings[i].Rating > gameRatings[j].Rating
	})

	// Суммируем топ-N игр
	totalRating := 0.0
	count := topGamesCount
	if len(gameRatings) < count {
		count = len(gameRatings)
	}
	for i := 0; i < count; i++ {
		totalRating += gameRatings[i].Rating
	}

	return totalRating
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

	// Рассчитываем рейтинг динамически
	userRating := h.calculateUserRating(userID, 10)
	user.Rating = userRating

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

	// Рассчитываем рейтинг динамически
	userRating := h.calculateUserRating(user.ID, 10)
	user.Rating = userRating

	profile := models.UserProfile{
		User:  user,
		Stats: stats,
	}

	utils.JSONResponse(w, http.StatusOK, profile)
}

func (h *ProfileHandler) UpdateLastSeen(userID int) error {
	now := time.Now()
	stats := models.UserStats{
		UserID:    userID,
		LastSeen:  now,
		UpdatedAt: now,
	}

	err := h.db.Where("user_id = ?", userID).FirstOrCreate(&stats, models.UserStats{UserID: userID}).Error
	if err != nil {
		return err
	}

	// Обновляем last_seen и updated_at
	return h.db.Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"last_seen":  now,
			"updated_at": now,
		}).Error
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

	var colorPtr *string
	if req.Color != "" {
		colorPtr = &req.Color
	}

	err := h.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("color", colorPtr).Error
	if err != nil {
		log.Printf("Error updating color for user %d: %v", userID, err)
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update color")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProfileHandler) FindUserColor(id int) (string, error) {
	var user models.User
	err := h.db.Select("color").First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	if user.Color != nil {
		return *user.Color, nil
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
type GameParticipant struct {
	UserID   int
	Nickname string
	Color    string
}

func (h *ProfileHandler) RecordGameResult(userID int, width, height, mines int, gameTime float64, won bool, chording bool, quickStart bool, roomID string, seed int64, creatorID int, participants []GameParticipant) error {
	// Если participants не передан, используем пустой слайс
	if participants == nil {
		participants = []GameParticipant{}
	}

	var err error
	// Сохраняем игру в историю (для побед и поражений)
	log.Printf("Сохранение игры в историю: userID=%d, roomID=%s, размер=%dx%d, мины=%d, время=%.2f сек, seed=%d, creatorID=%d, won=%v",
		userID, roomID, width, height, mines, gameTime, seed, creatorID, won)
	gameHistory := models.UserGameHistory{
		UserID:     userID,
		RoomID:     roomID,
		Width:      width,
		Height:     height,
		Mines:      mines,
		GameTime:   gameTime,
		Seed:       seed,
		CreatorID:  creatorID,
		Chording:   chording,
		QuickStart: quickStart,
		CreatedAt:  time.Now(),
	}
	err = h.db.Create(&gameHistory).Error
	if err != nil {
		log.Printf("Error saving game to history: %v", err)
	} else if len(participants) > 0 {
		// Сохраняем участников игры
		for _, participant := range participants {
			var colorPtr *string
			if participant.Color != "" {
				colorPtr = &participant.Color
			}
			gameParticipant := models.GameParticipant{
				GameHistoryID: gameHistory.ID,
				UserID:        participant.UserID,
				Nickname:      participant.Nickname,
				Color:         colorPtr,
			}
			err = h.db.Where("game_history_id = ? AND user_id = ?", gameHistory.ID, participant.UserID).
				FirstOrCreate(&gameParticipant).Error
			if err != nil {
				log.Printf("Error saving game participant: %v", err)
			}
		}
	}

	if won {
		// Проверяем, может ли игра дать рейтинг
		// Если указан seed (не равен 0), игра нерейтинговая
		if seed != 0 {
			log.Printf("Игра не дает рейтинг: указан seed=%d (игра нерейтинговая)", seed)
		} else if !rating.IsRatingEligible(float64(width), float64(height), float64(mines), gameTime) {
			log.Printf("Игра не дает рейтинг: плотность=%.2f%% (мин. 10%%)",
				float64(mines)/(float64(width)*float64(height))*100)
		} else {
			// Вычисляем рейтинг за игру по формуле: R = K * d / ln(t + 1)
			gameRating := rating.CalculateGameRating(float64(width), float64(height), float64(mines), gameTime)
			
			// Если используется Chording, рейтинг умножается на 0.8
			if chording {
				gameRating = gameRating * 0.8
				log.Printf("Chording enabled: рейтинг умножен на 0.8")
			}
			
			// Если используется QuickStart, рейтинг умножается на 0.9
			if quickStart {
				gameRating = gameRating * 0.9
				log.Printf("QuickStart enabled: рейтинг умножен на 0.9")
			}

			log.Printf("Field %dx%d with %d mines, time=%.2f, chording=%v, quickStart=%v: gameRating=%.2f",
				width, height, mines, gameTime, chording, quickStart, gameRating)
			// Рейтинг теперь рассчитывается динамически как сумма лучших игр, не сохраняем в БД
		}
	} else {
		// For lost games, don't update rating
		log.Printf("Game lost - no rating update")
	}

	// Update game statistics
	now := time.Now()
	if won {
		stats := models.UserStats{UserID: userID}
		err = h.db.Where("user_id = ?", userID).FirstOrCreate(&stats).Error
		if err != nil {
			return err
		}
		err = h.db.Model(&models.UserStats{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"games_played": gorm.Expr("games_played + ?", 1),
				"games_won":    gorm.Expr("games_won + ?", 1),
				"updated_at":   now,
			}).Error
		return err
	} else {
		stats := models.UserStats{UserID: userID}
		err = h.db.Where("user_id = ?", userID).FirstOrCreate(&stats).Error
		if err != nil {
			return err
		}
		err = h.db.Model(&models.UserStats{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"games_played": gorm.Expr("games_played + ?", 1),
				"games_lost":   gorm.Expr("games_lost + ?", 1),
				"updated_at":   now,
			}).Error
		return err
	}
}

func (h *ProfileHandler) findUserByID(id int) (models.User, error) {
	return h.FindUserByID(id)
}

// FindUserByID находит пользователя по ID (публичный метод)
func (h *ProfileHandler) FindUserByID(id int) (models.User, error) {
	var user models.User
	err := h.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}
	return user, err
}

func (h *ProfileHandler) findUserByUsername(username string) (models.User, error) {
	var user models.User
	err := h.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}
	return user, err
}

func (h *ProfileHandler) getUserStats(userID int) (models.UserStats, error) {
	var stats models.UserStats
	err := h.db.Where("user_id = ?", userID).First(&stats).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserStats{}, err
	}
	return stats, err
}

func (h *ProfileHandler) createUserStats(userID int) (models.UserStats, error) {
	now := time.Now()
	stats := models.UserStats{
		UserID:      userID,
		GamesPlayed: 0,
		GamesWon:    0,
		GamesLost:   0,
		LastSeen:    now,
		UpdatedAt:   now,
	}
	err := h.db.Create(&stats).Error
	if err != nil {
		return models.UserStats{}, err
	}
	return stats, nil
}

// GetLeaderboard возвращает список всех игроков, отсортированных по рейтингу
func (h *ProfileHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	type LeaderboardEntry struct {
		ID          int     `json:"id"`
		Username    string  `json:"username"`
		Color       string  `json:"color,omitempty"`
		Rating      float64 `json:"rating"`
		GamesPlayed int     `json:"gamesPlayed"`
		GamesWon    int     `json:"gamesWon"`
		GamesLost   int     `json:"gamesLost"`
	}

	var users []models.User
	err := h.db.Order("username ASC").Find(&users).Error
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var leaderboard []LeaderboardEntry
	for _, user := range users {
		// Рассчитываем рейтинг динамически для каждого пользователя
		userRating := h.calculateUserRating(user.ID, 10)

		entry := LeaderboardEntry{
			ID:       user.ID,
			Username: user.Username,
			Rating:   userRating,
		}
		if user.Color != nil {
			entry.Color = *user.Color
		}

		// Получаем статистику пользователя
		var stats models.UserStats
		h.db.Where("user_id = ?", user.ID).First(&stats)
		entry.GamesPlayed = stats.GamesPlayed
		entry.GamesWon = stats.GamesWon
		entry.GamesLost = stats.GamesLost

		leaderboard = append(leaderboard, entry)
	}

	// Сортируем по рейтингу (по убыванию), затем по username
	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].Rating != leaderboard[j].Rating {
			return leaderboard[i].Rating > leaderboard[j].Rating
		}
		return leaderboard[i].Username < leaderboard[j].Username
	})

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

	// Получаем все игры пользователя для расчета рейтинга
	type GameHistory struct {
		ID        int     `json:"id"`
		Width     int     `json:"width"`
		Height    int     `json:"height"`
		Mines     int     `json:"mines"`
		GameTime  float64 `json:"gameTime"`
		Rating    float64 `json:"rating"`
		CreatedAt string  `json:"createdAt"`
	}

	var historyRecords []models.UserGameHistory
	err = h.db.Where("user_id = ?", userID).
		Find(&historyRecords).Error

	if err != nil {
		log.Printf("Error querying top games: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var games []GameHistory
	for _, record := range historyRecords {
		// Рассчитываем рейтинг для игры
		var gameRating float64
		if rating.IsRatingEligible(float64(record.Width), float64(record.Height), float64(record.Mines), record.GameTime) {
			gameRating = rating.CalculateGameRating(float64(record.Width), float64(record.Height), float64(record.Mines), record.GameTime)
			
			// Применяем модификаторы
			if record.Chording {
				gameRating = gameRating * 0.8
			}
			if record.QuickStart {
				gameRating = gameRating * 0.9
			}
		}

		games = append(games, GameHistory{
			ID:        record.ID,
			Width:     record.Width,
			Height:    record.Height,
			Mines:     record.Mines,
			GameTime:  record.GameTime,
			Rating:    gameRating,
			CreatedAt: record.CreatedAt.Format(time.RFC3339),
		})
	}

	// Сортируем по рейтингу (по убыванию) и берем топ-10
	sort.Slice(games, func(i, j int) bool {
		return games[i].Rating > games[j].Rating
	})
	if len(games) > 10 {
		games = games[:10]
	}

	utils.JSONResponse(w, http.StatusOK, games)
}

// GetRecentGames возвращает последние 10 игр пользователя с информацией об участниках
func (h *ProfileHandler) GetRecentGames(w http.ResponseWriter, r *http.Request) {
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

	// Получаем последние 10 игр
	type GameParticipantInfo struct {
		UserID   int    `json:"userId"`
		Nickname string `json:"nickname"`
		Color    string `json:"color,omitempty"`
	}

	type RecentGame struct {
		ID           int                   `json:"id"`
		Width        int                   `json:"width"`
		Height       int                   `json:"height"`
		Mines        int                   `json:"mines"`
		GameTime     float64               `json:"gameTime"`
		Rating       float64               `json:"rating"`
		CreatedAt    string                `json:"createdAt"`
		Participants []GameParticipantInfo `json:"participants"`
	}

	var historyRecords []models.UserGameHistory
	err = h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&historyRecords).Error

	if err != nil {
		log.Printf("Error querying recent games: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var games []RecentGame
	for _, record := range historyRecords {
		// Рассчитываем рейтинг для игры
		var gameRating float64
		if rating.IsRatingEligible(float64(record.Width), float64(record.Height), float64(record.Mines), record.GameTime) {
			gameRating = rating.CalculateGameRating(float64(record.Width), float64(record.Height), float64(record.Mines), record.GameTime)
			
			// Применяем модификаторы
			if record.Chording {
				gameRating = gameRating * 0.8
			}
			if record.QuickStart {
				gameRating = gameRating * 0.9
			}
		}

		game := RecentGame{
			ID:           record.ID,
			Width:        record.Width,
			Height:       record.Height,
			Mines:        record.Mines,
			GameTime:     record.GameTime,
			Rating:       gameRating,
			CreatedAt:    record.CreatedAt.Format(time.RFC3339),
			Participants: []GameParticipantInfo{},
		}

		// Получаем участников игры
		var participants []models.GameParticipant
		err := h.db.Where("game_history_id = ?", record.ID).Find(&participants).Error
		if err == nil {
			for _, p := range participants {
				participant := GameParticipantInfo{
					UserID:   p.UserID,
					Nickname: p.Nickname,
				}
				if p.Color != nil {
					participant.Color = *p.Color
				}
				game.Participants = append(game.Participants, participant)
			}
		}

		games = append(games, game)
	}

	utils.JSONResponse(w, http.StatusOK, games)
}
