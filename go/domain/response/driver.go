package response

import (
	"github.com/taco-labs/taco/go/domain/entity"
)

type DriverResponse struct {
	Id         string `json:"id"`
	DriverType string `json:"driverType"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	BirthDay   string `json:"birthday"`
	Phone      string `json:"phone"`
	Gender     string `json:"gender"`
	AppOs      string `json:"appOs"`
	AppVersion string `json:"osVersion"`
}

type DriverSignupResponse struct {
	Token  string         `json:"token"`
	Driver DriverResponse `json:"driver"`
}

func DriverToResponse(driver entity.Driver) DriverResponse {
	return DriverResponse{
		Id:         driver.Id,
		DriverType: string(driver.DriverType),
		FirstName:  driver.FirstName,
		LastName:   driver.LastName,
		BirthDay:   driver.BirthDay,
		Phone:      driver.Phone,
		Gender:     driver.Gender,
		AppOs:      string(driver.AppOs),
		AppVersion: driver.AppVersion,
	}
}
