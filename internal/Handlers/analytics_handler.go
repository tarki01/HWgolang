package handler

import (
	middleware "HWGO/internal/MidWare"
	"net/http"
	"strconv"
)

func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	monthly, err := h.analyticsSvc.GetMonthlyAnalytics(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "analytics error")
		return
	}

	creditAnalytics, err := h.analyticsSvc.GetCreditAnalytics(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "credit analytics error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"monthly": monthly,
		"credits": creditAnalytics,
	})
}

func (h *Handler) GetBalanceForecast(w http.ResponseWriter, r *http.Request) {
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

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil {
			days = n
		}
	}

	forecast, err := h.analyticsSvc.GetBalanceForecast(userID, accountID, days)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, forecast)
}
