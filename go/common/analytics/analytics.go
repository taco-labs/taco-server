package analytics

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

type LogType string

const (
	LogType_UserSignup LogType = "User_Signup"

	LogType_UserTaxiCallRequest LogType = "User_TaxiCallRequest"

	LogType_UserCancelTaxiCallRequest LogType = "User_CancelTaxiCallRequest"

	LogType_UserTaxiCallRequestFailed LogType = "User_taxiCallRequestFailed"

	LogType_UserPaymentDone LogType = "User_PaymentDone"

	LogType_DriverSignup LogType = "Driver_Signup"

	LogType_DriverOnDuty LogType = "Driver_OnDuty"

	LogType_DriverLocation LogType = "Driver_Location"

	LogType_DriverTaxiCallTicketDistribution LogType = "Driver_TaxiCallTicketDistribution"

	LogType_DriverTaxiCallTicketAccept LogType = "Driver_TaxiCallTicketAccept"

	LogType_DriverTaxiCallTicketReject LogType = "Driver_TaxiCallTicketReject"

	LogType_DriverTaxiCallCancel LogType = "Driver_TaxiCallCancel"

	LogType_DriverTaxiToArrival LogType = "Driver_TaxiToArrival"

	LogType_DriverTaxiDone LogType = "Driver_TaxiDone"
)

func WriteAnalyticsLog(ctx context.Context, eventTime time.Time, logType LogType, payload any) {
	// TODO (taekyeom)  중복 방지를 위한 log id 필드 추가 필요
	logger.Info("analytics",
		zap.Time("timestamp", eventTime),
		zap.String("kind", "analytics"),
		zap.String("logType", string(logType)),
		zap.Any("payload", payload),
	)
}

func init() {
	logger, _ = zap.NewProduction(
		zap.IncreaseLevel(zapcore.InfoLevel),
	)
}
