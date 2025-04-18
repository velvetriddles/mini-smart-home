package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Мок для Redis клиента
type mockRedisClient struct {
	mock.Mock
}

func (m *mockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *mockRedisClient) Exists(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *mockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Тест функции GenerateJTI
func TestGenerateJTI(t *testing.T) {
	server := &Server{
		config: &Config{},
		logger: zap.NewNop(),
	}

	userID := "test-user-id"
	jti1 := server.GenerateJTI(userID)
	jti2 := server.GenerateJTI(userID)

	assert.Contains(t, jti1, userID, "JTI должен содержать userID")
	assert.NotEqual(t, jti1, jti2, "Последовательные JTI должны быть различными")
}

// Тест функции GenerateJWT
func TestGenerateJWT(t *testing.T) {
	// Создание мок-сервера с настройками
	server := &Server{
		config: &Config{
			JwtSecret: "test-secret",
			JwtTTL:    time.Hour,
		},
		logger: zap.NewNop(),
	}

	// Тестовые данные
	userID := "test-user-id"
	username := "testuser"
	roles := []string{"user", "admin"}

	// Генерация токенов
	accessToken, refreshToken, expiresAt, err := server.GenerateJWT(userID, username, roles)

	// Проверка отсутствия ошибок
	assert.NoError(t, err, "GenerateJWT не должен возвращать ошибку")

	// Проверка, что токены не пустые
	assert.NotEmpty(t, accessToken, "Access token не должен быть пустым")
	assert.NotEmpty(t, refreshToken, "Refresh token не должен быть пустым")

	// Проверка времени истечения
	expectedExpiry := time.Now().Add(time.Hour)
	assert.InDelta(t, expiresAt.Unix(), expectedExpiry.Unix(), 1, "Время истечения должно быть около часа")

	// Парсинг access токена для проверки содержимого
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, err, "Ошибка при парсинге токена")

	// Проверка метода подписи
	assert.IsType(t, &jwt.SigningMethodHMAC{}, token.Method, "Неожиданный метод подписи")

	// Проверка claims
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok, "Не удалось извлечь claims")

	// Проверка содержимого claims
	assert.Equal(t, userID, claims["sub"], "Неверный sub в токене")
	assert.Equal(t, username, claims["name"], "Неверное имя в токене")

	// Проверка ролей
	assert.Contains(t, claims, "roles", "Отсутствуют роли в токене")
	roleClaims, ok := claims["roles"].([]interface{})
	assert.True(t, ok, "Роли неверного типа")
	assert.Len(t, roleClaims, len(roles), "Неверное количество ролей")

	// Проверка jti
	assert.Contains(t, claims, "jti", "Отсутствует jti в токене")

	// Проверка времени создания и истечения
	assert.Contains(t, claims, "iat", "Отсутствует iat в токене")
	assert.Contains(t, claims, "exp", "Отсутствует exp в токене")
}

// Тест функции ValidateJWT
func TestValidateJWT(t *testing.T) {
	// Создание мок Redis клиента
	mockRedis := new(mockRedisClient)
	existsCmd := redis.NewIntCmd(context.Background())
	existsCmd.SetVal(0) // Токен не отозван
	mockRedis.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(existsCmd)

	// Создание сервера с мок Redis
	server := &Server{
		config: &Config{
			JwtSecret: "test-secret",
			JwtTTL:    time.Hour,
		},
		redisClient: mockRedis,
		logger:      zap.NewNop(),
	}

	// Генерация тестового токена
	claims := jwt.MapClaims{
		"sub":   "test-user-id",
		"name":  "testuser",
		"roles": []string{"user", "admin"},
		"jti":   "test-jti",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	// Валидация токена
	validToken, validClaims, err := server.ValidateJWT(tokenString)

	// Проверка отсутствия ошибок
	assert.NoError(t, err, "ValidateJWT не должен возвращать ошибку для валидного токена")
	assert.NotNil(t, validToken, "Валидный токен не должен быть nil")
	assert.NotNil(t, validClaims, "Валидные claims не должны быть nil")

	// Проверка, что Redis Exists был вызван
	mockRedis.AssertCalled(t, "Exists", mock.Anything, "revoked:test-jti")

	// Тест с отозванным токеном
	mockRedis.ExpectedCalls = nil
	existsRevokedCmd := redis.NewIntCmd(context.Background())
	existsRevokedCmd.SetVal(1) // Токен отозван
	mockRedis.On("Exists", mock.Anything, mock.AnythingOfType("string")).Return(existsRevokedCmd)

	// Валидация отозванного токена
	_, _, err = server.ValidateJWT(tokenString)
	assert.Error(t, err, "ValidateJWT должен возвращать ошибку для отозванного токена")
	assert.Contains(t, err.Error(), "revoked", "Ошибка должна содержать информацию об отзыве")
}

// Тест функции RevokeToken
func TestRevokeToken(t *testing.T) {
	// Создание мок Redis клиента
	mockRedis := new(mockRedisClient)
	setCmd := redis.NewStatusCmd(context.Background())
	setCmd.SetVal("OK")
	mockRedis.On("Set", mock.Anything, mock.AnythingOfType("string"),
		mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(setCmd)

	// Создание сервера с мок Redis
	server := &Server{
		config: &Config{
			JwtSecret: "test-secret",
			JwtTTL:    time.Hour,
		},
		redisClient: mockRedis,
		logger:      zap.NewNop(),
	}

	// Генерация тестового токена
	claims := jwt.MapClaims{
		"sub": "test-user-id",
		"jti": "test-jti",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	// Отзыв токена
	err := server.RevokeToken(context.Background(), tokenString)

	// Проверка отсутствия ошибок
	assert.NoError(t, err, "RevokeToken не должен возвращать ошибку для валидного токена")

	// Проверка, что Redis Set был вызван с правильными параметрами
	mockRedis.AssertCalled(t, "Set", mock.Anything, "revoked:test-jti",
		"1", mock.AnythingOfType("time.Duration"))
}
