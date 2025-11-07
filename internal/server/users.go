package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserReq defines the request body for creating a user
type CreateUserReq struct {
	Username string `json:"username" validate:"required"`
	FullName string `json:"full_name,omitempty"`
	Password string `json:"password" validate:"required,min=6"`
	RoleID   int32  `json:"role_id" validate:"required,gt=0"`
}

// UpdateUserReq defines the request body for updating a user
type UpdateUserReq struct {
	FullName string `json:"full_name,omitempty"`
	RoleID   *int32 `json:"role_id,omitempty"`
}

// CreateUser handles POST /api/v1/users
func (s *Server) CreateUser(c echo.Context) error {
	var req CreateUserReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "hash_error", "Failed to hash password.")
	}

	ctx := c.Request().Context()
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		FullName:     sql.NullString{String: req.FullName, Valid: req.FullName != ""},
		PasswordHash: string(hashedPassword),
		RoleID:       sql.NullInt32{Int32: req.RoleID, Valid: true},
	})
	if err != nil {
		// Check for duplicate username
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"" {
			return RespondError(c, http.StatusConflict, "duplicate_username", "A user with this username already exists.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to create user.")
	}

	// Don't return password hash
	user.PasswordHash = ""

	return RespondSuccess(c, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/:id
func (s *Server) GetUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	user, err := s.queries.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "User with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to retrieve user.")
	}

	// Don't return password hash
	user.PasswordHash = ""

	return RespondSuccess(c, http.StatusOK, user)
}

// ListUsers handles GET /api/v1/users
func (s *Server) ListUsers(c echo.Context) error {
	ctx := c.Request().Context()

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := s.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to fetch users.")
	}

	if users == nil {
		users = []db.User{}
	}

	// Remove password hashes from response
	for i := range users {
		users[i].PasswordHash = ""
	}

	return RespondSuccess(c, http.StatusOK, users)
}

// UpdateUser handles PUT /api/v1/users/:id
func (s *Server) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid UUID.")
	}

	var req UpdateUserReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", "The request body is not valid.")
	}

	ctx := c.Request().Context()

	params := db.UpdateUserParams{
		ID: id,
	}

	if req.FullName != "" {
		params.FullName = sql.NullString{String: req.FullName, Valid: true}
	}
	if req.RoleID != nil {
		params.RoleID = sql.NullInt32{Int32: *req.RoleID, Valid: true}
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", "User with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to update user.")
	}

	// Don't return password hash
	user.PasswordHash = ""

	return RespondSuccess(c, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/:id
func (s *Server) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", "The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	err = s.queries.DeleteUser(ctx, id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", "Failed to delete user.")
	}

	return c.NoContent(http.StatusNoContent)
}