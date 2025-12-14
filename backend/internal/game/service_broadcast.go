package game

import (
	"log"

	gorillaWS "github.com/gorilla/websocket"
)

// BroadcastGameState отправляет состояние игры всем игрокам
func (s *Service) BroadcastGameState(room *Room) {
	binaryData, err := EncodeGameStateProtobuf(room.GameState)
	if err != nil {
		log.Printf("Ошибка кодирования gameState: %v", err)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

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
						log.Printf("Ошибка отправки состояния игры игроку %s: %v", id, err)
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				}
			}
		}
	}
}

// BroadcastCellUpdates отправляет обновления клеток всем игрокам
func (s *Service) BroadcastCellUpdates(room *Room, changedCells map[[2]int]bool, gameOver, gameWon bool, revealed, hintsUsed int, loserPlayerID, loserNickname string) {
	if len(changedCells) == 0 && !gameOver && !gameWon {
		log.Printf("BroadcastCellUpdates: пропуск (changedCells=%d, gameOver=%v, gameWon=%v)", len(changedCells), gameOver, gameWon)
		return
	}

	log.Printf("BroadcastCellUpdates: отправка обновлений (changedCells=%d, gameOver=%v, gameWon=%v, revealed=%d)", len(changedCells), gameOver, gameWon, revealed)
	updates := CollectCellUpdates(room, changedCells)
	binaryData, err := EncodeCellUpdateProtobuf(updates, gameOver, gameWon, revealed, hintsUsed, loserPlayerID, loserNickname)
	if err != nil {
		log.Printf("Ошибка кодирования обновлений клеток: %v", err)
		s.BroadcastGameState(room)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

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
						log.Printf("Ошибка отправки обновлений клеток игроку %s: %v", id, err)
					}
					if muVal, ok := mu.(interface{ Lock(); Unlock() }); ok {
						muVal.Unlock()
					}
				}
			}
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
			log.Printf("Ошибка кодирования чата: %v", err)
			return
		}
	} else {
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

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
						log.Printf("Ошибка отправки сообщения чата игроку %s: %v", id, err)
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
			log.Printf("Ошибка кодирования курсора: %v", err)
			return
		}
	} else {
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
						log.Printf("Ошибка отправки сообщения игроку %s: %v", id, err)
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
		log.Printf("Ошибка кодирования списка игроков: %v", err)
		return
	}

	room.Mu.RLock()
	playerIDs := make([]string, 0, len(room.Players))
	for id := range room.Players {
		playerIDs = append(playerIDs, id)
	}
	room.Mu.RUnlock()

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
						log.Printf("Ошибка отправки списка игроков: %v", err)
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
		log.Printf("Ошибка кодирования gameState: %v", err)
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
				log.Printf("Ошибка отправки состояния игры: %v", err)
			}
		}
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
		log.Printf("Ошибка кодирования списка игроков: %v", err)
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
				log.Printf("Ошибка отправки списка игроков: %v", err)
			}
		}
	}
}

