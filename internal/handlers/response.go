package handlers

import (
	"github.com/labstack/echo/v4"
)

// ساختار استاندارد خطا
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// ساختار موفقیت
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// هندلر برای خطا
func RespondError(c echo.Context, code int, err string, details string) error {
	return c.JSON(code, ErrorResponse{
		Error:   err,
		Details: details,
	})
}

// هندلر برای موفقیت
func RespondSuccess(c echo.Context, code int, data interface{}) error {
	return c.JSON(code, SuccessResponse{
		Data: data,
	})
}
