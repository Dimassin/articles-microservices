package http

import (
	"net/http"

	"auth/internal/adapter/transport/http/handler"
)

func SetupRouter(authHandler *handler.AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/validate", authHandler.Validate)

	return mux
}
