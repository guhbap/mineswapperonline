package handlers

import (
	"net/http"

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
		Name     string `json:"name"`
		Password string `json:"password"`
		Rows     int    `json:"rows"`
		Cols     int    `json:"cols"`
		Mines    int    `json:"mines"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := utils.ValidateRoomParams(req.Name, req.Rows, req.Cols, req.Mines); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	room := h.roomManager.CreateRoom(req.Name, req.Password, req.Rows, req.Cols, req.Mines)
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

