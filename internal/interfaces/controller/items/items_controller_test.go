package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"Aicon-assignment/internal/domain/entity"
	domainErrors "Aicon-assignment/internal/domain/errors"
	"Aicon-assignment/internal/usecase"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockItemUsecase is a mock implementation of ItemUsecase for testing
type MockItemUsecase struct {
	mock.Mock
}

func (m *MockItemUsecase) GetAllItems(ctx context.Context) ([]*entity.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) GetItemByID(ctx context.Context, id int64) (*entity.Item, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) CreateItem(ctx context.Context, input usecase.CreateItemInput) (*entity.Item, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) PartialUpdateItem(ctx context.Context, id int64, input usecase.UpdateItemInput) (*entity.Item, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) DeleteItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemUsecase) GetCategorySummary(ctx context.Context) (*usecase.CategorySummary, error) {
	args := m.Called(ctx)
	return args.Get(0).(*usecase.CategorySummary), args.Error(1)
}

func TestItemHandler_PatchItem(t *testing.T) {
	e := echo.New()

	t.Run("Successfully update item name", func(t *testing.T) {
		mockUsecase := new(MockItemUsecase)
		handler := NewItemHandler(mockUsecase)

		itemID := int64(1)
		updateInput := usecase.UpdateItemInput{
			Name: stringPtr("Updated Item Name"),
		}

		expectedItem := &entity.Item{
			ID:       itemID,
			Name:     "Updated Item Name",
			Category: "時計",
			Brand:    "ROLEX",
		}

		mockUsecase.On("PartialUpdateItem", mock.Anything, itemID, updateInput).Return(expectedItem, nil)

		requestBody, _ := json.Marshal(updateInput)
		req := httptest.NewRequest(http.MethodPatch, "/items/1", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/items/:id")
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(itemID, 10))

		err := handler.PatchItem(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response entity.Item
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "Updated Item Name", response.Name)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("Successfully update item brand and price", func(t *testing.T) {
		mockUsecase := new(MockItemUsecase)
		handler := NewItemHandler(mockUsecase)

		itemID := int64(2)
		updateInput := usecase.UpdateItemInput{
			Brand:         stringPtr("Updated Brand"),
			PurchasePrice: intPtr(2000000),
		}

		expectedItem := &entity.Item{
			ID:            itemID,
			Name:          "ロレックス デイトナ",
			Category:      "時計",
			Brand:         "Updated Brand",
			PurchasePrice: 2000000,
		}

		mockUsecase.On("PartialUpdateItem", mock.Anything, itemID, updateInput).Return(expectedItem, nil)

		requestBody, _ := json.Marshal(updateInput)
		req := httptest.NewRequest(http.MethodPatch, "/items/2", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/items/:id")
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(itemID, 10))

		err := handler.PatchItem(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response entity.Item
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "Updated Brand", response.Brand)
		assert.Equal(t, 2000000, response.PurchasePrice)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("Item not found", func(t *testing.T) {
		mockUsecase := new(MockItemUsecase)
		handler := NewItemHandler(mockUsecase)

		itemID := int64(999)
		updateInput := usecase.UpdateItemInput{
			Name: stringPtr("Non-existent Item"),
		}

		mockUsecase.On("PartialUpdateItem", mock.Anything, itemID, updateInput).Return((*entity.Item)(nil), domainErrors.ErrItemNotFound)

		requestBody, _ := json.Marshal(updateInput)
		req := httptest.NewRequest(http.MethodPatch, "/items/999", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/items/:id")
		c.SetParamNames("id")
		c.SetParamValues(strconv.FormatInt(itemID, 10))

		err := handler.PatchItem(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		var response ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "item not found", response.Error)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("Invalid item ID", func(t *testing.T) {
		mockUsecase := new(MockItemUsecase)
		handler := NewItemHandler(mockUsecase)

		req := httptest.NewRequest(http.MethodPatch, "/items/invalid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/items/:id")
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		err := handler.PatchItem(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "invalid item ID", response.Error)
	})

	t.Run("No fields provided for update", func(t *testing.T) {
		mockUsecase := new(MockItemUsecase)
		handler := NewItemHandler(mockUsecase)

		updateInput := usecase.UpdateItemInput{} // 空の入力

		requestBody, _ := json.Marshal(updateInput)
		req := httptest.NewRequest(http.MethodPatch, "/items/1", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/items/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := handler.PatchItem(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Contains(t, response.Error, "at least one field")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockUsecase := new(MockItemUsecase)
		handler := NewItemHandler(mockUsecase)

		req := httptest.NewRequest(http.MethodPatch, "/items/1", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/items/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := handler.PatchItem(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response ErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "invalid request format", response.Error)
	})
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
