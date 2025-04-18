# C4 модель архитектуры Smart-Home

## Context Diagram

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml

title Контекстная диаграмма системы "Smart-Home"

Person(user, "Пользователь", "Владелец умного дома")
System(smarthome, "Smart-Home", "Система управления умным домом")
System_Ext(wirenboard, "Wirenboard", "Контроллер устройств умного дома")

Rel(user, smarthome, "Управляет устройствами через веб-интерфейс/голосовые команды")
Rel(smarthome, wirenboard, "Отправляет команды и получает статусы устройств через MQTT")
Rel(wirenboard, smarthome, "Отправляет статусы устройств через MQTT")

@enduml
```

## Container Diagram

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

title Диаграмма контейнеров системы "Smart-Home"

Person(user, "Пользователь", "Владелец умного дома")
System_Boundary(smarthome, "Smart-Home") {
    Container(web, "Web SPA", "React, TypeScript", "Веб-интерфейс для управления умным домом")
    Container(api_gateway, "API Gateway", "Go, gRPC-gateway", "API Gateway для внешних запросов с REST -> gRPC преобразованием")
    Container(auth_service, "Auth Service", "Go, gRPC", "Сервис аутентификации и авторизации")
    Container(device_service, "Device Service", "Go, gRPC", "Сервис управления устройствами")
    Container(voice_service, "Voice Service", "Go, gRPC", "Сервис голосового управления")
    
    ContainerDb(postgres, "PostgreSQL", "База данных", "Хранение информации о пользователях и устройствах")
    ContainerDb(redis, "Redis", "Кэш", "Хранение отозванных JWT токенов")
    Container(kafka, "Kafka", "Брокер сообщений", "Обмен сообщениями между микросервисами")
    Container(mqtt_broker, "MQTT Broker", "Mosquitto", "Брокер для обмена сообщениями с устройствами")
}

System_Ext(wirenboard, "Wirenboard", "Контроллер устройств умного дома")

Rel(user, web, "Взаимодействует через", "HTTPS")
Rel(web, api_gateway, "Отправляет запросы на", "REST/WebSockets")
Rel(api_gateway, auth_service, "Проверяет аутентификацию и авторизацию через", "gRPC")
Rel(api_gateway, device_service, "Перенаправляет запросы на управление устройствами через", "gRPC")
Rel(api_gateway, voice_service, "Перенаправляет голосовые команды через", "gRPC")
Rel(voice_service, device_service, "Преобразует голосовые команды в управляющие сигналы через", "gRPC")

Rel(auth_service, postgres, "Хранит пользователей в", "SQL")
Rel(auth_service, redis, "Сохраняет отозванные токены в", "Redis API")
Rel(device_service, postgres, "Хранит состояния устройств в", "SQL")
Rel(device_service, kafka, "Публикует изменения статусов и команды в", "Kafka API")
Rel(kafka, api_gateway, "Уведомляет о изменениях через", "Kafka API")
Rel(device_service, mqtt_broker, "Отправляет команды и получает статусы через", "MQTT")
Rel(mqtt_broker, wirenboard, "Передает команды и получает статусы через", "MQTT")

@enduml
``` 