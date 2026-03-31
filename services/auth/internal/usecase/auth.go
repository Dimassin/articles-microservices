package usecase

import (
	"auth/internal/domain"
	"auth/internal/ports"
	"context"
	"log"
	"time"
)

type AuthUsecase struct {
	userRepo         ports.UserRepository
	tokenMgr         ports.TokenManager
	hasher           ports.PasswordHasher
	eventPublisher   ports.EventPublisher
	refreshTokenRepo ports.RefreshTokenRepository
}

func NewAuthUsecase(
	userRepo ports.UserRepository,
	tokenMgr ports.TokenManager,
	hasher ports.PasswordHasher,
	eventPublisher ports.EventPublisher,
	refreshTokenRepo ports.RefreshTokenRepository,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:         userRepo,
		tokenMgr:         tokenMgr,
		hasher:           hasher,
		eventPublisher:   eventPublisher,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Register - регистрация нового пользователя
func (uc *AuthUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// 1. Проверяем, существует ли пользователь с таким email
	existing, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// 2. Проверяем сложность пароля
	if len(req.Password) < 6 {
		return nil, domain.ErrWeakPassword
	}

	// 3. Хешируем пароль
	hashedPassword, err := uc.hasher.Hash(ctx, req.Password)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	// 4. Создаем пользователя
	user := &domain.User{
		Email:    req.Email,
		Password: hashedPassword,
		Username: req.Username,
	}

	if err := uc.eventPublisher.PublishUserCreated(ctx, user.ID, user.Email, user.Username); err != nil {
		// Логируем ошибку, но не блокируем регистрацию
		log.Printf("Failed to publish user created event: %v", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5. Генерируем токены
	accessToken, err := uc.tokenMgr.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenMgr.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refreshTokenTTL := 720 * time.Hour
	expiresAt := time.Now().Add(refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, user.ID, refreshToken, expiresAt); err != nil {
		log.Printf("Failed to save refresh token: %v", err)
	}

	// 6. Возвращаем ответ
	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID,
		Email:        user.Email,
		Username:     user.Username,
	}, nil
}

// Login - вход пользователя
func (uc *AuthUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// 1. Находим пользователя по email
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 2. Проверяем пароль
	if err := uc.hasher.Compare(ctx, user.Password, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 3. Генерируем токены
	accessToken, err := uc.tokenMgr.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	refreshToken, err := uc.tokenMgr.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	refreshTokenTTL := 720 * time.Hour
	expiresAt := time.Now().Add(refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, user.ID, refreshToken, expiresAt); err != nil {
		log.Printf("Failed to save refresh token: %v", err)
	}

	// 4. Возвращаем ответ
	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID,
		Email:        user.Email,
		Username:     user.Username,
	}, nil
}

// ValidateToken - проверяет токен и возвращает информацию о пользователе
func (uc *AuthUsecase) ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error) {
	claims, err := uc.tokenMgr.VerifyToken(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (uc *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	userID, err := uc.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := uc.tokenMgr.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	newRefreshToken, err := uc.tokenMgr.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	refreshTokenTTL := 720 * time.Hour
	expiresAt := time.Now().Add(refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, user.ID, newRefreshToken, expiresAt); err != nil {
		return nil, domain.ErrInternalServer
	}

	if err := uc.refreshTokenRepo.Revoke(ctx, refreshToken); err != nil {
		log.Printf("Failed to revoke old refresh token: %v", err)
	}

	return &domain.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		UserID:       user.ID,
		Email:        user.Email,
		Username:     user.Username,
	}, nil
}
