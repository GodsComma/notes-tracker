-- Relax the attachments ownership constraint to allow a third mode:
-- user-owned direct library uploads (user_id set, class_id + source_id NULL).
-- SQLite does not support ALTER CONSTRAINT, so we recreate the table.

PRAGMA foreign_keys = OFF;

CREATE TABLE attachments_new (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER REFERENCES users(id)    ON DELETE CASCADE,
    class_id   INTEGER REFERENCES classes(id)  ON DELETE CASCADE,
    source_id  INTEGER REFERENCES sources(id)  ON DELETE CASCADE,
    filename   TEXT    NOT NULL,
    mime_type  TEXT    NOT NULL DEFAULT '',
    size_bytes INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now')),
    CHECK (
        (class_id IS NOT NULL AND source_id IS NULL  AND user_id IS NULL) OR
        (class_id IS NULL     AND source_id IS NOT NULL AND user_id IS NULL) OR
        (class_id IS NULL     AND source_id IS NULL  AND user_id IS NOT NULL)
    )
);

INSERT INTO attachments_new (id, class_id, source_id, filename, mime_type, size_bytes, created_at)
SELECT                        id, class_id, source_id, filename, mime_type, size_bytes, created_at
FROM attachments;

DROP TABLE attachments;
ALTER TABLE attachments_new RENAME TO attachments;

PRAGMA foreign_keys = ON;
