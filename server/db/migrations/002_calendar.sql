CREATE TABLE IF NOT EXISTS class_meetings (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  class_id    INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
  day_of_week INTEGER NOT NULL CHECK(day_of_week BETWEEN 0 AND 6),
  start_time  TEXT NOT NULL,
  end_time    TEXT NOT NULL,
  location    TEXT NOT NULL DEFAULT ''
);

ALTER TABLE sources ADD COLUMN due_date TEXT NOT NULL DEFAULT '';
