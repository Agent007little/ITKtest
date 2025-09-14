package repository

import (
	"context"

	"ITKtest/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockWalletRepository мок репозитория
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) CreateWallet(ctx context.Context, walletID uuid.UUID) error {
	args := m.Called(ctx, walletID)
	return args.Error(0)
}

func (m *MockWalletRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, amount int64, operationType models.OperationType) error {
	args := m.Called(ctx, walletID, amount, operationType)
	return args.Error(0)
}