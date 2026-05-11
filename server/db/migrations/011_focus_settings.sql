CREATE TABLE attachment_focus_settings (
    attachment_id          INTEGER PRIMARY KEY REFERENCES attachments(id) ON DELETE CASCADE,
    duration_min           INTEGER NOT NULL DEFAULT 25,
    theme                  TEXT    NOT NULL DEFAULT 'default',
    ruler                  TEXT    NOT NULL DEFAULT 'off',
    suppress_notifications INTEGER NOT NULL DEFAULT 1,
    annotations            TEXT    NOT NULL DEFAULT 'on',
    auto_hide_drawer       INTEGER NOT NULL DEFAULT 1,
    updated_at             DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
