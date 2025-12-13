/**
 * Protobuf сообщения для WebSocket
 * Использует protobufjs для динамической загрузки .proto файла
 */

import protobuf from 'protobufjs'

let root: protobuf.Root | null = null
let WebSocketMessage: protobuf.Type | null = null
let ClientMessage: protobuf.Type | null = null

// Загружаем .proto файл
async function loadProto(): Promise<void> {
  if (root) return // Уже загружен

  try {
    // Пытаемся загрузить из public директории
    root = await protobuf.load('/messages.proto')
    WebSocketMessage = root.lookupType('messages.WebSocketMessage')
    ClientMessage = root.lookupType('messages.ClientMessage')
  } catch (error) {
    console.error('Ошибка загрузки .proto файла из /messages.proto:', error)
    // Пытаемся загрузить из относительного пути
    try {
      root = await protobuf.load('messages.proto')
      WebSocketMessage = root.lookupType('messages.WebSocketMessage')
      ClientMessage = root.lookupType('messages.ClientMessage')
    } catch (error2) {
      console.error('Ошибка загрузки .proto файла из messages.proto:', error2)
      throw error2
    }
  }
}

// Кодирует сообщение в protobuf формат
export async function encodeProtobufMessage(message: any): Promise<ArrayBuffer> {
  await loadProto()

  if (!WebSocketMessage) {
    throw new Error('WebSocketMessage type not loaded')
  }

  // Создаем объект сообщения
  const msgObj: any = {}

  if (message.type === 'gameState' && message.gameState) {
    // В proto файле поле называется game_state (snake_case)
    msgObj.game_state = convertGameStateToProtobuf(message.gameState)
  } else if (message.type === 'chat' && message.chat) {
    msgObj.chat = {
      playerId: message.playerId || '',
      nickname: message.nickname || '',
      color: message.color || '',
      text: message.chat.text || '',
      isSystem: message.chat.isSystem || false,
      action: message.chat.action || '',
      row: message.chat.row !== undefined ? message.chat.row : -1,
      col: message.chat.col !== undefined ? message.chat.col : -1
    }
  } else if (message.type === 'cursor' && message.cursor) {
    msgObj.cursor = {
      playerId: message.playerId || message.cursor.pid || '',
      nickname: message.nickname || '',
      color: message.color || '',
      x: message.cursor.x,
      y: message.cursor.y
    }
  } else if (message.type === 'players' && message.players) {
    msgObj.players = {
      players: message.players.map((p: any) => ({
        id: p.id || '',
        nickname: p.nickname || '',
        color: p.color || ''
      }))
    }
  } else if (message.type === 'pong') {
    msgObj.pong = {}
  } else if (message.type === 'error' && message.error) {
    msgObj.error = {
      error: message.error
    }
  } else if (message.type === 'cellUpdate' && message.cellUpdates) {
    // В proto файле поле называется cell_update (snake_case)
    msgObj.cell_update = {
      gameOver: message.gameOver || false,
      gameWon: message.gameWon || false,
      revealed: message.revealed !== undefined ? message.revealed : -1,
      hintsUsed: message.hintsUsed !== undefined ? message.hintsUsed : -1,
      loserPlayerId: message.loserPlayerId || '',
      loserNickname: message.loserNickname || '',
      updates: message.cellUpdates.map((update: any) => ({
        row: update.row,
        col: update.col,
        type: update.type
      }))
    }
  }

  const errMsg = WebSocketMessage.verify(msgObj)
  if (errMsg) {
    throw new Error(`Ошибка валидации protobuf сообщения: ${errMsg}`)
  }

  const messageInstance = WebSocketMessage.create(msgObj)
  const buffer = WebSocketMessage.encode(messageInstance).finish()
  return buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength) as ArrayBuffer
}

// Декодирует protobuf сообщение
export async function decodeProtobufMessage(data: ArrayBuffer): Promise<any> {
  await loadProto()

  if (!WebSocketMessage) {
    throw new Error('WebSocketMessage type not loaded')
  }

  if (data.byteLength === 0) {
    throw new Error('Empty buffer')
  }

  const buffer = new Uint8Array(data)

  try {
    const message = WebSocketMessage.decode(buffer)
    const obj = WebSocketMessage.toObject(message, {
      longs: String,
      enums: Number, // Конвертируем enums в числа, а не строки
      bytes: String,
      defaults: true,
      arrays: true,
      objects: true,
      oneofs: true
    })

    // Преобразуем в формат WebSocketMessage
    // Protobufjs конвертирует snake_case в camelCase, но проверяем оба варианта
    if (obj.gameState || obj.game_state) {
    return {
      type: 'gameState',
      gameState: convertProtobufToGameState(obj.gameState || obj.game_state)
    }
  } else if (obj.chat) {
    return {
      type: 'chat',
      playerId: obj.chat.playerId,
      nickname: obj.chat.nickname,
      color: obj.chat.color,
      chat: {
        text: obj.chat.text,
        isSystem: obj.chat.isSystem,
        action: obj.chat.action,
        row: obj.chat.row >= 0 ? obj.chat.row : undefined,
        col: obj.chat.col >= 0 ? obj.chat.col : undefined
      }
    }
  } else if (obj.cursor) {
    return {
      type: 'cursor',
      playerId: obj.cursor.playerId,
      nickname: obj.cursor.nickname,
      color: obj.cursor.color,
      cursor: {
        pid: obj.cursor.playerId,
        x: obj.cursor.x,
        y: obj.cursor.y
      }
    }
  } else if (obj.players) {
    return {
      type: 'players',
      players: obj.players.players
    }
  } else if (obj.pong) {
    return {
      type: 'pong'
    }
  } else if (obj.error) {
    return {
      type: 'error',
      error: obj.error.error
    }
  } else if (obj.cellUpdate || obj.cell_update) {
    const cellUpdate = obj.cellUpdate || obj.cell_update
    return {
      type: 'cellUpdate',
      gameOver: cellUpdate.gameOver,
      gameWon: cellUpdate.gameWon,
      revealed: cellUpdate.revealed,
      hintsUsed: cellUpdate.hintsUsed,
      loserPlayerId: cellUpdate.loserPlayerId,
      loserNickname: cellUpdate.loserNickname,
      cellUpdates: cellUpdate.updates
    }
  }

    return null
  } catch (error) {
    // Если декодирование не удалось, выбрасываем ошибку
    throw new Error(`Failed to decode protobuf message: ${error}`)
  }
}

