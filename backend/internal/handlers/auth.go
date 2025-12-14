package handlers

import (
	"errors"
	"log"
	mathrand "math/rand"
	"net/http"
	"time"

	"minesweeperonline/internal/auth"
	"minesweeperonline/internal/config"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/utils"

	"gorm.io/gorm"
)

type AuthHandler struct {
	db             *database.DB
	profileHandler *ProfileHandler
	config         *config.Config
}

func NewAuthHandler(db *database.DB, profileHandler *ProfileHandler, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:             db,
		profileHandler: profileHandler,
		config:         cfg,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.RegisterRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateRegisterParams(req.Username, req.Email, req.Password); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Проверка существования пользователя
	if h.userExists(req.Username, req.Email) {
		utils.JSONError(w, http.StatusConflict, "Username or email already exists")
		return
	}

	// Создание пользователя
	user, err := h.createUser(req.Username, req.Email, req.Password)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Генерация токена
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.JSONResponse(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  h.userToMap(user),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.LoginRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateAuthParams(req.Username, req.Password); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Поиск и проверка пользователя
	user, err := h.findUserByUsername(req.Username)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Генерация токена
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.JSONResponse(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  h.userToMap(user),
	})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	user, err := h.findUserByID(userID)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Рассчитываем рейтинг динамически
	if h.profileHandler != nil {
		userRating := h.profileHandler.calculateUserRating(userID, 100)
		user.Rating = userRating
	}

	// Проверяем, является ли пользователь администратором
	type UserResponse struct {
		models.User
		IsAdmin bool `json:"isAdmin"`
	}

	response := UserResponse{
		User:    user,
		IsAdmin: h.config.AdminEmail != "" && user.Email == h.config.AdminEmail,
	}

	utils.JSONResponse(w, http.StatusOK, response)
}

// Вспомогательные методы

func (h *AuthHandler) userExists(username, email string) bool {
	var count int64
	h.db.Model(&models.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count)
	return count > 0
}

func (h *AuthHandler) createUser(username, email, password string) (models.User, error) {
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	// Цвета по умолчанию для новых пользователей
	defaultColors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8",
		"#F7DC6F", "#BB8FCE", "#85C1E2", "#F8B739", "#52BE80",
		"#E74C3C", "#3498DB", "#9B59B6", "#1ABC9C", "#F39C12",
		"#E67E22", "#95A5A6", "#34495E", "#16A085", "#27AE60",
	}

	// Выбираем случайный цвет из списка
	mathrand.Seed(time.Now().UnixNano())
	defaultColor := defaultColors[mathrand.Intn(len(defaultColors))]

	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Color:        &defaultColor,
		CreatedAt:    time.Now(),
	}

	err = h.db.Create(&user).Error
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (h *AuthHandler) findUserByUsername(username string) (models.User, error) {
	var user models.User
	err := h.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}
	return user, err
}

func (h *AuthHandler) findUserByID(id int) (models.User, error) {
	var user models.User
	err := h.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}

	return user, err
}

func (h *AuthHandler) findUserByEmail(email string) (models.User, error) {
	var user models.User
	err := h.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}
	return user, err
}

// userToMap преобразует User в map с добавлением isAdmin
func (h *AuthHandler) userToMap(user models.User) map[string]interface{} {
	isAdmin := h.config.AdminEmail != "" && user.Email == h.config.AdminEmail
	userMap := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"rating":    user.Rating,
		"createdAt": user.CreatedAt,
		"isAdmin":   isAdmin,
	}
	if user.Color != nil {
		userMap["color"] = *user.Color
	}
	return userMap
}

// ResetPasswordByAdmin позволяет администратору сбросить пароль пользователя
func (h *AuthHandler) ResetPasswordByAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Проверяем, что запрос от администратора
	userID := r.Context().Value("userID").(int)
	adminUser, err := h.findUserByID(userID)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверяем, что пользователь - администратор
	if h.config.AdminEmail == "" || adminUser.Email != h.config.AdminEmail {
		utils.JSONError(w, http.StatusForbidden, "Only admin can reset passwords")
		return
	}

	var req struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация нового пароля
	if req.NewPassword == "" || len(req.NewPassword) < 6 {
		utils.JSONError(w, http.StatusBadRequest, "New password must be at least 6 characters")
		return
	}

	// Находим пользователя по username или email
	var targetUser models.User
	if req.Username != "" {
		targetUser, err = h.findUserByUsername(req.Username)
	} else if req.Email != "" {
		targetUser, err = h.findUserByEmail(req.Email)
	} else {
		utils.JSONError(w, http.StatusBadRequest, "Username or email is required")
		return
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.JSONError(w, http.StatusNotFound, "User not found")
		} else {
			log.Printf("Error finding user: %v", err)
			utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Хешируем новый пароль
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Обновляем пароль
	err = h.db.Model(&models.User{}).
		Where("id = ?", targetUser.ID).
		Update("password_hash", newPasswordHash).Error
	if err != nil {
		log.Printf("Error updating password: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	log.Printf("Admin %s reset password for user %s (ID: %d)", adminUser.Username, targetUser.Username, targetUser.ID)

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Password reset successfully",
	})
}

// RequestPasswordReset генерирует токен для сброса пароля (без отправки email)
// В реальном приложении здесь должна быть отправка email с токеном
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		utils.JSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Проверяем, существует ли пользователь с таким email
	user, err := h.findUserByEmail(req.Email)
	if err != nil {
		// Для безопасности не сообщаем, существует ли пользователь
		utils.JSONResponse(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Генерируем токен сброса пароля (действителен 1 час)
	resetToken, err := auth.GeneratePasswordResetToken(user.ID, user.Email)
	if err != nil {
		log.Printf("Error generating reset token: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// В реальном приложении здесь должна быть отправка email с токеном
	// Для разработки возвращаем токен в ответе (в продакшене это небезопасно!)
	log.Printf("Password reset token for user %s (email: %s): %s", user.Username, user.Email, resetToken)

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"message":    "Password reset token generated",
		"resetToken": resetToken, // В продакшене убрать это поле!
		"expiresIn":  "1 hour",
		"note":       "In production, this token would be sent via email",
	})
}

// ResetPasswordByToken сбрасывает пароль по токену
func (h *AuthHandler) ResetPasswordByToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		utils.JSONError(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	if len(req.NewPassword) < 6 {
		utils.JSONError(w, http.StatusBadRequest, "New password must be at least 6 characters")
		return
	}

	// Валидируем токен
	claims, err := auth.ValidatePasswordResetToken(req.Token)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid or expired reset token")
		return
	}

	// Находим пользователя
	user, err := h.findUserByID(claims.UserID)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Хешируем новый пароль
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Обновляем пароль
	err = h.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("password_hash", newPasswordHash).Error
	if err != nil {
		log.Printf("Error updating password: %v", err)
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	log.Printf("Password reset by token for user %s (ID: %d)", user.Username, user.ID)

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Password reset successfully",
	})
}
