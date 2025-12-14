package handlers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"minesweeperonline/internal/auth"
	"minesweeperonline/internal/cache"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/game"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/rating"
	"minesweeperonline/internal/utils"

	"gorm.io/gorm"
)

type ProfileHandler struct {
	db    *database.DB
	cache *cache.Cache
}

func NewProfileHandler(db *database.DB) *ProfileHandler {
	// Создаем кеш с TTL 5 минут для профилей и статистики
	// Для лидерборда используем более короткий TTL (1 минута)
	return &ProfileHandler{
		db:    db,
		cache: cache.NewCache(5 * time.Minute),
	}
}

// calculateGameRating рассчитывает рейтинг для одной игры с учетом модификаторов
func (h *ProfileHandler) calculateGameRating(width, height, mines int, gameTime float64, chording, quickStart bool) float64 {
	if !rating.IsRatingEligible(float64(width), float64(height), float64(mines), gameTime) {
		return 0.0
	}

	gameRating := rating.CalculateGameRating(float64(width), float64(height), float64(mines), gameTime)

	// Применяем модификаторы
	if chording {
		gameRating = gameRating * 0.8
	}
	if quickStart {
		gameRating = gameRating * 0.9
	}

	return gameRating
}

// calculateUserRating рассчитывает рейтинг пользователя как сумму рейтинга лучших игр
// topGamesCount - количество лучших игр для суммирования (по умолчанию 10)
func (h *ProfileHandler) calculateUserRating(userID int, topGamesCount int) float64 {
	if topGamesCount <= 0 {
		topGamesCount = 100 // По умолчанию топ-100
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
		// Учитываем только выигранные игры
		if !record.Won {
			continue
		}

		// Пропускаем игры с явно указанным пользователем seed (нерейтинговые)
		if record.HasCustomSeed {
			continue
		}

		// Рассчитываем рейтинг для игры
		gameRating := h.calculateGameRating(record.Width, record.Height, record.Mines, record.GameTime, record.Chording, record.QuickStart)

		if gameRating > 0 {
			gameRatings = append(gameRatings, GameRating{Rating: gameRating})
		}
	}

	// Сортируем по рейтингу (по убыванию)
	sort.Slice(gameRatings, func(i, j int) bool {
		return gameRatings[i].Rating > gameRatings[j].Rating
	})

	// Суммируем топ-N игр с весовыми коэффициентами
	// Первая игра (лучшая) дает 100% рейтинга (коэффициент 1.0)
	// Вторая - 95% (0.95)
	// Третья - 90.25% (0.95^2)
	// N-я игра дает 0.95^(n-1) процентов
	totalRating := 0.0
	count := topGamesCount
	if len(gameRatings) < count {
		count = len(gameRatings)
	}
	for i := 0; i < count; i++ {
		// Коэффициент для i-й игры: 0.95^i
		// i=0 (первая игра) -> коэффициент = 1.0 (100%)
		// i=1 (вторая игра) -> коэффициент = 0.95 (95%)
		// i=2 (третья игра) -> коэффициент = 0.9025 (90.25%)
		coefficient := math.Pow(0.95, float64(i))
		totalRating += gameRatings[i].Rating * coefficient
	}

	return totalRating
}

