# Диаграммы последовательности взаимодействия

## Аутентификация пользователя

```plantuml
@startuml
title Аутентификация пользователя

actor Пользователь
participant "Web\n(React SPA)" as Web
participant "API Gateway" as Gateway
participant "Auth Service" as Auth
database "PostgreSQL\n(users)" as DB
database "Redis\n(revoked-jwt)" as Redis

Пользователь -> Web: Логин/пароль
Web -> Gateway: POST /api/v1/auth/login
Gateway -> Auth: gRPC Login()
Auth -> DB: Проверка учётных данных
DB --> Auth: Данные пользователя
Auth -> Auth: Генерация JWT\n(access_token и refresh_token)
Auth --> Gateway: LoginResponse{accessToken, refreshToken, user}
Gateway --> Web: {accessToken, refreshToken, user}
Web -> Web: Сохранение токенов\nв localStorage
Web --> Пользователь: Успешная аутентификация

@enduml
```

## Обновление токена доступа

```plantuml
@startuml
title Обновление токена доступа

actor Пользователь
participant "Web\n(React SPA)" as Web
participant "API Gateway" as Gateway
participant "Auth Service" as Auth
database "PostgreSQL\n(users)" as DB
database "Redis\n(revoked-jwt)" as Redis

Пользователь -> Web: Запрос данных
Web -> Web: Проверка наличия\nи срока действия токена

alt Токен истек или близок к истечению
    Web -> Gateway: POST /api/v1/auth/refresh\n{refreshToken}
    Gateway -> Auth: gRPC Refresh()
    Auth -> Auth: Проверка refresh_token
    
    alt refresh_token валиден
        Auth -> DB: Получение данных пользователя
        DB --> Auth: Данные пользователя
        Auth -> Redis: Отзыв старого refresh_token
        Auth -> Auth: Генерация новых токенов
        Auth --> Gateway: RefreshResponse{accessToken, refreshToken, expiresAt}
        Gateway --> Web: {accessToken, refreshToken, expiresAt}
        Web -> Web: Обновление токенов в localStorage
        Web -> Gateway: Повтор исходного запроса с новым токеном
        Gateway --> Web: Данные
        Web --> Пользователь: Отображение данных
    else refresh_token невалиден
        Auth --> Gateway: 401 Unauthorized
        Gateway --> Web: 401 Unauthorized
        Web -> Web: Перенаправление на страницу входа
        Web --> Пользователь: Требуется повторная аутентификация
    end
else Токен действителен
    Web -> Gateway: Запрос с токеном
    Gateway -> Auth: gRPC ValidateToken()
    Auth --> Gateway: ValidateTokenResponse{valid, user}
    Gateway --> Web: Данные
    Web --> Пользователь: Отображение данных
end

@enduml
```

## Автоматическое обновление токена при получении 401

```plantuml
@startuml
title Автоматическое обновление токена при получении 401

actor Пользователь
participant "Web\n(React SPA)" as Web
participant "Fetch\nInterceptor" as Interceptor
participant "API Gateway" as Gateway
participant "Auth Service" as Auth
database "Redis\n(revoked-jwt)" as Redis

Пользователь -> Web: Действие, требующее запрос к API
Web -> Interceptor: fetch('/api/ресурс')
Interceptor -> Gateway: Запрос с access_token
Gateway -> Auth: gRPC ValidateToken()
Auth -> Auth: Проверка токена\n(истек срок действия)
Auth --> Gateway: 401 Unauthorized

Gateway --> Interceptor: 401 Unauthorized
Interceptor -> Interceptor: Проверка наличия refresh_token
Interceptor -> Gateway: POST /api/v1/auth/refresh\n{refreshToken}
Gateway -> Auth: gRPC Refresh()
Auth -> Auth: Проверка refresh_token
Auth -> Redis: Отзыв старого refresh_token
Auth -> Auth: Генерация новых токенов
Auth --> Gateway: RefreshResponse{accessToken, refreshToken, expiresAt}
Gateway --> Interceptor: {accessToken, refreshToken, expiresAt}

Interceptor -> Interceptor: Обновление токенов в localStorage
Interceptor -> Gateway: Повтор исходного запроса с новым токеном
Gateway -> Auth: gRPC ValidateToken()
Auth --> Gateway: ValidateTokenResponse{valid, user}
Gateway --> Interceptor: Данные
Interceptor --> Web: Данные
Web --> Пользователь: Отображение результата

@enduml
```

## Получение списка устройств

