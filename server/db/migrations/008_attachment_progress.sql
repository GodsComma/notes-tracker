CREATE TABLE IF NOT EXISTS attachment_progress (
  attachment_id INTEGER PRIMARY KEY REFERENCES attachments(id) ON DELETE CASCADE,
  last_page     INTEGER NOT NULL DEFAULT 1,
  updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
