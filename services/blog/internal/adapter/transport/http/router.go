package http

import (
	"net/http"

	"blog/internal/adapter/transport/http/handler"
)

func SetupRouter(articleHandler *handler.ArticleHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/articles", articleHandler.CreateArticle)
	mux.HandleFunc("GET /api/articles/{id}", articleHandler.GetArticle)
	mux.HandleFunc("GET /api/articles", articleHandler.ListArticles)
	mux.HandleFunc("PUT /api/articles/{id}", articleHandler.UpdateArticle)
	mux.HandleFunc("DELETE /api/articles/{id}", articleHandler.DeleteArticle)

	return mux
}
