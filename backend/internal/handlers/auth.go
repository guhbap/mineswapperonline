package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"minesweeperonline/internal/auth"
	"minesweeperonline/internal/database"
	"minesweeperonline/internal/models"
	"minesweeperonline/internal/utils"
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
	var id int
	err := h.db.QueryRow("SELECT id FROM users WHERE username = $1 OR email = $2", username, email).Scan(&id)
	return err == nil
}

func (h *AuthHandler) createUser(username, email, password string) (models.User, error) {
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	var user models.User
	err = h.db.QueryRow(
		"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, username, email, color, created_at",
		username, email, passwordHash,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Color, &user.CreatedAt)

	return user, err
}

func (h *AuthHandler) findUserByUsername(username string) (models.User, error) {
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, email, password_hash, color, created_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Color, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return models.User{}, err
	}
	return user, err
}

func (h *AuthHandler) findUserByID(id int) (models.User, error) {
	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, email, color, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Color, &user.CreatedAt)

	return user, err
}
