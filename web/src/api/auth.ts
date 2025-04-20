// API-функции для аутентификации

/**
 * Выполняет вход в систему
 */
export async function login(email: string, password: string) {
  const response = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username: email, password }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || 'Ошибка входа');
  }

  return response.json(); // Возвращает { accessToken, refreshToken, expiresAt, user }
}

/**
 * Выполняет выход из системы
 */
export async function logout(token: string) {
  const response = await fetch('/api/v1/auth/logout', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || 'Ошибка выхода');
  }

  return true;
}

/**
 * Обновляет токен доступа
 */
export async function refreshToken(refreshToken: string) {
  const response = await fetch('/api/v1/auth/refresh', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refreshToken }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || 'Ошибка обновления токена');
  }

  return response.json(); // Возвращает { accessToken, refreshToken, expiresAt }
} 