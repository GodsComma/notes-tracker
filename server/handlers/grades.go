package handlers

import (
	"database/sql"
	"net/http"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type GradeHandler struct {
	DB *db.DB
}

// --- Rubric ---

func (h *GradeHandler) GetRubric(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	rubric, err := h.DB.GetRubric(classID, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "no rubric")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rubric)
}

func (h *GradeHandler) PutRubric(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	var body struct {
		GradingType string `json:"grading_type"`
		PlusMinus   bool   `json:"plus_minus"`
		Thresholds  string `json:"thresholds"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.GradingType != "letter" && body.GradingType != "pass_fail" {
		writeError(w, http.StatusBadRequest, "grading_type must be 'letter' or 'pass_fail'")
		return
	}
	rubric, err := h.DB.UpsertRubric(classID, userID, body.GradingType, body.PlusMinus, body.Thresholds)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "class not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rubric)
}

func (h *GradeHandler) DeleteRubric(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	if err := h.DB.DeleteRubric(classID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Categories ---

func (h *GradeHandler) AddCategory(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	rubricID, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid rubric id")
		return
	}
	var body struct {
		Name          string  `json:"name"`
		Weight        float64 `json:"weight"`
		IsExtraCredit bool    `json:"is_extra_credit"`
		Possible      float64 `json:"possible"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.Weight <= 0 {
		writeError(w, http.StatusBadRequest, "weight must be greater than 0")
		return
	}
	if body.Possible <= 0 {
		body.Possible = 100
	}
	cat, err := h.DB.AddCategory(rubricID, userID, body.Name, body.Weight, body.IsExtraCredit, body.Possible)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "rubric not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *GradeHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var raw map[string]interface{}
	if err := decode(r, &raw); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var (
		name          *string
		weight        *float64
		isExtraCredit *bool
		earned        *float64
		earnedSet     bool
		possible      *float64
		sortOrder     *int
	)
	if v, ok := raw["name"].(string); ok {
		name = &v
	}
	if v, ok := raw["weight"].(float64); ok {
		weight = &v
	}
	if v, ok := raw["is_extra_credit"].(bool); ok {
		isExtraCredit = &v
	}
	if v, ok := raw["possible"].(float64); ok {
		possible = &v
	}
	if v, ok := raw["sort_order"].(float64); ok {
		i := int(v)
		sortOrder = &i
	}
	if rawEarned, present := raw["earned"]; present {
		earnedSet = true
		if rawEarned != nil {
			if v, ok := rawEarned.(float64); ok {
				earned = &v
			}
		}
	}
	cat, err := h.DB.UpdateCategory(id, userID, name, weight, isExtraCredit, earned, earnedSet, possible, sortOrder)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *GradeHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.DB.DeleteCategory(id, userID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
