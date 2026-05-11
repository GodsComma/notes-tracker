package handlers

import (
	"database/sql"
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type SemesterHandler struct{ DB *db.DB }

func (h *SemesterHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	semesters, err := h.DB.ListSemesters(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if semesters == nil {
		semesters = []db.Semester{}
	}
	writeJSON(w, http.StatusOK, semesters)
}

func (h *SemesterHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	var body struct {
		Name      string `json:"name"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := decode(r, &body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	s, err := h.DB.CreateSemester(userID, body.Name, body.StartDate, body.EndDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *SemesterHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Name      string `json:"name"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	s, err := h.DB.UpdateSemester(id, userID, body.Name, body.StartDate, body.EndDate)
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

func (h *SemesterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.DB.DeleteSemester(id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
