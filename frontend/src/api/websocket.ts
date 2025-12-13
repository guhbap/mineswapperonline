import type { Ref } from 'vue'
import { decodeProtobufMessage, encodeClientMessage } from '../utils/protobufMessages'

export interface WebSocketMessage {
  type: string
  playerId?: string
  nickname?: string
  color?: string
  cursor?: {
    pid: string // playerId сокращено до pid
    x: number
    y: number
  }
  cellClick?: {
    row: number
    col: number
    flag: boolean
  }
  hint?: {
    row: number
    col: number
  }
  gameState?: {
    b: Cell[][] // board
    r: number // rows
    c: number // cols
    m: number // mines
    go: boolean // gameOver
    gw: boolean // gameWon
    rv: number // revealed
    hu: number // hintsUsed
    sc?: Array<{ r: number; c: number }> // safeCells - безопасные ячейки для режима без угадываний
    hints?: Array<{ r: number; c: number; t: string }> // cellHints - подсказки для ячеек (MINE, SAFE, UNKNOWN)
    lpid?: string // loserPlayerId
    ln?: string // loserNickname
  }
  players?: Array<{
    id: string
    nickname: string
    color: string
  }>
  chat?: {
    text: string
    isSystem?: boolean
    action?: string
    row?: number
    col?: number
  }
  error?: string
  cellUpdates?: Array<{
    row: number
    col: number
    type: number
  }>
  gameOver?: boolean
  gameWon?: boolean
  revealed?: number
  hintsUsed?: number
  loserPlayerId?: string
  loserNickname?: string
}

export interface Cell {
  m: boolean // isMine
  r: boolean // isRevealed
  f: boolean // isFlagged
  n: number // neighborMines
  fc?: string // flagColor - цвет игрока, который поставил флаг
}

export interface IWebSocketClient {
  connect(): void
  send(message: WebSocketMessage): void
  sendNickname(nickname: string): void
  sendCursor(x: number, y: number): void
  sendCellClick(row: number, col: number, flag: boolean): void
  sendHint(row: number, col: number): void
  sendNewGame(): void
  disconnect(): void
  isConnected(): boolean
}

export class WebSocketClient implements IWebSocketClient {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private cursorThrottleTimer: ReturnType<typeof setTimeout> | null = null
  private lastCursorPosition: { x: number; y: number } | null = null
  private pendingCursorPosition: { x: number; y: number } | null = null
  private cursorThrottleDelay = 100 // Отправляем позицию курсора каждые 100ms
  private lastCursorSendTime = 0
  private pingInterval: ReturnType<typeof setInterval> | null = null
  private pingIntervalDelay = 30000 // Отправляем ping каждые 30 секунд
  private lastPongTime = 0
  private pongTimeout: ReturnType<typeof setTimeout> | null = null
  private isIntentionallyDisconnected = false

  constructor(
    private url: string,
    private onMessage: (msg: WebSocketMessage) => void,
    private onOpen?: () => void,
    private onClose?: () => void,
    private onError?: (error: Event) => void
  ) {}

  connect() {
    // Сбрасываем флаг намеренного отключения при попытке подключения
    this.isIntentionallyDisconnected = false

    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        const timestamp = new Date().toISOString()
        console.log(`[WS CONN ${timestamp}] WebSocket подключен:`, this.url)
        this.reconnectAttempts = 0
        this.lastPongTime = Date.now()
        this.isIntentionallyDisconnected = false // Сбрасываем флаг при успешном подключении
        this.startPingInterval()
        this.onOpen?.()
      }

