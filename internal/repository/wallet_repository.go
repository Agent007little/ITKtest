package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ITKtest/internal/models"

	"github.com/google/uuid"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context, walletID uuid.UUID) error
	GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error)
	UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, amount int64, operationType models.OperationType) error
}

type walletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) CreateWallet(ctx context.Context, walletID uuid.UUID) error {
	query := `
		INSERT INTO wallets (id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, walletID, 0, time.Now(), time.Now())
	return err
}

func (r *walletRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, balance, created_at, updated_at
		FROM wallets
		WHERE id = $1
	`
	
	var wallet models.Wallet
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(
		&wallet.ID,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, err
	}
	
	return &wallet, nil
}

func (r *walletRepository) UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, amount int64, operationType models.OperationType) error {
	switch operationType {
	case models.DEPOSIT, models.WITHDRAW:
		// Valid operation types
	default:
		return errors.New("invalid operation type")
	}
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the wallet row for update
	var currentBalance int64
	err = tx.QueryRowContext(ctx, "SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", walletID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	// Calculate new balance
	var newBalance int64
	switch operationType {
	case models.DEPOSIT:
		newBalance = currentBalance + amount
	case models.WITHDRAW:
		newBalance = currentBalance - amount
		if newBalance < 0 {
			return errors.New("insufficient funds")
		}
	default:
		return errors.New("invalid operation type")
	}

	// Update balance
	_, err = tx.ExecContext(ctx, 
		"UPDATE wallets SET balance = $1, updated_at = $2 WHERE id = $3",
		newBalance, time.Now(), walletID)
	if err != nil {
		return err
	}

	return tx.Commit()
}