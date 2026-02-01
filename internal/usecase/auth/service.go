package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
	"backend-gin/internal/infrastructure/config"
	"backend-gin/pkg/apperror"
)

const (
	refreshTokenPrefix = "refresh_token:"
	userTokensPrefix   = "user_tokens:"
)

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Role   entity.Role `json:"role"`
	jwt.RegisteredClaims
}

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	ValidateToken(tokenString string) (*Claims, error)
}

type service struct {
	userRepo    repository.UserRepository
	redisClient *redis.Client
	jwtConfig   config.JWTConfig
	bcryptCost  int
}

func NewService(userRepo repository.UserRepository, redisClient *redis.Client, jwtConfig config.JWTConfig, securityConfig config.SecurityConfig) Service {
	return &service{
		userRepo:    userRepo,
		redisClient: redisClient,
		jwtConfig:   jwtConfig,
		bcryptCost:  securityConfig.BcryptCost,
	}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to check existing user")
	}
	if existingUser != nil {
		return nil, apperror.Conflict("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to hash password")
	}

	user := &entity.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         entity.RoleUser,
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to create user")
	}

	return s.generateTokenPair(ctx, createdUser)
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrInvalidCredentials
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	return s.generateTokenPair(ctx, user)
}

func (s *service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*AuthResponse, error) {
	if s.redisClient == nil {
		return nil, apperror.InternalServer("refresh token not supported without Redis")
	}

	key := refreshTokenPrefix + req.RefreshToken
	userIDStr, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperror.Unauthorized("invalid or expired refresh token")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to validate refresh token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.Unauthorized("user not found")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to invalidate old refresh token")
	}

	s.removeUserToken(ctx, userID, req.RefreshToken)

	return s.generateTokenPair(ctx, user)
}

func (s *service) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	if s.redisClient == nil {
		return nil
	}

	key := refreshTokenPrefix + refreshToken
	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to invalidate refresh token")
	}

	s.removeUserToken(ctx, userID, refreshToken)

	return nil
}

func (s *service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if s.redisClient == nil {
		return nil
	}

	userTokensKey := userTokensPrefix + userID.String()
	tokens, err := s.redisClient.SMembers(ctx, userTokensKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user tokens")
	}

	for _, token := range tokens {
		s.redisClient.Del(ctx, refreshTokenPrefix+token)
	}

	s.redisClient.Del(ctx, userTokensKey)

	return nil
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.Unauthorized("invalid token signing method")
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, apperror.Unauthorized("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token claims")
	}

	return claims, nil
}

func (s *service) generateTokenPair(ctx context.Context, user *entity.User) (*AuthResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to generate access token")
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to generate refresh token")
	}

	if s.redisClient != nil {
		key := refreshTokenPrefix + refreshToken
		if err := s.redisClient.Set(ctx, key, user.ID.String(), s.jwtConfig.RefreshExpiration).Err(); err != nil {
			return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to store refresh token")
		}

		userTokensKey := userTokensPrefix + user.ID.String()
		s.redisClient.SAdd(ctx, userTokensKey, refreshToken)
		s.redisClient.Expire(ctx, userTokensKey, s.jwtConfig.RefreshExpiration)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessExpiration.Seconds()),
		TokenType:    "Bearer",
		User: UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role.String(),
		},
	}, nil
}

func (s *service) generateAccessToken(user *entity.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtConfig.AccessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "backend-gin",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

func (s *service) generateRefreshToken() (string, error) {
	length := s.jwtConfig.RefreshTokenLength
	if length == 0 {
		length = 64
	}
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *service) removeUserToken(ctx context.Context, userID uuid.UUID, token string) {
	if s.redisClient == nil {
		return
	}
	userTokensKey := userTokensPrefix + userID.String()
	s.redisClient.SRem(ctx, userTokensKey, token)
}
