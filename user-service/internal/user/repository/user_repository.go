package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/virhanali/filmnesia/user-service/internal/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at;
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = "user"
	}

	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		log.Printf("Error creating user in DB: %v. Query: %s", err, query)
		return nil, err
	}

	log.Printf("User created successfully with ID: %s", user.ID.String())
	return user, nil
}

func scanUser(rowScanner interface{ Scan(...interface{}) error }) (*domain.User, error) {
	var user domain.User
	var idBytes []byte

	err := rowScanner.Scan(
		&idBytes,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error scanning user row: %v", err)
		return nil, err
	}

	parsedID, err := uuid.FromBytes(idBytes)
	if err != nil {
		parsedID, err = uuid.Parse(string(idBytes))
		if err != nil {
			log.Printf("Error parsing UUID from DB: %v", err)
		}
	}
	user.ID = parsedID

	return &user, nil
}

func (r *pgUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1;
	`
	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with email: %s", email)
			return nil, nil
		}
		log.Printf("Error getting user by email from DB: %v. Email: %s", err, email)
		return nil, err
	}
	return &user, nil
}

func (r *pgUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1;
	`
	row := r.db.QueryRowContext(ctx, query, username)
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with username: %s", username)
			return nil, nil
		}
		log.Printf("Error getting user by username from DB: %v. Username: %s", err, username)
		return nil, err
	}
	return &user, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1;
	`
	row := r.db.QueryRowContext(ctx, query, id)
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with ID: %s", id.String())
			return nil, nil
		}
		log.Printf("Error getting user by ID from DB: %v. ID: %s", err, id.String())
		return nil, err
	}
	return &user, nil
}

func (r *pgUserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, role = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, username, email, password_hash, role, created_at, updated_at;
	`
	user.UpdatedAt = time.Now()
	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.UpdatedAt,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error updating user in DB: %v. Query: %s", err, query)
		return nil, err
	}
	log.Printf("User updated successfully with ID: %s", user.ID.String())
	return user, nil
}

func (r *pgUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1;
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Error deleting user from DB: %v. Query: %s", err, query)
		return err
	}
	log.Printf("User deleted successfully with ID: %s", id.String())
	return nil
}
