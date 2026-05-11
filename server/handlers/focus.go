package handlers

import (
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type FocusHandler struct {
	DB *db.DB
}

func (h *FocusHandler) Log(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	var body struct {
		Date         string `json:"date"`
		DurationSec  int    `json:"duration_sec"`
		AttachmentID int64  `json:"attachment_id"`
		Filename     string `json:"filename"`
		Completed    bool   `json:"completed"`
	}
	if err := decode(r, &body); err != nil || body.Date == "" || body.DurationSec < 180 {
		writeError(w, http.StatusBadRequest, "invalid session")
		return
	}
	s, err := h.DB.LogFocusSession(userID, body.Date, body.DurationSec, body.AttachmentID, body.Filename, body.Completed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *FocusHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	sessions, err := h.DB.ListFocusSessions(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}
