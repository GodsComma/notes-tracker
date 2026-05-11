BEGIN;

CREATE TABLE users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       TEXT    NOT NULL UNIQUE,
    password    TEXT    NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Placeholder owner so existing rows can be back-filled.
-- The real owner registers via /api/v1/auth/register on first launch.
INSERT INTO users (id, email, password) VALUES (1, 'owner@localhost', 'locked');

ALTER TABLE semesters      ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1 REFERENCES users(id);
ALTER TABLE focus_sessions ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1 REFERENCES users(id);

CREATE INDEX idx_semesters_user      ON semesters(user_id);
CREATE INDEX idx_focus_sessions_user ON focus_sessions(user_id);

COMMIT;
