package handler

import (
	middleware "HWGO/internal/MidWare"
	"net/http"
)

type IssueCardRequest struct {
	AccountID int `json:"account_id"`
}

func (h *Handler) IssueCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req IssueCardRequest
	if err := decode(r, &req); err != nil || req.AccountID == 0 {
		writeError(w, http.StatusBadRequest, "account_id is required")
		return
	}

	card, err := h.cardSvc.IssueCard(userID, req.AccountID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, card)
}

func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
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

	cards, err := h.cardSvc.GetCards(userID, accountID)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cards)
}
