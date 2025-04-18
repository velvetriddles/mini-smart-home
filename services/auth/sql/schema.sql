-- Включение расширения для работы с UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    pass_hash TEXT NOT NULL,
    roles TEXT[] NOT NULL DEFAULT ARRAY['user']::TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Вставка тестового пользователя admin/admin123
-- Пароль захеширован с использованием bcrypt
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin') THEN
        INSERT INTO users (
            id, 
            username, 
            email, 
            pass_hash, 
            roles
        ) VALUES (
            '00000000-0000-0000-0000-000000000000',
            'admin',
            'admin@example.com',
            -- Хеш пароля 'admin123', сгенерированный bcrypt
            '$2a$10$BnBfAGNJlm3HFDL3Ym0n3.9bw0BNfC6qZheHJbCl4FI70n1RCe4P2',
            ARRAY['admin', 'user']::TEXT[]
        );
    END IF;
END
$$; 