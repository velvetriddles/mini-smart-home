package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	smarthomev1 "github.com/velvetriddles/mini-smart-home/services/auth/proto/smarthome/v1"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// RedisClientInterface определяет интерфейс для работы с Redis
type RedisClientInterface interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Close() error
}

// Config содержит настройки сервера
type Config struct {
	GrpcPort    string
	HttpPort    string
	PostgresDSN string
	RedisDSN    string
	JwtSecret   string
	JwtTTL      time.Duration
}

// Server представляет собой сервер аутентификации
type Server struct {
	config      *Config
	db          *sql.DB
	redisClient RedisClientInterface
	grpcServer  *grpc.Server
	httpServer  *http.Server
	logger      *zap.Logger
	smarthomev1.UnimplementedAuthServiceServer
	healthpb.UnimplementedHealthServer
}

// NewServer создает новый экземпляр сервера
func NewServer(config *Config) (*Server, error) {
	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Подключение к PostgreSQL
	db, err := sql.Open("postgres", config.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Проверка соединения с PostgreSQL
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Инициализация схемы данных
	if err = initDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Подключение к Redis
	opt, err := redis.ParseURL(config.RedisDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis DSN: %w", err)
	}

	redisClient := redis.NewClient(opt)

	// Проверка соединения с Redis
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	// Настройка Prometheus метрик для gRPC
	grpc_prometheus.EnableHandlingTimeHistogram()

	// Создание gRPC сервера с middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
			grpc_logging.UnaryServerInterceptor(logger),
		)),
	)

	// Регистрация Prometheus метрик для gRPC
	grpc_prometheus.Register(grpcServer)

	// Создание HTTP сервера для health checks
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Создание сервера
	server := &Server{
		config:      config,
		db:          db,
		redisClient: redisClient,
		grpcServer:  grpcServer,
		httpServer: &http.Server{
			Addr:    ":" + config.HttpPort,
			Handler: router,
		},
		logger: logger,
	}

	// Регистрация сервисов gRPC
	// В реальном приложении здесь был бы вызов pb.RegisterAuthServiceServer(grpcServer, server)
	// Но мы используем только health сервис для примера
	healthpb.RegisterHealthServer(grpcServer, server)
	reflection.Register(grpcServer)

	// Настройка HTTP маршрутов
	router.Get("/healthz", server.HealthCheck)
	router.Get("/metrics", promhttp.Handler().ServeHTTP)

	return server, nil
}

// initDB инициализирует схему базы данных и добавляет тестового пользователя
func initDB(db *sql.DB) error {
	// Выполнение миграции из schema.sql
	schemaSQL, err := readSQLFile("sql/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	_, err = db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}

	// Проверка наличия тестового пользователя admin
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", "admin").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for admin user: %w", err)
	}

	// Добавление тестового пользователя, если его нет
	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost+2)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		_, err = db.Exec(
			"INSERT INTO users (id, username, email, pass_hash, roles) VALUES ($1, $2, $3, $4, $5)",
			"00000000-0000-0000-0000-000000000000",
			"admin",
			"admin@example.com",
			string(hashedPassword),
			[]string{"admin", "user"},
		)
		if err != nil {
			return fmt.Errorf("failed to insert admin user: %w", err)
		}

		log.Println("Default admin user created")
	}

	return nil
}

// readSQLFile читает файл SQL
func readSQLFile(filename string) (string, error) {
	// В реальном приложении здесь будет чтение из файла
	// Для упрощения в демонстрационных целях используем встроенную схему

	// Схема базы данных
	schema := `
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
`
	return schema, nil
}

// Start запускает сервер
func (s *Server) Start() error {
	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":"+s.config.GrpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.config.GrpcPort, err)
	}

	// Запуск gRPC сервера в горутине
	go func() {
		s.logger.Info("Starting gRPC server", zap.String("port", s.config.GrpcPort))
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Запуск HTTP сервера для health check и metrics в горутине
	go func() {
		s.logger.Info("Starting HTTP server", zap.String("port", s.config.HttpPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to serve HTTP", zap.Error(err))
		}
	}()

	return nil
}

// Stop останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	// Остановка HTTP сервера
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	// Graceful stop для gRPC сервера
	s.grpcServer.GracefulStop()

	// Закрытие соединений с БД
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("PostgreSQL connection close error: %w", err)
	}

	if err := s.redisClient.Close(); err != nil {
		return fmt.Errorf("Redis connection close error: %w", err)
	}

	// Синхронизация логгера перед завершением
	if err := s.logger.Sync(); err != nil {
		return fmt.Errorf("logger sync error: %w", err)
	}

	return nil
}

