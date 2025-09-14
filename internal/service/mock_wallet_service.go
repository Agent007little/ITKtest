package service

import (
	"context"

	"ITKtest/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockWalletService мок сервиса
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) ProcessWalletOperation(ctx context.Context, req models.WalletOperationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockWalletService) GetWalletBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}