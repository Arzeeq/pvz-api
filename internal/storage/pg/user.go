package pg

import (
	"context"
	"fmt"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStorage struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewUserStorage(pool *pgxpool.Pool) *UserStorage {
	return &UserStorage{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (s *UserStorage) CreateUser(ctx context.Context, payload dto.PostRegisterJSONBody) (*dto.User, error) {
	query, args, err := s.builder.
		Insert("users").
		Columns("email", "password_hash", "role").
		Values(payload.Email, payload.Password, payload.Role).
		Suffix("RETURNING id, email, role").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user dto.User
	err = s.pool.QueryRow(ctx, query, args...).Scan(&user.Id, &user.Email, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (s *UserStorage) GetUserPassword(ctx context.Context, email string) (string, error) {
	query, args, err := s.builder.
		Select("password_hash").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}

	var hashedPassword string
	err = s.pool.QueryRow(ctx, query, args...).Scan(&hashedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	return hashedPassword, nil
}

func (s *UserStorage) GetUserByEmail(ctx context.Context, email string) (*dto.User, error) {
	query, args, err := s.builder.
		Select("id", "email", "role").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user dto.User
	err = s.pool.QueryRow(ctx, query, args...).Scan(&user.Id, &user.Email, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