// HealthCheck обрабатывает HTTP-запрос для проверки работоспособности
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Проверка доступности PostgreSQL
	if err := s.db.Ping(); err != nil {
		http.Error(w, "Database connection error", http.StatusServiceUnavailable)
		return
	}

	// Проверка доступности Redis
	if _, err := s.redisClient.Ping(r.Context()).Result(); err != nil {
		http.Error(w, "Redis connection error", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Check реализует метод health.HealthServer для gRPC health check
func (s *Server) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	// Проверка доступности PostgreSQL
	if err := s.db.Ping(); err != nil {
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	// Проверка доступности Redis
	if _, err := s.redisClient.Ping(ctx).Result(); err != nil {
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

// Watch implements the Watch method for the health server
func (s *Server) Watch(req *healthpb.HealthCheckRequest, stream healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health watch is not implemented")
}

// GenerateJTI создает уникальный идентификатор для JWT-токена
func (s *Server) GenerateJTI(userID string) string {
	return fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
}

// GenerateJWT генерирует JWT токен для пользователя
func (s *Server) GenerateJWT(userID, username string, roles []string) (string, string, time.Time, error) {
	// Время истечения токена
	expiresAt := time.Now().Add(s.config.JwtTTL)

	// Создание уникального JTI (JWT ID)
	jti := s.GenerateJTI(userID)

	// Claims для токена доступа
	claims := jwt.MapClaims{
		"sub":   userID,
		"name":  username,
		"roles": roles,
		"jti":   jti,
		"iat":   time.Now().Unix(),
		"exp":   expiresAt.Unix(),
	}

	// Создание токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подпись токена
	accessToken, err := token.SignedString([]byte(s.config.JwtSecret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	// Для refresh токена используем тот же JTI, но с другим сроком действия
	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"jti": jti + "-refresh",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.config.JwtTTL * 2).Unix(), // Refresh токен живет дольше
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(s.config.JwtSecret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, signedRefreshToken, expiresAt, nil
}

// ValidateJWT проверяет JWT токен
func (s *Server) ValidateJWT(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	// Парсинг токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.config.JwtSecret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Проверка валидности токена
	if !token.Valid {
		return nil, nil, fmt.Errorf("invalid token")
	}

	// Извлечение claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("failed to extract claims")
	}

	// Проверка, не отозван ли токен
	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("missing jti claim")
	}

	ctx := context.Background()
	exists, err := s.redisClient.Exists(ctx, "revoked:"+jti).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check if token is revoked: %w", err)
	}

	if exists > 0 {
		return nil, nil, fmt.Errorf("token has been revoked")
	}

	return token, claims, nil
}

// RevokeToken добавляет токен в черный список
func (s *Server) RevokeToken(ctx context.Context, tokenString string) error {
	// Парсинг токена без валидации
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})

	// Обработка ошибок парсинга
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	// Извлечение claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("failed to extract claims")
	}

	// Получение JTI и времени истечения
	jti, ok := claims["jti"].(string)
	if !ok {
		return fmt.Errorf("missing jti claim")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing exp claim")
	}

	exp := time.Unix(int64(expFloat), 0)
	ttl := time.Until(exp)

	// Если токен уже истек, устанавливаем небольшой TTL
	if ttl <= 0 {
		ttl = time.Hour
	}

	// Добавление токена в черный список
	err = s.redisClient.Set(ctx, "revoked:"+jti, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// Login реализует метод Login из AuthService
func (s *Server) Login(ctx context.Context, req *smarthomev1.LoginRequest) (*smarthomev1.LoginResponse, error) {
	// Проверка входных данных
	if req.Username == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	// Поиск пользователя в базе данных
	var user struct {
		ID       string
		Username string
		Email    string
		PassHash string
		Roles    []string
	}

	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, pass_hash, roles FROM users 
         WHERE username = $1 OR email = $1`,
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PassHash, &user.Roles)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "invalid username or password")
		}
		s.logger.Error("Database error during login", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "database error")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(req.Password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	// Генерация JWT токена
	accessToken, refreshToken, expiresAt, err := s.GenerateJWT(user.ID, user.Username, user.Roles)
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	// Возврат ответа
	return &smarthomev1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
		User: &smarthomev1.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Roles:    user.Roles,
		},
	}, nil
}

// Logout реализует метод Logout из AuthService
func (s *Server) Logout(ctx context.Context, req *smarthomev1.LogoutRequest) (*smarthomev1.LogoutResponse, error) {
	// Проверка входных данных
	if req.AccessToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
	}

	// Отзыв токена
	err := s.RevokeToken(ctx, req.AccessToken)
	if err != nil {
		s.logger.Error("Error revoking token", zap.Error(err))
		return &smarthomev1.LogoutResponse{
			Success: false,
			Message: "Failed to revoke token",
		}, nil
	}

	return &smarthomev1.LogoutResponse{
		Success: true,
		Message: "Token revoked successfully",
	}, nil
}

// Refresh реализует метод Refresh из AuthService
func (s *Server) Refresh(ctx context.Context, req *smarthomev1.RefreshRequest) (*smarthomev1.RefreshResponse, error) {
	// Проверка входных данных
	if req.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required")
	}

	// Валидация refresh токена
	_, claims, err := s.ValidateJWT(req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	// Извлечение данных пользователя
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid token payload")
	}

	// Получение информации о пользователе из базы данных
	var user struct {
		ID       string
		Username string
		Email    string
		Roles    []string
	}

	err = s.db.QueryRowContext(
		ctx,
		"SELECT id, username, email, roles FROM users WHERE id = $1",
		sub,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Roles)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		s.logger.Error("Database error during token refresh", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "database error")
	}

	// Отзыв текущего refresh токена
	err = s.RevokeToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Warn("Error revoking refresh token", zap.Error(err))
	}

	// Генерация новых токенов
	accessToken, refreshToken, expiresAt, err := s.GenerateJWT(user.ID, user.Username, user.Roles)
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &smarthomev1.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// ValidateToken реализует метод ValidateToken из AuthService
func (s *Server) ValidateToken(ctx context.Context, req *smarthomev1.ValidateTokenRequest) (*smarthomev1.ValidateTokenResponse, error) {
	// Проверка входных данных
	if req.AccessToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
	}

	// Валидация токена
	_, claims, err := s.ValidateJWT(req.AccessToken)
	if err != nil {
		return &smarthomev1.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	// Извлечение данных пользователя
	userID, ok := claims["sub"].(string)
	if !ok {
		return &smarthomev1.ValidateTokenResponse{
			Valid: false,
			Error: "invalid token payload",
		}, nil
	}

	username, ok := claims["name"].(string)
	if !ok {
		return &smarthomev1.ValidateTokenResponse{
			Valid: false,
			Error: "invalid token payload",
		}, nil
	}

	// Извлечение ролей пользователя
	rolesClaim, ok := claims["roles"]
	if !ok {
		return &smarthomev1.ValidateTokenResponse{
			Valid: false,
			Error: "invalid token payload",
		}, nil
	}

	// Преобразование ролей из интерфейса в []string
	var roles []string

	if rolesArray, ok := rolesClaim.([]interface{}); ok {
		for _, role := range rolesArray {
			if roleStr, ok := role.(string); ok {
				roles = append(roles, roleStr)
			}
		}
	}

	// Получение актуальной информации о пользователе из базы данных
	var email string
	err = s.db.QueryRowContext(
		ctx,
		"SELECT email FROM users WHERE id = $1",
		userID,
	).Scan(&email)

	if err != nil {
		if err == sql.ErrNoRows {
			return &smarthomev1.ValidateTokenResponse{
				Valid: false,
				Error: "user not found",
			}, nil
		}

		s.logger.Error("Database error while validating token", zap.Error(err))

		// В случае ошибки базы данных, но валидного токена,
		// все равно возвращаем информацию из токена
		return &smarthomev1.ValidateTokenResponse{
			Valid: true,
			User: &smarthomev1.User{
				Id:       userID,
				Username: username,
				Roles:    roles,
			},
		}, nil
	}

	// Возврат успешного ответа
	return &smarthomev1.ValidateTokenResponse{
		Valid: true,
		User: &smarthomev1.User{
			Id:       userID,
			Username: username,
			Email:    email,
			Roles:    roles,
		},
	}, nil
}
