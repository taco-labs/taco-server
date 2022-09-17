package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type UserPaymentRepository interface {
	// Payment entity
	GetUserPayment(context.Context, string) (entity.UserPayment, error)
	ListUserPayment(context.Context, string) ([]entity.UserPayment, error) // TODO (taekyeom) pagination?
	CreateUserPayment(context.Context, entity.UserPayment) error
	DeleteUserPayment(context.Context, string) error

	// Default payment entity
	GetDefaultPaymentByUserId(context.Context, string) (entity.UserDefaultPayment, error)
	UpsertDefaultPayment(context.Context, entity.UserDefaultPayment) error
}

type userPaymentRepository struct{}

func (u userPaymentRepository) GetUserPayment(ctx context.Context, paymentId string) (entity.UserPayment, error) {
	db := GetQueryContext(ctx)

	userPayment := entity.UserPayment{
		Id: paymentId,
	}

	err := db.NewSelect().Model(&userPayment).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.UserPayment{}, value.ErrUserNotFound
	}
	if err != nil {
		return entity.UserPayment{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return userPayment, nil
}

func (u userPaymentRepository) ListUserPayment(ctx context.Context, userId string) ([]entity.UserPayment, error) {
	db := GetQueryContext(ctx)

	payments := []entity.UserPayment{}

	err := db.NewSelect().Model(&payments).Where("user_id = ?", userId).Scan(ctx)

	if err != nil {
		return []entity.UserPayment{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return payments, nil
}

func (u userPaymentRepository) CreateUserPayment(ctx context.Context, userPayment entity.UserPayment) error {
	db := GetQueryContext(ctx)

	res, err := db.NewInsert().Model(&userPayment).Exec(ctx)

	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%v: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (u userPaymentRepository) DeleteUserPayment(ctx context.Context, userPaymentId string) error {
	db := GetQueryContext(ctx)

	res, err := db.NewDelete().Model(&entity.UserPayment{}).WherePK(userPaymentId).Exec(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return value.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%v: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return nil
}

func (u userPaymentRepository) GetDefaultPaymentByUserId(ctx context.Context, userId string) (entity.UserDefaultPayment, error) {
	db := GetQueryContext(ctx)

	userDefaultPayment := entity.UserDefaultPayment{
		UserId: userId,
	}

	err := db.NewSelect().Model(&userDefaultPayment).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.UserDefaultPayment{}, value.ErrUserNotFound
	}
	if err != nil {
		return entity.UserDefaultPayment{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return userDefaultPayment, nil
}

func (u userPaymentRepository) UpsertDefaultPayment(ctx context.Context, userDefaultPayment entity.UserDefaultPayment) error {
	db := GetQueryContext(ctx)

	_, err := db.NewInsert().
		Model(&userDefaultPayment).
		On("CONFLICT (user_id) UPDATE").
		Set("payment_id = EXCLUDED.payment_id").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return nil
}

func NewUserPaymentRepository() userPaymentRepository {
	return userPaymentRepository{}
}
