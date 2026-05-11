CREATE TABLE IF NOT EXISTS focus_sessions (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  date          TEXT NOT NULL,
  duration_sec  INTEGER NOT NULL,
  attachment_id INTEGER NOT NULL,
  filename      TEXT NOT NULL DEFAULT '',
  completed     INTEGER NOT NULL DEFAULT 0,
  created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
