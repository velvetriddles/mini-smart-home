import { refreshToken } from './auth';

// Хранилище для refresh токена
let refreshTokenValue: string | null = null;

// Флаг, указывающий, что процесс обновления токена уже идет
let isRefreshing = false;

// Очередь запросов, ожидающих обновления токена
let refreshQueue: Array<() => void> = [];

// Инициализирует refresh токен
export function initRefreshToken(token: string) {
  refreshTokenValue = token;
}

// Очищает refresh токен
export function clearRefreshToken() {
  refreshTokenValue = null;
}

// Выполняет запрос с автоматическим обновлением токена при необходимости
export async function fetchWithTokenRefresh<T>(
  url: string, 
  options: RequestInit & { token?: string } = {}
): Promise<T> {
  try {
    // Проверяем, есть ли токен в опциях
    if (options.token) {
      options.headers = {
        ...options.headers,
        'Authorization': `Bearer ${options.token}`
      };
    }

    // Делаем запрос
    const response = await fetch(url, options);
    
    // Если 401 Unauthorized и у нас есть refresh токен
    if (response.status === 401 && refreshTokenValue) {
      // Пытаемся обновить токен и повторить запрос
      const newToken = await handleTokenRefresh();
      
      if (newToken) {
        // Обновляем заголовок Authorization с новым токеном
        options.headers = {
          ...options.headers,
          'Authorization': `Bearer ${newToken}`
        };
        
        // Повторяем запрос с новым токеном
        const newResponse = await fetch(url, options);
        
        if (!newResponse.ok) {
          throw new Error(`HTTP error: ${newResponse.status} ${newResponse.statusText}`);
        }
        
        return await newResponse.json();
      } else {
        throw new Error('Не удалось обновить токен');
      }
    }
    
    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status} ${response.statusText}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Ошибка запроса с обновлением токена:', error);
    throw error;
  }
}

// Обрабатывает обновление токена
async function handleTokenRefresh(): Promise<string | null> {
  // Проверяем, что у нас есть refresh токен
  if (!refreshTokenValue) {
    return null;
  }
  
  // Если уже идет процесс обновления, добавляем в очередь
  if (isRefreshing) {
    return new Promise<string | null>((resolve) => {
      refreshQueue.push(() => {
        // После обновления токена получаем текущее значение из localStorage
        const currentToken = localStorage.getItem('token');
        resolve(currentToken);
      });
    });
  }
  
  // Устанавливаем флаг, что обновление токена начато
  isRefreshing = true;
  
  try {
    // Пытаемся обновить токен
    const tokenResponse = await refreshToken(refreshTokenValue);
    
    if (tokenResponse && tokenResponse.accessToken) {
      // Сохраняем новые токены
      localStorage.setItem('token', tokenResponse.accessToken);
      refreshTokenValue = tokenResponse.refreshToken;
      
      // Вызываем все колбэки из очереди
      refreshQueue.forEach(callback => callback());
      refreshQueue = [];
      
      return tokenResponse.accessToken;
    }
    
    return null;
  } catch (error) {
    console.error('Ошибка обновления токена:', error);
    // Очищаем токены при ошибке
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    refreshTokenValue = null;
    
    // В реальном приложении здесь можно перенаправить на страницу логина
    window.location.href = '/login';
    
    return null;
  } finally {
    isRefreshing = false;
  }
} 