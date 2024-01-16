package main

import (
	"fmt"
	"go-chat/pkg/database"
	history "go-chat/pkg/log"
	"go-chat/routes"
	"net/http"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	database.InitDB()
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatal("error loading file .env")
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Auth-Token"},
	}))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(history.LogRequest)

	routes.EndpointRoutes(e)

	format := "0.0.0.0:%s"
	port := fmt.Sprintf(format, os.Getenv("WEB_PORT"))
	e.Start(port)
}
