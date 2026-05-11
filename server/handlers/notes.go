package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type NoteHandler struct{ DB *db.DB }

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	note, err := h.DB.GetNote(sourceID, userID)
	if err == sql.ErrNoRows {
		// Return empty note rather than 404 — editor always needs something
		writeJSON(w, http.StatusOK, db.Note{SourceID: sourceID, Body: ""})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	note, err := h.DB.UpsertNote(sourceID, userID, body.Body)
	if errors.Is(err, db.ErrNoteLocked) {
		writeError(w, http.StatusConflict, "note is locked")
		return
	}
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) SetLock(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	var body struct {
		Locked bool `json:"locked"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	note, err := h.DB.SetNoteLock(sourceID, userID, body.Locked)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, note)
}
