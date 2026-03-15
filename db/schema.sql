CREATE TABLE companies (
    id UUID PRIMARY KEY,
    name VARCHAR(15) NOT NULL,
    description TEXT,
    employees INT NOT NULL CHECK (employees >= 0),
    registered BOOL NOT NULL,
    company_type VARCHAR(50) NOT NULL CHECK (company_type IN ('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID NOT NULL PRIMARY KEY,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
)