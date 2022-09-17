package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/taco-labs/taco/go/domain/entity"
	"github.com/taco-labs/taco/go/domain/value"
)

type UserRepository interface {
	FindById(context.Context, string) (entity.User, error)
	FindByUserUniqueKey(context.Context, string) (entity.User, error)
	CreateUser(context.Context, entity.User) error
	UpdateUser(context.Context, entity.User) error
	DeleteUser(context.Context, entity.User) error
}

type userRepository struct{}

func (u userRepository) FindById(ctx context.Context, userId string) (entity.User, error) {
	db := GetQueryContext(ctx)

	user := entity.User{
		Id: userId,
	}

	err := db.NewSelect().Model(&user).WherePK().Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.User{}, value.ErrUserNotFound
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return user, nil
}

func (u userRepository) FindByUserUniqueKey(ctx context.Context, userUniqueKey string) (entity.User, error) {
	db := GetQueryContext(ctx)

	user := entity.User{}

	err := db.NewSelect().Model(&user).Where("user_unique_key = ?", userUniqueKey).Scan(ctx)

	if errors.Is(sql.ErrNoRows, err) {
		return entity.User{}, value.ErrUserNotFound
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("%v: %v", value.ErrDBInternal, err)
	}

	return user, nil
}

func (u userRepository) CreateUser(ctx context.Context, user entity.User) error {
	db := GetQueryContext(ctx)

	res, err := db.NewInsert().Model(&user).Exec(ctx)

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

func (u userRepository) UpdateUser(ctx context.Context, user entity.User) error {
	db := GetQueryContext(ctx)

	res, err := db.NewUpdate().Model(&user).WherePK().Exec(ctx)

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

func (u userRepository) DeleteUser(ctx context.Context, user entity.User) error {
	db := GetQueryContext(ctx)

	res, err := db.NewDelete().Model(&user).WherePK().Exec(ctx)

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

func NewUserRepository() userRepository {
	return userRepository{}
}
