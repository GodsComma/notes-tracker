package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	sql *sql.DB
}

// --- Types ---

type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type Semester struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	CreatedAt string `json:"created_at"`
}

type Class struct {
	ID         int64  `json:"id"`
	SemesterID int64  `json:"semester_id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Instructor string `json:"instructor"`
	Schedule   string `json:"schedule"`
	Color      string `json:"color"`
}

type Source struct {
	ID        int64  `json:"id"`
	ClassID   int64  `json:"class_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	DueDate   string `json:"due_date"`
	Reference string `json:"reference"`
	SortOrder int    `json:"sort_order"`
	Priority  int    `json:"priority"`
	Completed bool   `json:"completed"`
}

type ClassMeeting struct {
	ID        int64  `json:"id"`
	ClassID   int64  `json:"class_id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Location  string `json:"location"`
}

type SourceWithClass struct {
	Source
	ClassName  string `json:"class_name"`
	ClassColor string `json:"class_color"`
}

type ClassMeetingWithClass struct {
	ClassMeeting
	ClassColor string `json:"class_color"`
	ClassName  string `json:"class_name"`
}

type CalendarEventProps struct {
	Type       string `json:"type"`
	SourceID   int64  `json:"sourceID,omitempty"`
	ClassID    int64  `json:"classID,omitempty"`
	ClassName  string `json:"className,omitempty"`
	SourceType string `json:"sourceType,omitempty"`
}

type CalendarEvent struct {
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	Start         string             `json:"start"`
	End           string             `json:"end,omitempty"`
	AllDay        bool               `json:"allDay"`
	Color         string             `json:"color"`
	ExtendedProps CalendarEventProps `json:"extendedProps"`
}

type Attachment struct {
	ID        int64  `json:"id"`
	ClassID   *int64 `json:"class_id"`
	SourceID  *int64 `json:"source_id"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	CreatedAt string `json:"created_at"`
}

type PDFLibraryItem struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	SizeBytes   int64  `json:"size_bytes"`
	CreatedAt   string `json:"created_at"`
	ClassID     int64  `json:"class_id"`
	ClassName   string `json:"class_name"`
	ClassColor  string `json:"class_color"`
	SourceID    *int64 `json:"source_id"`
	SourceTitle string `json:"source_title"`
}

type FocusSession struct {
	ID           int64  `json:"id"`
	Date         string `json:"date"`
	DurationSec  int    `json:"duration_sec"`
	AttachmentID int64  `json:"attachment_id"`
	Filename     string `json:"filename"`
	Completed    bool   `json:"completed"`
	CreatedAt    string `json:"created_at"`
}

type GradeRubric struct {
	ID          int64           `json:"id"`
	ClassID     int64           `json:"class_id"`
	GradingType string          `json:"grading_type"`
	PlusMinus   bool            `json:"plus_minus"`
	Thresholds  string          `json:"thresholds"`
	Categories  []GradeCategory `json:"categories"`
	CreatedAt   string          `json:"created_at"`
}

type GradeCategory struct {
	ID            int64    `json:"id"`
	RubricID      int64    `json:"rubric_id"`
	Name          string   `json:"name"`
	Weight        float64  `json:"weight"`
	IsExtraCredit bool     `json:"is_extra_credit"`
	Earned        *float64 `json:"earned"`
	Possible      float64  `json:"possible"`
	SortOrder     int      `json:"sort_order"`
}

