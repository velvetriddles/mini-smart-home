import React, { createContext, useState, useEffect, useContext } from 'react';
import { login as apiLogin, logout as apiLogout } from '../api/auth';
import { initRefreshToken, clearRefreshToken } from '../api/interceptors';

interface User {
  id: string;
  username: string;
  email: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth должен использоваться внутри AuthProvider');
  }
  return context;
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // При инициализации проверяем localStorage на наличие сохраненного токена
  useEffect(() => {
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');
    const storedRefreshToken = localStorage.getItem('refreshToken');
    
    if (storedToken && storedUser) {
      setToken(storedToken);
      try {
        setUser(JSON.parse(storedUser));
        
        // Инициализируем refresh токен, если он есть
        if (storedRefreshToken) {
          initRefreshToken(storedRefreshToken);
        }
      } catch (e) {
        console.error('Ошибка парсинга данных пользователя:', e);
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        clearRefreshToken();
      }
    }
    
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await apiLogin(email, password);
      
      if (response && response.accessToken) {
        localStorage.setItem('token', response.accessToken);
        
        // Сохраняем refresh токен
        if (response.refreshToken) {
          localStorage.setItem('refreshToken', response.refreshToken);
          initRefreshToken(response.refreshToken);
        }
        
        if (response.user) {
          localStorage.setItem('user', JSON.stringify(response.user));
          setUser(response.user);
        }
        setToken(response.accessToken);
      } else {
        throw new Error('Некорректный ответ от сервера');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    
    try {
      if (token) {
        await apiLogout(token);
      }
    } catch (err) {
      console.error('Ошибка выхода:', err);
    } finally {
      // Очистка состояния и localStorage даже при ошибке
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('refreshToken');
      clearRefreshToken();
      setUser(null);
      setToken(null);
      setIsLoading(false);
    }
  };

  const clearError = () => setError(null);

  return (
    <AuthContext.Provider value={{ 
      user, 
      token, 
      isLoading, 
      error, 
      login, 
      logout,
      clearError
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContext; 