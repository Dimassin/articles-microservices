package main

import (
	"auth/internal/adapter/kafka"
	"database/sql"
	"log"
	"net"
	"net/http"
	"time"

	"auth/config"
	"auth/internal/adapter/hasher"
	"auth/internal/adapter/jwtadapter"
	"auth/internal/adapter/repository/postgres"
	httptransport "auth/internal/adapter/transport/http"
	"auth/internal/adapter/transport/http/handler"
	"auth/internal/usecase"

	grpcserver "auth/internal/adapter/transport/grpc" // ваш адаптер

	pb "github.com/Dimassin/articles-microservices/proto/auth"
	"google.golang.org/grpc" // библиотека gRPC
)

func main() {
	// 1. Загружаем конфиг
	cfg := &config.Config{
		DB: config.DBConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "auth",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:          "qwerty",
			AccessTokenTTL:  "15m",
			RefreshTokenTTL: "720h",
		},
		Server: config.ServerConfig{
			Port: "8080",
		},
	}

	// 2. Подключаемся к БД
	connStr := "host=" + cfg.DB.Host + " port=" + cfg.DB.Port + " user=" + cfg.DB.User +
		" password=" + cfg.DB.Password + " dbname=" + cfg.DB.DBName + " sslmode=" + cfg.DB.SSLMode
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connected")

	// 3. Создаем репозиторий
	userRepo := postgres.NewUserRepository(db)

	// 4. Создаем хешер
	passwordHasher := hasher.NewBcryptHasher()

	// 5. Создаем JWT менеджер
	accessTTL, err := time.ParseDuration(cfg.JWT.AccessTokenTTL)
	if err != nil {
		log.Fatal("Invalid access token TTL:", err)
	}
	refreshTTL, err := time.ParseDuration(cfg.JWT.RefreshTokenTTL)
	if err != nil {
		log.Fatal("Invalid refresh token TTL:", err)
	}

	jwtManager := jwtadapter.NewJWTManager(cfg.JWT.Secret, accessTTL, refreshTTL)

	kafkaProducer := kafka.NewEventProducer(
		[]string{"localhost:9092"},
		"user-events",
	)
	defer kafkaProducer.Close()

	// 6. Создаем usecase
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager, passwordHasher, kafkaProducer)

	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("Failed to listen on gRPC port:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, grpcserver.NewAuthGrpcServer(authUsecase))

	go func() {
		log.Println("gRPC server starting on port 50051")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()

	// 7. Создаем handler и роутер
	authHandler := handler.NewAuthHandler(authUsecase)
	router := httptransport.SetupRouter(authHandler)

	// 8. Запускаем сервер
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	log.Println("Server starting on port", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
