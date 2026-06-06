CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accounts (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance    NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    currency   VARCHAR(3) NOT NULL DEFAULT 'RUB',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cards (
    id               SERIAL PRIMARY KEY,
    account_id       INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    encrypted_number TEXT NOT NULL,
    encrypted_expiry TEXT NOT NULL,
    cvv_hash         VARCHAR(255) NOT NULL,
    hmac_number      VARCHAR(255) NOT NULL,
    last_four        VARCHAR(4) NOT NULL,
    created_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id               SERIAL PRIMARY KEY,
    from_account_id  INTEGER REFERENCES accounts(id),
    to_account_id    INTEGER REFERENCES accounts(id),
    amount           NUMERIC(18, 2) NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    description      TEXT,
    created_at       TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS credits (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES users(id),
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    principal       NUMERIC(18, 2) NOT NULL,
    interest_rate   NUMERIC(6, 2) NOT NULL,
    term_months     INTEGER NOT NULL,
    monthly_payment NUMERIC(18, 2) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_schedules (
    id           SERIAL PRIMARY KEY,
    credit_id    INTEGER NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
    payment_date TIMESTAMP NOT NULL,
    amount       NUMERIC(18, 2) NOT NULL,
    principal    NUMERIC(18, 2) NOT NULL,
    interest     NUMERIC(18, 2) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid_at      TIMESTAMP,
    created_at   TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_from ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to ON transactions(to_account_id);
CREATE INDEX IF NOT EXISTS idx_credits_user_id ON credits(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_credit_id ON payment_schedules(credit_id);
CREATE INDEX IF NOT EXISTS idx_payment_schedules_date ON payment_schedules(payment_date);
