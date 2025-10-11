package server

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateDosageFormReq defines the request body for creating a new dosage form.
type CreateDosageFormReq struct {
	Name string `json:"name"`
}

// NewCreateDosageFormHandler creates a handler for POST /api/v1/dosage_forms.
func NewCreateDosageFormHandler(queries *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req CreateDosageFormReq
		if err := c.Bind(&req); err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
		}

		// Basic validation
		if req.Name == "" {
			return RespondError(c, http.StatusBadRequest, "validation_error", "Dosage form name cannot be empty.")
		}

		ctx := c.Request().Context()
		dosageForm, err := queries.CreateDosageForm(ctx, req.Name)
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create dosage form.")
		}

		return RespondSuccess(c, http.StatusCreated, dosageForm)
	}
}

// NewListDosageFormsHandler creates a handler for GET /api/v1/dosage_forms.
func NewListDosageFormsHandler(queries *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		dosageForms, err := queries.ListDosageForms(ctx)
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve dosage forms.")
		}

		if dosageForms == nil {
			dosageForms = []db.DosageForm{}
		}

		return RespondSuccess(c, http.StatusOK, dosageForms)
	}
}

// NewGetDosageFormHandler creates a handler for GET /api/v1/dosage_forms/:id.
func NewGetDosageFormHandler(queries *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
		}

		ctx := c.Request().Context()
		dosageForm, err := queries.GetDosageForm(ctx, int32(id))
		if err != nil {
			if err == sql.ErrNoRows {
				return RespondError(c, http.StatusNotFound, "not_found", "Dosage form with the specified ID was not found.")
			}
			return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve dosage form.")
		}

		return RespondSuccess(c, http.StatusOK, dosageForm)
	}
}