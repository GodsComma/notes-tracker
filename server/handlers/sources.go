package handlers

import (
	"database/sql"
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type SourceHandler struct {
	DB      *db.DB
	DataDir string
}

func (h *SourceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s, err := h.DB.GetSource(id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *SourceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	sources, err := h.DB.ListSources(classID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sources == nil {
		sources = []db.Source{}
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *SourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	var body struct {
		Type      string `json:"type"`
		Title     string `json:"title"`
		Date      string `json:"date"`
		DueDate   string `json:"due_date"`
		Reference string `json:"reference"`
		Priority  int    `json:"priority"`
	}
	if err := decode(r, &body); err != nil || body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if body.Type == "" {
		body.Type = "Lecture"
	}
	s, err := h.DB.CreateSource(classID, userID, body.Type, body.Title, body.Date, body.DueDate, body.Reference, body.Priority)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "class not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *SourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := h.DB.GetSource(id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		Type      *string `json:"type"`
		Title     *string `json:"title"`
		Date      *string `json:"date"`
		DueDate   *string `json:"due_date"`
		Reference *string `json:"reference"`
		Priority  *int    `json:"priority"`
		Completed *bool   `json:"completed"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	pick := func(p *string, fallback string) string {
		if p != nil {
			return *p
		}
		return fallback
	}
	pickInt := func(p *int, fallback int) int {
		if p != nil {
			return *p
		}
		return fallback
	}
	pickBool := func(p *bool, fallback bool) bool {
		if p != nil {
			return *p
		}
		return fallback
	}
	s, err := h.DB.UpdateSource(id, userID,
		pick(body.Type, existing.Type),
		pick(body.Title, existing.Title),
		pick(body.Date, existing.Date),
		pick(body.DueDate, existing.DueDate),
		pick(body.Reference, existing.Reference),
		pickInt(body.Priority, existing.Priority),
		pickBool(body.Completed, existing.Completed),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *SourceHandler) ListForSemester(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	semID, ok := parseID(r, "semesterID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid semester id")
		return
	}
	sources, err := h.DB.ListSourcesForSemester(semID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *SourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if h.DataDir != "" {
		if atts, err := h.DB.ListAttachmentsForSource(id, userID); err == nil {
			RemoveAttachmentFiles(h.DataDir, atts)
		}
	}
	if err := h.DB.DeleteSource(id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SourceHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.DB.ReorderSources(classID, userID, body.IDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
