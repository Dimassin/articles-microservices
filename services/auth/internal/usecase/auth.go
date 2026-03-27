package usecase

import (
	"auth/internal/domain"
	"auth/internal/ports"
	"context"
	"log"
)

type AuthUsecase struct {
	userRepo       ports.UserRepository
	tokenMgr       ports.TokenManager
	hasher         ports.PasswordHasher
	eventPublisher ports.EventPublisher
}

func NewAuthUsecase(
	userRepo ports.UserRepository,
	tokenMgr ports.TokenManager,
	hasher ports.PasswordHasher,
	eventPublisher ports.EventPublisher,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:       userRepo,
		tokenMgr:       tokenMgr,
		hasher:         hasher,
		eventPublisher: eventPublisher,
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