      this.ws.onmessage = async (event) => {
        try {
          let buffer: ArrayBuffer
          const timestamp = new Date().toISOString()

          // Преобразуем данные в ArrayBuffer
          if (event.data instanceof ArrayBuffer) {
            buffer = event.data
            console.log(`[WS RECV ${timestamp}] Бинарное сообщение (ArrayBuffer):`, {
              size: buffer.byteLength,
              bytes: Array.from(new Uint8Array(buffer.slice(0, Math.min(20, buffer.byteLength))))
            })
          } else if (event.data instanceof Blob) {
            buffer = await event.data.arrayBuffer()
            console.log(`[WS RECV ${timestamp}] Бинарное сообщение (Blob):`, {
              size: buffer.byteLength,
              bytes: Array.from(new Uint8Array(buffer.slice(0, Math.min(20, buffer.byteLength))))
            })
          } else {
            // Все сообщения теперь должны быть в protobuf формате
            // Если пришло текстовое сообщение, пытаемся декодировать как protobuf
            console.warn(`[WS RECV ${timestamp}] Получено текстовое сообщение, ожидается бинарный protobuf формат`)
            // Пытаемся декодировать как protobuf (на случай если это base64 или другой формат)
            try {
              const textData = event.data as string
              // Если это JSON (для обратной совместимости), парсим его
              if (textData.startsWith('{')) {
                const msg: WebSocketMessage = JSON.parse(textData)
                console.log(`[WS RECV ${timestamp}] JSON сообщение (fallback):`, {
                  type: msg.type,
                  data: msg
                })
                this.onMessage(msg)
                return
              }
            } catch (error) {
              console.error(`[WS RECV ${timestamp}] Ошибка обработки текстового сообщения:`, error)
            }
            return
          }

          if (buffer.byteLength === 0) {
            console.warn(`[WS RECV ${timestamp}] Пустое бинарное сообщение`)
            return
          }

          console.log(`[WS RECV ${timestamp}] Бинарное сообщение (размер: ${buffer.byteLength} байт)`)

          // Все сообщения теперь в protobuf формате
          let decodedMsg: any = null
          try {
            decodedMsg = await decodeProtobufMessage(buffer)
          } catch (error) {
            console.error(`[WS RECV ${timestamp}] Ошибка декодирования protobuf сообщения:`, error)
            // Логируем первые байты для отладки
            const firstBytes = Array.from(new Uint8Array(buffer.slice(0, Math.min(10, buffer.byteLength))))
            console.error(`[WS RECV ${timestamp}] Первые байты сообщения:`, firstBytes)
          }

          if (decodedMsg) {
            // Дополнительное логирование для chat сообщений
            if (decodedMsg.type === 'chat') {
              console.log(`[WS RECV ${timestamp}] Детали chat сообщения:`, {
                playerId: decodedMsg.playerId,
                nickname: decodedMsg.nickname,
                color: decodedMsg.color,
                text: decodedMsg.chat?.text,
                isSystem: decodedMsg.chat?.isSystem,
                action: decodedMsg.chat?.action,
                row: decodedMsg.chat?.row,
                col: decodedMsg.chat?.col
              })
            }
            // Обрабатываем pong сообщение
            if (decodedMsg.type === 'pong') {
              console.log(`[WS RECV ${timestamp}] pong`)
              this.lastPongTime = Date.now()
              if (this.pongTimeout) {
                clearTimeout(this.pongTimeout)
                this.pongTimeout = null
              }
              return
            }

            // Преобразуем DecodedMessage в WebSocketMessage
            const wsMsg: WebSocketMessage = {
              type: decodedMsg.type,
              ...(decodedMsg.playerId ? { playerId: decodedMsg.playerId } : {}),
              ...(decodedMsg.nickname ? { nickname: decodedMsg.nickname } : {}),
              ...(decodedMsg.color ? { color: decodedMsg.color } : {}),
              ...(decodedMsg.cursor ? { cursor: decodedMsg.cursor } : {}),
              ...(decodedMsg.players ? { players: decodedMsg.players } : {}),
              ...(decodedMsg.chat ? { chat: decodedMsg.chat } : {}),
              ...(decodedMsg.error ? { error: decodedMsg.error } : {}),
              ...(decodedMsg.gameState ? { gameState: decodedMsg.gameState } : {}),
              ...(decodedMsg.cellUpdates ? { cellUpdates: decodedMsg.cellUpdates } : {}),
              ...(decodedMsg.gameOver !== undefined ? { gameOver: decodedMsg.gameOver } : {}),
              ...(decodedMsg.gameWon !== undefined ? { gameWon: decodedMsg.gameWon } : {}),
              ...(decodedMsg.revealed !== undefined ? { revealed: decodedMsg.revealed } : {}),
              ...(decodedMsg.hintsUsed !== undefined ? { hintsUsed: decodedMsg.hintsUsed } : {}),
              ...(decodedMsg.loserPlayerId ? { loserPlayerId: decodedMsg.loserPlayerId } : {}),
              ...(decodedMsg.loserNickname ? { loserNickname: decodedMsg.loserNickname } : {})
            }

            console.log(`[WS RECV ${timestamp}] Декодированное бинарное сообщение:`, {
              type: wsMsg.type,
              data: wsMsg,
              cellUpdatesCount: wsMsg.cellUpdates?.length || 0,
              playersCount: wsMsg.players?.length || 0
            })

            this.onMessage(wsMsg)
          } else {
            console.warn(`[WS RECV ${timestamp}] Не удалось декодировать protobuf сообщение`)
          }
        } catch (error) {
          // Игнорируем ошибку парсинга для бинарных сообщений
          console.error(`[WS RECV] Ошибка обработки сообщения:`, error)
        }
      }

