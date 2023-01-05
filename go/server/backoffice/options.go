package backoffice

import "github.com/labstack/echo/v4"

type backofficeOption func(*backofficeServer)

func WithEndpoint(endpoint string) backofficeOption {
	return func(bs *backofficeServer) {
		bs.endpoint = endpoint
	}
}

func WithPort(port int) backofficeOption {
	return func(bs *backofficeServer) {
		bs.port = port
	}
}

func WithDriverApp(driverApp driverApp) backofficeOption {
	return func(bs *backofficeServer) {
		bs.app.driver = driverApp
	}
}

func WithTaxicallApp(taxicallApp taxicallApp) backofficeOption {
	return func(bs *backofficeServer) {
		bs.app.taxicall = taxicallApp
	}
}

func WithUserApp(userApp userApp) backofficeOption {
	return func(bs *backofficeServer) {
		bs.app.user = userApp
	}
}

func WithMiddleware(middleware echo.MiddlewareFunc) backofficeOption {
	return func(bs *backofficeServer) {
		bs.middlewares = append(bs.middlewares, middleware)
	}
}
