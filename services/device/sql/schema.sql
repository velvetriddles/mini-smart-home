-- Schema для Device Service
-- Для будущей миграции с in-memory хранилища на PostgreSQL

-- Таблица устройств
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    room VARCHAR(100),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица статусов устройств
CREATE TABLE IF NOT EXISTS device_statuses (
    device_id VARCHAR(36) PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    online BOOLEAN DEFAULT TRUE,
    battery_level INTEGER,
    signal_strength REAL,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица параметров устройств
CREATE TABLE IF NOT EXISTS device_parameters (
    device_id VARCHAR(36) REFERENCES devices(id) ON DELETE CASCADE,
    param_key VARCHAR(50) NOT NULL,
    param_value TEXT NOT NULL,
    PRIMARY KEY (device_id, param_key)
);

-- Таблица тегов устройств
CREATE TABLE IF NOT EXISTS device_tags (
    device_id VARCHAR(36) REFERENCES devices(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    PRIMARY KEY (device_id, tag)
);

-- Индексы для ускорения запросов
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_devices_room ON devices(room);
CREATE INDEX IF NOT EXISTS idx_device_statuses_online ON device_statuses(online);

-- Комментарии к таблицам
COMMENT ON TABLE devices IS 'Устройства умного дома';
COMMENT ON TABLE device_statuses IS 'Текущие статусы устройств';
COMMENT ON TABLE device_parameters IS 'Динамические параметры устройств';
COMMENT ON TABLE device_tags IS 'Теги для группировки устройств';

-- Пока SQL-скрипт содержит только схему без данных,
-- так как для первой итерации используется in-memory хранилище 