package httpHandler

import (
	"encoding/json"
	"net/http"
	"schedule/internal/controller/httpHandler/models"
)

func (h *Handler) writeAndLogErr(w http.ResponseWriter, err error, status int) {
	if status < 400 {
		h.l.Debug(err)
	} else if status < 500 {
		h.l.Warn(err)
	} else {
		h.l.Error(err)
	}

	if status < 500 {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()}); err != nil {
			h.l.Error(err)
		}
	}

	w.WriteHeader(status)
}

func (h *Handler) writeJson(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.l.Error(err)
	}
}
