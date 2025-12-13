/**
 * Декодирует бинарные сообщения WebSocket
 */

// Типы бинарных сообщений
const MessageTypeGameState = 0
const MessageTypeChat = 1
const MessageTypeCursor = 2
const MessageTypePlayers = 3
const MessageTypePong = 4
const MessageTypeError = 5
const MessageTypeCellUpdate = 6

// Типы клеток
const CellTypeClosed = 0  // Закрыта
const CellTypeMine = 9   // Мина
const CellTypeSafe = 10  // Зеленая (SAFE)
const CellTypeUnknown = 11 // Желтая (UNKNOWN)
const CellTypeDanger = 12 // Красная (MINE)

export interface DecodedMessage {
  type: string
  playerId?: string
  nickname?: string
  color?: string
  cursor?: {
    pid: string
    x: number
    y: number
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

/**
 * Декодирует бинарное сообщение чата
 */
function decodeChatBinary(data: ArrayBuffer): DecodedMessage {
  const view = new DataView(data)
  let offset = 1 // Пропускаем тип сообщения

  // Читаем PlayerID
  const pidLen = view.getUint8(offset)
  offset += 1
  const pidBytes = new Uint8Array(data, offset, 5)
  offset += 5
  const playerId = pidLen > 0 ? new TextDecoder().decode(pidBytes.slice(0, pidLen)) : ''

  // Читаем Nickname
  const nicknameLen = view.getUint8(offset)
  offset += 1
  let nickname = ''
  if (nicknameLen > 0) {
    const nicknameBytes = new Uint8Array(data, offset, nicknameLen)
    offset += nicknameLen
    nickname = new TextDecoder().decode(nicknameBytes)
  }

  // Читаем Color
  const colorLen = view.getUint8(offset)
  offset += 1
  let color = ''
  if (colorLen > 0) {
    const colorBytes = new Uint8Array(data, offset, colorLen)
    offset += colorLen
    color = new TextDecoder().decode(colorBytes)
  }

  // Читаем Text
  const textLen = view.getUint8(offset)
  offset += 1
  let text = ''
  if (textLen > 0) {
    const textBytes = new Uint8Array(data, offset, textLen)
    offset += textLen
    text = new TextDecoder().decode(textBytes)
  }

  // Читаем флаги
  const flags = view.getUint8(offset)
  offset += 1
  const isSystem = (flags & (1 << 0)) !== 0
  const hasAction = (flags & (1 << 1)) !== 0
  const hasRowCol = (flags & (1 << 2)) !== 0

  let action = ''
  if (hasAction) {
    const actionLen = view.getUint8(offset)
    offset += 1
    if (actionLen > 0) {
      const actionBytes = new Uint8Array(data, offset, actionLen)
      offset += actionLen
      action = new TextDecoder().decode(actionBytes)
    }
  }

  let row = -1
  let col = -1
  if (hasRowCol) {
    row = view.getUint16(offset, true)
    offset += 2
    col = view.getUint16(offset, true)
    offset += 2
  }

  return {
    type: 'chat',
    playerId,
    nickname,
    color,
    chat: {
      text,
      isSystem,
      ...(action ? { action } : {}),
      ...(row >= 0 && col >= 0 ? { row, col } : {})
    }
  }
}

/**
 * Декодирует бинарное сообщение курсора
 */
function decodeCursorBinary(data: ArrayBuffer): DecodedMessage {
  const view = new DataView(data)
  let offset = 1 // Пропускаем тип сообщения

  // Читаем PlayerID
  const pidLen = view.getUint8(offset)
  offset += 1
  const pidBytes = new Uint8Array(data, offset, 5)
  offset += 5
  const playerId = pidLen > 0 ? new TextDecoder().decode(pidBytes.slice(0, pidLen)) : ''

  // Читаем Nickname
  const nicknameLen = view.getUint8(offset)
  offset += 1
  let nickname = ''
  if (nicknameLen > 0) {
    const nicknameBytes = new Uint8Array(data, offset, nicknameLen)
    offset += nicknameLen
    nickname = new TextDecoder().decode(nicknameBytes)
  }

  // Читаем Color
  const colorLen = view.getUint8(offset)
  offset += 1
  let color = ''
  if (colorLen > 0) {
    const colorBytes = new Uint8Array(data, offset, colorLen)
    offset += colorLen
    color = new TextDecoder().decode(colorBytes)
  }

  // Читаем X, Y (float64)
  const x = view.getFloat64(offset, true)
  offset += 8
  const y = view.getFloat64(offset, true)
  offset += 8

  return {
    type: 'cursor',
    playerId,
    nickname,
    color,
    cursor: {
      pid: playerId,
      x,
      y
    }
  }
}

/**
 * Декодирует бинарное сообщение списка игроков
 */
function decodePlayersBinary(data: ArrayBuffer): DecodedMessage {
  const view = new DataView(data)
  let offset = 1 // Пропускаем тип сообщения

  const playerCount = view.getUint8(offset)
  offset += 1

  const players: Array<{ id: string; nickname: string; color: string }> = []

  for (let i = 0; i < playerCount; i++) {
    // Читаем ID
    const idLen = view.getUint8(offset)
    offset += 1
    const idBytes = new Uint8Array(data, offset, 5)
    offset += 5
    const id = idLen > 0 ? new TextDecoder().decode(idBytes.slice(0, idLen)) : ''

    // Читаем Nickname
    const nicknameLen = view.getUint8(offset)
    offset += 1
    let nickname = ''
    if (nicknameLen > 0) {
      const nicknameBytes = new Uint8Array(data, offset, nicknameLen)
      offset += nicknameLen
      nickname = new TextDecoder().decode(nicknameBytes)
    }

    // Читаем Color
    const colorLen = view.getUint8(offset)
    offset += 1
    let color = ''
    if (colorLen > 0) {
      const colorBytes = new Uint8Array(data, offset, colorLen)
      offset += colorLen
      color = new TextDecoder().decode(colorBytes)
    }

    players.push({ id, nickname, color })
  }

  return {
    type: 'players',
    players
  }
}

/**
 * Декодирует бинарное сообщение об ошибке
 */
function decodeErrorBinary(data: ArrayBuffer): DecodedMessage {
  const view = new DataView(data)
  let offset = 1 // Пропускаем тип сообщения

  const errorLen = view.getUint8(offset)
  offset += 1
  let error = ''
  if (errorLen > 0) {
    const errorBytes = new Uint8Array(data, offset, errorLen)
    error = new TextDecoder().decode(errorBytes)
  }

  return {
    type: 'error',
    error
  }
}

/**
 * Декодирует бинарное сообщение обновления клеток
 */
function decodeCellUpdateBinary(data: ArrayBuffer): DecodedMessage {
  const view = new DataView(data)
  let offset = 1 // Пропускаем тип сообщения

  // Читаем флаги
  const flags = view.getUint8(offset)
  offset += 1
  const gameOver = (flags & (1 << 0)) !== 0
  const gameWon = (flags & (1 << 1)) !== 0
  const hasRevealed = (flags & (1 << 2)) !== 0
  const hasHintsUsed = (flags & (1 << 3)) !== 0

  let loserPlayerId = ''
  let loserNickname = ''

  // Читаем GameOver данные
  if (gameOver) {
    const pidLen = view.getUint8(offset)
    offset += 1
    const pidBytes = new Uint8Array(data, offset, 5)
    offset += 5
    loserPlayerId = pidLen > 0 ? new TextDecoder().decode(pidBytes.slice(0, pidLen)) : ''

    const nicknameLen = view.getUint8(offset)
    offset += 1
    if (nicknameLen > 0) {
      const nicknameBytes = new Uint8Array(data, offset, nicknameLen)
      offset += nicknameLen
      loserNickname = new TextDecoder().decode(nicknameBytes)
    }
  }

  // Читаем Revealed
  let revealed = -1
  if (hasRevealed) {
    revealed = view.getUint16(offset, true)
    offset += 2
  }

  // Читаем HintsUsed
  let hintsUsed = -1
  if (hasHintsUsed) {
    hintsUsed = view.getUint8(offset)
    offset += 1
  }

  // Читаем количество обновленных клеток
  const updateCount = view.getUint16(offset, true)
  offset += 2

  const cellUpdates: Array<{ row: number; col: number; type: number }> = []

  for (let i = 0; i < updateCount; i++) {
    const row = view.getUint16(offset, true)
    offset += 2
    const col = view.getUint16(offset, true)
    offset += 2
    const type = view.getUint8(offset)
    offset += 1

    cellUpdates.push({ row, col, type })
  }

  return {
    type: 'cellUpdate',
    cellUpdates,
    gameOver,
    gameWon,
    ...(revealed >= 0 ? { revealed } : {}),
    ...(hintsUsed >= 0 ? { hintsUsed } : {}),
    ...(loserPlayerId ? { loserPlayerId } : {}),
    ...(loserNickname ? { loserNickname } : {})
  }
}

/**
 * Декодирует бинарное сообщение по типу
 */
export function decodeBinaryMessage(data: ArrayBuffer): DecodedMessage | null {
  if (data.byteLength === 0) {
    return null
  }

  const messageType = new Uint8Array(data)[0]

  switch (messageType) {
    case MessageTypeGameState:
      // gameState обрабатывается отдельно в websocket.ts
      return null
    case MessageTypeChat:
      return decodeChatBinary(data)
    case MessageTypeCursor:
      return decodeCursorBinary(data)
    case MessageTypePlayers:
      return decodePlayersBinary(data)
    case MessageTypePong:
      return { type: 'pong' }
    case MessageTypeError:
      return decodeErrorBinary(data)
    case MessageTypeCellUpdate:
      return decodeCellUpdateBinary(data)
    default:
      console.warn('Неизвестный тип бинарного сообщения:', messageType)
      return null
  }
}

