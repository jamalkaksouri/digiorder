// internal/server/audit.go - Complete Audit Logging System

package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
	"github.com/sqlc-dev/pqtype"
)

// AuditLogFilter for querying audit logs
type AuditLogFilter struct {
	UserID     string `query:"user_id"`
	EntityType string `query:"entity_type"`
	EntityID   string `query:"entity_id"`
	Action     string `query:"action"`
	StartDate  string `query:"start_date"`
	EndDate    string `query:"end_date"`
	Limit      int    `query:"limit"`
	Offset     int    `query:"offset"`
}

// logAudit creates an audit log entry
func (s *Server) logAudit(ctx context.Context, userID uuid.UUID, action, entityType, entityID string, 
	oldValues, newValues map[string]interface{}, ipAddress, userAgent string) error {
	
	var oldJSON, newJSON pqtype.NullRawMessage

	if oldValues != nil {
		data, err := json.Marshal(oldValues)
		if err == nil {
			oldJSON = pqtype.NullRawMessage{
				RawMessage: data,
				Valid:      true,
			}
		}
	}

	if newValues != nil {
		data, err := json.Marshal(newValues)
		if err == nil {
			newJSON = pqtype.NullRawMessage{
				RawMessage: data,
				Valid:      true,
			}
		}
	}

	_, err := s.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     uuid.NullUUID{UUID: userID, Valid: true},
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldJSON,
		NewValues:  newJSON,
		IpAddress:  sql.NullString{String: ipAddress, Valid: true},
		UserAgent:  sql.NullString{String: userAgent, Valid: true},
	})

	return err
}

// GetAuditLogs handles GET /api/v1/audit-logs
func (s *Server) GetAuditLogs(c echo.Context) error {
	var filter AuditLogFilter
	if err := c.Bind(&filter); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request", 
			"Invalid query parameters.")
	}

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	ctx := c.Request().Context()

	var logs []db.AuditLog
	var err error

	// Apply filters
	if filter.UserID != "" {
		userID, err := uuid.Parse(filter.UserID)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_user_id", 
				"Invalid user ID format.")
		}
		logs, err = s.queries.GetAuditLogsByUser(ctx, db.GetAuditLogsByUserParams{
			UserID: uuid.NullUUID{UUID: userID, Valid: true},
			Limit:  int32(filter.Limit),
			Offset: int32(filter.Offset),
		})
	} else if filter.EntityType != "" && filter.EntityID != "" {
		logs, err = s.queries.GetAuditLogsByEntity(ctx, db.GetAuditLogsByEntityParams{
			EntityType: filter.EntityType,
			EntityID:   filter.EntityID,
			Limit:      int32(filter.Limit),
			Offset:     int32(filter.Offset),
		})
	} else if filter.Action != "" {
		logs, err = s.queries.GetAuditLogsByAction(ctx, db.GetAuditLogsByActionParams{
			Action: filter.Action,
			Limit:  int32(filter.Limit),
			Offset: int32(filter.Offset),
		})
	} else {
		logs, err = s.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{
			Limit:  int32(filter.Limit),
			Offset: int32(filter.Offset),
		})
	}

	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", 
			"Failed to retrieve audit logs.")
	}

	if logs == nil {
		logs = []db.AuditLog{}
	}

	// Enrich with user information
	enrichedLogs := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		enriched := map[string]interface{}{
			"id":          log.ID,
			"user_id":     log.UserID.UUID,
			"action":      log.Action,
			"entity_type": log.EntityType,
			"entity_id":   log.EntityID,
			"old_values":  json.RawMessage(log.OldValues.RawMessage),
			"new_values":  json.RawMessage(log.NewValues.RawMessage),
			"ip_address":  log.IpAddress.String,
			"user_agent":  log.UserAgent.String,
			"created_at":  log.CreatedAt,
		}

		// Get username if available
		if log.UserID.Valid {
			user, err := s.queries.GetUser(ctx, log.UserID.UUID)
			if err == nil {
				enriched["username"] = user.Username
			}
		}

		enrichedLogs[i] = enriched
	}

	return RespondSuccess(c, http.StatusOK, enrichedLogs)
}

// GetAuditLog handles GET /api/v1/audit-logs/:id
func (s *Server) GetAuditLog(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id", 
			"The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	log, err := s.queries.GetAuditLog(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found", 
				"Audit log with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error", 
			"Failed to retrieve audit log.")
	}

	enriched := map[string]interface{}{
		"id":          log.ID,
		"user_id":     log.UserID.UUID,
		"action":      log.Action,
		"entity_type": log.EntityType,
		"entity_id":   log.EntityID,
		"old_values":  json.RawMessage(log.OldValues.RawMessage),
		"new_values":  json.RawMessage(log.NewValues.RawMessage),
		"ip_address":  log.IpAddress.String,
		"user_agent":  log.UserAgent.String,
		"created_at":  log.CreatedAt,
	}

	// Get username
	if log.UserID.Valid {
		user, err := s.queries.GetUser(ctx, log.UserID.UUID)
		if err == nil {
			enriched["username"] = user.Username
		}
	}

	return RespondSuccess(c, http.StatusOK, enriched)
}

// GetUserActivity handles GET /api/v1/users/:user_id/activity
func (s *Server) GetUserActivity(c echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_user_id", 
			"The provided user ID is not a valid UUID.")
	}

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

	ctx := c.Request().Context()
	logs, err := s.queries.GetAuditLogsByUser(ctx, db.GetAuditLogsByUserParams{
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", 
			"Failed to retrieve user activity.")
	}

	if logs == nil {
		logs = []db.AuditLog{}
	}

	// Format response
	activities := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		activities[i] = map[string]interface{}{
			"id":          log.ID,
			"action":      log.Action,
			"entity_type": log.EntityType,
			"entity_id":   log.EntityID,
			"ip_address":  log.IpAddress.String,
			"created_at":  log.CreatedAt,
		}
	}

	return RespondSuccess(c, http.StatusOK, activities)
}

// GetEntityHistory handles GET /api/v1/audit-logs/entity/:type/:id
func (s *Server) GetEntityHistory(c echo.Context) error {
	entityType := c.Param("type")
	entityID := c.Param("id")

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

	ctx := c.Request().Context()
	logs, err := s.queries.GetAuditLogsByEntity(ctx, db.GetAuditLogsByEntityParams{
		EntityType: entityType,
		EntityID:   entityID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", 
			"Failed to retrieve entity history.")
	}

	if logs == nil {
		logs = []db.AuditLog{}
	}

	// Format response with user information
	history := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		h := map[string]interface{}{
			"id":         log.ID,
			"action":     log.Action,
			"old_values": json.RawMessage(log.OldValues.RawMessage),
			"new_values": json.RawMessage(log.NewValues.RawMessage),
			"ip_address": log.IpAddress.String,
			"created_at": log.CreatedAt,
		}

		// Add username
		if log.UserID.Valid {
			user, err := s.queries.GetUser(ctx, log.UserID.UUID)
			if err == nil {
				h["username"] = user.Username
				h["user_id"] = user.ID
			}
		}

		history[i] = h
	}

	return RespondSuccess(c, http.StatusOK, history)
}

// GetAuditStats handles GET /api/v1/audit-logs/stats
func (s *Server) GetAuditStats(c echo.Context) error {
	ctx := c.Request().Context()

	// Get statistics
	stats, err := s.queries.GetAuditLogStats(ctx)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error", 
			"Failed to retrieve audit statistics.")
	}

	return RespondSuccess(c, http.StatusOK, stats)
}