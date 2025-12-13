package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"minesweeperonline/internal/game"
	"minesweeperonline/internal/utils"
)

type RoomHandler struct {
	roomManager *game.RoomManager
}

func NewRoomHandler(roomManager *game.RoomManager) *RoomHandler {
	return &RoomHandler{roomManager: roomManager}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Name       string `json:"name"`
		Password   string `json:"password"`
		Rows       int    `json:"rows"`
		Cols       int    `json:"cols"`
		Mines      int    `json:"mines"`
		GameMode   string `json:"gameMode"`
		QuickStart bool   `json:"quickStart"`
		Chording   bool   `json:"chording"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateRoomParams(req.Name, req.Rows, req.Cols, req.Mines); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Получаем creatorID из контекста (если пользователь авторизован)
	creatorID := 0
	if userID := r.Context().Value("userID"); userID != nil {
		if id, ok := userID.(int); ok {
			creatorID = id
		}
	}

	// Валидация gameMode
	gameMode := req.GameMode
	if gameMode != "classic" && gameMode != "training" && gameMode != "fair" {
		gameMode = "classic" // По умолчанию
	}

	room := h.roomManager.CreateRoom(req.Name, req.Password, req.Rows, req.Cols, req.Mines, creatorID, gameMode, req.QuickStart, req.Chording)
	log.Printf("Создана комната: %s (ID: %s, CreatorID: %d, GameMode: %s, QuickStart: %v, Chording: %v)", req.Name, room.ID, creatorID, gameMode, req.QuickStart, req.Chording)
	utils.JSONResponse(w, http.StatusOK, room.ToResponse())
}

func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms := h.roomManager.GetRoomsList()
	utils.JSONResponse(w, http.StatusOK, rooms)
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		RoomID   string `json:"roomId"`
		Password string `json:"password"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	room := h.roomManager.GetRoom(req.RoomID)
	if room == nil {
		utils.JSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	if !room.ValidatePassword(req.Password) {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	utils.JSONResponse(w, http.StatusOK, room.ToResponse())
}

func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		utils.JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем userID из контекста (требуется авторизация)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Получаем roomID из URL
	vars := mux.Vars(r)
	roomID := vars["id"]
	if roomID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Room ID required")
		return
	}

	// Используем map для проверки, было ли поле password передано
	var reqMap map[string]interface{}
	if err := utils.DecodeJSON(r, &reqMap); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Извлекаем значения из map
	name, _ := reqMap["name"].(string)
	rowsFloat, _ := reqMap["rows"].(float64)
	colsFloat, _ := reqMap["cols"].(float64)
	minesFloat, _ := reqMap["mines"].(float64)
	rows := int(rowsFloat)
	cols := int(colsFloat)
	mines := int(minesFloat)
	gameMode := "classic"
	if gameModeVal, exists := reqMap["gameMode"]; exists {
		if gameModeStr, ok := gameModeVal.(string); ok {
			if gameModeStr == "classic" || gameModeStr == "training" || gameModeStr == "fair" {
				gameMode = gameModeStr
			}
		}
	}
	
	// Извлекаем quickStart (по умолчанию false, если не указан)
	quickStart := false
	if quickStartVal, exists := reqMap["quickStart"]; exists {
		if quickStartBool, ok := quickStartVal.(bool); ok {
			quickStart = quickStartBool
		}
	}
	
	// Извлекаем chording (по умолчанию false, если не указан)
	chording := false
	if chordingVal, exists := reqMap["chording"]; exists {
		if chordingBool, ok := chordingVal.(bool); ok {
			chording = chordingBool
		}
	}

	// Проверяем, было ли передано поле password
	passwordProvided := false
	password := ""
	if pwd, exists := reqMap["password"]; exists {
		passwordProvided = true
		if pwdStr, ok := pwd.(string); ok {
			password = pwdStr
		}
	}

	if err := utils.ValidateRoomParams(name, rows, cols, mines); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Проверяем, что комната существует и пользователь является создателем
	room := h.roomManager.GetRoom(roomID)
	if room == nil {
		utils.JSONError(w, http.StatusNotFound, "Room not found")
		return
	}

	// Проверяем, является ли пользователь создателем комнаты
	isCreator := room.IsCreator(userID)

	if !isCreator {
		utils.JSONError(w, http.StatusForbidden, "Only room creator can update room settings")
		return
	}

	// Обрабатываем пароль
	if !passwordProvided {
		password = "__KEEP__"
	}

	// Обновляем комнату
	if err := h.roomManager.UpdateRoom(roomID, name, password, rows, cols, mines, gameMode, quickStart, chording); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Получаем обновленную комнату
	room = h.roomManager.GetRoom(roomID)
	utils.JSONResponse(w, http.StatusOK, room.ToResponse())
}

