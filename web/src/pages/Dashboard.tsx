import React, { useState, FormEvent, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useDevices } from '../contexts/DevicesContext';
import Button from '../components/Button';
import Card from '../components/Card';
import Spinner from '../components/Spinner';

const Dashboard: React.FC = () => {
  const { user, logout } = useAuth();
  const { 
    devices, 
    isLoading, 
    error, 
    fetchDevices,
    toggleDevice,
    sendVoiceCommand,
    voiceResponse,
    voiceLoading
  } = useDevices();
  
  const [voiceText, setVoiceText] = useState('');
  const navigate = useNavigate();

  // Автоматически перенаправлять на логин, если пользователь не авторизован
  useEffect(() => {
    if (!user) {
      navigate('/login');
    }
  }, [user, navigate]);

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const handleVoiceSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!voiceText.trim()) return;
    
    try {
      await sendVoiceCommand(voiceText);
      setVoiceText(''); // Очищаем поле ввода после отправки
    } catch (err) {
      console.error('Ошибка голосовой команды:', err);
    }
  };

  const handleDeviceToggle = async (id: string, currentState: boolean) => {
    await toggleDevice(id, !currentState);
  };

  const refreshDevices = () => {
    fetchDevices();
  };

  if (!user) {
    return null; // Или можно отрендерить спиннер загрузки
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <div className="flex-shrink-0 flex items-center">
                <h1 className="text-xl font-bold text-gray-900">Умный дом</h1>
              </div>
            </div>
            <div className="flex items-center">
              <span className="text-sm text-gray-500 mr-4">{user.email}</span>
              <Button variant="outline" onClick={handleLogout}>
                Выйти
              </Button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="mb-6">
          <Card title="Голосовое управление">
            <form onSubmit={handleVoiceSubmit} className="space-y-4">
              <div>
                <label htmlFor="voiceCommand" className="block text-sm font-medium text-gray-700">
                  Введите голосовую команду
                </label>
                <div className="mt-1">
                  <input
                    id="voiceCommand"
                    name="voiceCommand"
                    type="text"
                    className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    placeholder="Например: включи свет на кухне"
                    value={voiceText}
                    onChange={(e) => setVoiceText(e.target.value)}
                  />
                </div>
              </div>
              
              <div className="flex justify-end">
                <Button type="submit" isLoading={voiceLoading}>
                  Отправить
                </Button>
              </div>
            </form>

            {voiceResponse && (
              <div className="mt-4 p-3 bg-gray-100 rounded-md">
                <h4 className="text-sm font-medium text-gray-700">Ответ системы:</h4>
                <p className="mt-1 text-sm text-gray-900">{voiceResponse}</p>
              </div>
            )}
          </Card>
        </div>
        
        <div className="mb-6 flex justify-between items-center">
          <h2 className="text-2xl font-semibold text-gray-900">Мои устройства</h2>
          <Button variant="outline" onClick={refreshDevices}>
            Обновить
          </Button>
        </div>

        {isLoading ? (
          <div className="flex justify-center my-12">
            <Spinner size="lg" />
          </div>
        ) : error ? (
          <div className="rounded-md bg-red-50 p-4 my-6">
            <div className="text-sm text-red-700">{error}</div>
          </div>
        ) : devices.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">Нет подключенных устройств</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {devices.map((device) => (
              <Card 
                key={device.id} 
                title={device.name}
                className="hover:shadow-lg transition-shadow"
              >
                <div className="mb-4">
                  <p className="text-sm text-gray-500">Тип: {device.type}</p>
                  <p className="text-sm text-gray-500">
                    Статус: <span className={device.state.on ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>
                      {device.state.on ? 'Включено' : 'Выключено'}
                    </span>
                  </p>
                </div>
                
                <div className="flex justify-end">
                  <Button 
                    onClick={() => handleDeviceToggle(device.id, device.state.on)}
                    variant={device.state.on ? 'success' : 'primary'}
                  >
                    {device.state.on ? 'Выключить' : 'Включить'}
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Dashboard; 