// Преобразует GameState в protobuf формат
function convertGameStateToProtobuf(gameState: any): any {
  const rows = gameState.b.map((row: any[]) => ({
    cells: row.map((cell: any) => ({
      isMine: cell.m,
      isRevealed: cell.r,
      isFlagged: cell.f,
      neighborMines: cell.n,
      flagColor: cell.fc || ''
    }))
  }))

  return {
    board: { rows },
    rows: gameState.r,
    cols: gameState.c,
    mines: gameState.m,
    gameOver: gameState.go,
    gameWon: gameState.gw,
    revealed: gameState.rv,
    hintsUsed: gameState.hu,
    safeCells: (gameState.sc || []).map((sc: any) => ({ row: sc.r, col: sc.c })),
    cellHints: (gameState.hints || []).map((h: any) => ({ row: h.r, col: h.c, type: h.t })),
    loserPlayerId: gameState.lpid || '',
    loserNickname: gameState.ln || ''
  }
}

// Преобразует protobuf GameState в формат приложения
function convertProtobufToGameState(gameState: any): any {
  const board = gameState.board.rows.map((row: any) =>
    row.cells.map((cell: any) => ({
      m: cell.isMine,
      r: cell.isRevealed,
      f: cell.isFlagged,
      n: cell.neighborMines,
      fc: cell.flagColor || undefined
    }))
  )

  return {
    b: board,
    r: gameState.rows,
    c: gameState.cols,
    m: gameState.mines,
    go: gameState.gameOver,
    gw: gameState.gameWon,
    rv: gameState.revealed,
    hu: gameState.hintsUsed,
    sc: gameState.safeCells?.map((sc: any) => ({ r: sc.row, c: sc.col })),
    hints: gameState.cellHints?.map((h: any) => ({ r: h.row, c: h.col, t: h.type })),
    lpid: gameState.loserPlayerId || undefined,
    ln: gameState.loserNickname || undefined
  }
}

// Кодирует клиентское сообщение в protobuf формат
export async function encodeClientMessage(message: any): Promise<ArrayBuffer> {
  await loadProto()

  if (!ClientMessage) {
    throw new Error('ClientMessage type not loaded')
  }

  const msgObj: any = {}

  if (message.type === 'nickname') {
    msgObj.nickname = message.nickname
  } else if (message.type === 'cursor' && message.cursor) {
    msgObj.cursor = {
      playerId: '',
      nickname: '',
      color: '',
      x: message.cursor.x,
      y: message.cursor.y
    }
  } else if (message.type === 'cellClick' && message.cellClick) {
    msgObj.cellClick = {
      row: message.cellClick.row,
      col: message.cellClick.col,
      flag: message.cellClick.flag
    }
  } else if (message.type === 'hint' && message.hint) {
    msgObj.hint = {
      row: message.hint.row,
      col: message.hint.col
    }
  } else if (message.type === 'newGame') {
    msgObj.newGame = {}
  } else if (message.type === 'chat' && message.chat) {
    msgObj.chat = {
      playerId: '',
      nickname: '',
      color: '',
      text: message.chat.text,
      isSystem: false,
      action: '',
      row: -1,
      col: -1
    }
  } else if (message.type === 'ping') {
    msgObj.ping = {}
  }

  const errMsg = ClientMessage.verify(msgObj)
  if (errMsg) {
    throw new Error(`Ошибка валидации protobuf сообщения: ${errMsg}`)
  }

  const messageInstance = ClientMessage.create(msgObj)
  const buffer = ClientMessage.encode(messageInstance).finish()
  return buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength) as ArrayBuffer
}

// Преобразует строковое имя типа клетки (enum) в число (fallback)
function convertCellTypeStringToNumber(typeStr: string): number {
  const typeMap: Record<string, number> = {
    'CELL_TYPE_NEIGHBOR_0': 0,
    'CELL_TYPE_NEIGHBOR_1': 1,
    'CELL_TYPE_NEIGHBOR_2': 2,
    'CELL_TYPE_NEIGHBOR_3': 3,
    'CELL_TYPE_NEIGHBOR_4': 4,
    'CELL_TYPE_NEIGHBOR_5': 5,
    'CELL_TYPE_NEIGHBOR_6': 6,
    'CELL_TYPE_NEIGHBOR_7': 7,
    'CELL_TYPE_NEIGHBOR_8': 8,
    'CELL_TYPE_MINE': 9,
    'CELL_TYPE_SAFE': 10,
    'CELL_TYPE_UNKNOWN': 11,
    'CELL_TYPE_DANGER': 12,
    'CELL_TYPE_CLOSED': 255
  }
  return typeMap[typeStr] ?? 255 // По умолчанию закрытая клетка
}