type Note struct {
	ID        int64  `json:"id"`
	SourceID  int64  `json:"source_id"`
	Body      string `json:"body"`
	Locked    bool   `json:"locked"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

var ErrNoteLocked = errors.New("note is locked")

type FocusSettings struct {
	DurationMin           int    `json:"duration_min"`
	Theme                 string `json:"theme"`
	Ruler                 string `json:"ruler"`
	SuppressNotifications bool   `json:"suppress_notifications"`
	Annotations           string `json:"annotations"`
	AutoHideDrawer        bool   `json:"auto_hide_drawer"`
}

type SearchResult struct {
	NoteID      int64  `json:"note_id"`
	SourceID    int64  `json:"source_id"`
	SourceTitle string `json:"source_title"`
	SourceDate  string `json:"source_date"`
	ClassID     int64  `json:"class_id"`
	ClassName   string `json:"class_name"`
	ClassColor  string `json:"class_color"`
	Snippet     string `json:"snippet"`
}

// --- Open / Migrate ---

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dataDir, "notes.db")
	sqlDB, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	d := &DB{sql: sqlDB}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

func (d *DB) migrate() error {
	entries, err := filepath.Glob("server/db/migrations/*.sql")
	if err != nil {
		return err
	}
	if _, err := d.sql.Exec(`CREATE TABLE IF NOT EXISTS _migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	for _, path := range entries {
		name := filepath.Base(path)
		var existing string
		err := d.sql.QueryRow(`SELECT name FROM _migrations WHERE name = ?`, name).Scan(&existing)
		if err == nil {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := d.sql.Exec(string(data)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := d.sql.Exec(`INSERT INTO _migrations (name) VALUES (?)`, name); err != nil {
			return err
		}
	}

	const reindexMark = "!fts-reindex-pdfquote-text"
	var marker string
	if d.sql.QueryRow(`SELECT name FROM _migrations WHERE name = ?`, reindexMark).Scan(&marker) == sql.ErrNoRows {
		if err := d.reindexAllFTS(); err != nil {
			return fmt.Errorf("fts reindex: %w", err)
		}
		if _, err := d.sql.Exec(`INSERT INTO _migrations (name) VALUES (?)`, reindexMark); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) reindexAllFTS() error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM notes_fts`); err != nil {
		return err
	}
	rows, err := tx.Query(`SELECT id, body FROM notes WHERE body != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var body string
		if err := rows.Scan(&id, &body); err != nil {
			return err
		}
		if text := extractText(body); text != "" {
			if _, err := tx.Exec(`INSERT INTO notes_fts (note_id, content) VALUES (?, ?)`, id, text); err != nil {
				return err
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()
}

// --- Users ---

func (d *DB) CreateUser(email, passwordHash string) (User, error) {
	res, err := d.sql.Exec(`INSERT INTO users (email, password) VALUES (?, ?)`, email, passwordHash)
	if err != nil {
		return User{}, err
	}
	id, _ := res.LastInsertId()
	return d.GetUserByID(id)
}

func (d *DB) GetUserByID(id int64) (User, error) {
	var u User
	err := d.sql.QueryRow(`SELECT id, email, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Email, &u.CreatedAt)
	return u, err
}

// GetUserByEmail returns the user and the raw password hash for bcrypt comparison.
func (d *DB) GetUserByEmail(email string) (User, string, error) {
	var u User
	var hash string
	err := d.sql.QueryRow(`SELECT id, email, password, created_at FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &hash, &u.CreatedAt)
	return u, hash, err
}

// --- Semesters ---

func (d *DB) ListSemesters(userID int64) ([]Semester, error) {
	rows, err := d.sql.Query(
		`SELECT id, name, start_date, end_date, created_at FROM semesters WHERE user_id = ? ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Semester
	for rows.Next() {
		var s Semester
		if err := rows.Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) CreateSemester(userID int64, name, startDate, endDate string) (Semester, error) {
	res, err := d.sql.Exec(
		`INSERT INTO semesters (user_id, name, start_date, end_date) VALUES (?, ?, ?, ?)`,
		userID, name, startDate, endDate,
	)
	if err != nil {
		return Semester{}, err
	}
	id, _ := res.LastInsertId()
	return d.getSemesterByID(id)
}

func (d *DB) GetSemester(id, userID int64) (Semester, error) {
	var s Semester
	err := d.sql.QueryRow(
		`SELECT id, name, start_date, end_date, created_at FROM semesters WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.CreatedAt)
	return s, err
}

// getSemesterByID is an internal helper used after ownership is already established.
func (d *DB) getSemesterByID(id int64) (Semester, error) {
	var s Semester
	err := d.sql.QueryRow(
		`SELECT id, name, start_date, end_date, created_at FROM semesters WHERE id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.StartDate, &s.EndDate, &s.CreatedAt)
	return s, err
}

func (d *DB) UpdateSemester(id, userID int64, name, startDate, endDate string) (Semester, error) {
	res, err := d.sql.Exec(
		`UPDATE semesters SET name=?, start_date=?, end_date=? WHERE id=? AND user_id=?`,
		name, startDate, endDate, id, userID,
	)
	if err != nil {
		return Semester{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Semester{}, sql.ErrNoRows
	}
	return d.getSemesterByID(id)
}

func (d *DB) DeleteSemester(id, userID int64) error {
	_, err := d.sql.Exec(`DELETE FROM semesters WHERE id=? AND user_id=?`, id, userID)
	return err
}

// --- Classes ---

func (d *DB) ListClasses(semesterID, userID int64) ([]Class, error) {
	rows, err := d.sql.Query(`
		SELECT c.id, c.semester_id, c.name, c.code, c.instructor, c.schedule, c.color
		FROM classes c
		JOIN semesters s ON s.id = c.semester_id
		WHERE c.semester_id = ? AND s.user_id = ?
		ORDER BY c.id`,
		semesterID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.SemesterID, &c.Name, &c.Code, &c.Instructor, &c.Schedule, &c.Color); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *DB) CreateClass(semesterID, userID int64, name, code, instructor, schedule, color string) (Class, error) {
	var count int
	if err := d.sql.QueryRow(
		`SELECT COUNT(*) FROM semesters WHERE id = ? AND user_id = ?`, semesterID, userID,
	).Scan(&count); err != nil || count == 0 {
		return Class{}, sql.ErrNoRows
	}
	res, err := d.sql.Exec(
		`INSERT INTO classes (semester_id, name, code, instructor, schedule, color) VALUES (?,?,?,?,?,?)`,
		semesterID, name, code, instructor, schedule, color,
	)
	if err != nil {
		return Class{}, err
	}
	id, _ := res.LastInsertId()
	return d.getClassByID(id)
}

func (d *DB) GetClass(id, userID int64) (Class, error) {
	var c Class
	err := d.sql.QueryRow(`
		SELECT c.id, c.semester_id, c.name, c.code, c.instructor, c.schedule, c.color
		FROM classes c
		JOIN semesters s ON s.id = c.semester_id
		WHERE c.id = ? AND s.user_id = ?`,
		id, userID,
	).Scan(&c.ID, &c.SemesterID, &c.Name, &c.Code, &c.Instructor, &c.Schedule, &c.Color)
	return c, err
}

// getClassByID is an internal helper used after ownership is already established.
func (d *DB) getClassByID(id int64) (Class, error) {
	var c Class
	err := d.sql.QueryRow(
		`SELECT id, semester_id, name, code, instructor, schedule, color FROM classes WHERE id=?`, id,
	).Scan(&c.ID, &c.SemesterID, &c.Name, &c.Code, &c.Instructor, &c.Schedule, &c.Color)
	return c, err
}

func (d *DB) UpdateClass(id, userID int64, name, code, instructor, schedule, color string) (Class, error) {
	res, err := d.sql.Exec(`
		UPDATE classes SET name=?, code=?, instructor=?, schedule=?, color=?
		WHERE id=? AND semester_id IN (SELECT id FROM semesters WHERE user_id=?)`,
		name, code, instructor, schedule, color, id, userID,
	)
	if err != nil {
		return Class{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Class{}, sql.ErrNoRows
	}
	return d.getClassByID(id)
}

func (d *DB) DeleteClass(id, userID int64) error {
	_, err := d.sql.Exec(`
		DELETE FROM classes WHERE id=? AND semester_id IN (SELECT id FROM semesters WHERE user_id=?)`,
		id, userID,
	)
	return err
}

// --- Sources ---

func (d *DB) ListSources(classID, userID int64) ([]Source, error) {
	rows, err := d.sql.Query(`
		SELECT s.id, s.class_id, s.type, s.title, s.date, s.due_date, s.reference, s.sort_order, s.priority, s.completed
		FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE s.class_id = ? AND sem.user_id = ?
		ORDER BY s.sort_order, s.date, s.id`,
		classID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSources(rows)
}

func (d *DB) CreateSource(classID, userID int64, srcType, title, date, dueDate, reference string, priority int) (Source, error) {
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM classes c JOIN semesters s ON s.id = c.semester_id
		WHERE c.id = ? AND s.user_id = ?`, classID, userID,
	).Scan(&count); err != nil || count == 0 {
		return Source{}, sql.ErrNoRows
	}
	var maxOrder int
	_ = d.sql.QueryRow(`SELECT COALESCE(MAX(sort_order),0) FROM sources WHERE class_id=?`, classID).Scan(&maxOrder)
	res, err := d.sql.Exec(
		`INSERT INTO sources (class_id, type, title, date, due_date, reference, sort_order, priority) VALUES (?,?,?,?,?,?,?,?)`,
		classID, srcType, title, date, dueDate, reference, maxOrder+1, priority,
	)
	if err != nil {
		return Source{}, err
	}
	id, _ := res.LastInsertId()
	return d.getSourceByID(id)
}

func (d *DB) GetSource(id, userID int64) (Source, error) {
	var s Source
	err := d.sql.QueryRow(`
		SELECT s.id, s.class_id, s.type, s.title, s.date, s.due_date, s.reference, s.sort_order, s.priority, s.completed
		FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE s.id = ? AND sem.user_id = ?`,
		id, userID,
	).Scan(&s.ID, &s.ClassID, &s.Type, &s.Title, &s.Date, &s.DueDate, &s.Reference, &s.SortOrder, &s.Priority, &s.Completed)
	return s, err
}

// getSourceByID is an internal helper used after ownership is already established.
func (d *DB) getSourceByID(id int64) (Source, error) {
	var s Source
	err := d.sql.QueryRow(
		`SELECT id, class_id, type, title, date, due_date, reference, sort_order, priority, completed FROM sources WHERE id=?`, id,
	).Scan(&s.ID, &s.ClassID, &s.Type, &s.Title, &s.Date, &s.DueDate, &s.Reference, &s.SortOrder, &s.Priority, &s.Completed)
	return s, err
}

func (d *DB) UpdateSource(id, userID int64, srcType, title, date, dueDate, reference string, priority int, completed bool) (Source, error) {
	res, err := d.sql.Exec(`
		UPDATE sources SET type=?, title=?, date=?, due_date=?, reference=?, priority=?, completed=?
		WHERE id=? AND class_id IN (
			SELECT c.id FROM classes c JOIN semesters s ON s.id = c.semester_id WHERE s.user_id=?
		)`,
		srcType, title, date, dueDate, reference, priority, completed, id, userID,
	)
	if err != nil {
		return Source{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Source{}, sql.ErrNoRows
	}
	return d.getSourceByID(id)
}

func (d *DB) DeleteSource(id, userID int64) error {
	_, err := d.sql.Exec(`
		DELETE FROM sources WHERE id=? AND class_id IN (
			SELECT c.id FROM classes c JOIN semesters s ON s.id = c.semester_id WHERE s.user_id=?
		)`,
		id, userID,
	)
	return err
}

func (d *DB) ReorderSources(classID, userID int64, ids []int64) error {
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM classes c JOIN semesters s ON s.id = c.semester_id
		WHERE c.id = ? AND s.user_id = ?`, classID, userID,
	).Scan(&count); err != nil || count == 0 {
		return sql.ErrNoRows
	}
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if _, err := tx.Exec(`UPDATE sources SET sort_order=? WHERE id=? AND class_id=?`, i, id, classID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --- Class Meetings ---

func (d *DB) ListMeetings(classID, userID int64) ([]ClassMeeting, error) {
	rows, err := d.sql.Query(`
		SELECT m.id, m.class_id, m.day_of_week, m.start_time, m.end_time, m.location
		FROM class_meetings m
		JOIN classes c ON c.id = m.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE m.class_id = ? AND s.user_id = ?
		ORDER BY m.day_of_week, m.start_time`,
		classID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClassMeeting
	for rows.Next() {
		var m ClassMeeting
		if err := rows.Scan(&m.ID, &m.ClassID, &m.DayOfWeek, &m.StartTime, &m.EndTime, &m.Location); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (d *DB) CreateMeeting(classID, userID int64, dayOfWeek int, startTime, endTime, location string) (ClassMeeting, error) {
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM classes c JOIN semesters s ON s.id = c.semester_id
		WHERE c.id = ? AND s.user_id = ?`, classID, userID,
	).Scan(&count); err != nil || count == 0 {
		return ClassMeeting{}, sql.ErrNoRows
	}
	res, err := d.sql.Exec(
		`INSERT INTO class_meetings (class_id, day_of_week, start_time, end_time, location) VALUES (?,?,?,?,?)`,
		classID, dayOfWeek, startTime, endTime, location,
	)
	if err != nil {
		return ClassMeeting{}, err
	}
	id, _ := res.LastInsertId()
	return d.getMeetingByID(id)
}

func (d *DB) GetMeeting(id, userID int64) (ClassMeeting, error) {
	var m ClassMeeting
	err := d.sql.QueryRow(`
		SELECT m.id, m.class_id, m.day_of_week, m.start_time, m.end_time, m.location
		FROM class_meetings m
		JOIN classes c ON c.id = m.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE m.id = ? AND s.user_id = ?`,
		id, userID,
	).Scan(&m.ID, &m.ClassID, &m.DayOfWeek, &m.StartTime, &m.EndTime, &m.Location)
	return m, err
}

// getMeetingByID is an internal helper used after ownership is already established.
func (d *DB) getMeetingByID(id int64) (ClassMeeting, error) {
	var m ClassMeeting
	err := d.sql.QueryRow(
		`SELECT id, class_id, day_of_week, start_time, end_time, location FROM class_meetings WHERE id=?`, id,
	).Scan(&m.ID, &m.ClassID, &m.DayOfWeek, &m.StartTime, &m.EndTime, &m.Location)
	return m, err
}

func (d *DB) UpdateMeeting(id, userID int64, dayOfWeek int, startTime, endTime, location string) (ClassMeeting, error) {
	res, err := d.sql.Exec(`
		UPDATE class_meetings SET day_of_week=?, start_time=?, end_time=?, location=?
		WHERE id=? AND class_id IN (
			SELECT c.id FROM classes c JOIN semesters s ON s.id = c.semester_id WHERE s.user_id=?
		)`,
		dayOfWeek, startTime, endTime, location, id, userID,
	)
	if err != nil {
		return ClassMeeting{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ClassMeeting{}, sql.ErrNoRows
	}
	return d.getMeetingByID(id)
}

func (d *DB) DeleteMeeting(id, userID int64) error {
	_, err := d.sql.Exec(`
		DELETE FROM class_meetings WHERE id=? AND class_id IN (
			SELECT c.id FROM classes c JOIN semesters s ON s.id = c.semester_id WHERE s.user_id=?
		)`,
		id, userID,
	)
	return err
}

// --- Calendar queries ---

func (d *DB) SourcesInRangeForClass(classID, userID int64, start, end string) ([]Source, error) {
	rows, err := d.sql.Query(`
		SELECT s.id, s.class_id, s.type, s.title, s.date, s.due_date, s.reference, s.sort_order, s.priority, s.completed
		FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE s.class_id = ? AND sem.user_id = ?
		  AND (
		    (s.date != '' AND s.date BETWEEN ? AND ?)
		    OR (s.due_date != '' AND s.due_date BETWEEN ? AND ?)
		  )
		ORDER BY s.date, s.due_date`,
		classID, userID, start, end, start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSources(rows)
}

func (d *DB) SourcesInRangeForSemester(semesterID, userID int64, start, end string) ([]SourceWithClass, error) {
	rows, err := d.sql.Query(`
		SELECT s.id, s.class_id, s.type, s.title, s.date, s.due_date, s.reference, s.sort_order, s.priority, s.completed,
		       c.name, c.color
		FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE c.semester_id = ? AND sem.user_id = ?
		  AND (
		    (s.date != '' AND s.date BETWEEN ? AND ?)
		    OR (s.due_date != '' AND s.due_date BETWEEN ? AND ?)
		  )
		ORDER BY s.date, s.due_date`,
		semesterID, userID, start, end, start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SourceWithClass
	for rows.Next() {
		var s SourceWithClass
		if err := rows.Scan(&s.ID, &s.ClassID, &s.Type, &s.Title, &s.Date, &s.DueDate, &s.Reference, &s.SortOrder, &s.Priority, &s.Completed, &s.ClassName, &s.ClassColor); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) ListSourcesForSemester(semesterID, userID int64) ([]SourceWithClass, error) {
	rows, err := d.sql.Query(`
		SELECT s.id, s.class_id, s.type, s.title, s.date, s.due_date, s.reference, s.sort_order, s.priority, s.completed,
		       c.name, c.color
		FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE c.semester_id = ? AND sem.user_id = ?
		ORDER BY
		  CASE WHEN s.priority = 0 THEN 9999 ELSE s.priority END,
		  CASE WHEN s.due_date = '' THEN '9999-99-99' ELSE s.due_date END,
		  s.title`,
		semesterID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SourceWithClass
	for rows.Next() {
		var s SourceWithClass
		if err := rows.Scan(&s.ID, &s.ClassID, &s.Type, &s.Title, &s.Date, &s.DueDate, &s.Reference, &s.SortOrder, &s.Priority, &s.Completed, &s.ClassName, &s.ClassColor); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []SourceWithClass{}
	}
	return out, rows.Err()
}

func (d *DB) ListMeetingsForSemester(semesterID, userID int64) ([]ClassMeetingWithClass, error) {
	rows, err := d.sql.Query(`
		SELECT m.id, m.class_id, m.day_of_week, m.start_time, m.end_time, m.location, c.color, c.name
		FROM class_meetings m
		JOIN classes c ON c.id = m.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE c.semester_id = ? AND sem.user_id = ?
		ORDER BY m.day_of_week, m.start_time`,
		semesterID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClassMeetingWithClass
	for rows.Next() {
		var m ClassMeetingWithClass
		if err := rows.Scan(&m.ID, &m.ClassID, &m.DayOfWeek, &m.StartTime, &m.EndTime, &m.Location, &m.ClassColor, &m.ClassName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func scanSources(rows *sql.Rows) ([]Source, error) {
	var out []Source
	for rows.Next() {
		var s Source
		if err := rows.Scan(&s.ID, &s.ClassID, &s.Type, &s.Title, &s.Date, &s.DueDate, &s.Reference, &s.SortOrder, &s.Priority, &s.Completed); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// --- Attachments ---

func (d *DB) scanAttachments(rows *sql.Rows) ([]Attachment, error) {
	var out []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.ClassID, &a.SourceID, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (d *DB) ListAttachmentsForClass(classID, userID int64) ([]Attachment, error) {
	rows, err := d.sql.Query(`
		SELECT a.id, a.class_id, a.source_id, a.filename, a.mime_type, a.size_bytes, a.created_at
		FROM attachments a
		JOIN classes c ON c.id = a.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.class_id = ? AND s.user_id = ?
		ORDER BY a.created_at`,
		classID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanAttachments(rows)
}

func (d *DB) ListAttachmentsForSource(sourceID, userID int64) ([]Attachment, error) {
	rows, err := d.sql.Query(`
		SELECT a.id, a.class_id, a.source_id, a.filename, a.mime_type, a.size_bytes, a.created_at
		FROM attachments a
		JOIN sources src ON src.id = a.source_id
		JOIN classes c ON c.id = src.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.source_id = ? AND s.user_id = ?
		ORDER BY a.created_at`,
		sourceID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanAttachments(rows)
}

func (d *DB) ListAllAttachmentsForClass(classID, userID int64) ([]Attachment, error) {
	rows, err := d.sql.Query(`
		SELECT a.id, a.class_id, a.source_id, a.filename, a.mime_type, a.size_bytes, a.created_at
		FROM attachments a
		LEFT JOIN sources src ON src.id = a.source_id
		JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE c.id = ? AND s.user_id = ?`,
		classID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.scanAttachments(rows)
}

func (d *DB) ListAllPDFs(userID int64) ([]PDFLibraryItem, error) {
	rows, err := d.sql.Query(`
		SELECT a.id, a.filename, a.size_bytes, a.created_at,
		       COALESCE(c.id, 0) AS cid,
		       COALESCE(c.name, ''),
		       COALESCE(c.color, ''),
		       a.source_id,
		       COALESCE(src.title, '')
		FROM attachments a
		LEFT JOIN sources src ON src.id = a.source_id
		LEFT JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.mime_type = 'application/pdf' AND s.user_id = ?
		ORDER BY a.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PDFLibraryItem
	for rows.Next() {
		var p PDFLibraryItem
		if err := rows.Scan(
			&p.ID, &p.Filename, &p.SizeBytes, &p.CreatedAt,
			&p.ClassID, &p.ClassName, &p.ClassColor,
			&p.SourceID, &p.SourceTitle,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) GetAttachment(id, userID int64) (Attachment, error) {
	var a Attachment
	err := d.sql.QueryRow(`
		SELECT a.id, a.class_id, a.source_id, a.filename, a.mime_type, a.size_bytes, a.created_at
		FROM attachments a
		LEFT JOIN sources src ON src.id = a.source_id
		LEFT JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.id = ? AND s.user_id = ?`,
		id, userID,
	).Scan(&a.ID, &a.ClassID, &a.SourceID, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt)
	return a, err
}

func (d *DB) CreateAttachment(classID, sourceID *int64, filename, mimeType string, sizeBytes int64) (Attachment, error) {
	res, err := d.sql.Exec(
		`INSERT INTO attachments (class_id, source_id, filename, mime_type, size_bytes) VALUES (?,?,?,?,?)`,
		classID, sourceID, filename, mimeType, sizeBytes,
	)
	if err != nil {
		return Attachment{}, err
	}
	id, _ := res.LastInsertId()
	return d.getAttachmentByID(id)
}

// getAttachmentByID is an internal helper used after the record is just created.
func (d *DB) getAttachmentByID(id int64) (Attachment, error) {
	var a Attachment
	err := d.sql.QueryRow(
		`SELECT id, class_id, source_id, filename, mime_type, size_bytes, created_at FROM attachments WHERE id=?`, id,
	).Scan(&a.ID, &a.ClassID, &a.SourceID, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt)
	return a, err
}

func (d *DB) UpdateAttachmentSize(id, sizeBytes int64) error {
	_, err := d.sql.Exec(`UPDATE attachments SET size_bytes=? WHERE id=?`, sizeBytes, id)
	return err
}

func (d *DB) DeleteAttachment(id int64) error {
	_, err := d.sql.Exec(`DELETE FROM attachments WHERE id=?`, id)
	return err
}

// --- Attachment Progress (last page viewed) ---

func (d *DB) GetAttachmentLastPage(attachmentID, userID int64) (int, error) {
	var lastPage int
	err := d.sql.QueryRow(`
		SELECT ap.last_page
		FROM attachment_progress ap
		JOIN attachments a ON a.id = ap.attachment_id
		LEFT JOIN sources src ON src.id = a.source_id
		LEFT JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE ap.attachment_id = ? AND s.user_id = ?`,
		attachmentID, userID,
	).Scan(&lastPage)
	if err == sql.ErrNoRows {
		return 1, nil
	}
	return lastPage, err
}

func (d *DB) SaveAttachmentLastPage(attachmentID, userID int64, lastPage int) error {
	// Verify ownership first
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM attachments a
		LEFT JOIN sources src ON src.id = a.source_id
		LEFT JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.id = ? AND s.user_id = ?`,
		attachmentID, userID,
	).Scan(&count); err != nil || count == 0 {
		return sql.ErrNoRows
	}
	if lastPage < 1 {
		lastPage = 1
	}
	_, err := d.sql.Exec(`
		INSERT INTO attachment_progress (attachment_id, last_page, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(attachment_id) DO UPDATE SET
			last_page = excluded.last_page,
			updated_at = excluded.updated_at
	`, attachmentID, lastPage)
	return err
}

// --- Focus Sessions ---

func (d *DB) LogFocusSession(userID int64, date string, durationSec int, attachmentID int64, filename string, completed bool) (FocusSession, error) {
	res, err := d.sql.Exec(
		`INSERT INTO focus_sessions (user_id, date, duration_sec, attachment_id, filename, completed) VALUES (?,?,?,?,?,?)`,
		userID, date, durationSec, attachmentID, filename, completed,
	)
	if err != nil {
		return FocusSession{}, err
	}
	id, _ := res.LastInsertId()
	return d.getFocusSessionByID(id)
}

func (d *DB) getFocusSessionByID(id int64) (FocusSession, error) {
	var s FocusSession
	err := d.sql.QueryRow(
		`SELECT id, date, duration_sec, attachment_id, filename, completed, created_at FROM focus_sessions WHERE id=?`, id,
	).Scan(&s.ID, &s.Date, &s.DurationSec, &s.AttachmentID, &s.Filename, &s.Completed, &s.CreatedAt)
	return s, err
}

func (d *DB) ListFocusSessions(userID int64) ([]FocusSession, error) {
	rows, err := d.sql.Query(
		`SELECT id, date, duration_sec, attachment_id, filename, completed, created_at
		 FROM focus_sessions WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FocusSession
	for rows.Next() {
		var s FocusSession
		if err := rows.Scan(&s.ID, &s.Date, &s.DurationSec, &s.AttachmentID, &s.Filename, &s.Completed, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []FocusSession{}
	}
	return out, rows.Err()
}

// --- Notes ---

func (d *DB) GetNote(sourceID, userID int64) (Note, error) {
	var n Note
	err := d.sql.QueryRow(`
		SELECT n.id, n.source_id, n.body, n.locked, n.created_at, n.updated_at
		FROM notes n
		JOIN sources s ON s.id = n.source_id
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE n.source_id = ? AND sem.user_id = ?`,
		sourceID, userID,
	).Scan(&n.ID, &n.SourceID, &n.Body, &n.Locked, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func (d *DB) UpsertNote(sourceID, userID int64, body string) (Note, error) {
	// Verify ownership and check lock in one query.
	var locked bool
	err := d.sql.QueryRow(`
		SELECT COALESCE(n.locked, 0)
		FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		LEFT JOIN notes n ON n.source_id = s.id
		WHERE s.id = ? AND sem.user_id = ?`,
		sourceID, userID,
	).Scan(&locked)
	if err == sql.ErrNoRows {
		return Note{}, sql.ErrNoRows
	}
	if err != nil {
		return Note{}, err
	}
	if locked {
		return Note{}, ErrNoteLocked
	}
	_, err = d.sql.Exec(`
		INSERT INTO notes (source_id, body, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(source_id) DO UPDATE SET
			body = excluded.body,
			updated_at = excluded.updated_at
	`, sourceID, body)
	if err != nil {
		return Note{}, err
	}
	n, err := d.GetNote(sourceID, userID)
	if err != nil {
		return Note{}, err
	}
	_ = d.syncFTS(n.ID, extractText(body))
	return n, nil
}

func (d *DB) SetNoteLock(sourceID, userID int64, locked bool) (Note, error) {
	// Verify ownership
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM sources s
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE s.id = ? AND sem.user_id = ?`,
		sourceID, userID,
	).Scan(&count); err != nil || count == 0 {
		return Note{}, sql.ErrNoRows
	}
	_, err := d.sql.Exec(`
		INSERT INTO notes (source_id, body, locked, updated_at)
		VALUES (?, '', ?, datetime('now'))
		ON CONFLICT(source_id) DO UPDATE SET locked = excluded.locked
	`, sourceID, locked)
	if err != nil {
		return Note{}, err
	}
	return d.GetNote(sourceID, userID)
}

func (d *DB) syncFTS(noteID int64, text string) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM notes_fts WHERE note_id = ?`, noteID); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO notes_fts (note_id, content) VALUES (?, ?)`, noteID, text); err != nil {
		return err
	}
	return tx.Commit()
}

// --- Search ---

func (d *DB) Search(userID int64, query string, classID int64) ([]SearchResult, error) {
	seen := map[int64]bool{}
	var out []SearchResult

	// 1. FTS search of note body content
	q1 := `
		SELECT CAST(nf.note_id AS INTEGER), snippet(notes_fts, 1, '<mark>', '</mark>', '…', 15),
		       n.source_id, s.title, COALESCE(s.date, ''), c.id, c.name, c.color
		FROM notes_fts nf
		JOIN notes n ON CAST(nf.note_id AS INTEGER) = n.id
		JOIN sources s ON s.id = n.source_id
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE notes_fts MATCH ? AND sem.user_id = ?`
	args1 := []any{query, userID}
	if classID > 0 {
		q1 += " AND c.id = ?"
		args1 = append(args1, classID)
	}
	q1 += " ORDER BY rank LIMIT 30"

	rows1, err := d.sql.Query(q1, args1...)
	if err != nil {
		return nil, err
	}
	defer rows1.Close()
	for rows1.Next() {
		var r SearchResult
		if err := rows1.Scan(&r.NoteID, &r.Snippet, &r.SourceID, &r.SourceTitle, &r.SourceDate, &r.ClassID, &r.ClassName, &r.ClassColor); err != nil {
			return nil, err
		}
		out = append(out, r)
		seen[r.SourceID] = true
	}
	if err := rows1.Err(); err != nil {
		return nil, err
	}

	// 2. Source title / reference LIKE match
	like := "%" + strings.ToLower(query) + "%"
	q2 := `
		SELECT n.id, n.source_id, s.title, COALESCE(s.date, ''), c.id, c.name, c.color
		FROM notes n
		JOIN sources s ON s.id = n.source_id
		JOIN classes c ON c.id = s.class_id
		JOIN semesters sem ON sem.id = c.semester_id
		WHERE (LOWER(s.title) LIKE ? OR LOWER(s.reference) LIKE ?) AND sem.user_id = ?`
	args2 := []any{like, like, userID}
	if classID > 0 {
		q2 += " AND c.id = ?"
		args2 = append(args2, classID)
	}
	q2 += " LIMIT 30"

	rows2, err := d.sql.Query(q2, args2...)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var r SearchResult
		if err := rows2.Scan(&r.NoteID, &r.SourceID, &r.SourceTitle, &r.SourceDate, &r.ClassID, &r.ClassName, &r.ClassColor); err != nil {
			return nil, err
		}
		if seen[r.SourceID] {
			continue
		}
		out = append(out, r)
		seen[r.SourceID] = true
	}
	return out, rows2.Err()
}

// --- Text extraction ---

func extractText(body string) string {
	if body == "" || !strings.HasPrefix(strings.TrimSpace(body), "{") {
		return body
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		return body
	}
	var sb strings.Builder
	walkNode(doc, &sb)
	return strings.TrimSpace(sb.String())
}

func walkNode(node map[string]any, sb *strings.Builder) {
	if t, _ := node["type"].(string); t == "text" {
		if text, _ := node["text"].(string); text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
		return
	}
	if t, _ := node["type"].(string); t == "pdfQuote" {
		if attrs, ok := node["attrs"].(map[string]any); ok {
			if quoted, _ := attrs["text"].(string); quoted != "" {
				sb.WriteString(quoted)
				sb.WriteString(" ")
			}
		}
	}
	if content, ok := node["content"].([]any); ok {
		for _, child := range content {
			if m, ok := child.(map[string]any); ok {
				walkNode(m, sb)
			}
		}
		sb.WriteString("\n")
	}
}

// --- Focus Settings ---

func (d *DB) GetFocusSettings(attachmentID, userID int64) (FocusSettings, error) {
	defaults := FocusSettings{DurationMin: 25, Theme: "default", Ruler: "off", SuppressNotifications: true, Annotations: "on", AutoHideDrawer: true}
	// Verify ownership
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM attachments a
		LEFT JOIN sources src ON src.id = a.source_id
		LEFT JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.id = ? AND s.user_id = ?`,
		attachmentID, userID,
	).Scan(&count); err != nil || count == 0 {
		return defaults, sql.ErrNoRows
	}
	var s FocusSettings
	err := d.sql.QueryRow(`
		SELECT duration_min, theme, ruler, suppress_notifications, annotations, auto_hide_drawer
		FROM attachment_focus_settings WHERE attachment_id = ?`, attachmentID,
	).Scan(&s.DurationMin, &s.Theme, &s.Ruler, &s.SuppressNotifications, &s.Annotations, &s.AutoHideDrawer)
	if err == sql.ErrNoRows {
		return defaults, nil
	}
	return s, err
}

func (d *DB) SaveFocusSettings(attachmentID, userID int64, s FocusSettings) error {
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM attachments a
		LEFT JOIN sources src ON src.id = a.source_id
		LEFT JOIN classes c ON c.id = COALESCE(a.class_id, src.class_id)
		JOIN semesters s ON s.id = c.semester_id
		WHERE a.id = ? AND s.user_id = ?`,
		attachmentID, userID,
	).Scan(&count); err != nil || count == 0 {
		return sql.ErrNoRows
	}
	_, err := d.sql.Exec(`
		INSERT INTO attachment_focus_settings
		  (attachment_id, duration_min, theme, ruler, suppress_notifications, annotations, auto_hide_drawer, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		ON CONFLICT(attachment_id) DO UPDATE SET
		  duration_min           = excluded.duration_min,
		  theme                  = excluded.theme,
		  ruler                  = excluded.ruler,
		  suppress_notifications = excluded.suppress_notifications,
		  annotations            = excluded.annotations,
		  auto_hide_drawer       = excluded.auto_hide_drawer,
		  updated_at             = excluded.updated_at`,
		attachmentID, s.DurationMin, s.Theme, s.Ruler, s.SuppressNotifications, s.Annotations, s.AutoHideDrawer,
	)
	return err
}

// --- Grades ---

func (d *DB) GetRubric(classID, userID int64) (GradeRubric, error) {
	var r GradeRubric
	err := d.sql.QueryRow(`
		SELECT gr.id, gr.class_id, gr.grading_type, gr.plus_minus, gr.thresholds, gr.created_at
		FROM grade_rubrics gr
		JOIN classes c ON c.id = gr.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE gr.class_id = ? AND s.user_id = ?`,
		classID, userID,
	).Scan(&r.ID, &r.ClassID, &r.GradingType, &r.PlusMinus, &r.Thresholds, &r.CreatedAt)
	if err != nil {
		return r, err
	}
	cats, err := d.listCategories(r.ID)
	if err != nil {
		return r, err
	}
	r.Categories = cats
	return r, nil
}

func (d *DB) listCategories(rubricID int64) ([]GradeCategory, error) {
	rows, err := d.sql.Query(
		`SELECT id, rubric_id, name, weight, is_extra_credit, earned, possible, sort_order
		 FROM grade_categories WHERE rubric_id=? ORDER BY sort_order, id`, rubricID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GradeCategory{}
	for rows.Next() {
		var c GradeCategory
		var earned sql.NullFloat64
		if err := rows.Scan(&c.ID, &c.RubricID, &c.Name, &c.Weight, &c.IsExtraCredit, &earned, &c.Possible, &c.SortOrder); err != nil {
			return nil, err
		}
		if earned.Valid {
			v := earned.Float64
			c.Earned = &v
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d *DB) UpsertRubric(classID, userID int64, gradingType string, plusMinus bool, thresholds string) (GradeRubric, error) {
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM classes c JOIN semesters s ON s.id = c.semester_id
		WHERE c.id = ? AND s.user_id = ?`, classID, userID,
	).Scan(&count); err != nil || count == 0 {
		return GradeRubric{}, sql.ErrNoRows
	}
	_, err := d.sql.Exec(`
		INSERT INTO grade_rubrics (class_id, grading_type, plus_minus, thresholds)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(class_id) DO UPDATE SET
		  grading_type = excluded.grading_type,
		  plus_minus   = excluded.plus_minus,
		  thresholds   = excluded.thresholds
	`, classID, gradingType, plusMinus, thresholds)
	if err != nil {
		return GradeRubric{}, err
	}
	return d.GetRubric(classID, userID)
}

func (d *DB) DeleteRubric(classID, userID int64) error {
	_, err := d.sql.Exec(`
		DELETE FROM grade_rubrics WHERE class_id=? AND class_id IN (
			SELECT c.id FROM classes c JOIN semesters s ON s.id = c.semester_id WHERE s.user_id=?
		)`,
		classID, userID,
	)
	return err
}

func (d *DB) AddCategory(rubricID, userID int64, name string, weight float64, isExtraCredit bool, possible float64) (GradeCategory, error) {
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM grade_rubrics gr
		JOIN classes c ON c.id = gr.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE gr.id = ? AND s.user_id = ?`, rubricID, userID,
	).Scan(&count); err != nil || count == 0 {
		return GradeCategory{}, sql.ErrNoRows
	}
	var maxOrder int
	_ = d.sql.QueryRow(`SELECT COALESCE(MAX(sort_order),-1) FROM grade_categories WHERE rubric_id=?`, rubricID).Scan(&maxOrder)
	res, err := d.sql.Exec(
		`INSERT INTO grade_categories (rubric_id, name, weight, is_extra_credit, possible, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rubricID, name, weight, isExtraCredit, possible, maxOrder+1,
	)
	if err != nil {
		return GradeCategory{}, err
	}
	id, _ := res.LastInsertId()
	return d.GetCategory(id)
}

func (d *DB) GetCategory(id int64) (GradeCategory, error) {
	var c GradeCategory
	var earned sql.NullFloat64
	err := d.sql.QueryRow(
		`SELECT id, rubric_id, name, weight, is_extra_credit, earned, possible, sort_order
		 FROM grade_categories WHERE id=?`, id,
	).Scan(&c.ID, &c.RubricID, &c.Name, &c.Weight, &c.IsExtraCredit, &earned, &c.Possible, &c.SortOrder)
	if err != nil {
		return c, err
	}
	if earned.Valid {
		v := earned.Float64
		c.Earned = &v
	}
	return c, nil
}

func (d *DB) UpdateCategory(
	id, userID int64,
	name *string,
	weight *float64,
	isExtraCredit *bool,
	earned *float64, earnedSet bool,
	possible *float64,
	sortOrder *int,
) (GradeCategory, error) {
	existing, err := d.GetCategory(id)
	if err != nil {
		return GradeCategory{}, err
	}
	// Verify ownership
	var count int
	if err := d.sql.QueryRow(`
		SELECT COUNT(*) FROM grade_categories gc
		JOIN grade_rubrics gr ON gr.id = gc.rubric_id
		JOIN classes c ON c.id = gr.class_id
		JOIN semesters s ON s.id = c.semester_id
		WHERE gc.id = ? AND s.user_id = ?`, id, userID,
	).Scan(&count); err != nil || count == 0 {
		return GradeCategory{}, sql.ErrNoRows
	}
	n := existing.Name
	if name != nil {
		n = *name
	}
	w := existing.Weight
	if weight != nil {
		w = *weight
	}
	ec := existing.IsExtraCredit
	if isExtraCredit != nil {
		ec = *isExtraCredit
	}
	p := existing.Possible
	if possible != nil {
		p = *possible
	}
	so := existing.SortOrder
	if sortOrder != nil {
		so = *sortOrder
	}
	var earnedArg interface{}
	if earnedSet {
		if earned == nil {
			earnedArg = nil
		} else {
			earnedArg = *earned
		}
	} else if existing.Earned != nil {
		earnedArg = *existing.Earned
	} else {
		earnedArg = nil
	}
	_, err = d.sql.Exec(
		`UPDATE grade_categories
		 SET name=?, weight=?, is_extra_credit=?, earned=?, possible=?, sort_order=?
		 WHERE id=?`,
		n, w, ec, earnedArg, p, so, id,
	)
	if err != nil {
		return GradeCategory{}, err
	}
	return d.GetCategory(id)
}

func (d *DB) DeleteCategory(id, userID int64) error {
	res, err := d.sql.Exec(`
		DELETE FROM grade_categories WHERE id=? AND rubric_id IN (
			SELECT gr.id FROM grade_rubrics gr
			JOIN classes c ON c.id = gr.class_id
			JOIN semesters s ON s.id = c.semester_id
			WHERE s.user_id = ?
		)`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
