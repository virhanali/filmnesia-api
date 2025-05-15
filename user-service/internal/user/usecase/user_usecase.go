package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/virhanali/filmnesia/user-service/internal/platform/hash"
	"github.com/virhanali/filmnesia/user-service/internal/user/domain"
	"github.com/virhanali/filmnesia/user-service/internal/user/repository"

	"github.com/google/uuid"
)

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrUsernameExists = errors.New("username already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidInput   = errors.New("invalid input")
)

type UserUsecase interface {
	Register(ctx context.Context, req domain.RegisterUserRequest) (*domain.UserResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
	GetByUsername(ctx context.Context, username string) (*domain.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: repo,
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
