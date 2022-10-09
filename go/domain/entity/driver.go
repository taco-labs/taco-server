package entity

import (
	"time"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"github.com/twpayne/go-geom"
	"github.com/uptrace/bun"
)

type Driver struct {
	DriverDto
	DriverLocationDto
}

type DriverDto struct {
	bun.BaseModel `bun:"table:driver"`

	Id                    string          `bun:"id,pk"`
	DriverType            enum.DriverType `bun:"driver_type"`
	FirstName             string          `bun:"first_name"`
	LastName              string          `bun:"last_name"`
	BirthDay              string          `bun:"birthday"`
	Phone                 string          `bun:"phone"`
	Gender                string          `bun:"gender"`
	AppOs                 enum.OsType     `bun:"app_os"`
	AppVersion            string          `bun:"app_version"`
	AppFcmToken           string          `bun:"app_fcm_token"`
	UserUniqueKey         string          `bun:"user_unique_key"`
	DriverLicenseId       string          `bun:"driver_license_id"`
	DriverLicenseImageUrl string          `bun:"driver_license_image_url"`
	Active                bool            `bun:"active"`
	CreateTime            time.Time       `bun:"create_time"`
	UpdateTime            time.Time       `bun:"update_time"`
	DeleteTime            time.Time       `bun:"delete_time"`
}

type DriverSettlementAccount struct {
	bun.BaseModel `bun:"table:driver_settlement_account"`

	DriverId      string    `bun:"driver_id,pk"`
	Bank          string    `bun:"bank"` // TODO(taekyeom) maybe enum?
	AccountNumber string    `bun:"account_number"`
	CreateTime    time.Time `bun:"create_time"`
	UpdateTime    time.Time `bun:"update_time"`
}

type DriverLocationDto struct {
	bun.BaseModel `bun:"driver_location"`

	Location *geom.Point `bun:"location"`
	DriverId string      `bun:"driver_id,pk"`
	OnDuty   bool        `bun:"on_duty"`
}

func NewDriverLocation(driverId string, latitude float64, longitude float64, onDuty bool) DriverLocationDto {
	point := geom.NewPoint(geom.XY).
		MustSetCoords([]float64{longitude, latitude}).
		SetSRID(value.SRID_SPHERE)

	return DriverLocationDto{
		DriverId: driverId,
		Location: point,
		OnDuty:   onDuty,
	}
}

func NewEmptyDriverLocation(driverId string) DriverLocationDto {
	point := geom.NewPointEmpty(geom.XY).SetSRID(value.SRID_SPHERE)
	return DriverLocationDto{
		DriverId: driverId,
		Location: point,
		OnDuty:   false,
	}
}
