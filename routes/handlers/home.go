package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	content, err := os.ReadFile("index.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not open requested file")
	}

	// Read environment variables
	serverPort := os.Getenv("SERVER_PORT")
	webAddress := os.Getenv("WEB_ADDRESS")

	// Replace placeholders in the HTML content
	htmlContent := string(content)
	htmlContent = replacePlaceholder(htmlContent, "{{SERVER_PORT}}", serverPort)
	htmlContent = replacePlaceholder(htmlContent, "{{WEB_ADDRESS}}", webAddress)

	return c.HTML(http.StatusOK, htmlContent)
}

func replacePlaceholder(content, placeholder, value string) string {
	return strings.ReplaceAll(content, placeholder, value)
}
