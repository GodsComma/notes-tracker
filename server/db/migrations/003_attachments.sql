CREATE TABLE IF NOT EXISTS attachments (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  class_id    INTEGER REFERENCES classes(id)  ON DELETE CASCADE,
  source_id   INTEGER REFERENCES sources(id)  ON DELETE CASCADE,
  filename    TEXT    NOT NULL,
  mime_type   TEXT    NOT NULL DEFAULT '',
  size_bytes  INTEGER NOT NULL DEFAULT 0,
  created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
  CHECK (
    (class_id IS NOT NULL AND source_id IS NULL) OR
    (class_id IS NULL     AND source_id IS NOT NULL)
  )
);
