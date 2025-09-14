package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ITKtest/internal/models"
	"ITKtest/internal/service"
	"ITKtest/responder"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletController_HandleWalletOperation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*service.MockWalletService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful deposit",
			requestBody: models.WalletOperationRequest{
				WalletID:      uuid.New(),
				OperationType: models.DEPOSIT,
				Amount:        1000,
			},
			mockSetup: func(m *service.MockWalletService) {
				m.On("ProcessWalletOperation", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful withdraw",
			requestBody: models.WalletOperationRequest{
				WalletID:      uuid.New(),
				OperationType: models.WITHDRAW,
				Amount:        500,
			},
			mockSetup: func(m *service.MockWalletService) {
				m.On("ProcessWalletOperation", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *service.MockWalletService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "negative amount",
			requestBody: models.WalletOperationRequest{
				WalletID:      uuid.New(),
				OperationType: models.DEPOSIT,
				Amount:        -100,
			},
			mockSetup:      func(m *service.MockWalletService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Amount must be positive",
		},
		{
			name: "invalid operation type",
			requestBody: map[string]interface{}{
				"walletId":      uuid.New().String(),
				"operationType": "INVALID",
				"amount":        1000,
			},
			mockSetup:      func(m *service.MockWalletService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Operation type must be DEPOSIT or WITHDRAW",
		},
		{
			name: "insufficient funds",
			requestBody: models.WalletOperationRequest{
				WalletID:      uuid.New(),
				OperationType: models.WITHDRAW,
				Amount:        1000,
			},
			mockSetup: func(m *service.MockWalletService) {
				m.On("ProcessWalletOperation", mock.Anything, mock.Anything).
					Return(errors.New("insufficient funds"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Insufficient funds",
		},
		{
			name: "internal server error",
			requestBody: models.WalletOperationRequest{
				WalletID:      uuid.New(),
				OperationType: models.DEPOSIT,
				Amount:        1000,
			},
			mockSetup: func(m *service.MockWalletService) {
				m.On("ProcessWalletOperation", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockWalletService{}
			resp := responder.NewJSONResponder()
			controller := NewWalletController(mockService, resp)

			tt.mockSetup(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.HandleWalletOperation(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "success", response["status"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletController_GetWalletBalance(t *testing.T) {
	walletID := uuid.New()
	tests := []struct {
		name            string
		walletID        string
		mockSetup       func(*service.MockWalletService)
		expectedStatus  int
		expectedError   string
		expectedBalance int64
	}{
		{
			name:     "successful balance retrieval",
			walletID: walletID.String(),
			mockSetup: func(m *service.MockWalletService) {
				m.On("GetWalletBalance", mock.Anything, walletID).Return(int64(2500), nil)
			},
			expectedStatus:  http.StatusOK,
			expectedBalance: 2500,
		},
		{
			name:           "invalid UUID",
			walletID:       "invalid-uuid",
			mockSetup:      func(m *service.MockWalletService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid wallet ID",
		},
		{
			name:     "wallet not found",
			walletID: walletID.String(),
			mockSetup: func(m *service.MockWalletService) {
				m.On("GetWalletBalance", mock.Anything, walletID).
					Return(int64(0), errors.New("wallet not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Wallet not found",
		},
		{
			name:     "internal server error",
			walletID: walletID.String(),
			mockSetup: func(m *service.MockWalletService) {
				m.On("GetWalletBalance", mock.Anything, walletID).
					Return(int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockWalletService{}
			resp := responder.NewJSONResponder()
			controller := NewWalletController(mockService, resp)

			tt.mockSetup(mockService)

			r := chi.NewRouter()
			r.Get("/api/v1/wallets/{walletId}", controller.GetWalletBalance)

			req := httptest.NewRequest("GET", "/api/v1/wallets/"+tt.walletID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response models.WalletBalanceResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, walletID, response.WalletID)
				assert.Equal(t, tt.expectedBalance, response.Balance)
			}

			mockService.AssertExpectations(t)
		})
	}
}
