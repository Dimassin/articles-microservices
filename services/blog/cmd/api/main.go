package main

import (
	"database/sql"
	"log"
	"net/http"

	"blog/config"
	"blog/internal/adapter/auth"
	"blog/internal/adapter/repository/postgres"
	bloghttp "blog/internal/adapter/transport/http"
	"blog/internal/adapter/transport/http/handler"
	"blog/internal/usecase"
)

func main() {
	cfg := &config.Config{
		DB: config.DBConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "blog",
			SSLMode:  "disable",
		},
		Server: config.ServerConfig{
			Port: "8081",
		},
	}

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

	articleRepo := postgres.NewArticleRepository(db)
	mockTokenValidator := auth.NewMockTokenValidator()
	articleUseCase := usecase.NewArticleUsecase(articleRepo, mockTokenValidator)
	articleHandler := handler.NewArticleHandler(articleUseCase)
	router := bloghttp.SetupRouter(articleHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	log.Println("Server starting on port", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
