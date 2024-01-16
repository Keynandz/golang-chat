package routes

import (
	"go-chat/routes/handlers"
	"github.com/labstack/echo/v4"
)

func EndpointRoutes(e *echo.Echo) {
	e.GET("/", handlers.Home)
	e.GET("/ws", handlers.HandleWebSocket)
}