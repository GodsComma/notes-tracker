package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"notes-tracker/server/db"
	"notes-tracker/server/handlers"
	authmw "notes-tracker/server/middleware"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	secret := []byte(jwtSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}

	database, err := db.Open(dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	auth := &handlers.AuthHandler{DB: database, Secret: secret}
	sem := &handlers.SemesterHandler{DB: database}
	cls := &handlers.ClassHandler{DB: database, DataDir: dataDir}
	src := &handlers.SourceHandler{DB: database, DataDir: dataDir}
	note := &handlers.NoteHandler{DB: database}
	srch := &handlers.SearchHandler{DB: database}
	cal := &handlers.CalendarHandler{DB: database}
	mtg := &handlers.MeetingHandler{DB: database}
	att := &handlers.AttachmentHandler{DB: database, DataDir: dataDir}
	focus := &handlers.FocusHandler{DB: database}
	grd := &handlers.GradeHandler{DB: database}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	// Public auth routes — no JWT required
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/logout", auth.Logout)
	})

	// All other API routes require a valid JWT
	r.Route("/api/v1/", func(r chi.Router) {
		r.Use(authmw.Authenticated(secret))

		// Current user
		r.Get("/auth/me", auth.Me)

		// Semesters
		r.Get("/semesters", sem.List)
		r.Post("/semesters", sem.Create)
		r.Patch("/semesters/{id}", sem.Update)
		r.Delete("/semesters/{id}", sem.Delete)

		// Classes nested under semester
		r.Get("/semesters/{semesterID}/classes", cls.List)
		r.Post("/semesters/{semesterID}/classes", cls.Create)
		r.Get("/classes/{id}", cls.Get)
		r.Patch("/classes/{id}", cls.Update)
		r.Delete("/classes/{id}", cls.Delete)

		// Sources nested under class
		r.Get("/classes/{classID}/sources", src.List)
		r.Post("/classes/{classID}/sources", src.Create)
		r.Patch("/classes/{classID}/sources/reorder", src.Reorder)
		r.Get("/sources/{id}", src.Get)
		r.Patch("/sources/{id}", src.Update)
		r.Delete("/sources/{id}", src.Delete)

		// All sources across a semester (todo panel)
		r.Get("/semesters/{semesterID}/sources", src.ListForSemester)

		// Notes (one per source)
		r.Get("/sources/{sourceID}/note", note.Get)
		r.Put("/sources/{sourceID}/note", note.Upsert)
		r.Patch("/sources/{sourceID}/note/lock", note.SetLock)

		// Meetings
		r.Get("/classes/{classID}/meetings", mtg.List)
		r.Post("/classes/{classID}/meetings", mtg.Create)
		r.Patch("/meetings/{id}", mtg.Update)
		r.Delete("/meetings/{id}", mtg.Delete)

		// Calendar event feeds
		r.Get("/classes/{id}/calendar", cal.ForClass)
		r.Get("/semesters/{semesterID}/calendar", cal.ForSemester)

		// Attachments
		r.Get("/classes/{classID}/attachments", att.ListForClass)
		r.Post("/classes/{classID}/attachments", att.UploadToClass)
		r.Get("/sources/{sourceID}/attachments", att.ListForSource)
		r.Post("/sources/{sourceID}/attachments", att.UploadToSource)
		r.Get("/attachments/{id}/file", att.ServeFile)
		r.Get("/attachments/{id}/progress", att.GetProgress)
		r.Put("/attachments/{id}/progress", att.SaveProgress)
		r.Get("/attachments/{id}/settings", att.GetFocusSettings)
		r.Put("/attachments/{id}/settings", att.SaveFocusSettings)
		r.Delete("/attachments/{id}", att.Delete)

		// Deep Read library — read-only aggregate of every PDF
		r.Get("/library/pdfs", att.ListLibraryPDFs)

		// Focus sessions
		r.Post("/focus-sessions", focus.Log)
		r.Get("/focus-sessions", focus.List)

		// Grades
		r.Get("/classes/{id}/rubric", grd.GetRubric)
		r.Put("/classes/{id}/rubric", grd.PutRubric)
		r.Delete("/classes/{id}/rubric", grd.DeleteRubric)
		r.Post("/rubrics/{id}/categories", grd.AddCategory)
		r.Patch("/grade-categories/{id}", grd.UpdateCategory)
		r.Delete("/grade-categories/{id}", grd.DeleteCategory)

		// Search
		r.Get("/search", srch.Search)
	})

	// Serve Vue SPA from client/dist in production
	if _, err := os.Stat("client/dist"); err == nil {
		distFS := os.DirFS("client/dist")
		fileServer := http.FileServer(http.FS(distFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if _, err := fs.Stat(distFS, path); err != nil {
				f, err := distFS.(fs.FS).Open("index.html")
				if err != nil {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				defer f.Close()
				body, _ := io.ReadAll(f)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(body)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	fmt.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
