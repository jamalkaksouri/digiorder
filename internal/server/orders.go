// internal/server/orders.go - FIXED VERSION
package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/labstack/echo/v4"
)

// CreateOrderReq defines the request body for creating an order
type CreateOrderReq struct {
	CreatedBy string `json:"created_by,omitempty"`
	Status    string `json:"status" validate:"required"`
	Notes     string `json:"notes,omitempty"`
}

// UpdateOrderStatusReq defines the request for updating order status
type UpdateOrderStatusReq struct {
	Status string `json:"status" validate:"required"`
}

// CreateOrderItemReq defines the request for creating an order item
// FIXED: Unit is now optional - will auto-populate from product
type CreateOrderItemReq struct {
	ProductID    string `json:"product_id" validate:"required"`
	RequestedQty int32  `json:"requested_qty" validate:"required,gt=0"`
	Unit         string `json:"unit,omitempty"` // Optional - auto-filled from product
	Note         string `json:"note,omitempty"`
}

// UpdateOrderItemReq defines the request for updating an order item
type UpdateOrderItemReq struct {
	RequestedQty int32  `json:"requested_qty" validate:"required,gt=0"`
	Unit         string `json:"unit,omitempty"`
	Note         string `json:"note,omitempty"`
}

// CreateOrder handles POST /api/v1/orders
func (s *Server) CreateOrder(c echo.Context) error {
	var req CreateOrderReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()

	params := db.CreateOrderParams{
		Status: req.Status,
		Notes:  sql.NullString{String: req.Notes, Valid: req.Notes != ""},
	}

	if req.CreatedBy != "" {
		createdByUUID, err := uuid.Parse(req.CreatedBy)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_user_id",
				"Created by user ID is not a valid UUID.")
		}
		params.CreatedBy = uuid.NullUUID{UUID: createdByUUID, Valid: true}
	}

	order, err := s.queries.CreateOrder(ctx, params)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to create order.")
	}

	return RespondSuccess(c, http.StatusCreated, order)
}

// GetOrder handles GET /api/v1/orders/:id
func (s *Server) GetOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	order, err := s.queries.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Order with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve order.")
	}

	return RespondSuccess(c, http.StatusOK, order)
}

// ListOrders handles GET /api/v1/orders
func (s *Server) ListOrders(c echo.Context) error {
	ctx := c.Request().Context()

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")
	userID := c.QueryParam("user_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var orders []db.Order

	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return RespondError(c, http.StatusBadRequest, "invalid_user_id",
				"The provided user ID is not a valid UUID.")
		}

		orders, err = s.queries.ListOrdersByUser(ctx, db.ListOrdersByUserParams{
			CreatedBy: uuid.NullUUID{UUID: userUUID, Valid: true},
			Limit:     int32(limit),
			Offset:    int32(offset),
		})
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, "db_error",
				"Failed to fetch orders.")
		}
	} else {
		orders, err = s.queries.ListOrders(ctx, db.ListOrdersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return RespondError(c, http.StatusInternalServerError, "db_error",
				"Failed to fetch orders.")
		}
	}

	if orders == nil {
		orders = []db.Order{}
	}

	return RespondSuccess(c, http.StatusOK, orders)
}

// UpdateOrderStatus handles PUT /api/v1/orders/:id/status
func (s *Server) UpdateOrderStatus(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	var req UpdateOrderStatusReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	order, err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     id,
		Status: req.Status,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Order with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to update order status.")
	}

	return RespondSuccess(c, http.StatusOK, order)
}

// DeleteOrder handles DELETE /api/v1/orders/:id
func (s *Server) DeleteOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	err = s.queries.DeleteOrder(ctx, id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to delete order.")
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateOrderItem handles POST /api/v1/orders/:order_id/items
// FIXED: Auto-populates unit from product, prevents duplicates
func (s *Server) CreateOrderItem(c echo.Context) error {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_order_id",
			"The provided order ID is not a valid UUID.")
	}

	var req CreateOrderItemReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_product_id",
			"The provided product ID is not a valid UUID.")
	}

	ctx := c.Request().Context()

	// FIXED: Get product to auto-populate unit
	product, err := s.queries.GetProduct(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "product_not_found",
				"Product with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to retrieve product.")
	}

	// FIXED: Check if product already exists in this order
	existingItems, err := s.queries.GetOrderItems(ctx, uuid.NullUUID{UUID: orderID, Valid: true})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to check existing order items.")
	}

	for _, item := range existingItems {
		if item.ProductID.UUID == productID {
			return RespondError(c, http.StatusConflict, "product_already_in_order",
				"This product already exists in the order. Please update its quantity instead of adding it again.")
		}
	}

	// FIXED: Use product unit if not provided
	unit := req.Unit
	if unit == "" && product.Unit.Valid {
		unit = product.Unit.String
	}

	orderItem, err := s.queries.CreateOrderItem(ctx, db.CreateOrderItemParams{
		OrderID:      uuid.NullUUID{UUID: orderID, Valid: true},
		ProductID:    uuid.NullUUID{UUID: productID, Valid: true},
		RequestedQty: req.RequestedQty,
		Unit:         sql.NullString{String: unit, Valid: unit != ""},
		Note:         sql.NullString{String: req.Note, Valid: req.Note != ""},
	})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to create order item.")
	}

	return RespondSuccess(c, http.StatusCreated, orderItem)
}

// GetOrderItems handles GET /api/v1/orders/:order_id/items
func (s *Server) GetOrderItems(c echo.Context) error {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_order_id",
			"The provided order ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	items, err := s.queries.GetOrderItems(ctx, uuid.NullUUID{UUID: orderID, Valid: true})
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to fetch order items.")
	}

	if items == nil {
		items = []db.OrderItem{}
	}

	return RespondSuccess(c, http.StatusOK, items)
}

// UpdateOrderItem handles PUT /api/v1/order_items/:id
func (s *Server) UpdateOrderItem(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	var req UpdateOrderItemReq
	if err := c.Bind(&req); err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_request",
			"The request body is not valid.")
	}

	if err := s.validator.Struct(req); err != nil {
		return RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
	}

	ctx := c.Request().Context()
	orderItem, err := s.queries.UpdateOrderItem(ctx, db.UpdateOrderItemParams{
		ID:           id,
		RequestedQty: req.RequestedQty,
		Unit:         sql.NullString{String: req.Unit, Valid: req.Unit != ""},
		Note:         sql.NullString{String: req.Note, Valid: req.Note != ""},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return RespondError(c, http.StatusNotFound, "not_found",
				"Order item with the specified ID was not found.")
		}
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to update order item.")
	}

	return RespondSuccess(c, http.StatusOK, orderItem)
}

// DeleteOrderItem handles DELETE /api/v1/order_items/:id
func (s *Server) DeleteOrderItem(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RespondError(c, http.StatusBadRequest, "invalid_id",
			"The provided ID is not a valid UUID.")
	}

	ctx := c.Request().Context()
	err = s.queries.DeleteOrderItem(ctx, id)
	if err != nil {
		return RespondError(c, http.StatusInternalServerError, "db_error",
			"Failed to delete order item.")
	}

	return c.NoContent(http.StatusNoContent)
}
