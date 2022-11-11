package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
	"github.com/uptrace/bun"
)

type SmsVerificationRepository interface {
	FindById(context.Context, bun.IDB, string) (entity.SmsVerification, error)
	Create(context.Context, bun.IDB, entity.SmsVerification) error
	Update(context.Context, bun.IDB, entity.SmsVerification) error
	Delete(context.Context, bun.IDB, entity.SmsVerification) error
}

type smsVerificationRepository struct{}

func (s smsVerificationRepository) FindById(ctx context.Context, db bun.IDB, id string) (entity.SmsVerification, error) {
	smsVerification := entity.SmsVerification{
		Id: id,
	}

	err := db.NewSelect().Model(&smsVerification).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.SmsVerification{}, value.ErrNotFound
	}
	if err != nil {
		return entity.SmsVerification{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return smsVerification, nil
}

func (s smsVerificationRepository) Create(ctx context.Context, db bun.IDB, smsVerification entity.SmsVerification) error {
	res, err := db.NewInsert().Model(&smsVerification).Exec(ctx)

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

func (s smsVerificationRepository) Update(ctx context.Context, db bun.IDB, smsVerification entity.SmsVerification) error {
	res, err := db.NewUpdate().Model(&smsVerification).WherePK().Exec(ctx)

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

func (s smsVerificationRepository) Delete(ctx context.Context, db bun.IDB, smsVerification entity.SmsVerification) error {
	res, err := db.NewDelete().Model(&smsVerification).WherePK().Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrDriverNotFound
	}
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

func NewSmsVerificationRepository() *smsVerificationRepository {
	return &smsVerificationRepository{}
}
