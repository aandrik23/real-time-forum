BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    target_type TEXT NOT NULL, -- 'post' or 'comment'
    target_id INTEGER NOT NULL,
    is_like BOOLEAN NOT NULL,  -- TRUE for like, FALSE for dislike
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, target_type, target_id), -- prevents multiple votes
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

COMMIT;