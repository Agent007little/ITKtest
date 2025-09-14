package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"ITKtest/internal/models"
	"ITKtest/internal/service"
	"ITKtest/responder"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WalletController struct {
	service   service.WalletService
	responder responder.Responder
}

func NewWalletController(service service.WalletService, responder responder.Responder) *WalletController {
	return &WalletController{
		service:   service,
		responder: responder,
	}
}

func (c *WalletController) HandleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var req models.WalletOperationRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.responder.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Amount <= 0 {
		c.responder.Error(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	if req.OperationType != models.DEPOSIT && req.OperationType != models.WITHDRAW {
		c.responder.Error(w, http.StatusBadRequest, "Operation type must be DEPOSIT or WITHDRAW")
		return
	}

	// Process operation
	if err := c.service.ProcessWalletOperation(r.Context(), req); err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			c.responder.Error(w, http.StatusBadRequest, "Insufficient funds")
			return
		}
		c.responder.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	c.responder.OutputJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (c *WalletController) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	walletIDStr := chi.URLParam(r, "walletId")
	
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.responder.Error(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	balance, err := c.service.GetWalletBalance(r.Context(), walletID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.responder.Error(w, http.StatusNotFound, "Wallet not found")
			return
		}
		c.responder.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := models.WalletBalanceResponse{
		WalletID: walletID,
		Balance:  balance,
	}

	c.responder.OutputJSON(w, http.StatusOK, response)
}