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
)

type AttachmentHandler struct {
	DB      *db.DB
	DataDir string
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._\-]`)

func sanitizeFilename(name string) string {
	name = filepath.Base(name) // strip any directory components from the client
	name = unsafeChars.ReplaceAllString(name, "_")
	// Prevent hidden files
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
// Called by ClassHandler and SourceHandler before cascade DB deletes.
func RemoveAttachmentFiles(dataDir string, attachments []db.Attachment) {
	for _, a := range attachments {
		dir := filepath.Join(dataDir, "uploads", strconv.FormatInt(a.ID, 10))
		_ = os.RemoveAll(dir)
	}
}

// --- List ---

func (h *AttachmentHandler) ListLibraryPDFs(w http.ResponseWriter, r *http.Request) {
	list, err := h.DB.ListAllPDFs()
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
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	list, err := h.DB.ListAttachmentsForClass(classID)
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
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	list, err := h.DB.ListAttachmentsForSource(sourceID)
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
	classID, ok := parseID(r, "classID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	h.upload(w, r, &classID, nil)
}

func (h *AttachmentHandler) UploadToSource(w http.ResponseWriter, r *http.Request) {
	sourceID, ok := parseID(r, "sourceID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid source id")
		return
	}
	h.upload(w, r, nil, &sourceID)
}

func (h *AttachmentHandler) upload(w http.ResponseWriter, r *http.Request, classID, sourceID *int64) {
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

	// Insert DB record first to claim an ID for the directory name.
	att, err := h.DB.CreateAttachment(classID, sourceID, safeName, mimeType, 0)
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

// --- Serve ---

func (h *AttachmentHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	att, err := h.DB.GetAttachment(id)
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
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	att, err := h.DB.GetAttachment(id)
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

// --- Progress (last viewed page) ---

func (h *AttachmentHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	lastPage, err := h.DB.GetAttachmentLastPage(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"last_page": lastPage})
}

func (h *AttachmentHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
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
	if err := h.DB.SaveAttachmentLastPage(id, body.LastPage); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
