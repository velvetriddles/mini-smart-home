package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

var (
	// TokenAuth создает новый JWT-аутентификатор
	// В реальной системе секретный ключ должен быть взят из конфигурации
	TokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
)

// JWT middleware для проверки токена
func JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем проверку для публичных эндпоинтов
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем токен из заголовка Authorization
		token, claims, err := jwtauth.FromContext(r.Context())

		if err != nil || token == nil {
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return
		}

		// Проверяем срок действия токена
		if expClaim, exists := claims["exp"]; exists {
			exp, ok := expClaim.(float64)
			if !ok || float64(time.Now().Unix()) > exp {
				http.Error(w, "Unauthorized: Token expired", http.StatusUnauthorized)
				return
			}
		}

		// Токен валиден, продолжаем обработку запроса
		next.ServeHTTP(w, r)
	})
}

// isPublicPath определяет, требует ли путь аутентификацию
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/health",
		"/metrics",
		"/ws", // WebSocket для тестирования
	}

	for _, pp := range publicPaths {
		if strings.HasPrefix(path, pp) {
			return true
		}
	}

	return false
}

// Генерирует тестовый JWT-токен для отладки
// В реальной системе токены будет генерировать auth-сервис
func GenerateToken(userID string, username string) (string, error) {
	claims := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	_, tokenString, err := TokenAuth.Encode(claims)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
