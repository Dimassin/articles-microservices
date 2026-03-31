package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"auth/internal/adapter/hasher"
	"auth/internal/adapter/jwtadapter"
	"auth/internal/adapter/kafka"
	"auth/internal/adapter/repository/postgres"
	httptransport "auth/internal/adapter/transport/http"
	"auth/internal/adapter/transport/http/handler"
	"auth/internal/usecase"

	grpcserver "auth/internal/adapter/transport/grpc"

	pb "github.com/Dimassin/articles-microservices/proto/auth"
	"google.golang.org/grpc"
)

func main() {
	// Читаем переменные окружения
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	jwtSecret := os.Getenv("JWT_SECRET")
	kafkaBroker := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")

	// Значения по умолчанию для локальной разработки
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	if dbName == "" {
		dbName = "auth"
	}
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	if kafkaTopic == "" {
		kafkaTopic = "user-events"
	}

	// Формируем строку подключения к БД
	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser +
		" password=" + dbPassword + " dbname=" + dbName + " sslmode=" + dbSSLMode

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connected")

	userRepo := postgres.NewUserRepository(db)
	passwordHasher := hasher.NewBcryptHasher()

	accessTTL := 15 * time.Minute
	refreshTTL := 720 * time.Hour
	jwtManager := jwtadapter.NewJWTManager(jwtSecret, accessTTL, refreshTTL)

	kafkaProducer := kafka.NewEventProducer([]string{kafkaBroker}, kafkaTopic)
	defer kafkaProducer.Close()

	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager, passwordHasher, kafkaProducer, refreshTokenRepo)

	// gRPC сервер
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

	// HTTP сервер
	authHandler := handler.NewAuthHandler(authUsecase)
	router := httptransport.SetupRouter(authHandler)
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Println("Server starting on port", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
