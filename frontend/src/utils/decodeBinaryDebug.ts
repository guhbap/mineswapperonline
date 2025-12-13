/**
 * Утилита для декодирования и отладки бинарных сообщений
 */

export function decodeBinaryMessageDebug(data: ArrayBuffer): void {
  if (data.byteLength === 0) {
    console.log('Пустое сообщение')
    return
  }

  const bytes = new Uint8Array(data)
  const messageType = bytes[0]

  console.log('=== Декодирование бинарного сообщения ===')
  console.log(`Тип сообщения: ${messageType}`)
  console.log(`Размер: ${data.byteLength} байт`)
  console.log(`Все байты:`, Array.from(bytes))

  switch (messageType) {
    case 0:
      console.log('Тип: gameState (binary)')
      decodeGameStateDebug(data)
      break
    case 1:
      console.log('Тип: chat')
      decodeChatDebug(data)
      break
    case 2:
      console.log('Тип: cursor')
      decodeCursorDebug(data)
      break
    case 3:
      console.log('Тип: players')
      decodePlayersDebug(data)
      break
    case 4:
      console.log('Тип: pong')
      break
    case 5:
      console.log('Тип: error')
      decodeErrorDebug(data)
      break
    case 6:
      console.log('Тип: cellUpdate')
      decodeCellUpdateDebug(data)
      break
    default:
      console.log(`Неизвестный тип: ${messageType}`)
  }
}

function decodeChatDebug(data: ArrayBuffer) {
  const view = new DataView(data)
  let offset = 1 // Пропускаем тип сообщения

  // PlayerID
  const pidLen = view.getUint8(offset)
  offset += 1
  const pidBytes = new Uint8Array(data, offset, 5)
  offset += 5
  const playerId = pidLen > 0 ? new TextDecoder().decode(pidBytes.slice(0, pidLen)) : ''
  console.log(`  PlayerID (длина ${pidLen}): "${playerId}"`)
  console.log(`  PlayerID байты [${offset - 5}-${offset - 1}]:`, Array.from(pidBytes.slice(0, pidLen)))

  // Nickname
  const nicknameLen = view.getUint8(offset)
  offset += 1
  let nickname = ''
  if (nicknameLen > 0) {
    const nicknameBytes = new Uint8Array(data, offset, nicknameLen)
    offset += nicknameLen
    nickname = new TextDecoder().decode(nicknameBytes)
  }
  console.log(`  Nickname (длина ${nicknameLen}): "${nickname}"`)
  if (nicknameLen > 0) {
    const nicknameBytes = new Uint8Array(data, offset - nicknameLen, nicknameLen)
    console.log(`  Nickname байты:`, Array.from(nicknameBytes))
  }

  // Color
  const colorLen = view.getUint8(offset)
  offset += 1
  let color = ''
  if (colorLen > 0) {
    const colorBytes = new Uint8Array(data, offset, colorLen)
    offset += colorLen
    color = new TextDecoder().decode(colorBytes)
  }
  console.log(`  Color (длина ${colorLen}): "${color}"`)
  if (colorLen > 0) {
    const colorBytes = new Uint8Array(data, offset - colorLen, colorLen)
    console.log(`  Color байты:`, Array.from(colorBytes))
  }

  // Text
  const textLen = view.getUint8(offset)
  offset += 1
  let text = ''
  if (textLen > 0) {
    const textBytes = new Uint8Array(data, offset, textLen)
    offset += textLen
    text = new TextDecoder().decode(textBytes)
  }
  console.log(`  Text (длина ${textLen}): "${text}"`)
  if (textLen > 0) {
    const textBytes = new Uint8Array(data, offset - textLen, textLen)
    console.log(`  Text байты:`, Array.from(textBytes))
  }

  // Флаги
  const flags = view.getUint8(offset)
  offset += 1
  const isSystem = (flags & (1 << 0)) !== 0
  const hasAction = (flags & (1 << 1)) !== 0
  const hasRowCol = (flags & (1 << 2)) !== 0
  console.log(`  Флаги (${flags.toString(2).padStart(8, '0')}):`, {
    isSystem,
    hasAction,
    hasRowCol
  })

  // Action
  if (hasAction) {
    const actionLen = view.getUint8(offset)
    offset += 1
    let action = ''
    if (actionLen > 0) {
      const actionBytes = new Uint8Array(data, offset, actionLen)
      offset += actionLen
      action = new TextDecoder().decode(actionBytes)
    }
    console.log(`  Action (длина ${actionLen}): "${action}"`)
  }

  // Row/Col
  if (hasRowCol) {
    const row = view.getUint16(offset, true)
    offset += 2
    const col = view.getUint16(offset, true)
    offset += 2
    console.log(`  Row: ${row}, Col: ${col}`)
  }

  console.log(`  Осталось байт: ${data.byteLength - offset}`)
}

function decodeGameStateDebug(data: ArrayBuffer) {
  console.log('  (gameState декодирование не реализовано в debug)')
}

function decodeCursorDebug(data: ArrayBuffer) {
  console.log('  (cursor декодирование не реализовано в debug)')
}

function decodePlayersDebug(data: ArrayBuffer) {
  console.log('  (players декодирование не реализовано в debug)')
}

function decodeErrorDebug(data: ArrayBuffer) {
  console.log('  (error декодирование не реализовано в debug)')
}

function decodeCellUpdateDebug(data: ArrayBuffer) {
  console.log('  (cellUpdate декодирование не реализовано в debug)')
}

