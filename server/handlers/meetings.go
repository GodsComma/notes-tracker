package handlers

import (
	"database/sql"
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type MeetingHandler struct{ DB *db.DB }

func (h *MeetingHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	meetings, err := h.DB.ListMeetings(classID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meetings == nil {
		meetings = []db.ClassMeeting{}
	}
	writeJSON(w, http.StatusOK, meetings)
}

func (h *MeetingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	var body struct {
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Location  string `json:"location"`
	}
	if err := decode(r, &body); err != nil || body.StartTime == "" || body.EndTime == "" {
		writeError(w, http.StatusBadRequest, "start_time and end_time are required")
		return
	}
	m, err := h.DB.CreateMeeting(classID, userID, body.DayOfWeek, body.StartTime, body.EndTime, body.Location)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "class not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *MeetingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Location  string `json:"location"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	m, err := h.DB.UpdateMeeting(id, userID, body.DayOfWeek, body.StartTime, body.EndTime, body.Location)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MeetingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.DB.DeleteMeeting(id, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
