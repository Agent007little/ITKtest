package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ITKtest/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWalletRepository_CompleteFlow(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)
	walletID := uuid.New()
	now := time.Now()
	ctx := context.Background()

	// 1. Тестируем создание кошелька
	t.Run("CreateWallet", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO wallets \(id, balance, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4\)`).
			WithArgs(walletID, 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateWallet(ctx, walletID)
		assert.NoError(t, err)
	})

	// 2. Тестируем получение кошелька (должен быть пустой)
	t.Run("GetWallet after creation", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(walletID, 0, now, now)

		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(walletID).
			WillReturnRows(rows)

		wallet, err := repo.GetWallet(ctx, walletID)
		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, walletID, wallet.ID)
		assert.Equal(t, int64(0), wallet.Balance)
	})

	// 3. Тестируем пополнение счета
	t.Run("Deposit funds", func(t *testing.T) {
		mock.ExpectBegin()
		rows := sqlmock.NewRows([]string{"balance"}).AddRow(0)
		mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
			WithArgs(walletID).
			WillReturnRows(rows)

		mock.ExpectExec(`UPDATE wallets SET balance = \$1, updated_at = \$2 WHERE id = \$3`).
			WithArgs(1000, sqlmock.AnyArg(), walletID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.UpdateWalletBalance(ctx, walletID, 1000, models.DEPOSIT)
		assert.NoError(t, err)
	})

	// 4. Проверяем баланс после пополнения
	t.Run("GetWallet after deposit", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(walletID, 1000, now, now)

		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(walletID).
			WillReturnRows(rows)

		wallet, err := repo.GetWallet(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), wallet.Balance)
	})

	// 5. Тестируем снятие средств (успешное)
	t.Run("Withdraw funds successfully", func(t *testing.T) {
		mock.ExpectBegin()
		rows := sqlmock.NewRows([]string{"balance"}).AddRow(1000)
		mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
			WithArgs(walletID).
			WillReturnRows(rows)

		mock.ExpectExec(`UPDATE wallets SET balance = \$1, updated_at = \$2 WHERE id = \$3`).
			WithArgs(500, sqlmock.AnyArg(), walletID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.UpdateWalletBalance(ctx, walletID, 500, models.WITHDRAW)
		assert.NoError(t, err)
	})

	// 6. Проверяем баланс после снятия
	t.Run("GetWallet after withdraw", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(walletID, 500, now, now)

		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(walletID).
			WillReturnRows(rows)

		wallet, err := repo.GetWallet(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(500), wallet.Balance)
	})

	// 7. Тестируем попытку снять больше чем есть (должна быть ошибка)
	t.Run("Withdraw insufficient funds", func(t *testing.T) {
		mock.ExpectBegin()
		rows := sqlmock.NewRows([]string{"balance"}).AddRow(500)
		mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
			WithArgs(walletID).
			WillReturnRows(rows)

		mock.ExpectRollback()

		err := repo.UpdateWalletBalance(ctx, walletID, 1000, models.WITHDRAW)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	// 8. Проверяем что баланс не изменился после неудачного снятия
	t.Run("GetWallet after failed withdraw", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(walletID, 500, now, now)

		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(walletID).
			WillReturnRows(rows)

		wallet, err := repo.GetWallet(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(500), wallet.Balance) // Баланс должен остаться прежним
	})

	// 9. Тестируем еще одно пополнение
	t.Run("Deposit more funds", func(t *testing.T) {
		mock.ExpectBegin()
		rows := sqlmock.NewRows([]string{"balance"}).AddRow(500)
		mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
			WithArgs(walletID).
			WillReturnRows(rows)

		mock.ExpectExec(`UPDATE wallets SET balance = \$1, updated_at = \$2 WHERE id = \$3`).
			WithArgs(1500, sqlmock.AnyArg(), walletID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.UpdateWalletBalance(ctx, walletID, 1000, models.DEPOSIT)
		assert.NoError(t, err)
	})

	// 10. Финальная проверка баланса
	t.Run("Final balance check", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance", "created_at", "updated_at"}).
			AddRow(walletID, 1500, now, now)

		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(walletID).
			WillReturnRows(rows)

		wallet, err := repo.GetWallet(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1500), wallet.Balance)
	})

	// Проверяем что все ожидания выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWalletRepository_EdgeCases(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWalletRepository(db)
	walletID := uuid.New()
	ctx := context.Background()

	// 1. Тестируем получение несуществующего кошелька
	t.Run("Get non-existent wallet", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, balance, created_at, updated_at FROM wallets WHERE id = \$1`).
			WithArgs(walletID).
			WillReturnError(sql.ErrNoRows)

		wallet, err := repo.GetWallet(ctx, walletID)
		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "not found")
	})

	// 2. Тестируем невалидную операцию - ТЕПЕРЬ БЕЗ ТРАНЗАКЦИИ
	t.Run("Invalid operation type", func(t *testing.T) {
		// Сначала создаем кошелек
		mock.ExpectExec(`INSERT INTO wallets \(id, balance, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4\)`).
			WithArgs(walletID, 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateWallet(ctx, walletID)
		assert.NoError(t, err)

		// Пытаемся выполнить невалидную операцию
		err = repo.UpdateWalletBalance(ctx, walletID, 100, "INVALID_OPERATION")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid operation type")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}