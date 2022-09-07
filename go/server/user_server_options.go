package server

import "github.com/labstack/echo/v4"

type userServerOption func(us *userServer)

func WithEndpoint(endpoint string) userServerOption {
	return func(us *userServer) {
		us.endpoint = endpoint
	}
}

func WithPort(port int) userServerOption {
	return func(us *userServer) {
		us.port = port
	}
}

func WithUserApp(userApp UserApp) userServerOption {
	return func(us *userServer) {
		us.app.user = userApp
	}
}

func WithMiddleware(middleware echo.MiddlewareFunc) userServerOption {
	return func(us *userServer) {
		us.middlewares = append(us.middlewares, middleware)
	}
}
