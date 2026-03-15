-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL PRIMARY KEY,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO users (id, username, password)
VALUES
    ('23ff21a2-cb39-43bd-a3ea-877060ea244d', 'admin', 'admin'),
    ('01196871-f51a-4668-8daa-41b59c6915ce', 'user', 'pass');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
