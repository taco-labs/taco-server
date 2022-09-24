package repository

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type TaxiCallRepository interface {
	// GetById(context.Context, string) (entity.TaxiCallRequest, error)
	// GetLatestByUserId(context.Context, string) (entity.TaxiCallRequest, error)
	// GetLatestByDriverId(context.Context, string) (entity.TaxiCallRequest, error)
	ListByUserId(context.Context, string) ([]entity.TaxiCallRequest, error)
	// ListByDriverId(context.Context, string) ([]entity.TaxiCallRequest, error)
}

type taxiCallRepository struct{}

func (t taxiCallRepository) ListByUserId(ctx context.Context, userId string) ([]entity.TaxiCallRequest, error) {
	db := GetQueryContext(ctx)

	resp := []entity.TaxiCallRequest{}

	// TODO (taekyeom) pagenation, ordering
	err := db.NewSelect().Model(&resp).Relation("CallHistory").Scan(ctx)

	if err != nil {
		return resp, fmt.Errorf("%w: error from db: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func NewTaxiCallRepository() taxiCallRepository {
	return taxiCallRepository{}
}
