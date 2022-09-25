package enum

type OsType string

var (
	OsType_UNKNOWN OsType = "UNKNOWN"
	OSType_WEB     OsType = "WEB"
	OsType_IOS     OsType = "IOS"
	OsType_AOS     OsType = "AOS"
)

func OsTypeFromString(osTypeStr string) OsType {
	switch osTypeStr {
	case string(OSType_WEB):
		return OSType_WEB
	case string(OsType_IOS):
		return OsType_IOS
	case string(OsType_AOS):
		return OsType_AOS
	default:
		return OsType_UNKNOWN
	}
}