      this.ws.onclose = (event) => {
        const timestamp = new Date().toISOString()
        console.log(`[WS CONN ${timestamp}] WebSocket закрыт:`, {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
          intentionallyDisconnected: this.isIntentionallyDisconnected
        })
        this.stopPingInterval()
        this.onClose?.()
        // Переподключаемся только если отключение было не намеренным
        if (!this.isIntentionallyDisconnected) {
          this.attemptReconnect()
        }
      }

      this.ws.onerror = (error) => {
        const timestamp = new Date().toISOString()
        console.error(`[WS ERROR ${timestamp}] Ошибка WebSocket:`, error)
        this.onError?.(error)
      }
    } catch (error) {
      console.error('Ошибка подключения WebSocket:', error)
      // Переподключаемся только если отключение было не намеренным
      if (!this.isIntentionallyDisconnected) {
        this.attemptReconnect()
      }
    }
  }

  private attemptReconnect() {
    // Не переподключаемся, если отключение было намеренным
    if (this.isIntentionallyDisconnected) {
      return
    }

    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      setTimeout(() => {
        // Проверяем флаг еще раз перед переподключением
        if (!this.isIntentionallyDisconnected) {
          console.log(`Попытка переподключения ${this.reconnectAttempts}...`)
          this.connect()
        }
      }, this.reconnectDelay * this.reconnectAttempts)
    }
  }

  private truncatePlayerId(playerId: string | undefined): string | undefined {
    if (!playerId) return playerId
    return playerId.length > 5 ? playerId.substring(0, 5) : playerId
  }

  async send(message: WebSocketMessage) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      // Ограничиваем playerId до 5 символов при отправке
      const optimizedMessage: WebSocketMessage = {
        ...message,
        playerId: this.truncatePlayerId(message.playerId),
        cursor: message.cursor ? {
          ...message.cursor,
          pid: this.truncatePlayerId(message.cursor.pid) || ''
        } : undefined,
        gameState: message.gameState ? {
          ...message.gameState,
          lpid: this.truncatePlayerId(message.gameState.lpid)
        } : undefined
      }
      const timestamp = new Date().toISOString()
      try {
        // Кодируем сообщение в protobuf формат
        const binaryData = await encodeClientMessage(optimizedMessage)
        console.log(`[WS SEND ${timestamp}] Protobuf сообщение:`, {
          type: optimizedMessage.type,
          data: optimizedMessage,
          size: binaryData.byteLength
        })
        this.ws.send(binaryData)
      } catch (error) {
        console.error(`[WS SEND ${timestamp}] Ошибка кодирования protobuf сообщения:`, error)
        // Все сообщения должны отправляться в protobuf формате
        // Если кодирование не удалось, не отправляем сообщение
        throw error
      }
    } else {
      console.warn(`[WS SEND] Попытка отправить сообщение при закрытом соединении. Тип: ${message.type}`)
    }
  }

  sendNickname(nickname: string) {
    console.log(`[WS SEND] Отправка nickname:`, nickname)
    this.send({ type: 'nickname', nickname })
  }

  sendCursor(x: number, y: number) {
    const now = Date.now()

    // Округляем координаты до 2 знаков после запятой для оптимизации
    const roundedX = Math.round(x * 100) / 100
    const roundedY = Math.round(y * 100) / 100

    // Сохраняем последнюю позицию (округленную)
    this.pendingCursorPosition = { x: roundedX, y: roundedY }

    // Проверяем, изменилась ли позиция значительно (минимум 5px)
    if (this.lastCursorPosition) {
      const dx = Math.abs(roundedX - this.lastCursorPosition.x)
      const dy = Math.abs(roundedY - this.lastCursorPosition.y)
      if (dx < 5 && dy < 5 && (now - this.lastCursorSendTime) < this.cursorThrottleDelay) {
        return // Позиция не изменилась значительно и не прошло достаточно времени
      }
    }

    // Если прошло достаточно времени с последней отправки, отправляем сразу
    if (now - this.lastCursorSendTime >= this.cursorThrottleDelay) {
      this.lastCursorPosition = { x: roundedX, y: roundedY }
      this.lastCursorSendTime = now
      this.send({ type: 'cursor', cursor: { pid: '', x: roundedX, y: roundedY } })
      this.pendingCursorPosition = null
      return
    }

    // Иначе планируем отправку через throttle
    if (!this.cursorThrottleTimer) {
      const delay = this.cursorThrottleDelay - (now - this.lastCursorSendTime)
      this.cursorThrottleTimer = setTimeout(() => {
        if (this.pendingCursorPosition) {
          this.lastCursorPosition = { ...this.pendingCursorPosition }
          this.lastCursorSendTime = Date.now()
          this.send({
            type: 'cursor',
            cursor: {
              pid: '',
              x: this.pendingCursorPosition.x,
              y: this.pendingCursorPosition.y
            }
          })
          this.pendingCursorPosition = null
        }
        this.cursorThrottleTimer = null
      }, delay)
    }
  }

  sendCellClick(row: number, col: number, flag: boolean) {
    console.log(`[WS SEND] Отправка cellClick:`, { row, col, flag })
    this.send({ type: 'cellClick', cellClick: { row, col, flag } })
  }

  sendHint(row: number, col: number) {
    console.log(`[WS SEND] Отправка hint:`, { row, col })
    this.send({ type: 'hint', hint: { row, col } })
  }

  sendNewGame() {
    console.log(`[WS SEND] Отправка newGame`)
    this.send({ type: 'newGame' })
  }

  sendChatMessage(text: string) {
    console.log(`[WS SEND] Отправка chat:`, text)
    this.send({
      type: 'chat',
      chat: {
        text: text,
        isSystem: false
      }
    })
  }

  private startPingInterval() {
    this.stopPingInterval()

    this.pingInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        // Отправляем ping сообщение
        console.log(`[WS PING] Отправка ping`)
        this.send({ type: 'ping' })

        // Устанавливаем таймаут для ожидания pong
        if (this.pongTimeout) {
          clearTimeout(this.pongTimeout)
        }

        this.pongTimeout = setTimeout(() => {
          // Если не получили pong в течение 10 секунд, считаем соединение разорванным
          const timeSinceLastPong = Date.now() - this.lastPongTime
          if (timeSinceLastPong > 10000) {
            console.warn('Не получен pong от сервера, переподключаемся...')
            if (this.ws) {
              this.ws.close()
            }
          }
        }, 10000)
      }
    }, this.pingIntervalDelay)
  }

  private stopPingInterval() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
    if (this.pongTimeout) {
      clearTimeout(this.pongTimeout)
      this.pongTimeout = null
    }
  }

  disconnect() {
    // Устанавливаем флаг, что отключение намеренное
    this.isIntentionallyDisconnected = true

    this.stopPingInterval()
    if (this.cursorThrottleTimer) {
      clearTimeout(this.cursorThrottleTimer)
      this.cursorThrottleTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.lastCursorPosition = null
    this.pendingCursorPosition = null
    this.lastCursorSendTime = 0
    this.lastPongTime = 0
    this.reconnectAttempts = 0
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