// buildUserProfile создает профиль пользователя с расчетом рейтинга и статистики
func (h *ProfileHandler) buildUserProfile(userID int) (models.UserProfile, error) {
	// Проверяем кеш
	cacheKey := fmt.Sprintf("profile:%d", userID)
	if cached, found := h.cache.Get(cacheKey); found {
		if profile, ok := cached.(models.UserProfile); ok {
			// Обновляем онлайн статус (он может измениться)
			profile.Stats.IsOnline = time.Since(profile.Stats.LastSeen) < 5*time.Minute
			return profile, nil
		}
	}

	// Получаем информацию о пользователе
	user, err := h.FindUserByID(userID)
	if err != nil {
		return models.UserProfile{}, err
	}

	// Получаем или создаем статистику пользователя
	stats, err := h.getUserStats(userID)
	if err != nil {
		stats, err = h.createUserStats(userID)
		if err != nil {
			return models.UserProfile{}, err
		}
	}

	// Проверяем онлайн статус (последний раз был онлайн менее 5 минут назад)
	stats.IsOnline = time.Since(stats.LastSeen) < 5*time.Minute

	// Рассчитываем рейтинг динамически
	userRating := h.calculateUserRating(userID, 100)
	user.Rating = userRating

	profile := models.UserProfile{
		User:  user,
		Stats: stats,
	}

	// Сохраняем в кеш
	h.cache.Set(cacheKey, profile)

	return profile, nil
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	profile, err := h.buildUserProfile(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(w, http.StatusNotFound, "User not found")
		} else {
			log.Printf("Error building user profile: %v", err)
			utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
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

	profile, err := h.buildUserProfile(user.ID)
	if err != nil {
		log.Printf("Error building user profile: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
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
	err = h.db.Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"last_seen":  now,
			"updated_at": now,
		}).Error

	// Инвалидируем кеш статистики и профиля (last_seen влияет на онлайн статус)
	if err == nil {
		h.cache.Delete(fmt.Sprintf("stats:%d", userID))
		h.cache.Delete(fmt.Sprintf("profile:%d", userID))
	}

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

	// Инвалидируем кеш пользователя и профиля
	h.cache.Delete(fmt.Sprintf("user:id:%d", userID))
	h.cache.Delete(fmt.Sprintf("profile:%d", userID))
	// Найдем username для инвалидации кеша по username
	user, _ := h.FindUserByID(userID)
	if user.Username != "" {
		h.cache.Delete(fmt.Sprintf("user:username:%s", user.Username))
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := r.Context().Value("userID").(int)

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация паролей
	if req.CurrentPassword == "" || req.NewPassword == "" {
		utils.JSONError(w, http.StatusBadRequest, "Current password and new password are required")
		return
	}

	if len(req.NewPassword) < 6 {
		utils.JSONError(w, http.StatusBadRequest, "New password must be at least 6 characters")
		return
	}

	// Получаем пользователя из БД
	user, err := h.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(w, http.StatusNotFound, "User not found")
		} else {
			log.Printf("Error finding user %d: %v", userID, err)
			utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Проверяем текущий пароль
	if !auth.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid current password")
		return
	}

	// Хешируем новый пароль
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password for user %d: %v", userID, err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Обновляем пароль в БД
	err = h.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", newPasswordHash).Error
	if err != nil {
		log.Printf("Error updating password for user %d: %v", userID, err)
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// Инвалидируем кеш пользователя (пароль не влияет на профиль, но для безопасности очистим)
	h.cache.Delete(fmt.Sprintf("user:id:%d", userID))
	if user.Username != "" {
		h.cache.Delete(fmt.Sprintf("user:username:%s", user.Username))
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
func (h *ProfileHandler) RecordGameResult(userID int, width, height, mines int, gameTime float64, won bool, chording bool, quickStart bool, roomID string, seed string, hasCustomSeed bool, creatorID int, participants []game.GameParticipant) error {
	// Если participants не передан, используем пустой слайс
	if participants == nil {
		participants = []game.GameParticipant{}
	}

	var err error
	// Сохраняем игру в историю (для побед и поражений)
	log.Printf("Сохранение игры в историю: userID=%d, roomID=%s, размер=%dx%d, мины=%d, время=%.2f сек, seed=%s (len=%d), creatorID=%d, won=%v",
		userID, roomID, width, height, mines, gameTime, seed, len(seed), creatorID, won)
	gameHistory := models.UserGameHistory{
		UserID:        userID,
		RoomID:        roomID,
		Width:         width,
		Height:        height,
		Mines:         mines,
		GameTime:      gameTime,
		Seed:          seed,
		HasCustomSeed: hasCustomSeed,
		CreatorID:     creatorID,
		Won:           won,
		Chording:      chording,
		QuickStart:    quickStart,
		CreatedAt:     time.Now(),
	}
	log.Printf("GameHistory перед сохранением: Seed=%s (len=%d), тип=%T", gameHistory.Seed, len(gameHistory.Seed), gameHistory.Seed)
	// Сохраняем через GORM, но проверяем тип колонки перед сохранением
	err = h.db.Create(&gameHistory).Error
	if err != nil {
		log.Printf("Error saving game to history: %v", err)
	} else {
		// Проверяем, что сохранилось
		var savedRecord models.UserGameHistory
		if err := h.db.First(&savedRecord, gameHistory.ID).Error; err == nil {
			log.Printf("Проверка сохраненного seed: ID=%d, Seed=%s (len=%d)", savedRecord.ID, savedRecord.Seed, len(savedRecord.Seed))
			if savedRecord.Seed != gameHistory.Seed {
				log.Printf("ОШИБКА: seed не совпадает! Ожидалось: %s (len=%d), получено: %s (len=%d)",
					gameHistory.Seed, len(gameHistory.Seed), savedRecord.Seed, len(savedRecord.Seed))
				// Пытаемся обновить через Raw SQL
				if err := h.db.Exec(`UPDATE user_game_history SET seed = ? WHERE id = ?`, gameHistory.Seed, gameHistory.ID).Error; err != nil {
					log.Printf("Ошибка обновления seed через Raw SQL: %v", err)
				} else {
					log.Printf("Seed обновлен через Raw SQL: %s", gameHistory.Seed)
				}
			}
		}

		if len(participants) > 0 {
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
	}

	if won {
		// Проверяем, может ли игра дать рейтинг
		// Если seed был указан пользователем явно, игра нерейтинговая
		if hasCustomSeed {
			log.Printf("Игра не дает рейтинг: указан seed=%s (игра нерейтинговая)", seed)
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
	return h.updateGameStats(userID, won)
}

// updateGameStats обновляет статистику игр пользователя
func (h *ProfileHandler) updateGameStats(userID int, won bool) error {
	stats := models.UserStats{UserID: userID}
	if err := h.db.Where("user_id = ?", userID).FirstOrCreate(&stats).Error; err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"games_played": gorm.Expr("games_played + ?", 1),
		"updated_at":   now,
	}

	if won {
		updates["games_won"] = gorm.Expr("games_won + ?", 1)
	} else {
		updates["games_lost"] = gorm.Expr("games_lost + ?", 1)
	}

	err := h.db.Model(&models.UserStats{}).
		Where("user_id = ?", userID).
		Updates(updates).Error

	// Инвалидируем кеш статистики, профиля и лидерборда
	if err == nil {
		h.cache.Delete(fmt.Sprintf("stats:%d", userID))
		h.cache.Delete(fmt.Sprintf("profile:%d", userID))
		h.cache.Delete("leaderboard") // Лидерборд зависит от статистики
	}

	return err
}

// FindUserByID находит пользователя по ID
func (h *ProfileHandler) FindUserByID(userID int) (models.User, error) {
	// Проверяем кеш
	cacheKey := fmt.Sprintf("user:id:%d", userID)
	if cached, found := h.cache.Get(cacheKey); found {
		if user, ok := cached.(models.User); ok {
			return user, nil
		}
	}

	var user models.User
	err := h.db.First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}

	if err == nil {
		// Сохраняем в кеш
		h.cache.Set(cacheKey, user)
		// Также кешируем по username для быстрого поиска
		if user.Username != "" {
			h.cache.Set(fmt.Sprintf("user:username:%s", user.Username), user)
		}
	}

	return user, err
}

func (h *ProfileHandler) findUserByUsername(username string) (models.User, error) {
	// Проверяем кеш
	cacheKey := fmt.Sprintf("user:username:%s", username)
	if cached, found := h.cache.Get(cacheKey); found {
		if user, ok := cached.(models.User); ok {
			return user, nil
		}
	}

	var user models.User
	err := h.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}

	if err == nil {
		// Сохраняем в кеш
		h.cache.Set(cacheKey, user)
		// Также кешируем по ID
		h.cache.Set(fmt.Sprintf("user:id:%d", user.ID), user)
	}

	return user, err
}

func (h *ProfileHandler) getUserStats(userID int) (models.UserStats, error) {
	// Проверяем кеш
	cacheKey := fmt.Sprintf("stats:%d", userID)
	if cached, found := h.cache.Get(cacheKey); found {
		if stats, ok := cached.(models.UserStats); ok {
			return stats, nil
		}
	}

	var stats models.UserStats
	err := h.db.Where("user_id = ?", userID).First(&stats).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserStats{}, err
	}

	if err == nil {
		// Сохраняем в кеш
		h.cache.Set(cacheKey, stats)
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

// LeaderboardEntry представляет запись в лидерборде
type LeaderboardEntry struct {
	ID          int     `json:"id"`
	Username    string  `json:"username"`
	Color       string  `json:"color,omitempty"`
	Rating      float64 `json:"rating"`
	GamesPlayed int     `json:"gamesPlayed"`
	GamesWon    int     `json:"gamesWon"`
	GamesLost   int     `json:"gamesLost"`
}

// GetLeaderboard возвращает список всех игроков, отсортированных по рейтингу
func (h *ProfileHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Проверяем кеш (лидерборд кешируем на 1 минуту)
	cacheKey := "leaderboard"
	if cached, found := h.cache.Get(cacheKey); found {
		if leaderboard, ok := cached.([]LeaderboardEntry); ok {
			utils.JSONResponse(w, http.StatusOK, leaderboard)
			return
		}
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
		userRating := h.calculateUserRating(user.ID, 100)

		entry := LeaderboardEntry{
			ID:       user.ID,
			Username: user.Username,
			Rating:   userRating,
		}
		if user.Color != nil {
			entry.Color = *user.Color
		}

		// Получаем статистику пользователя (используем кешированный метод)
		stats, err := h.getUserStats(user.ID)
		if err == nil {
			entry.GamesPlayed = stats.GamesPlayed
			entry.GamesWon = stats.GamesWon
			entry.GamesLost = stats.GamesLost
		}

		leaderboard = append(leaderboard, entry)
	}

	// Сортируем по рейтингу (по убыванию), затем по username
	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].Rating != leaderboard[j].Rating {
			return leaderboard[i].Rating > leaderboard[j].Rating
		}
		return leaderboard[i].Username < leaderboard[j].Username
	})

	// Сохраняем в кеш (используем основной кеш, но лидерборд обновляется часто)
	h.cache.Set(cacheKey, leaderboard)

	utils.JSONResponse(w, http.StatusOK, leaderboard)
}

// getUserIDFromRequest получает userID из запроса (из username параметра или контекста)
func (h *ProfileHandler) getUserIDFromRequest(r *http.Request) (int, error) {
	username := r.URL.Query().Get("username")
	if username != "" {
		user, err := h.findUserByUsername(username)
		if err != nil {
			return 0, err
		}
		return user.ID, nil
	}

	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		return 0, errors.New("unauthorized")
	}
	return userIDValue.(int), nil
}

// GetTopGames возвращает топ-100 лучших игр пользователя по начисленному рейтингу
func (h *ProfileHandler) GetTopGames(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(w, http.StatusNotFound, "User not found")
		} else {
			utils.JSONError(w, http.StatusUnauthorized, "Unauthorized")
		}
		return
	}

	// Получаем все игры пользователя для расчета рейтинга
	type GameHistory struct {
		ID                int     `json:"id"`
		Width             int     `json:"width"`
		Height            int     `json:"height"`
		Mines             int     `json:"mines"`
		GameTime          float64 `json:"gameTime"`
		Rating            float64 `json:"rating"`            // Рейтинг игры (до применения коэффициента)
		RatingPercent     float64 `json:"ratingPercent"`     // Процент засчитанного рейтинга (0.95^позиция * 100)
		RatingContributed float64 `json:"ratingContributed"` // Конкретно полученный рейтинг (рейтинг * коэффициент)
		Won               bool    `json:"won"`
		CreatedAt         string  `json:"createdAt"`
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
		// Рассчитываем рейтинг только для выигранных игр без явно указанного seed
		var gameRating float64
		if record.Won && !record.HasCustomSeed {
			gameRating = h.calculateGameRating(record.Width, record.Height, record.Mines, record.GameTime, record.Chording, record.QuickStart)
		}

		games = append(games, GameHistory{
			ID:        record.ID,
			Width:     record.Width,
			Height:    record.Height,
			Mines:     record.Mines,
			GameTime:  record.GameTime,
			Rating:    gameRating,
			Won:       record.Won,
			CreatedAt: record.CreatedAt.Format(time.RFC3339),
		})
	}

	// Фильтруем только выигранные игры с рейтингом > 0
	var wonGames []GameHistory
	for _, game := range games {
		if game.Won && game.Rating > 0 {
			wonGames = append(wonGames, game)
		}
	}

	// Сортируем по рейтингу (по убыванию) и берем топ-10
	sort.Slice(wonGames, func(i, j int) bool {
		return wonGames[i].Rating > wonGames[j].Rating
	})
	if len(wonGames) > 100 {
		wonGames = wonGames[:100]
	}

	// Вычисляем процент засчитанного и полученный рейтинг для каждой игры
	for i := range wonGames {
		// Коэффициент для i-й игры: 0.95^i
		// i=0 (первая игра) -> коэффициент = 1.0 (100%)
		// i=1 (вторая игра) -> коэффициент = 0.95 (95%)
		// i=2 (третья игра) -> коэффициент = 0.9025 (90.25%)
		coefficient := math.Pow(0.95, float64(i))
		wonGames[i].RatingPercent = coefficient * 100.0
		wonGames[i].RatingContributed = wonGames[i].Rating * coefficient
	}

	utils.JSONResponse(w, http.StatusOK, wonGames)
}

