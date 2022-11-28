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

type PaymentRepository interface {
	// Payment registration entity
	GetUserPaymentRegistrationRequest(context.Context, bun.IDB, int) (entity.UserPaymentRegistrationRequest, error)
	CreateUserPaymentRegistrationRequest(context.Context, bun.IDB, entity.UserPaymentRegistrationRequest) (entity.UserPaymentRegistrationRequest, error)
	DeleteUserPaymentRegistrationRequest(context.Context, bun.IDB, entity.UserPaymentRegistrationRequest) error

	// Payment entity
	GetUserPayment(context.Context, bun.IDB, string) (entity.UserPayment, error)
	ListUserPayment(context.Context, bun.IDB, string) ([]entity.UserPayment, error) // TODO (taekyeom) pagination?
	CreateUserPayment(context.Context, bun.IDB, entity.UserPayment) error
	DeleteUserPayment(context.Context, bun.IDB, string) error
	BatchDeleteUserPayment(context.Context, bun.IDB, string) ([]entity.UserPayment, error)
	UpdateUserPayment(context.Context, bun.IDB, entity.UserPayment) error

	// Payment order
	GetPaymentOrder(context.Context, bun.IDB, string) (entity.UserPaymentOrder, error)
	CreatePaymentOrder(context.Context, bun.IDB, entity.UserPaymentOrder) error

	// Failed order
	CreateFailedOrder(context.Context, bun.IDB, entity.UserPaymentFailedOrder) error
	DeleteFailedOrder(context.Context, bun.IDB, entity.UserPaymentFailedOrder) error
	GetFailedOrdersByUserId(context.Context, bun.IDB, string) ([]entity.UserPaymentFailedOrder, error)
}

type userPaymentRepository struct{}

func (u userPaymentRepository) GetUserPaymentRegistrationRequest(ctx context.Context, db bun.IDB, requestId int) (entity.UserPaymentRegistrationRequest, error) {
	resp := entity.UserPaymentRegistrationRequest{
		RequestId: requestId,
	}

	err := db.NewSelect().Model(&resp).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.UserPaymentRegistrationRequest{}, value.ErrNotFound
	}
	if err != nil {
		return entity.UserPaymentRegistrationRequest{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (u userPaymentRepository) CreateUserPaymentRegistrationRequest(ctx context.Context, db bun.IDB, request entity.UserPaymentRegistrationRequest) (entity.UserPaymentRegistrationRequest, error) {
	res, err := db.NewInsert().Model(&request).ExcludeColumn("request_id").Returning("*").Exec(ctx)

	if err != nil {
		return entity.UserPaymentRegistrationRequest{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return entity.UserPaymentRegistrationRequest{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}
	if rowsAffected != 1 {
		return entity.UserPaymentRegistrationRequest{}, fmt.Errorf("%w: invalid rows affected %d", value.ErrDBInternal, rowsAffected)
	}

	return request, nil
}

func (u userPaymentRepository) DeleteUserPaymentRegistrationRequest(ctx context.Context, db bun.IDB, request entity.UserPaymentRegistrationRequest) error {
	res, err := db.NewDelete().Model(&request).WherePK().Exec(ctx)

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

func (u userPaymentRepository) GetUserPayment(ctx context.Context, db bun.IDB, paymentId string) (entity.UserPayment, error) {
	userPayment := entity.UserPayment{
		Id: paymentId,
	}

	err := db.NewSelect().
		Model(&userPayment).
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

	err := db.NewSelect().
		Model(&payments).
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
	res, err := db.NewDelete().Model(&entity.UserPayment{
		Id: userPaymentId,
	}).WherePK().Exec(ctx)

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

func (u userPaymentRepository) BatchDeleteUserPayment(ctx context.Context, db bun.IDB, userId string) ([]entity.UserPayment, error) {
	resp := []entity.UserPayment{}

	_, err := db.NewDelete().Model(&resp).Returning("*").Where("user_id = ?", userId).Exec(ctx)

	if err != nil {
		return []entity.UserPayment{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}

func (u userPaymentRepository) UpdateUserPayment(ctx context.Context, db bun.IDB, userPayment entity.UserPayment) error {
	res, err := db.NewUpdate().Model(&userPayment).ExcludeColumn("default_payment").WherePK().Exec(ctx)

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

func NewUserPaymentRepository() *userPaymentRepository {
	return &userPaymentRepository{}
}

func (u userPaymentRepository) GetPaymentOrder(ctx context.Context, db bun.IDB, orderId string) (entity.UserPaymentOrder, error) {
	userPaymentOrder := entity.UserPaymentOrder{
		OrderId: orderId,
	}

	err := db.NewSelect().Model(&userPaymentOrder).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.UserPaymentOrder{}, value.ErrUserNotFound
	}
	if err != nil {
		return entity.UserPaymentOrder{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return userPaymentOrder, nil

}
func (u userPaymentRepository) CreatePaymentOrder(ctx context.Context, db bun.IDB, userPaymentOrder entity.UserPaymentOrder) error {
	res, err := db.NewInsert().Model(&userPaymentOrder).Exec(ctx)

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

func (u userPaymentRepository) CreateFailedOrder(ctx context.Context, db bun.IDB, failedOrder entity.UserPaymentFailedOrder) error {
	res, err := db.NewInsert().Model(&failedOrder).Exec(ctx)

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

func (u userPaymentRepository) DeleteFailedOrder(ctx context.Context, db bun.IDB, failedOrder entity.UserPaymentFailedOrder) error {
	res, err := db.NewDelete().Model(&failedOrder).WherePK().Exec(ctx)

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

func (u userPaymentRepository) GetFailedOrdersByUserId(ctx context.Context, db bun.IDB, userId string) ([]entity.UserPaymentFailedOrder, error) {
	var resp []entity.UserPaymentFailedOrder

	err := db.NewSelect().Model(&resp).Where("user_id = ?", userId).Scan(ctx)

	if err != nil {
		return []entity.UserPaymentFailedOrder{}, fmt.Errorf("%w: %v", value.ErrDBInternal, err)
	}

	return resp, nil
}
