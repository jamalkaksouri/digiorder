package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	db "github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQueries is a mock implementation of db.Queries
type MockQueries struct {
	mock.Mock
}

func (m *MockQueries) CreateProduct(ctx interface{}, arg db.CreateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockQueries) ListProducts(ctx interface{}, arg db.ListProductsParams) ([]db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockQueries) GetProduct(ctx interface{}, id uuid.UUID) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func TestCreateProduct(t *testing.T) {
	// Setup
	e := echo.New()
	mockDB := &sql.DB{}
	mockQueries := new(MockQueries)
	
	server := &Server{
		db:        mockDB,
		queries:   mockQueries,
		router:    e,
		validator: validator.New(),
	}

	// Test case: Success
	t.Run("Success", func(t *testing.T) {
		productJSON := `{
			"name": "Test Product",
			"brand": "Test Brand",
			"dosage_form_id": 1,
			"strength": "100mg",
			"unit": "tablet",
			"category_id": 1,
			"description": "Test description"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(productJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedProduct := db.Product{
			ID:   uuid.New(),
			Name: "Test Product",
		}

		mockQueries.On("CreateProduct", mock.Anything, mock.Anything).Return(expectedProduct, nil)

		// Execute
		err := server.CreateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response SuccessResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NotNil(t, response.Data)
	})

	// Test case: Validation error
	t.Run("ValidationError", func(t *testing.T) {
		productJSON := `{
			"brand": "Test Brand"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(productJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := server.CreateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	// Test case: Invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := `{"name": "Test`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(invalidJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := server.CreateProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestListProducts(t *testing.T) {
	// Setup
	e := echo.New()
	mockDB := &sql.DB{}
	mockQueries := new(MockQueries)
	
	server := &Server{
		db:        mockDB,
		queries:   mockQueries,
		router:    e,
		validator: validator.New(),
	}

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products?limit=10&offset=0", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedProducts := []db.Product{
			{ID: uuid.New(), Name: "Product 1"},
			{ID: uuid.New(), Name: "Product 2"},
		}

		mockQueries.On("ListProducts", mock.Anything, mock.Anything).Return(expectedProducts, nil)

		// Execute
		err := server.ListProducts(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response SuccessResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NotNil(t, response.Data)
	})
}

func TestGetProduct(t *testing.T) {
	// Setup
	e := echo.New()
	mockDB := &sql.DB{}
	mockQueries := new(MockQueries)
	
	server := &Server{
		db:        mockDB,
		queries:   mockQueries,
		router:    e,
		validator: validator.New(),
	}

	productID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+productID.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())

		expectedProduct := db.Product{
			ID:   productID,
			Name: "Test Product",
		}

		mockQueries.On("GetProduct", mock.Anything, productID).Return(expectedProduct, nil)

		// Execute
		err := server.GetProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+productID.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(productID.String())

		mockQueries.On("GetProduct", mock.Anything, productID).Return(db.Product{}, sql.ErrNoRows)

		// Execute
		err := server.GetProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products/invalid-id", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("invalid-id")

		// Execute
		err := server.GetProduct(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}