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
	Name string `json:"name" validate:"required"`
}

// CreateDosageForm handles POST /api/v1/dosage_forms
func (s *Server) CreateDosageForm(c echo.Context) error {
	var req CreateDosageFormReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	dosageForm, err := s.queries.CreateDosageForm(ctx, req.Name)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create dosage form.")
	}

	return RespondSuccess(c, http.StatusCreated, dosageForm)
}

// ListDosageForms handles GET /api/v1/dosage_forms
func (s *Server) ListDosageForms(c echo.Context) error {
	ctx := c.Request().Context()
	dosageForms, err := s.queries.ListDosageForms(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve dosage forms.")
	}

	if dosageForms == nil {
		dosageForms = []db.DosageForm{}
	}

	return RespondSuccess(c, http.StatusOK, dosageForms)
}

// GetDosageForm handles GET /api/v1/dosage_forms/:id
func (s *Server) GetDosageForm(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid number.")
	}

	ctx := c.Request().Context()
	dosageForm, err := s.queries.GetDosageForm(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "Dosage form with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve dosage form.")
	}

	return RespondSuccess(c, http.StatusOK, dosageForm)
}