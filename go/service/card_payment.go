package service

import (
	"context"

	"github.com/ktk1012/taco/go/domain/entity"
	"github.com/ktk1012/taco/go/domain/value"
)

type CardPaymentService interface {
	RegisterCard(context.Context, entity.UserPayment) (string, error)
	Transaction(context.Context, entity.UserPayment, value.Payment) error // TODO(taekyeom) 결제 기록 별도 보관 필요
}
