BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    body TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    delivered_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_messages_conv_id_id
ON messages(conversation_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_messages_delivered_at
ON messages(delivered_at);

COMMIT;
