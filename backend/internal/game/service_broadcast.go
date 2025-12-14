package game

import (
	"log"

	gorillaWS "github.com/gorilla/websocket"
)

// BroadcastGameState отправляет состояние игры всем игрокам
func (s *Service) BroadcastGameState(room *Room) {
	binaryData, err := EncodeGameStateProtobuf(room.GameState)
	if err != nil {
		log.Printf("[WS OUT] Ошибка кодирования gameState: %v", err)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	log.Printf("[WS OUT] BroadcastGameState: отправка всем игрокам (количество=%d), размер=%d байт", len(playerIDs), len(binaryData))

	for _, id := range playerIDs {
		wsPlayer := s.wsManager.GetWSPlayer(id)
		if wsPlayer != nil {
			mu := wsPlayer.GetMu()
			conn := wsPlayer.GetConn()
			if conn != nil {
				if wsConn, ok := conn.(*gorillaWS.Conn); ok {
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Lock()
					}
					if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
						log.Printf("[WS OUT] Ошибка отправки gameState игроку %s: %v", id, err)
					} else {
						log.Printf("[WS OUT] Игрок %s: отправлен gameState, размер=%d байт", id, len(binaryData))
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				}
			} else {
				log.Printf("[WS OUT] Игрок %s: соединение nil, пропуск отправки gameState", id)
			}
		} else {
			log.Printf("[WS OUT] Игрок %s: wsPlayer не найден, пропуск отправки gameState", id)
		}
	}
}

// BroadcastCellUpdates отправляет обновления клеток всем игрокам
func (s *Service) BroadcastCellUpdates(room *Room, changedCells map[[2]int]bool, gameOver, gameWon bool, revealed, hintsUsed int, loserPlayerID, loserNickname string) {
	if len(changedCells) == 0 && !gameOver && !gameWon {
		log.Printf("[WS OUT] BroadcastCellUpdates: пропуск (changedCells=%d, gameOver=%v, gameWon=%v)", len(changedCells), gameOver, gameWon)
		return
	}

	log.Printf("[WS OUT] BroadcastCellUpdates: отправка обновлений (changedCells=%d, gameOver=%v, gameWon=%v, revealed=%d)", len(changedCells), gameOver, gameWon, revealed)
	updates := CollectCellUpdates(room, changedCells)
	log.Printf("[WS OUT] BroadcastCellUpdates: собрано обновлений клеток: %d", len(updates))
	binaryData, err := EncodeCellUpdateProtobuf(updates, gameOver, gameWon, revealed, hintsUsed, loserPlayerID, loserNickname)
	if err != nil {
		log.Printf("[WS OUT] Ошибка кодирования обновлений клеток: %v", err)
		s.BroadcastGameState(room)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	log.Printf("[WS OUT] BroadcastCellUpdates: отправка всем игрокам (количество=%d), размер=%d байт", len(playerIDs), len(binaryData))

	for _, id := range playerIDs {
		log.Printf("[WS OUT] BroadcastCellUpdates: поиск wsPlayer для id=%s", id)
		wsPlayer := s.wsManager.GetWSPlayer(id)
		if wsPlayer != nil {
			log.Printf("[WS OUT] BroadcastCellUpdates: wsPlayer найден для id=%s", id)
			mu := wsPlayer.GetMu()
			conn := wsPlayer.GetConn()
			if conn != nil {
				if wsConn, ok := conn.(*gorillaWS.Conn); ok {
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Lock()
					}
					if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
						log.Printf("[WS OUT] Ошибка отправки cellUpdate игроку %s: %v", id, err)
					} else {
						log.Printf("[WS OUT] Игрок %s: отправлен cellUpdate, размер=%d байт, обновлений=%d", id, len(binaryData), len(updates))
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				} else {
					log.Printf("[WS OUT] Игрок %s: соединение не является *gorillaWS.Conn, пропуск", id)
				}
			} else {
				log.Printf("[WS OUT] Игрок %s: соединение nil, пропуск отправки cellUpdate", id)
			}
		} else {
			log.Printf("[WS OUT] Игрок %s: wsPlayer не найден в wsPlayers, пропуск отправки cellUpdate", id)
		}
	}
}

// BroadcastToAll отправляет сообщение всем игрокам
func (s *Service) BroadcastToAll(room *Room, msg Message) {
	var binaryData []byte
	var err error
	if msg.Type == "chat" && msg.Chat != nil {
		binaryData, err = EncodeChatProtobuf(&msg)
		if err != nil {
			log.Printf("[WS OUT] Ошибка кодирования чата: %v", err)
			return
		}
	} else {
		log.Printf("[WS OUT] BroadcastToAll: неизвестный тип сообщения или пустой чат: type=%s", msg.Type)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	log.Printf("[WS OUT] BroadcastToAll (chat): отправка всем игрокам (количество=%d), размер=%d байт, текст=%s", len(playerIDs), len(binaryData), msg.Chat.Text)

	for _, id := range playerIDs {
		wsPlayer := s.wsManager.GetWSPlayer(id)
		if wsPlayer != nil {
			mu := wsPlayer.GetMu()
			conn := wsPlayer.GetConn()
			if conn != nil {
				if wsConn, ok := conn.(*gorillaWS.Conn); ok {
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Lock()
					}
					if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
						log.Printf("[WS OUT] Ошибка отправки chat игроку %s: %v", id, err)
					} else {
						log.Printf("[WS OUT] Игрок %s: отправлен chat, размер=%d байт", id, len(binaryData))
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				}
			}
		}
	}
}

