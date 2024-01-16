package handlers

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	content, err := os.ReadFile("index.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not open requested file")
	}

	return c.HTML(http.StatusOK, string(content))
}
