PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS semesters (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL,
  start_date TEXT NOT NULL DEFAULT '',
  end_date   TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS classes (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  semester_id INTEGER NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  code        TEXT NOT NULL DEFAULT '',
  instructor  TEXT NOT NULL DEFAULT '',
  schedule    TEXT NOT NULL DEFAULT '',
  color       TEXT NOT NULL DEFAULT '#6366f1'
);

CREATE TABLE IF NOT EXISTS sources (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  class_id   INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
  type       TEXT NOT NULL DEFAULT 'Lecture' CHECK(type IN ('Lecture','Reading','Problem Set','Discussion','Lab','Other')),
  title      TEXT NOT NULL,
  date       TEXT NOT NULL DEFAULT '',
  reference  TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS notes (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  source_id  INTEGER NOT NULL UNIQUE REFERENCES sources(id) ON DELETE CASCADE,
  body       TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
  note_id UNINDEXED,
  content,
  tokenize='porter ascii'
);
