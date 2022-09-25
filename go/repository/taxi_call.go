package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type TaxiCallRepository interface {
	GetById(context.Context, string) (entity.TaxiCallRequest, error)
	GetLatestByUserId(context.Context, string) (entity.TaxiCallRequest, error)
	// GetLatestByDriverId(context.Context, string) (entity.TaxiCallRequest, error)
	ListByUserId(context.Context, string) ([]entity.TaxiCallRequest, error)
	// ListByDriverId(context.Context, string) ([]entity.TaxiCallRequest, error)
	Create(context.Context, entity.TaxiCallRequest) error
	Update(context.Context, entity.TaxiCallRequest) error
}

type taxiCallRepository struct{}

func (t taxiCallRepository) GetById(ctx context.Context, taxiCallRequestId string) (entity.TaxiCallRequest, error) {
	db := GetQueryContext(ctx)

	resp := entity.TaxiCallRequest{
		Id: taxiCallRequestId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallRequest{}, value.ErrUserNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) GetLatestByUserId(ctx context.Context, userId string) (entity.TaxiCallRequest, error) {
	db := GetQueryContext(ctx)

	resp := entity.TaxiCallRequest{}

	err := db.NewSelect().Model(&resp).Where("user_id = ?", userId).OrderExpr("create_time DESC").Limit(1).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.TaxiCallRequest{}, value.ErrUserNotFound
	}
	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) ListByUserId(ctx context.Context, userId string) ([]entity.TaxiCallRequest, error) {
	db := GetQueryContext(ctx)

	resp := []entity.TaxiCallRequest{}

	// TODO (taekyeom) pagenation, ordering
	err := db.NewSelect().Model(&resp).Where("user_id = ?", userId).Scan(ctx)

	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (t taxiCallRepository) Create(ctx context.Context, taxiCallRequest entity.TaxiCallRequest) error {
	db := GetQueryContext(ctx)

	_, err := db.NewInsert().Model(&taxiCallRequest).Exec(ctx)

	// TODO (taekyeom) handle already exists
	if err != nil {
		return fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return nil
}

func (t taxiCallRepository) Update(ctx context.Context, taxiCallRequest entity.TaxiCallRequest) error {
	db := GetQueryContext(ctx)

	res, err := db.NewUpdate().Model(&taxiCallRequest).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func NewTaxiCallRepository() taxiCallRepository {
	return taxiCallRepository{}
}
