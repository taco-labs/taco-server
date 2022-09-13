package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/value"
)

type UserPaymentRepository interface {
	GetUserPayment(context.Context, string) (entity.UserPayment, error)
	ListUserPayment(context.Context, string) ([]entity.UserPayment, error) // TODO (taekyeom) pagination?
	CreateUserPayment(context.Context, entity.UserPayment) error
	DeleteUserPayment(context.Context, string) error
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

func NewUserPaymentRepository() userPaymentRepository {
	return userPaymentRepository{}
}
