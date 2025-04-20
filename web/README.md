# Веб-интерфейс для Smart Home

Минимальный, но рабочий фронтенд для управления умным домом с использованием React, TypeScript и Tailwind CSS.

## Функциональность

- Аутентификация пользователя (получение JWT)
- Просмотр списка устройств и их статусов
- Включение/выключение устройств
- Отправка текстовых голосовых команд

## Технологии

- **Vite + React 18 + TypeScript** - основа приложения
- **Tailwind CSS** - стилизация
- **React Router 6** - маршрутизация
- **Context API** - управление состоянием (AuthContext + DevicesContext)
- **Fetch API** - работа с сетевыми запросами
- **localStorage** - хранение JWT-токена

## Структура проекта

```
src/
 ├── api/                // тонкие врапперы над fetch
 │   ├── auth.ts
 │   └── devices.ts
 ├── components/
 │   ├── Button.tsx
 │   ├── Card.tsx
 │   └── Spinner.tsx
 ├── contexts/
 │   ├── AuthContext.tsx
 │   └── DevicesContext.tsx
 ├── pages/
 │   ├── Login.tsx
 │   └── Dashboard.tsx
 ├── hooks/
 │   └── useFetch.ts
 ├── App.tsx
 └── main.tsx
```

## Запуск для разработки

```bash
# Установка зависимостей
npm install

# Запуск в режиме разработки
npm run dev
```

Сервер для разработки будет доступен по адресу http://localhost:5173

## Сборка для продакшн

```bash
# Сборка проекта
npm run build

# Предпросмотр собранного проекта
npm run preview
```

## Docker

```bash
# Сборка Docker-образа
docker build -t smarthome-web .

# Запуск контейнера
docker run -p 80:80 smarthome-web
```

## API

Приложение ожидает, что API Gateway доступен по адресу указанному в `vite.config.ts` (по умолчанию http://localhost:8080) и предоставляет следующие эндпоинты:

- **POST /api/v1/auth/login** - аутентификация пользователя
- **GET /api/v1/devices** - получение списка устройств
- **POST /api/v1/devices/{id}/control** - управление устройством
- **POST /api/v1/voice** - отправка голосовой команды 