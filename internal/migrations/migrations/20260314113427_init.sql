-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY,
    name VARCHAR(15) NOT NULL,
    description TEXT,
    employees INT NOT NULL CHECK (employees >= 0),
    registered BOOL NOT NULL,
    company_type VARCHAR(50) NOT NULL CHECK (company_type IN ('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE companies;
-- +goose StatementEnd
