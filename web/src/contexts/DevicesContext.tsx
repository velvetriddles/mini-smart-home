import React, { createContext, useContext, useState, useEffect } from 'react';
import { Device, getDevices, toggleDevice as apiToggleDevice, sendVoiceCommand as apiSendVoiceCommand } from '../api/devices';
import { useAuth } from './AuthContext';

interface DevicesContextType {
  devices: Device[];
  isLoading: boolean;
  error: string | null;
  fetchDevices: () => Promise<void>;
  toggleDevice: (id: string, on: boolean) => Promise<void>;
  sendVoiceCommand: (text: string) => Promise<{ response: string }>;
  voiceResponse: string | null;
  voiceLoading: boolean;
}

const DevicesContext = createContext<DevicesContextType | null>(null);

export const useDevices = () => {
  const context = useContext(DevicesContext);
  if (!context) {
    throw new Error('useDevices должен использоваться внутри DevicesProvider');
  }
  return context;
};

export const DevicesProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { token, logout } = useAuth();
  
  const [devices, setDevices] = useState<Device[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [voiceResponse, setVoiceResponse] = useState<string | null>(null);
  const [voiceLoading, setVoiceLoading] = useState(false);

  // Загрузка устройств при монтировании, если есть токен
  useEffect(() => {
    if (token) {
      fetchDevices();
    }
  }, [token]);

  const fetchDevices = async () => {
    if (!token) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const fetchedDevices = await getDevices(token);
      setDevices(fetchedDevices);
    } catch (err: any) {
      console.error('Ошибка получения устройств:', err);
      
      // Если получена ошибка 401, выполняем выход
      if (err.message && err.message.includes('401')) {
        logout();
      } else {
        setError('Не удалось загрузить устройства');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const toggleDevice = async (id: string, on: boolean) => {
    if (!token) return;
    
    setError(null);
    
    try {
      // Оптимистичное обновление UI
      setDevices(prev => prev.map(device => 
        device.id === id ? { ...device, state: { ...device.state, on } } : device
      ));
      
      // Вызов API
      const updatedDevice = await apiToggleDevice(id, on, token);
      
      // Обновление состояния, если что-то изменилось в ответе
      setDevices(prev => prev.map(device => 
        device.id === id ? updatedDevice : device
      ));
    } catch (err) {
      console.error('Ошибка переключения устройства:', err);
      setError('Не удалось переключить устройство');
      
      // Восстановление предыдущего состояния в случае ошибки
      fetchDevices();
    }
  };

  const sendVoiceCommand = async (text: string) => {
    if (!token) throw new Error('Требуется авторизация');
    
    setVoiceLoading(true);
    setVoiceResponse(null);
    
    try {
      const response = await apiSendVoiceCommand(text, token);
      setVoiceResponse(response.response);
      
      // Обновляем список устройств, так как голосовая команда могла изменить их состояние
      fetchDevices();
      
      return response;
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Ошибка обработки голосовой команды';
      console.error('Ошибка голосовой команды:', err);
      setVoiceResponse(errorMsg);
      throw err;
    } finally {
      setVoiceLoading(false);
    }
  };

  return (
    <DevicesContext.Provider value={{
      devices,
      isLoading,
      error,
      fetchDevices,
      toggleDevice,
      sendVoiceCommand,
      voiceResponse,
      voiceLoading
    }}>
      {children}
    </DevicesContext.Provider>
  );
};

export default DevicesContext; 