package enum

type DriverType string

var (
	DriverType_UNKNOWN DriverType = "UNKNOWN"

	DriverType_INDIVIDUAL DriverType = "INDIVIDUAL"

	DriverType_COORPERATE DriverType = "COORPERATE"
)

func DriverTypeFromString(driverTypeStr string) DriverType {
	switch driverTypeStr {
	case string(DriverType_INDIVIDUAL):
		return DriverType_INDIVIDUAL
	case string(DriverType_COORPERATE):
		return DriverType_COORPERATE
	default:
		return DriverType_UNKNOWN
	}
}
