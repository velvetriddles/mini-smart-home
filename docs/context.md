# Smart‑Home Demo (Diploma Edition)

**Тезисно**  
• Monorepo на Go 1.22, фронт — React+Vite.  
• Микросервисы: api‑gateway, auth, device, voice, web (SPA).  
• Транспорты: gRPC (внутри), REST+WS (внешне).  
• Шины: Kafka (commands, statusChanged); MQTT (телеметрия Wirenboard).  
• DB: PostgreSQL (device_state, users), Redis (revoked‑jwt).  
• Развёртывание: Kind + Helm‑umbrella, локальный registry :5000.  
• Генерация stubs — `buf` (grpc + grpc‑gateway + OpenAPI).  

**Папки monorepo**  
services/ libs/ proto/ web/ infra/ docs/ .github/ go.work


**Железное правило**  
– Один Go‑модуль на каждый сервис (`services/<name>/go.mod`).  
– Общий код — только в `libs/`.  
– Контракты хранятся в `/proto`, а в сервисах лежит ТОЛЬКО сген‑код.  

Сервисы по умолчанию слушают: gateway :8080, auth :9090, voice :9100, device :9200.  
Kafka topics: `commands`, `statusChanged`.  
MQTT topics: `wb/device/+/status`, `wb/device/+/command`.  

**Что нужно получить на выходе**  
1. Папочная структура.  
2. Скелеты кода (main.go + Di‑wire или fx ⟂ без бизнес‑логики).  
3. Multi‑stage Dockerfile для каждого сервиса.  
4. Helm‑umbrella c динамическим `.Values.services[].replicaCount`.  
5. GitHub Actions: тесты → build → push → helm upgrade в Kind.  
6. Фронт получает device‑state по WS, отправляет REST → Gateway.  

_Этого достаточно, чтобы показать архитектуру и масштабирование на дипломе._
