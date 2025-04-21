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
		return nil, ErrBuildQuery
	}

	var pvz dto.PVZ
	err = s.pool.QueryRow(ctx, query, args...).Scan(&pvz.Id, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return nil, fmt.Errorf("failed to create PVZ: %w", err)
	}

	return &pvz, nil
}

func (s *PVZStorage) GetPVZs(ctx context.Context, params dto.GetPvzParams) ([]dto.PVZ, error) {
	offset := (*params.Page - 1) * (*params.Limit)
	query, args, err := s.builder.
		Select("pvz.id", "pvz.registration_date", "pvz.city").
		From("pvz").
		Join("receptions ON pvz.id = receptions.pvz_id").
		Where(squirrel.And{
			squirrel.GtOrEq{"receptions.date_time": *params.StartDate},
			squirrel.LtOrEq{"receptions.date_time": *params.EndDate},
		}).
		GroupBy("pvz.id").
		OrderBy("pvz.registration_date").
		Offset(uint64(offset)).
		Limit(uint64(*params.Limit)).
		ToSql()

	if err != nil {
		return nil, ErrBuildQuery
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var pvzs []dto.PVZ
	for rows.Next() {
		var pvz dto.PVZ
		if err := rows.Scan(
			&pvz.Id,
			&pvz.RegistrationDate,
			&pvz.City,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PVZ: %w", err)
		}
		pvzs = append(pvzs, pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return pvzs, nil
}

func (s *PVZStorage) GetAllPVZs(ctx context.Context) []dto.PVZ {
	query, args, err := s.builder.
		Select("pvz.id", "pvz.registration_date", "pvz.city").
		From("pvz").
		ToSql()

	if err != nil {
		return nil
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var pvzs []dto.PVZ
	for rows.Next() {
		var pvz dto.PVZ
		if err := rows.Scan(
			&pvz.Id,
			&pvz.RegistrationDate,
			&pvz.City,
		); err != nil {
			return nil
		}
		pvzs = append(pvzs, pvz)
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return pvzs
}
