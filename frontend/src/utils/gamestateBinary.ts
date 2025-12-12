/**
 * Декодирует бинарный формат gameState
 * Формат:
 * - 2 байта: Rows (uint16, little-endian)
 * - 2 байта: Cols (uint16, little-endian)
 * - 2 байта: Mines (uint16, little-endian)
 * - 2 байта: Revealed (uint16, little-endian)
 * - 1 байт: HintsUsed (количество использованных подсказок, 0-3)
 * - 1 байт: Флаги (бит 0: GameOver, бит 1: GameWon)
 * - 1 байт: Длина LoserPlayerID (0-5)
 * - 5 байт: LoserPlayerID (ASCII)
 * - 1 байт: Длина LoserNickname
 * - N байт: LoserNickname (UTF-8)
 * - Rows*Cols байт: Board (каждая Cell = 1 байт)
 * - 1 байт: Количество флагов с цветами
 * - Для каждого флага: 2 байта cellKey (uint16), 1 байт длина цвета, N байт цвет
 * - 2 байта: Количество SafeCells (uint16)
 * - Для каждой SafeCell: 2 байта row (uint16), 2 байта col (uint16)
 */
export function decodeGameStateBinary(data: ArrayBuffer): {
  b: Array<Array<{ m: boolean; r: boolean; f: boolean; n: number; fc?: string }>> // board
  r: number // rows
  c: number // cols
  m: number // mines
  go: boolean // gameOver
  gw: boolean // gameWon
  rv: number // revealed
  hu: number // hintsUsed
  sc?: Array<{ r: number; c: number }> // safeCells
  hints?: Array<{ r: number; c: number; t: string }> // cellHints - подсказки для ячеек (MINE, SAFE, UNKNOWN)
  lpid?: string // loserPlayerId
  ln?: string // loserNickname
} {
  const view = new DataView(data)
  let offset = 0

  // Читаем размеры (uint16, little-endian)
  const rows = view.getUint16(offset, true)
  offset += 2
  const cols = view.getUint16(offset, true)
  offset += 2
  const mines = view.getUint16(offset, true)
  offset += 2
  const revealed = view.getUint16(offset, true)
  offset += 2

  // Читаем HintsUsed
  const hintsUsed = view.getUint8(offset)
  offset += 1

  // Читаем флаги
  const flags = view.getUint8(offset)
  offset += 1
  const gameOver = (flags & (1 << 0)) !== 0
  const gameWon = (flags & (1 << 1)) !== 0

  // Читаем LoserPlayerID
  const loserPIDLen = view.getUint8(offset)
  offset += 1
  const pidBytes = new Uint8Array(data, offset, 5)
  offset += 5
  let loserPlayerID: string | undefined
  if (loserPIDLen > 0 && loserPIDLen <= 5) {
    loserPlayerID = new TextDecoder('ascii').decode(pidBytes.slice(0, loserPIDLen))
  }

  // Читаем LoserNickname
  const nicknameLen = view.getUint8(offset)
  offset += 1
  let loserNickname: string | undefined
  if (nicknameLen > 0) {
    const nicknameBytes = new Uint8Array(data, offset, nicknameLen)
    offset += nicknameLen
    loserNickname = new TextDecoder('utf-8').decode(nicknameBytes)
  }

  // Инициализируем Board
  const board: Array<Array<{ m: boolean; r: boolean; f: boolean; n: number; fc?: string }>> = []
  for (let i = 0; i < rows; i++) {
    board[i] = []
    for (let j = 0; j < cols; j++) {
      const cellByte = view.getUint8(offset)
      offset += 1
      
      const cell = {
        m: (cellByte & (1 << 0)) !== 0, // IsMine
        r: (cellByte & (1 << 1)) !== 0, // IsRevealed
        f: (cellByte & (1 << 2)) !== 0, // IsFlagged
        n: (cellByte >> 3) & 0x0F // NeighborMines (бит 3-6)
      }
      board[i][j] = cell
    }
  }

  // Читаем цвета флагов (если есть данные)
  if (offset < data.byteLength) {
    const flagCount = view.getUint8(offset)
    offset += 1
    
    // Читаем цвета флагов
    for (let i = 0; i < flagCount && offset < data.byteLength; i++) {
      // Читаем cellKey (2 байта, little-endian)
      const cellKey = view.getUint16(offset, true)
      offset += 2
      
      // Читаем длину цвета
      if (offset >= data.byteLength) break
      const colorLen = view.getUint8(offset)
      offset += 1
      
      // Читаем цвет
      if (colorLen > 0 && colorLen <= 7 && offset + colorLen <= data.byteLength) {
        const colorBytes = new Uint8Array(data, offset, colorLen)
        offset += colorLen
        const color = new TextDecoder('utf-8').decode(colorBytes)
        
        // Применяем цвет к соответствующей ячейке
        const row = Math.floor(cellKey / cols)
        const col = cellKey % cols
        if (row >= 0 && row < rows && col >= 0 && col < cols) {
          board[row][col].fc = color
        }
      }
    }
  }

  // Читаем SafeCells (если есть данные)
  let safeCells: Array<{ r: number; c: number }> | undefined
  if (offset < data.byteLength) {
    const safeCellsCount = view.getUint16(offset, true)
    offset += 2
    
    if (safeCellsCount > 0) {
      safeCells = []
      for (let i = 0; i < safeCellsCount && offset + 4 <= data.byteLength; i++) {
        const row = view.getUint16(offset, true)
        offset += 2
        const col = view.getUint16(offset, true)
        offset += 2
        safeCells.push({ r: row, c: col })
      }
    }
  }

  return {
    b: board,
    r: rows,
    c: cols,
    m: mines,
    go: gameOver,
    gw: gameWon,
    rv: revealed,
    hu: hintsUsed,
    sc: safeCells,
    lpid: loserPlayerID,
    ln: loserNickname
  }
}

