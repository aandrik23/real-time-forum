BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    actor_id INTEGER,
    type TEXT NOT NULL,
    target_id INTEGER,
    payload TEXT,
    read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_read_created
ON notifications(user_id, read, created_at DESC);

COMMIT;
