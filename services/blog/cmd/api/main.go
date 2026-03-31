package main

import (
	"blog/internal/ports"
	"database/sql"
	"log"
	"net/http"
	"os"

	"blog/internal/adapter/auth"
	"blog/internal/adapter/repository/postgres"
	bloghttp "blog/internal/adapter/transport/http"
	"blog/internal/adapter/transport/http/handler"
	"blog/internal/usecase"
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
	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")

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
		dbName = "blog"
	}
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:50051"
	}

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

	articleRepo := postgres.NewArticleRepository(db)

	localValidator := auth.NewJWTValidator(jwtSecret)
	grpcValidator, err := auth.NewGRPCTokenValidator(authGrpcAddr)
	if err != nil {
		log.Printf("Warning: failed to create gRPC validator: %v", err)
		grpcValidator = nil
	}

	var tokenValidator ports.TokenValidator
	if grpcValidator != nil {
		tokenValidator = auth.NewFallbackValidator(localValidator, grpcValidator)
	} else {
		tokenValidator = localValidator
	}

	articleUseCase := usecase.NewArticleUsecase(articleRepo, tokenValidator)
	articleHandler := handler.NewArticleHandler(articleUseCase)
	router := bloghttp.SetupRouter(articleHandler)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8081"
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
