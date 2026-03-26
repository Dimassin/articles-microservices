package handler

import (
	"blog/internal/domain"
	"blog/internal/usecase"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type ArticleHandler struct {
	articleUsecase *usecase.ArticleUsecase
}

func NewArticleHandler(uc *usecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{
		articleUsecase: uc,
	}
}

func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(prefix):]

	resp, err := h.articleUsecase.CreateArticle(r.Context(), token, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.articleUsecase.GetArticle(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrArticleNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	resp, err := h.articleUsecase.ListArticles(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ArticleHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req domain.UpdateArticleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(prefix):]

	resp, err := h.articleUsecase.UpdateArticle(r.Context(), token, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.Is(err, domain.ErrArticleNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ArticleHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}
	token := authHeader[len(prefix):]

	err := h.articleUsecase.DeleteArticle(r.Context(), token, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.Is(err, domain.ErrArticleNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
