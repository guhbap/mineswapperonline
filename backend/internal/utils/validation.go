package utils

// ValidateRoomParams валидирует параметры комнаты
func ValidateRoomParams(name string, rows, cols, mines int) error {
	if name == "" {
		return ErrRoomNameRequired
	}
	if rows < 5 || rows > 50 || cols < 5 || cols > 50 {
		return ErrInvalidDimensions
	}
	maxMines := (rows * cols) - 15
	if mines < 1 || mines > maxMines {
		return ErrInvalidMinesCount
	}
	return nil
}

// ValidateAuthParams валидирует параметры авторизации
func ValidateAuthParams(username, password string) error {
	if username == "" || password == "" {
		return ErrAuthRequired
	}
	return nil
}

// ValidateRegisterParams валидирует параметры регистрации
func ValidateRegisterParams(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return ErrAuthRequired
	}
	if len(password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}
