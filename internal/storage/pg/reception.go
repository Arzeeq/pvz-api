package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrNoActiveReception = errors.New("no active receptions found in pvz")
var ErrBuildQuery = errors.New("failed to build query")

type ReceptionStorage struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewReceptionStorage(pool *pgxpool.Pool) (*ReceptionStorage, error) {
	if pool == nil {
		return nil, errors.New("nil values in NewPVZStorage constructor")
	}

	return &ReceptionStorage{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (s *ReceptionStorage) CreateReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	query, args, err := s.builder.
		Insert("receptions").
		Columns("pvz_id", "status").
		Values(pvzID, "in_progress").
		Suffix("RETURNING id, date_time, pvz_id, status").
		ToSql()
	if err != nil {
		return nil, ErrBuildQuery
	}

	var reception dto.Reception
	err = s.pool.QueryRow(ctx, query, args...).Scan(
		&reception.Id,
		&reception.DateTime,
		&reception.PvzId,
		&reception.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reception: %w", err)
	}

	return &reception, nil
}

func (s *ReceptionStorage) CloseReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	query, args, err := s.builder.
		Update("receptions").
		Set("status", dto.Close).
		Where(squirrel.Eq{
			"pvz_id": pvzID,
			"status": dto.InProgress,
		}).
		Suffix("RETURNING id, date_time, pvz_id, status").
		ToSql()
	if err != nil {
		return nil, ErrBuildQuery
	}

	var reception dto.Reception
	err = s.pool.QueryRow(ctx, query, args...).Scan(
		&reception.Id,
		&reception.DateTime,
		&reception.PvzId,
		&reception.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to close reception: %w", err)
	}

	return &reception, nil
}

// func (s *ReceptionStorage) GetActiveReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
// 	query, args, err := s.builder.
// 		Select("id", "date_time", "pvz_id", "status").
// 		From("receptions").
// 		Where(squirrel.Eq{
// 			"pvz_id": pvzID,
// 			"status": dto.InProgress,
// 		}).
// 		ToSql()
// 	if err != nil {
// 		return nil, ErrBuildQuery
// 	}

// 	var reception dto.Reception
// 	err = s.pool.QueryRow(ctx, query, args...).Scan(
// 		&reception.Id,
// 		&reception.DateTime,
// 		&reception.PvzId,
// 		&reception.Status,
// 	)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return nil, ErrNoActiveReception
// 		}
// 		return nil, fmt.Errorf("failed to get active reception: %w", err)
// 	}

// 	return &reception, nil
// }

func (s *ReceptionStorage) GetPVZReceptionsFiltered(ctx context.Context, pvzID openapi_types.UUID, startDate, endDate time.Time) []dto.Reception {
	query, args, err := s.builder.
		Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.And{
			squirrel.GtOrEq{"date_time": startDate},
			squirrel.LtOrEq{"date_time": endDate},
		}).
		Where(squirrel.Eq{"pvz_id": pvzID}).
		ToSql()

	if err != nil {
		return nil
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var receptions []dto.Reception
	for rows.Next() {
		var r dto.Reception
		if err := rows.Scan(&r.Id, &r.DateTime, &r.PvzId, &r.Status); err != nil {
			return nil
		}
		receptions = append(receptions, r)
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return receptions
}
