package handler

import (
	"net/http"

	models "HWGO/internal/Core"
	middleware "HWGO/internal/MidWare"
)

func (h *Handler) CreateCredit(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateCreditRequest
	if err := decode(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	credit, schedule, err := h.creditSvc.CreateCredit(userID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"credit":   credit,
		"schedule": schedule,
	})
}

func (h *Handler) GetCreditSchedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	creditID, err := parseID(r, "creditId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid credit id")
		return
	}

	schedule, err := h.creditSvc.GetSchedule(userID, creditID)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

func (h *Handler) GetUserCredits(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	credits, err := h.creditSvc.GetUserCredits(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch credits")
		return
	}

	writeJSON(w, http.StatusOK, credits)
}
