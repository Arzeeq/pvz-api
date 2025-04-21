package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/internal/metrics"
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
		return nil, errors.New("nil values in NewProductStorage constructor")
	}

	return &ProductStorage{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (s *ProductStorage) GetReceptionProducts(ctx context.Context, receptionId openapi_types.UUID) []dto.Product {
	query, args, err := s.builder.
		Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionId}).
		OrderBy("date_time").
		ToSql()
	if err != nil {
		return nil
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil
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
			return nil
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return products
}

func (s *ProductStorage) CreateProduct(ctx context.Context, productDto dto.PostProductsJSONBody) (*dto.Product, error) {
	receptionQuery, receptionArgs, err := s.builder.
		Select("id").
		From("receptions").
		Where(squirrel.Eq{
			"pvz_id": productDto.PvzId,
			"status": "in_progress",
		}).
		ToSql()
	if err != nil {
		return nil, ErrBuildQuery
	}

	var receptionID openapi_types.UUID
	err = s.pool.QueryRow(ctx, receptionQuery, receptionArgs...).Scan(&receptionID)
	if err != nil {
		return nil, err
	}

	productQuery, productArgs, err := s.builder.
		Insert("products").
		Columns("type", "reception_id").
		Values(string(productDto.Type), receptionID).
		Suffix("RETURNING id, date_time, type, reception_id").
		ToSql()
	if err != nil {
		return nil, ErrBuildQuery
	}

	var product dto.Product
	err = s.pool.QueryRow(ctx, productQuery, productArgs...).Scan(
		&product.Id,
		&product.DateTime,
		&product.Type,
		&product.ReceptionId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	metrics.ProductsAddedTotal.Inc()
	return &product, nil
}

func (s *ProductStorage) GetLastProduct(ctx context.Context, pvzId openapi_types.UUID) (*dto.Product, error) {
	receptionQuery, receptionArgs, err := s.builder.
		Select("id").
		From("receptions").
		Where(squirrel.Eq{
			"pvz_id": pvzId,
			"status": dto.InProgress,
		}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, ErrBuildQuery
	}

	var receptionID openapi_types.UUID
	err = s.pool.QueryRow(ctx, receptionQuery, receptionArgs...).Scan(&receptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reception: %w", err)
	}

	productSelectQuery, productSelectArgs, err := s.builder.
		Select("id", "date_time", "reception_id", "type").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, ErrBuildQuery
	}

	var product dto.Product
	err = s.pool.QueryRow(ctx, productSelectQuery, productSelectArgs...).Scan(
		&product.Id,
		&product.DateTime,
		&product.ReceptionId,
		&product.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get last product: %w", err)
	}

	return &product, nil
}

func (s *ProductStorage) DeleteProduct(ctx context.Context, productID openapi_types.UUID) error {
	query, args, err := s.builder.
		Delete("products").
		Where(squirrel.Eq{"id": productID}).
		ToSql()
	if err != nil {
		return ErrBuildQuery
	}

	_, err = s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
