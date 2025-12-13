package main

// Типы бинарных сообщений
const (
	MessageTypeGameState  = byte(0)
	MessageTypeChat       = byte(1)
	MessageTypeCursor     = byte(2)
	MessageTypePlayers    = byte(3)
	MessageTypePong       = byte(4)
	MessageTypeError      = byte(5)
	MessageTypeCellUpdate = byte(6)
)

// Типы клеток для бинарного формата
const (
	CellTypeClosed  = byte(255) // Закрыта (используем 255 вместо 0, чтобы не конфликтовать с количеством мин)
	CellTypeMine    = byte(9)   // Мина
	CellTypeSafe    = byte(10)  // Зеленая (SAFE) - для режима обучения
	CellTypeUnknown = byte(11)  // Желтая (UNKNOWN) - для режима обучения
	CellTypeDanger  = byte(12)  // Красная (MINE) - для режима обучения
	// Значения 0-8 используются для открытых клеток с количеством соседних мин (0-8)
)
