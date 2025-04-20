// API-функции для работы с устройствами
import { fetchWithTokenRefresh } from './interceptors';

// Типы данных для устройств
export interface Device {
  id: string;
  name: string;
  type: string;
  state: {
    on: boolean;
    [key: string]: any;
  };
}

/**
 * Получает список всех устройств
 */
export async function getDevices(token: string): Promise<Device[]> {
  return fetchWithTokenRefresh<Device[]>('/api/v1/devices', {
    method: 'GET',
    token: token,
  });
}

/**
 * Получает информацию об одном устройстве
 */
export async function getDevice(id: string, token: string): Promise<Device> {
  return fetchWithTokenRefresh<Device>(`/api/v1/devices/${id}`, {
    method: 'GET',
    token: token,
  });
}

/**
 * Изменяет состояние устройства (включает/выключает)
 */
export async function toggleDevice(id: string, on: boolean, token: string): Promise<Device> {
  return fetchWithTokenRefresh<Device>(`/api/v1/devices/${id}/control`, {
    method: 'POST',
    token: token,
    body: { command: 'toggle', parameters: { on } },
  });
}

/**
 * Отправляет голосовую команду
 */
export async function sendVoiceCommand(text: string, token: string): Promise<{ response: string }> {
  return fetchWithTokenRefresh<{ response: string }>('/api/v1/voice', {
    method: 'POST',
    token: token,
    body: { text },
  });
} 