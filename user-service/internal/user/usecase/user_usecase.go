package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/virhanali/filmnesia/user-service/internal/config"
	"github.com/virhanali/filmnesia/user-service/internal/platform/hash"
	"github.com/virhanali/filmnesia/user-service/internal/platform/messagebroker"
	"github.com/virhanali/filmnesia/user-service/internal/user/domain"
	"github.com/virhanali/filmnesia/user-service/internal/user/repository"

	"github.com/google/uuid"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserUsecase interface {
	Register(ctx context.Context, req domain.RegisterUserRequest) (*domain.UserResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
	GetByUsername(ctx context.Context, username string) (*domain.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	Login(ctx context.Context, req domain.LoginUserRequest) (*domain.LoginUserResponse, error)
}

type userUsecase struct {
	userRepo  repository.UserRepository
	appConfig config.Config
	publisher *messagebroker.RabbitMQPublisher
}

func NewUserUsecase(repo repository.UserRepository, appConfig config.Config, publisher *messagebroker.RabbitMQPublisher) UserUsecase {
	return &userUsecase{
		userRepo:  repo,
		appConfig: appConfig,
		publisher: publisher,
	}
}

func (uc *userUsecase) Register(ctx context.Context, req domain.RegisterUserRequest) (*domain.UserResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	existingUserByUsername, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		log.Printf("Error getting user by username: %v", err)
		return nil, err
	}
	if existingUserByUsername != nil {
		return nil, ErrUsernameExists
	}

	existingUserByEmail, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("Error getting user by email: %v", err)
		return nil, err
	}
	if existingUserByEmail != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, err
	}

	newUser := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	createdUser, err := uc.userRepo.Create(ctx, newUser)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	if uc.publisher != nil {
		event := domain.UserRegisteredEvent{
			UserID:       createdUser.ID,
			Email:        createdUser.Email,
			Username:     createdUser.Username,
			RegisteredAt: time.Now(),
		}

		exchangeName := "user_events"
		routingKey := "user.registered"

		err := uc.publisher.PublishUserRegisteredEvent(ctx, exchangeName, routingKey, event)
		if err != nil {
			log.Printf("Error publishing user registered event: %v", err)
		} else {
			log.Printf("User registered event published successfully.")
		}
	}

	return createdUser.ToUserResponse(), nil
}

func (uc *userUsecase) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToUserResponse(), nil
}

func (uc *userUsecase) GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("Error getting user by email: %v", err)
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToUserResponse(), nil
}

func (uc *userUsecase) GetByUsername(ctx context.Context, username string) (*domain.UserResponse, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		log.Printf("Error getting user by username: %v", err)
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToUserResponse(), nil
}

func (uc *userUsecase) UpdateUser(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Username != nil {
		newUserName := strings.TrimSpace(*req.Username)
		if newUserName == "" {
			return nil, ErrInvalidInput
		}
		existingUserByUsername, err := uc.userRepo.GetByUsername(ctx, newUserName)
		if err != nil {
			log.Printf("Error getting user by username: %v", err)
			return nil, err
		}
		if existingUserByUsername != nil && existingUserByUsername.ID != id {
			return nil, ErrUsernameExists
		}
		user.Username = newUserName
	}

	if req.Email != nil {
		newUserEmail := strings.TrimSpace(strings.ToLower(*req.Email))
		if newUserEmail == "" {
			return nil, ErrInvalidInput
		}
		existingUserByEmail, err := uc.userRepo.GetByEmail(ctx, newUserEmail)
		if err != nil {
			log.Printf("Error getting user by email: %v", err)
			return nil, err
		}
		if existingUserByEmail != nil && existingUserByEmail.ID != id {
			return nil, ErrEmailExists
		}
		user.Email = newUserEmail
	}

	user.UpdatedAt = time.Now()
	updatedUser, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return nil, err
	}
	return updatedUser.ToUserResponse(), nil
}

func (uc *userUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	userExisting, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
		return err
	}
	if userExisting == nil {
		return ErrUserNotFound
	}
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		log.Printf("Error deleting user: %v", err)
		return err
	}
	return nil
}

func (uc *userUsecase) Login(ctx context.Context, req domain.LoginUserRequest) (*domain.LoginUserResponse, error) {
	if req.Password == "" {
		return nil, ErrInvalidInput
	}

	var user *domain.User
	var err error
	found := false

	if req.Email != nil && *req.Email != "" {
		email := strings.TrimSpace(strings.ToLower(*req.Email))
		user, err = uc.userRepo.GetByEmail(ctx, email)
		if err != nil {
			log.Printf("Error when GetByEmail for login (email: %s): %v", email, err)
			return nil, err
		}
		if user != nil {
			found = true
		}
	}

	if !found && req.Username != nil && *req.Username != "" {
		username := strings.TrimSpace(*req.Username)
		user, err = uc.userRepo.GetByUsername(ctx, username)
		if err != nil {
			log.Printf("Error when GetByUsername for login (username: %s): %v", username, err)
			return nil, err
		}
		if user != nil {
			found = true
		}
	}

	if !found {
		if (req.Email == nil || *req.Email == "") && (req.Username == nil || *req.Username == "") {
			return nil, ErrInvalidInput
		}
		return nil, ErrInvalidCredentials
	}

	if !hash.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	expirationTime := time.Now().Add(time.Duration(uc.appConfig.JWTExpirationHours) * time.Hour)

	claims := &domain.AppClaims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "filmnesia-user-service",
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.appConfig.JWTSecretKey))
	if err != nil {
		log.Printf("Error when signing JWT token: %v", err)
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	userResponse := user.ToUserResponse()
	loginResponse := &domain.LoginUserResponse{
		AccessToken: tokenString,
		User:        *userResponse,
	}

	return loginResponse, nil
}