// BroadcastToOthers отправляет сообщение всем игрокам кроме отправителя
func (s *Service) BroadcastToOthers(room *Room, senderID string, msg Message) {
	if room.GetPlayerCount() <= 1 {
		return
	}

	var binaryData []byte
	var err error
	if msg.Type == "cursor" && msg.Cursor != nil {
		binaryData, err = EncodeCursorProtobuf(&msg)
		if err != nil {
			log.Printf("[WS OUT] Ошибка кодирования курсора: %v", err)
			return
		}
	} else {
		log.Printf("[WS OUT] BroadcastToOthers: неизвестный тип сообщения: type=%s", msg.Type)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		if id != senderID {
			playerIDs = append(playerIDs, id)
		}
	}
	room.Mu.RUnlock()

	// Курсор логируем реже, чтобы не засорять логи
	// log.Printf("[WS OUT] BroadcastToOthers (cursor): отправка игрокам (количество=%d), размер=%d байт", len(playerIDs), len(binaryData))

	for _, id := range playerIDs {
		wsPlayer := s.wsManager.GetWSPlayer(id)
		if wsPlayer != nil {
			mu := wsPlayer.GetMu()
			conn := wsPlayer.GetConn()
			if conn != nil {
				if wsConn, ok := conn.(*gorillaWS.Conn); ok {
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Lock()
					}
					if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
						log.Printf("[WS OUT] Ошибка отправки cursor игроку %s: %v", id, err)
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				}
			}
		}
	}
}

// BroadcastPlayerList отправляет список игроков всем игрокам
func (s *Service) BroadcastPlayerList(room *Room) {
	room.Mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, player := range room.Players {
		playersList = append(playersList, map[string]string{
			"id":       player.ID,
			"nickname": player.Nickname,
			"color":    player.Color,
		})
	}
	room.Mu.RUnlock()

	binaryData, err := EncodePlayersProtobuf(playersList)
	if err != nil {
		log.Printf("[WS OUT] Ошибка кодирования списка игроков: %v", err)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

	log.Printf("[WS OUT] BroadcastPlayerList: отправка всем игрокам (количество=%d), размер=%d байт, игроков в списке=%d", len(playerIDs), len(binaryData), len(playersList))

	for _, id := range playerIDs {
		wsPlayer := s.wsManager.GetWSPlayer(id)
		if wsPlayer != nil {
			mu := wsPlayer.GetMu()
			conn := wsPlayer.GetConn()
			if conn != nil {
				if wsConn, ok := conn.(*gorillaWS.Conn); ok {
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Lock()
					}
					if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
						log.Printf("[WS OUT] Ошибка отправки players игроку %s: %v", id, err)
					} else {
						log.Printf("[WS OUT] Игрок %s: отправлен players, размер=%d байт", id, len(binaryData))
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				}
			}
		}
	}
}

// SendGameStateToPlayer отправляет состояние игры конкретному игроку
func (s *Service) SendGameStateToPlayer(room *Room, player WSPlayer) {
	binaryData, err := EncodeGameStateProtobuf(room.GameState)
	if err != nil {
		log.Printf("[WS OUT] Ошибка кодирования gameState: %v", err)
		return
	}

	mu := player.GetMu()
	conn := player.GetConn()
	if conn != nil {
		if wsConn, ok := conn.(*gorillaWS.Conn); ok {
			if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
				muVal.Lock()
				defer muVal.Unlock()
			}
			if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
				log.Printf("[WS OUT] Ошибка отправки gameState игроку: %v", err)
			} else {
				log.Printf("[WS OUT] Отправлен gameState игроку, размер=%d байт", len(binaryData))
			}
		}
	} else {
		log.Printf("[WS OUT] Соединение nil, пропуск отправки gameState")
	}
}

// SendPlayerListToPlayer отправляет список игроков конкретному игроку
func (s *Service) SendPlayerListToPlayer(room *Room, player WSPlayer) {
	room.Mu.RLock()
	playersList := make([]map[string]string, 0, len(room.Players))
	for _, p := range room.Players {
		playersList = append(playersList, map[string]string{
			"id":       p.ID,
			"nickname": p.Nickname,
			"color":    p.Color,
		})
	}
	room.Mu.RUnlock()

	binaryData, err := EncodePlayersProtobuf(playersList)
	if err != nil {
		log.Printf("[WS OUT] Ошибка кодирования списка игроков: %v", err)
		return
	}

	mu := player.GetMu()
	conn := player.GetConn()
	if conn != nil {
		if wsConn, ok := conn.(*gorillaWS.Conn); ok {
			if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
				muVal.Lock()
				defer muVal.Unlock()
			}
			if err := wsConn.WriteMessage(gorillaWS.BinaryMessage, binaryData); err != nil {
				log.Printf("[WS OUT] Ошибка отправки players игроку: %v", err)
			} else {
				log.Printf("[WS OUT] Отправлен players игроку, размер=%d байт, игроков в списке=%d", len(binaryData), len(playersList))
			}
		}
	} else {
		log.Printf("[WS OUT] Соединение nil, пропуск отправки players")
	}
}

