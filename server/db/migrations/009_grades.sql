CREATE TABLE IF NOT EXISTS grade_rubrics (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  class_id     INTEGER NOT NULL UNIQUE REFERENCES classes(id) ON DELETE CASCADE,
  grading_type TEXT    NOT NULL DEFAULT 'letter',
  plus_minus   INTEGER NOT NULL DEFAULT 0,
  thresholds   TEXT    NOT NULL DEFAULT '',
  created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS grade_categories (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  rubric_id       INTEGER NOT NULL REFERENCES grade_rubrics(id) ON DELETE CASCADE,
  name            TEXT    NOT NULL,
  weight          REAL    NOT NULL,
  is_extra_credit INTEGER NOT NULL DEFAULT 0,
  earned          REAL,
  possible        REAL    NOT NULL DEFAULT 100,
  sort_order      INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_grade_categories_rubric_id ON grade_categories(rubric_id);
