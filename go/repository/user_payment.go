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

type UserPaymentRepository interface {
	// Payment entity
	GetUserPayment(context.Context, bun.IDB, string) (entity.UserPayment, error)
	ListUserPayment(context.Context, bun.IDB, string) ([]entity.UserPayment, error) // TODO (taekyeom) pagination?
	CreateUserPayment(context.Context, bun.IDB, entity.UserPayment) error
	DeleteUserPayment(context.Context, bun.IDB, string) error

	// Default payment entity
	GetDefaultPaymentByUserId(context.Context, bun.IDB, string) (entity.UserDefaultPayment, error)
	UpsertDefaultPayment(context.Context, bun.IDB, entity.UserDefaultPayment) error
}

type userPaymentRepository struct{}

func (u userPaymentRepository) GetUserPayment(ctx context.Context, db bun.IDB, paymentId string) (entity.UserPayment, error) {
	defaultPaymentSubq := db.NewSelect().Model(&entity.UserDefaultPayment{}).Where("payment_id = ?", paymentId)

	userPayment := entity.UserPayment{
		Id: paymentId,
	}

	err := db.NewSelect().
		Model(&userPayment).
		ColumnExpr("user_payment.*").
		ColumnExpr("exists(select 1 from (?) s where s.payment_id = user_payment.id) as default_payment", defaultPaymentSubq).
		WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.UserPayment{}, value.ErrNotFound
	}
	if err != nil {
		return entity.UserPayment{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return userPayment, nil
}

func (u userPaymentRepository) ListUserPayment(ctx context.Context, db bun.IDB, userId string) ([]entity.UserPayment, error) {
	var payments []entity.UserPayment

	defaultPaymentSubq := db.NewSelect().Model(&entity.UserDefaultPayment{UserId: userId}).WherePK()

	err := db.NewSelect().
		Model(&payments).
		ColumnExpr("user_payment.*").
		ColumnExpr("exists(select 1 from (?) s where s.payment_id = user_payment.id) as default_payment", defaultPaymentSubq).
		Where("user_id = ?", userId).Scan(ctx)

	if err != nil {
		return []entity.UserPayment{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return payments, nil
}

func (u userPaymentRepository) CreateUserPayment(ctx context.Context, db bun.IDB, userPayment entity.UserPayment) error {
	res, err := db.NewInsert().Model(&userPayment).ExcludeColumn("default_payment").Exec(ctx)

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

func (u userPaymentRepository) DeleteUserPayment(ctx context.Context, db bun.IDB, userPaymentId string) error {
	res, err := db.NewDelete().Model(&entity.UserPayment{}).WherePK(userPaymentId).Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrUserNotFound
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

func (u userPaymentRepository) GetDefaultPaymentByUserId(ctx context.Context, db bun.IDB, userId string) (entity.UserDefaultPayment, error) {
	userDefaultPayment := entity.UserDefaultPayment{
		UserId: userId,
	}

	err := db.NewSelect().Model(&userDefaultPayment).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.UserDefaultPayment{}, value.ErrUserNotFound
	}
	if err != nil {
		return entity.UserDefaultPayment{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return userDefaultPayment, nil
}

func (u userPaymentRepository) UpsertDefaultPayment(ctx context.Context, db bun.IDB, userDefaultPayment entity.UserDefaultPayment) error {
	_, err := db.NewInsert().
		Model(&userDefaultPayment).
		On("CONFLICT (user_id) DO UPDATE").
		Set("payment_id = EXCLUDED.payment_id").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewUserPaymentRepository() *userPaymentRepository {
	return &userPaymentRepository{}
}
