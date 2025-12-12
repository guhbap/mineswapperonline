package handlers

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"time"

	"minesweeperonline/internal/auth"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/utils"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *database.DB
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{db: db}
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
		User:  user,
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
		User:  user,
	})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	user, err := h.findUserByID(userID)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.JSONResponse(w, http.StatusOK, user)
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
	rand.Seed(time.Now().UnixNano())
	defaultColor := defaultColors[rand.Intn(len(defaultColors))]

	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Color:        &defaultColor,
		Rating:       0.0,
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
	// Убеждаемся, что рейтинг не nil
	if user.Rating < 0 {
		user.Rating = 0.0
	}
	return user, err
}
