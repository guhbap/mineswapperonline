import type { Ref } from 'vue'

export interface WebSocketMessage {
  type: string
  playerId?: string
  nickname?: string
  color?: string
  cursor?: {
    playerId: string
    x: number
    y: number
  }
  cellClick?: {
    row: number
    col: number
    flag: boolean
  }
  gameState?: {
    board: Cell[][]
    rows: number
    cols: number
    mines: number
    gameOver: boolean
    gameWon: boolean
    revealed: number
    loserPlayerId?: string
    loserNickname?: string
  }
  players?: Array<{
    id: string
    nickname: string
    color: string
  }>
}

export interface Cell {
  isMine: boolean
  isRevealed: boolean
  isFlagged: boolean
  neighborMines: number
}

export interface IWebSocketClient {
  connect(): void
  send(message: WebSocketMessage): void
  sendNickname(nickname: string): void
  sendCursor(x: number, y: number): void
  sendCellClick(row: number, col: number, flag: boolean): void
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
  private cursorThrottleDelay = 50 // Отправляем позицию курсора каждые 50ms
  private lastCursorSendTime = 0

  constructor(
    private url: string,
    private onMessage: (msg: WebSocketMessage) => void,
    private onOpen?: () => void,
    private onClose?: () => void,
    private onError?: (error: Event) => void
  ) {}

  connect() {
    try {
      this.ws = new WebSocket(this.url)
      
      this.ws.onopen = () => {
        this.reconnectAttempts = 0
        this.onOpen?.()
      }

      this.ws.onmessage = (event) => {
        try {
          const msg: WebSocketMessage = JSON.parse(event.data)
          this.onMessage(msg)
        } catch (error) {
          console.error('Ошибка парсинга сообщения:', error, event.data)
        }
      }

      this.ws.onclose = () => {
        this.onClose?.()
        this.attemptReconnect()
      }

      this.ws.onerror = (error) => {
        this.onError?.(error)
      }
    } catch (error) {
      console.error('Ошибка подключения WebSocket:', error)
      this.attemptReconnect()
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      setTimeout(() => {
        console.log(`Попытка переподключения ${this.reconnectAttempts}...`)
        this.connect()
      }, this.reconnectDelay * this.reconnectAttempts)
    }
  }

  send(message: WebSocketMessage) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }

  sendNickname(nickname: string) {
    this.send({ type: 'nickname', nickname })
  }

  sendCursor(x: number, y: number) {
    const now = Date.now()
    
    // Сохраняем последнюю позицию
    this.pendingCursorPosition = { x, y }

    // Проверяем, изменилась ли позиция значительно (минимум 3px)
    if (this.lastCursorPosition) {
      const dx = Math.abs(x - this.lastCursorPosition.x)
      const dy = Math.abs(y - this.lastCursorPosition.y)
      if (dx < 3 && dy < 3 && (now - this.lastCursorSendTime) < this.cursorThrottleDelay) {
        return // Позиция не изменилась значительно и не прошло достаточно времени
      }
    }

    // Если прошло достаточно времени с последней отправки, отправляем сразу
    if (now - this.lastCursorSendTime >= this.cursorThrottleDelay) {
      this.lastCursorPosition = { x, y }
      this.lastCursorSendTime = now
      this.send({ type: 'cursor', cursor: { playerId: '', x, y } })
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
              playerId: '', 
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

  sendNewGame() {
    this.send({ type: 'newGame' })
  }

  disconnect() {
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
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