// GetRecentGames возвращает последние 10 игр пользователя с информацией об участниках
func (h *ProfileHandler) GetRecentGames(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(w, http.StatusNotFound, "User not found")
		} else {
			utils.JSONError(w, http.StatusUnauthorized, "Unauthorized")
		}
		return
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
		Won          bool                  `json:"won"`
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
		// Рассчитываем рейтинг только для выигранных игр без явно указанного seed
		var gameRating float64
		if record.Won && !record.HasCustomSeed {
			gameRating = h.calculateGameRating(record.Width, record.Height, record.Mines, record.GameTime, record.Chording, record.QuickStart)
		}

		game := RecentGame{
			ID:           record.ID,
			Width:        record.Width,
			Height:       record.Height,
			Mines:        record.Mines,
			GameTime:     record.GameTime,
			Rating:       gameRating,
			Won:          record.Won,
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

// GetGameDetails возвращает детальную информацию об игре по ID
func (h *ProfileHandler) GetGameDetails(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("id")
	if gameID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Game ID parameter is required")
		return
	}

	var gameHistory models.UserGameHistory
	err := h.db.First(&gameHistory, gameID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(w, http.StatusNotFound, "Game not found")
		} else {
			log.Printf("Error getting game details: %v", err)
			utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Получаем информацию о создателе комнаты
	var creator models.User
	creatorErr := h.db.First(&creator, gameHistory.CreatorID).Error
	if creatorErr != nil {
		log.Printf("Error getting creator: %v", creatorErr)
	}

	// Получаем участников игры
	var participants []models.GameParticipant
	h.db.Where("game_history_id = ?", gameHistory.ID).Find(&participants)

	// Получаем информацию о пользователях-участниках
	type ParticipantInfo struct {
		UserID   int    `json:"userId"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Color    string `json:"color,omitempty"`
	}

	participantInfos := make([]ParticipantInfo, 0, len(participants))
	for _, p := range participants {
		var user models.User
		if err := h.db.First(&user, p.UserID).Error; err == nil {
			participantInfo := ParticipantInfo{
				UserID:   p.UserID,
				Username: user.Username,
				Nickname: p.Nickname,
			}
			if p.Color != nil {
				participantInfo.Color = *p.Color
			}
			participantInfos = append(participantInfos, participantInfo)
		}
	}

	// Рассчитываем рейтинг игры
	var gameRating float64
	if gameHistory.Won && !gameHistory.HasCustomSeed {
		gameRating = h.calculateGameRating(
			gameHistory.Width,
			gameHistory.Height,
			gameHistory.Mines,
			gameHistory.GameTime,
			gameHistory.Chording,
			gameHistory.QuickStart,
		)
	}

	// Формируем ответ
	type GameDetailsResponse struct {
		ID            int               `json:"id"`
		RoomID        string            `json:"roomId"`
		Width         int               `json:"width"`
		Height        int               `json:"height"`
		Mines         int               `json:"mines"`
		Seed          string            `json:"seed"`
		HasCustomSeed bool              `json:"hasCustomSeed"`
		CreatorID     int               `json:"creatorId"`
		CreatorName   string            `json:"creatorName"`
		Won           bool              `json:"won"`
		Chording      bool              `json:"chording"`
		QuickStart    bool              `json:"quickStart"`
		StartTime     string            `json:"startTime"`
		Duration      float64           `json:"duration"`
		Rating        float64           `json:"rating"`
		Participants  []ParticipantInfo `json:"participants"`
	}

	response := GameDetailsResponse{
		ID:            gameHistory.ID,
		RoomID:        gameHistory.RoomID,
		Width:         gameHistory.Width,
		Height:        gameHistory.Height,
		Mines:         gameHistory.Mines,
		Seed:          gameHistory.Seed,
		HasCustomSeed: gameHistory.HasCustomSeed,
		CreatorID:     gameHistory.CreatorID,
		CreatorName:   creator.Username,
		Won:           gameHistory.Won,
		Chording:      gameHistory.Chording,
		QuickStart:    gameHistory.QuickStart,
		StartTime:     gameHistory.CreatedAt.Format(time.RFC3339),
		Duration:      gameHistory.GameTime,
		Rating:        gameRating,
		Participants:  participantInfos,
	}

	utils.JSONResponse(w, http.StatusOK, response)
}
