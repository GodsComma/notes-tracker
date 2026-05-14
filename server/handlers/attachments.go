package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"notes-tracker/server/db"
	authmw "notes-tracker/server/middleware"
)

type AttachmentHandler struct {
	DB      *db.DB
	DataDir string
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._\-]`)

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = unsafeChars.ReplaceAllString(name, "_")
	name = strings.TrimLeft(name, ".")
	if name == "" {
		name = "file"
	}
	return name
}

func (h *AttachmentHandler) attachmentDir(id int64) string {
	return filepath.Join(h.DataDir, "uploads", strconv.FormatInt(id, 10))
}

func (h *AttachmentHandler) attachmentPath(a db.Attachment) string {
	return filepath.Join(h.attachmentDir(a.ID), a.Filename)
}

// RemoveAttachmentFiles deletes the on-disk files for a list of attachments.
func RemoveAttachmentFiles(dataDir string, attachments []db.Attachment) {
	for _, a := range attachments {
		dir := filepath.Join(dataDir, "uploads", strconv.FormatInt(a.ID, 10))
		_ = os.RemoveAll(dir)
	}
}

// --- List ---

func (h *AttachmentHandler) ListLibraryPDFs(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	list, err := h.DB.ListAllPDFs(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []db.PDFLibraryItem{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *AttachmentHandler) ListForClass(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	list, err := h.DB.ListAttachmentsForClass(classID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []db.Attachment{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *AttachmentHandler) ListForSource(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	list, err := h.DB.ListAttachmentsForSource(sourceID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []db.Attachment{}
	}
	writeJSON(w, http.StatusOK, list)
}

// --- Upload ---

func (h *AttachmentHandler) UploadToClass(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	if _, err := h.DB.GetClass(classID, userID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "class not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.upload(w, r, &classID, nil, nil)
}

func (h *AttachmentHandler) UploadToSource(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	if _, err := h.DB.GetSource(sourceID, userID); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "source not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.upload(w, r, nil, &sourceID, nil)
}

func (h *AttachmentHandler) UploadToLibrary(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	h.upload(w, r, nil, nil, &userID)
}

// upload is the shared upload implementation. Exactly one of classID, sourceID, ownerUserID must be non-nil.
func (h *AttachmentHandler) upload(w http.ResponseWriter, r *http.Request, classID, sourceID, ownerUserID *int64) {
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field in form")
		return
	}
	defer file.Close()

	safeName := sanitizeFilename(header.Filename)
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	att, err := h.DB.CreateAttachment(classID, sourceID, ownerUserID, safeName, mimeType, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dir := h.attachmentDir(att.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		_ = h.DB.DeleteAttachment(att.ID)
		writeError(w, http.StatusInternalServerError, "failed to create storage directory")
		return
	}

	dst, err := os.Create(h.attachmentPath(att))
	if err != nil {
		_ = os.RemoveAll(dir)
		_ = h.DB.DeleteAttachment(att.ID)
		writeError(w, http.StatusInternalServerError, "failed to create file on disk")
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		_ = os.RemoveAll(dir)
		_ = h.DB.DeleteAttachment(att.ID)
		writeError(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	_ = h.DB.UpdateAttachmentSize(att.ID, size)
	att.SizeBytes = size

	writeJSON(w, http.StatusCreated, att)
}

// --- Assign ---

func (h *AttachmentHandler) AssignAttachment(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		ClassID  *int64 `json:"class_id"`
		SourceID *int64 `json:"source_id"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.ClassID == nil && body.SourceID == nil {
		writeError(w, http.StatusBadRequest, "class_id or source_id is required")
		return
	}
	if body.ClassID != nil && body.SourceID != nil {
		writeError(w, http.StatusBadRequest, "provide either class_id or source_id, not both")
		return
	}
	// Verify the target belongs to the same user
	if body.ClassID != nil {
		if _, err := h.DB.GetClass(*body.ClassID, userID); err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "class not found")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.SourceID != nil {
		if _, err := h.DB.GetSource(*body.SourceID, userID); err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "source not found")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	att, err := h.DB.AssignAttachment(id, userID, body.ClassID, body.SourceID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "attachment not found or already assigned")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, att)
}

// --- Rename ---

func (h *AttachmentHandler) RenameAttachment(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Filename string `json:"filename"`
	}
	if err := decode(r, &body); err != nil || strings.TrimSpace(body.Filename) == "" {
		writeError(w, http.StatusBadRequest, "filename is required")
		return
	}
	safeName := sanitizeFilename(body.Filename)
	att, err := h.DB.RenameAttachment(id, userID, safeName)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, att)
}

// --- Serve ---

func (h *AttachmentHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	att, err := h.DB.GetAttachment(id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	path := h.attachmentPath(att)
	f, err := os.Open(path)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found on disk")
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot stat file")
		return
	}

	mimeType := att.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	disposition := r.URL.Query().Get("disposition")
	if disposition == "attachment" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, att.Filename))
	} else {
		w.Header().Set("Content-Disposition", "inline")
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	http.ServeContent(w, r, att.Filename, stat.ModTime(), f)
}

// --- Delete ---

func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	att, err := h.DB.GetAttachment(id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = os.RemoveAll(h.attachmentDir(att.ID))
	if err := h.DB.DeleteAttachment(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Focus Settings ---

func (h *AttachmentHandler) GetFocusSettings(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s, err := h.DB.GetFocusSettings(id, userID)
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

func (h *AttachmentHandler) SaveFocusSettings(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body db.FocusSettings
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.DB.SaveFocusSettings(id, userID, body); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Progress (last viewed page) ---

func (h *AttachmentHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	lastPage, err := h.DB.GetAttachmentLastPage(id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"last_page": lastPage})
}

func (h *AttachmentHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	userID := authmw.UserIDFromCtx(r.Context())
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		LastPage int `json:"last_page"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.DB.SaveAttachmentLastPage(id, userID, body.LastPage); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
