package handlers

import (
	"database/sql"
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type ClassHandler struct {
	DB      *db.DB
	DataDir string
}

func (h *ClassHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	c, err := h.DB.GetClass(id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	semID, ok := parseID(r, "semesterID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid semester id")
		return
	}
	classes, err := h.DB.ListClasses(semID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if classes == nil {
		classes = []db.Class{}
	}
	writeJSON(w, http.StatusOK, classes)
}

func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	semID, ok := parseID(r, "semesterID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid semester id")
		return
	}
	var body struct {
		Name       string `json:"name"`
		Code       string `json:"code"`
		Instructor string `json:"instructor"`
		Schedule   string `json:"schedule"`
		Color      string `json:"color"`
	}
	if err := decode(r, &body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.Color == "" {
		body.Color = "#6366f1"
	}
	c, err := h.DB.CreateClass(semID, userID, body.Name, body.Code, body.Instructor, body.Schedule, body.Color)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "semester not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *ClassHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Name       *string `json:"name"`
		Code       *string `json:"code"`
		Instructor *string `json:"instructor"`
		Schedule   *string `json:"schedule"`
		Color      *string `json:"color"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	existing, err := h.DB.GetClass(id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	pick := func(p *string, fallback string) string {
		if p != nil {
			return *p
		}
		return fallback
	}
	c, err := h.DB.UpdateClass(
		id, userID,
		pick(body.Name, existing.Name),
		pick(body.Code, existing.Code),
		pick(body.Instructor, existing.Instructor),
		pick(body.Schedule, existing.Schedule),
		pick(body.Color, existing.Color),
	)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if h.DataDir != "" {
		if atts, err := h.DB.ListAllAttachmentsForClass(id, userID); err == nil {
			RemoveAttachmentFiles(h.DataDir, atts)
		}
	}
	if err := h.DB.DeleteClass(id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
