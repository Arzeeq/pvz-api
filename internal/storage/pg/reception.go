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

func (s *ReceptionStorage) GetPVZReceptionsFiltered(
	ctx context.Context,
	pvzID openapi_types.UUID,
	startDate, endDate time.Time,
) ([]dto.Reception, error) {
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
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query receptions: %w", err)
	}
	defer rows.Close()

	var receptions []dto.Reception
	for rows.Next() {
		var r dto.Reception
		if err := rows.Scan(&r.Id, &r.DateTime, &r.PvzId, &r.Status); err != nil {
			return nil, fmt.Errorf("failed to scan reception: %w", err)
		}
		receptions = append(receptions, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return receptions, nil
}
