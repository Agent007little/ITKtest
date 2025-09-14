package service

import (
	"context"
	"errors"
	"testing"

	"ITKtest/internal/models"
	"ITKtest/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletService_ProcessWalletOperation(t *testing.T) {
	walletID := uuid.New()
	tests := []struct {
		name          string
		request       models.WalletOperationRequest
		mockSetup     func(*repository.MockWalletRepository)
		expectedError string
	}{
		{
			name: "successful deposit with new wallet",
			request: models.WalletOperationRequest{
				WalletID:      walletID,
				OperationType: models.DEPOSIT,
				Amount:        1000,
			},
			mockSetup: func(m *repository.MockWalletRepository) {
				m.On("GetWallet", mock.Anything, walletID).Return((*models.Wallet)(nil), errors.New("not found"))
				m.On("CreateWallet", mock.Anything, walletID).Return(nil)
				m.On("UpdateWalletBalance", mock.Anything, walletID, int64(1000), models.DEPOSIT).Return(nil)
			},
		},
		{
			name: "successful withdraw with existing wallet",
			request: models.WalletOperationRequest{
				WalletID:      walletID,
				OperationType: models.WITHDRAW,
				Amount:        500,
			},
			mockSetup: func(m *repository.MockWalletRepository) {
				existingWallet := &models.Wallet{ID: walletID, Balance: 1000}
				m.On("GetWallet", mock.Anything, walletID).Return(existingWallet, nil)
				m.On("UpdateWalletBalance", mock.Anything, walletID, int64(500), models.WITHDRAW).Return(nil)
			},
		},
		{
			name: "invalid amount",
			request: models.WalletOperationRequest{
				WalletID:      walletID,
				OperationType: models.DEPOSIT,
				Amount:        -100,
			},
			mockSetup:     func(m *repository.MockWalletRepository) {},
			expectedError: "amount must be positive",
		},
		{
			name: "insufficient funds",
			request: models.WalletOperationRequest{
				WalletID:      walletID,
				OperationType: models.WITHDRAW,
				Amount:        1500,
			},
			mockSetup: func(m *repository.MockWalletRepository) {
				existingWallet := &models.Wallet{ID: walletID, Balance: 1000}
				m.On("GetWallet", mock.Anything, walletID).Return(existingWallet, nil)
				m.On("UpdateWalletBalance", mock.Anything, walletID, int64(1500), models.WITHDRAW).
					Return(errors.New("insufficient funds"))
			},
			expectedError: "insufficient funds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repository.MockWalletRepository{}
			service := NewWalletService(mockRepo)

			tt.mockSetup(mockRepo)

			err := service.ProcessWalletOperation(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_GetWalletBalance(t *testing.T) {
	walletID := uuid.New()
	tests := []struct {
		name           string
		walletID       uuid.UUID
		mockSetup      func(*repository.MockWalletRepository)
		expectedBalance int64
		expectedError  string
	}{
		{
			name:     "successful balance retrieval",
			walletID: walletID,
			mockSetup: func(m *repository.MockWalletRepository) {
				existingWallet := &models.Wallet{ID: walletID, Balance: 2500}
				m.On("GetWallet", mock.Anything, walletID).Return(existingWallet, nil)
			},
			expectedBalance: 2500,
		},
		{
			name:     "wallet not found",
			walletID: walletID,
			mockSetup: func(m *repository.MockWalletRepository) {
				m.On("GetWallet", mock.Anything, walletID).Return((*models.Wallet)(nil), errors.New("not found"))
			},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repository.MockWalletRepository{}
			service := NewWalletService(mockRepo)

			tt.mockSetup(mockRepo)

			balance, err := service.GetWalletBalance(context.Background(), tt.walletID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, int64(0), balance)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}