import { useState, useEffect } from 'react';
import { fetchWithTokenRefresh } from '../api/interceptors';

// Опции для запроса
export interface UseFetchOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: any;
  token?: string;
}

// Результат запроса
export interface UseFetchResult<T> {
  data: T | null;
  loading: boolean;
  error: Error | null;
  refetch: () => void;
}

export function useFetch<T>(url: string, options: UseFetchOptions = {}): UseFetchResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [refetchIndex, setRefetchIndex] = useState(0);

  const refetch = () => setRefetchIndex(prev => prev + 1);

  useEffect(() => {
    let isMounted = true;
    const abortController = new AbortController();
    
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const requestOptions: RequestInit & { token?: string } = {
          method: options.method || 'GET',
          headers: {
            'Content-Type': 'application/json',
            ...options.headers,
          },
          signal: abortController.signal,
        };

        // Добавляем тело запроса, если оно есть
        if (options.body) {
          requestOptions.body = JSON.stringify(options.body);
        }

        // Добавляем JWT-токен, если он предоставлен
        if (options.token) {
          requestOptions.token = options.token;
        }

        // Используем fetchWithTokenRefresh для автоматического обновления токена
        const result = await fetchWithTokenRefresh<T>(url, requestOptions);
        
        if (isMounted) {
          setData(result);
        }
      } catch (err) {
        if (err.name !== 'AbortError' && isMounted) {
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    fetchData();

    return () => {
      isMounted = false;
      abortController.abort();
    };
  }, [url, options.method, options.token, refetchIndex]);

  return { data, loading, error, refetch };
} 