```plantuml
@startuml
title Получение списка устройств

actor Пользователь
participant "Web\n(React SPA)" as Web
participant "API Gateway" as Gateway
participant "Auth Service" as Auth
participant "Device Service" as Device
database "PostgreSQL\n(device_state)" as DB

Пользователь -> Web: Запрос списка устройств
Web -> Gateway: GET /api/v1/devices\nAuthorization: Bearer <token>
Gateway -> Auth: gRPC ValidateToken()
Auth --> Gateway: ValidateTokenResponse{valid, user}

alt Токен валиден
    Gateway -> Device: gRPC ListDevices()
    Device -> DB: Запрос списка устройств
    DB --> Device: Список устройств
    Device --> Gateway: ListDevicesResponse{devices}
    Gateway --> Web: {devices}
    Web --> Пользователь: Отображение устройств
else Токен невалиден
    Gateway --> Web: 401 Unauthorized
    Web --> Пользователь: Ошибка авторизации
end

@enduml
```

## Управление устройством

```plantuml
@startuml
title Управление устройством

actor Пользователь
participant "Web\n(React SPA)" as Web
participant "API Gateway" as Gateway
participant "Auth Service" as Auth
participant "Device Service" as Device
queue "Kafka\n(commands)" as Kafka
participant "MQTT\nCloudMQTT" as MQTT
participant "Wirenboard\nDevice" as WB

Пользователь -> Web: Нажатие кнопки управления\n(переключение света)
Web -> Gateway: POST /api/v1/devices/{id}/control\nAuthorization: Bearer <token>
Gateway -> Auth: gRPC ValidateToken()
Auth --> Gateway: ValidateTokenResponse{valid, user}

alt Токен валиден
    Gateway -> Device: gRPC ControlDevice()
    Device -> Kafka: Отправка команды\n(deviceId, command)
    Device --> Gateway: ControlDeviceResponse{success}
    Gateway --> Web: {success}
    Web --> Пользователь: Подтверждение выполнения команды
    
    Kafka --> Device: Обработка команды
    Device -> MQTT: Публикация в\nwb/device/{id}/command
    MQTT --> WB: Получение команды
    WB -> WB: Выполнение действия\n(вкл/выкл свет)
    WB -> MQTT: Публикация в\nwb/device/{id}/status
    MQTT --> Device: Получение обновленного статуса
    Device -> Kafka: Публикация в\nstatusChanged
    
    Kafka --> Web: WebSocket уведомление\nоб изменении статуса
    Web --> Пользователь: Обновление UI
else Токен невалиден
    Gateway --> Web: 401 Unauthorized
    Web --> Пользователь: Ошибка авторизации
end

@enduml
```

## Голосовое управление

```plantuml
@startuml
title Голосовое управление

actor Пользователь
participant "Web\n(React SPA)" as Web
participant "API Gateway" as Gateway
participant "Auth Service" as Auth
participant "Voice Service" as Voice
participant "Device Service" as Device
queue "Kafka\n(commands)" as Kafka
participant "MQTT\nCloudMQTT" as MQTT
participant "Wirenboard\nDevice" as WB

Пользователь -> Web: Голосовая команда
Web -> Gateway: POST /api/v1/voice\nAuthorization: Bearer <token>
Gateway -> Auth: gRPC ValidateToken()
Auth --> Gateway: ValidateTokenResponse{valid, user}

alt Токен валиден
    Gateway -> Voice: gRPC ProcessVoice()
    Voice -> Voice: Обработка голосовой команды
    Voice -> Device: gRPC ControlDevice()
    Device -> Kafka: Отправка команды\n(deviceId, command)
    Device --> Voice: ControlDeviceResponse{success}
    Voice --> Gateway: ProcessVoiceResponse{success}
    Gateway --> Web: {success}
    Web --> Пользователь: Подтверждение выполнения команды
    
    Kafka --> Device: Обработка команды
    Device -> MQTT: Публикация в\nwb/device/{id}/command
    MQTT --> WB: Получение команды
    WB -> WB: Выполнение действия\n(вкл/выкл свет)
    WB -> MQTT: Публикация в\nwb/device/{id}/status
    MQTT --> Device: Получение обновленного статуса
    Device -> Kafka: Публикация в\nstatusChanged
    
    Kafka --> Web: WebSocket уведомление\nоб изменении статуса
    Web --> Пользователь: Обновление UI
else Токен невалиден
    Gateway --> Web: 401 Unauthorized
    Web --> Пользователь: Ошибка авторизации
end

@enduml
``` 