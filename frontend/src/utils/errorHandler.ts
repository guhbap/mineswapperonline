/**
 * Утилита для обработки ошибок API с русской локализацией
 */

interface ErrorResponse {
  error?: string
  message?: string
}

/**
 * Извлекает сообщение об ошибке из ответа API и переводит его на русский
 */
export function getErrorMessage(error: any, defaultMessage?: string): string {
  // Если передан объект ошибки axios
  if (error?.response?.data) {
    const data = error.response.data
    
    // Если data - это строка, пытаемся распарсить как JSON
    if (typeof data === 'string') {
      try {
        const parsed = JSON.parse(data) as ErrorResponse
        if (parsed.error) {
          return translateError(parsed.error)
        }
        if (parsed.message) {
          return translateError(parsed.message)
        }
      } catch {
        // Если не JSON, возвращаем строку как есть (если она уже на русском)
        return data
      }
    }
    
    // Если data - это объект
    if (typeof data === 'object') {
      const errorObj = data as ErrorResponse
      if (errorObj.error) {
        return translateError(errorObj.error)
      }
      if (errorObj.message) {
        return translateError(errorObj.message)
      }
      // Если объект не содержит error/message, но это объект с ошибкой
      // Пытаемся найти строковое значение
      const values = Object.values(data)
      if (values.length > 0 && typeof values[0] === 'string') {
        return translateError(values[0])
      }
    }
  }
  
  // Если есть message в самой ошибке
  if (error?.message) {
    return translateError(error.message)
  }
  
  // Возвращаем сообщение по умолчанию или общее сообщение
  return defaultMessage || 'Произошла ошибка. Попробуйте снова.'
}

/**
 * Переводит английские сообщения об ошибках на русский
 */
function translateError(message: string): string {
  const translations: Record<string, string> = {
    // Ошибки авторизации
    'Invalid username or password': 'Неверное имя пользователя или пароль',
    'Username or email already exists': 'Пользователь с таким именем или email уже существует',
    'User not found': 'Пользователь не найден',
    
    // Ошибки валидации
    'Invalid request body': 'Неверный формат запроса',
    'Method not allowed': 'Метод не разрешен',
    'Username parameter is required': 'Требуется параметр имени пользователя',
    'room name required': 'Требуется название комнаты',
    'rows and cols must be between 5 and 50': 'Количество строк и столбцов должно быть от 5 до 50',
    'mines must be between 1 and (rows*cols-1)': 'Количество мин должно быть от 1 до (строки × столбцы - 1)',
    'username and password are required': 'Требуются имя пользователя и пароль',
    'password must be at least 6 characters': 'Пароль должен содержать минимум 6 символов',
    
    // Ошибки комнат
    'Room not found': 'Комната не найдена',
    'Invalid password': 'Неверный пароль',
    
    // Общие ошибки
    'Internal server error': 'Внутренняя ошибка сервера',
    'Network Error': 'Ошибка сети. Проверьте подключение к интернету',
    'timeout': 'Превышено время ожидания ответа',
    'Request failed with status code 401': 'Ошибка авторизации',
    'Request failed with status code 403': 'Доступ запрещен',
    'Request failed with status code 404': 'Ресурс не найден',
    'Request failed with status code 500': 'Внутренняя ошибка сервера',
  }
  
  // Проверяем точное совпадение
  if (translations[message]) {
    return translations[message]
  }
  
  // Проверяем частичное совпадение для статус-кодов
  for (const [key, value] of Object.entries(translations)) {
    if (message.includes(key)) {
      return value
    }
  }
  
  // Если сообщение уже на русском или не найдено в словаре, возвращаем как есть
  return message
}

