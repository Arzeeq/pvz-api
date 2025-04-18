package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PVZStorage struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewPVZStorage(pool *pgxpool.Pool) (*PVZStorage, error) {
	if pool == nil {
		return nil, errors.New("nil values in NewPVZStorage constructor")
	}

	return &PVZStorage{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (s *PVZStorage) CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error) {
	columns := []string{"city"}
	values := []interface{}{payload.City}

	if payload.Id != nil {
		columns = append(columns, "id")
		values = append(values, payload.Id)
	}

	if payload.RegistrationDate != nil {
		columns = append(columns, "registration_date")
		values = append(values, payload.RegistrationDate)
	}

	query, args, err := s.builder.
		Insert("pvz").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id, registration_date, city").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var pvz dto.PVZ
	err = s.pool.QueryRow(ctx, query, args...).Scan(&pvz.Id, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return nil, fmt.Errorf("failed to create PVZ: %w", err)
	}

	return &pvz, nil
}

func (s *PVZStorage) GetPVZ(ctx context.Context, page, limit int) ([]dto.PVZ, error) {
	offset := (page - 1) * limit

	query, args, err := s.builder.
		Select("id", "registration_date", "city").
		From("pvz").
		OrderBy("registration_date").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build PVZ query: %w", err)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query PVZs: %w", err)
	}
	defer rows.Close()

	var pvzs []dto.PVZ
	for rows.Next() {
		var pvz dto.PVZ
		if err := rows.Scan(&pvz.Id, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, fmt.Errorf("failed to scan PVZ: %w", err)
		}

		pvzs = append(pvzs, pvz)
	}

	return pvzs, nil
}
