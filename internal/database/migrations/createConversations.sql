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

CREATE TABLE IF NOT EXISTS conversation_reads (
    conversation_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    last_read_msg_id INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (conversation_id, user_id)
);

COMMIT;