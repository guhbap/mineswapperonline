import type { Ref } from 'vue'

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
        this.reconnectAttempts = 0
        this.lastPongTime = Date.now()
        this.isIntentionallyDisconnected = false // Сбрасываем флаг при успешном подключении
        this.startPingInterval()
        this.onOpen?.()
      }

      this.ws.onmessage = async (event) => {
        try {
          // Проверяем, является ли это бинарным сообщением
          if (event.data instanceof ArrayBuffer) {
            const buffer = event.data
            const messageType = new Uint8Array(buffer)[0]
            
            // Тип 0 = gameState binary
            if (messageType === 0) {
              const { decodeGameStateBinary } = await import('../utils/gamestateBinary')
              const gameState = decodeGameStateBinary(buffer.slice(1)) // Пропускаем первый байт (тип)
              this.onMessage({
                type: 'gameState',
                gameState
              } as WebSocketMessage)
              return
            }
            // Другие бинарные сообщения (например, pong) игнорируем
            return
          }

          // Обрабатываем JSON сообщения
          if (event.data instanceof Blob) {
            // Blob может быть бинарным сообщением
            const arrayBuffer = await event.data.arrayBuffer()
            const messageType = new Uint8Array(arrayBuffer)[0]
            
            if (messageType === 0) {
              const { decodeGameStateBinary } = await import('../utils/gamestateBinary')
              const gameState = decodeGameStateBinary(arrayBuffer.slice(1))
              this.onMessage({
                type: 'gameState',
                gameState
              } as WebSocketMessage)
            }
            return
          }

          const msg: WebSocketMessage = JSON.parse(event.data)

          // Обрабатываем pong сообщение
          if (msg.type === 'pong') {
            this.lastPongTime = Date.now()
            if (this.pongTimeout) {
              clearTimeout(this.pongTimeout)
              this.pongTimeout = null
            }
            return
          }

          this.onMessage(msg)
        } catch (error) {
          // Если не JSON, возможно это бинарное сообщение
          // Игнорируем ошибку парсинга для бинарных сообщений
          console.warn('Ошибка обработки сообщения:', error)
        }
      }

      this.ws.onclose = () => {
        this.stopPingInterval()
        this.onClose?.()
        // Переподключаемся только если отключение было не намеренным
        if (!this.isIntentionallyDisconnected) {
          this.attemptReconnect()
        }
      }

      this.ws.onerror = (error) => {
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

  send(message: WebSocketMessage) {
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
      this.ws.send(JSON.stringify(optimizedMessage))
    }
  }

  sendNickname(nickname: string) {
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
    this.send({ type: 'cellClick', cellClick: { row, col, flag } })
  }

  sendHint(row: number, col: number) {
    this.send({ type: 'hint', hint: { row, col } })
  }

  sendNewGame() {
    this.send({ type: 'newGame' })
  }

  sendChatMessage(text: string) {
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

