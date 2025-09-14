package service

import (
	"context"
	"errors"

	"ITKtest/internal/models"
	"ITKtest/internal/repository"

	"github.com/google/uuid"
)

type WalletService interface {
	ProcessWalletOperation(ctx context.Context, req models.WalletOperationRequest) error
	GetWalletBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
}

type walletService struct {
	repo repository.WalletRepository
}

func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{repo: repo}
}

func (s *walletService) ProcessWalletOperation(ctx context.Context, req models.WalletOperationRequest) error {
	// Validate amount
	if req.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Check if wallet exists, create if not
	_, err := s.repo.GetWallet(ctx, req.WalletID)
	if err != nil {
		// If wallet doesn't exist, create it
		if err := s.repo.CreateWallet(ctx, req.WalletID); err != nil {
			return err
		}
	}

	// Update wallet balance
	return s.repo.UpdateWalletBalance(ctx, req.WalletID, req.Amount, req.OperationType)
}

func (s *walletService) GetWalletBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	wallet, err := s.repo.GetWallet(ctx, walletID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}