package driver

import "github.com/labstack/echo/v4"

type driverServerOption func(*driverServer)

func WithEndpoint(endpoint string) driverServerOption {
	return func(ds *driverServer) {
		ds.endpoint = endpoint
	}
}

func WithPort(port int) driverServerOption {
	return func(ds *driverServer) {
		ds.port = port
	}
}

func WithDriverApp(driverApp driverApp) driverServerOption {
	return func(ds *driverServer) {
		ds.app.driver = driverApp
	}
}

func WithMiddleware(middleware echo.MiddlewareFunc) driverServerOption {
	return func(ds *driverServer) {
		ds.middlewares = append(ds.middlewares, middleware)
	}
}
