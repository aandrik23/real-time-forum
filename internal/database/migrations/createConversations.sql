BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    last_message_id INTEGER,
    last_message_at INTEGER,
    last_read_msg_id INTEGER DEFAULT 0,
    UNIQUE(user1_id, user2_id)
);

COMMIT;