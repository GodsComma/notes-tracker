package handlers

import (
	"net/http"
	"strconv"

	"notes-tracker/server/db"
)

type SearchHandler struct{ DB *db.DB }

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []db.SearchResult{})
		return
	}
	var classID int64
	if s := r.URL.Query().Get("class_id"); s != "" {
		classID, _ = strconv.ParseInt(s, 10, 64)
	}
	results, err := h.DB.Search(q, classID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if results == nil {
		results = []db.SearchResult{}
	}
	writeJSON(w, http.StatusOK, results)
}
