package handler

import (
	"net/http"
	"strconv"

	models "HWGO/internal/Core"
	middleware "HWGO/internal/MidWare"

	"github.com/gorilla/mux"
)

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateAccountRequest
	if err := decode(r, &req); err != nil {
		req.Currency = "RUB"
	}

	acc, err := h.accountSvc.CreateAccount(userID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, acc)
}

func (h *Handler) GetUserAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.accountSvc.GetUserAccounts(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch accounts")
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID, err := parseID(r, "accountId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	acc, err := h.accountSvc.GetAccount(userID, accountID)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, acc)
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID, err := parseID(r, "accountId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req models.DepositRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	acc, err := h.accountSvc.Deposit(userID, accountID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, acc)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountID, err := parseID(r, "accountId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req models.DepositRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	acc, err := h.accountSvc.Withdraw(userID, accountID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, acc)
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.TransferRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.accountSvc.Transfer(userID, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseID(r *http.Request, key string) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars[key])
}
