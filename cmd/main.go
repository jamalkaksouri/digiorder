package main

import (
	"net/http"

	"github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
    // ست کردن error handler پیشفرض
e.HTTPErrorHandler = func(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "internal_server_error"

	// بررسی نوع خطا
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		} else {
			msg = http.StatusText(code)
		}
	}

	// جواب استاندارد JSON
	if !c.Response().Committed {
		c.JSON(code, ErrorResponse{
			Error:   msg,
			Details: err.Error(),
		})
	}
}


	database := db.Connect()
	defer database.Close()

	e.POST("/api/v1/products", handlers.NewCreateProductHandler(database, queries))
	e.GET("/api/v1/products", handlers.NewListProductsHandler(queries))

	e.Logger.Fatal(e.Start(":8080"))
}
