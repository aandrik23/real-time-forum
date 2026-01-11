BEGIN TRANSACTION;

INSERT INTO users (username, email, password, role) VALUES
('alice', 'alice@example.com', 'hashedpassword1', 'user'),
('bob', 'bob@example.com', 'hashedpassword2', 'admin'),
('charlie', 'charlie@example.com', 'hashedpassword3', 'user');


COMMIT;