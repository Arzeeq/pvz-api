package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type ProductStorage struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewProductStorage(pool *pgxpool.Pool) (*ProductStorage, error) {
	if pool == nil {
		return nil, errors.New("nil values in NewPVZStorage constructor")
	}

	return &ProductStorage{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (s *ProductStorage) GetReceptionProducts(ctx context.Context, receptionId openapi_types.UUID) ([]dto.Product, error) {
	query, args, err := s.builder.
		Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionId}).
		OrderBy("date_time ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []dto.Product
	for rows.Next() {
		var product dto.Product
		if err := rows.Scan(
			&product.Id,
			&product.DateTime,
			&product.Type,
			&product.ReceptionId,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return products, nil
}
