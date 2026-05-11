package handlers

import (
	"fmt"
	"net/http"
	"time"

	"notes-tracker/server/db"
)

type CalendarHandler struct{ DB *db.DB }

var sourceTypeColors = map[string]string{
	"Lecture":     "#3b82f6",
	"Reading":     "#a855f7",
	"Problem Set": "#f97316",
	"Discussion":  "#22c55e",
	"Lab":         "#14b8a6",
	"Other":       "#94a3b8",
}

const deadlineColor = "#f43f5e"

// ForClass handles GET /api/classes/:id/calendar?month=YYYY-MM
func (h *CalendarHandler) ForClass(w http.ResponseWriter, r *http.Request) {
	classID, ok := parseID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}
	start, end, err := monthRange(r.URL.Query().Get("month"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "month must be YYYY-MM")
		return
	}

	cls, err := h.DB.GetClass(classID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sem, err := h.DB.GetSemester(cls.SemesterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sources, err := h.DB.SourcesInRangeForClass(classID, start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	meetings, err := h.DB.ListMeetings(classID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	events := buildClassEvents(sources, meetings, cls.Color, cls.Name, start, end, sem.StartDate, sem.EndDate)
	writeJSON(w, http.StatusOK, events)
}

// ForSemester handles GET /api/semesters/:id/calendar?month=YYYY-MM
func (h *CalendarHandler) ForSemester(w http.ResponseWriter, r *http.Request) {
	semID, ok := parseID(r, "semesterID")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid semester id")
		return
	}
	start, end, err := monthRange(r.URL.Query().Get("month"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "month must be YYYY-MM")
		return
	}

	sem, err := h.DB.GetSemester(semID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sources, err := h.DB.SourcesInRangeForSemester(semID, start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	meetings, err := h.DB.ListMeetingsForSemester(semID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var events []db.CalendarEvent
	for _, s := range sources {
		if s.Date != "" && dateBetween(s.Date, start, end) {
			events = append(events, sourceEvent(s.Source, s.ClassColor, s.ClassName))
		}
		if s.DueDate != "" && dateBetween(s.DueDate, start, end) {
			events = append(events, deadlineEvent(s.Source, s.ClassColor, s.ClassName))
		}
	}
	// Clip recurring meetings to the intersection of the requested month
	// and the semester's term. Empty term bounds → no meeting events.
	if mStart, mEnd, ok := clipRange(start, end, sem.StartDate, sem.EndDate); ok {
		for _, m := range meetings {
			for _, d := range datesForDayOfWeek(m.DayOfWeek, mStart, mEnd) {
				events = append(events, meetingEvent(m.ClassMeeting, m.ClassColor, m.ClassName, d))
			}
		}
	}
	if events == nil {
		events = []db.CalendarEvent{}
	}
	writeJSON(w, http.StatusOK, events)
}

// --- helpers ---

func buildClassEvents(sources []db.Source, meetings []db.ClassMeeting, classColor, className, start, end, semStart, semEnd string) []db.CalendarEvent {
	var events []db.CalendarEvent
	for _, s := range sources {
		if s.Date != "" && dateBetween(s.Date, start, end) {
			events = append(events, sourceEvent(s, classColor, className))
		}
		if s.DueDate != "" && dateBetween(s.DueDate, start, end) {
			events = append(events, deadlineEvent(s, classColor, className))
		}
	}
	// Clip recurring meetings to the intersection of the requested month
	// and the semester's term. Empty term bounds → no meeting events.
	if mStart, mEnd, ok := clipRange(start, end, semStart, semEnd); ok {
		for _, m := range meetings {
			for _, d := range datesForDayOfWeek(m.DayOfWeek, mStart, mEnd) {
				events = append(events, meetingEvent(m, classColor, className, d))
			}
		}
	}
	if events == nil {
		events = []db.CalendarEvent{}
	}
	return events
}

// clipRange returns the inclusive intersection of [aStart,aEnd] and [bStart,bEnd].
// If either b-side bound is empty or the ranges don't overlap, ok=false.
// YYYY-MM-DD strings are lexicographically orderable, so plain string compare works.
func clipRange(aStart, aEnd, bStart, bEnd string) (start, end string, ok bool) {
	if bStart == "" || bEnd == "" {
		return "", "", false
	}
	start = aStart
	if bStart > start {
		start = bStart
	}
	end = aEnd
	if bEnd < end {
		end = bEnd
	}
	if start > end {
		return "", "", false
	}
	return start, end, true
}

func sourceEvent(s db.Source, classColor, className string) db.CalendarEvent {
	color, ok := sourceTypeColors[s.Type]
	if !ok {
		color = classColor
	}
	return db.CalendarEvent{
		ID:     fmt.Sprintf("source-%d", s.ID),
		Title:  s.Title,
		Start:  s.Date,
		AllDay: true,
		Color:  color,
		ExtendedProps: db.CalendarEventProps{
			Type:       "source",
			SourceID:   s.ID,
			ClassID:    s.ClassID,
			ClassName:  className,
			SourceType: s.Type,
		},
	}
}

func deadlineEvent(s db.Source, classColor, className string) db.CalendarEvent {
	return db.CalendarEvent{
		ID:     fmt.Sprintf("deadline-%d", s.ID),
		Title:  "Due: " + s.Title,
		Start:  s.DueDate,
		AllDay: true,
		Color:  deadlineColor,
		ExtendedProps: db.CalendarEventProps{
			Type:       "deadline",
			SourceID:   s.ID,
			ClassID:    s.ClassID,
			ClassName:  className,
			SourceType: s.Type,
		},
	}
}

func meetingEvent(m db.ClassMeeting, classColor, className, date string) db.CalendarEvent {
	return db.CalendarEvent{
		ID:    fmt.Sprintf("meeting-%d-%s", m.ID, date),
		Title: className,
		Start: date + "T" + m.StartTime,
		End:   date + "T" + m.EndTime,
		Color: classColor + "99", // ~60% opacity via alpha in hex
		ExtendedProps: db.CalendarEventProps{
			Type:      "meeting",
			ClassID:   m.ClassID,
			ClassName: className,
		},
	}
}

func monthRange(month string) (start, end string, err error) {
	if month == "" {
		now := time.Now()
		month = fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "", "", err
	}
	start = t.Format("2006-01-02")
	// Last day = first day of next month minus one day
	last := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	end = last.Format("2006-01-02")
	return start, end, nil
}

func datesForDayOfWeek(dow int, start, end string) []string {
	s, _ := time.Parse("2006-01-02", start)
	e, _ := time.Parse("2006-01-02", end)
	var dates []string
	for d := s; !d.After(e); d = d.AddDate(0, 0, 1) {
		if int(d.Weekday()) == dow {
			dates = append(dates, d.Format("2006-01-02"))
		}
	}
	return dates
}

func dateBetween(date, start, end string) bool {
	return date >= start && date <= end
}
