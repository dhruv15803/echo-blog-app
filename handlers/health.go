package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "server ok"}); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
