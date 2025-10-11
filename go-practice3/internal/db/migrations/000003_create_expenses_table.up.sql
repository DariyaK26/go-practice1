CREATE TABLE IF NOT EXISTS expenses (
                                        id           BIGSERIAL PRIMARY KEY,
                                        user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                        category_id  BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
                                        amount       NUMERIC(12,2) NOT NULL CHECK (amount > 0),
                                        currency     CHAR(3) NOT NULL,
                                        spent_at     TIMESTAMPTZ NOT NULL,
                                        created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                        note         TEXT
);


CREATE INDEX IF NOT EXISTS idx_expenses_user_id ON expenses(user_id);
CREATE INDEX IF NOT EXISTS idx_expenses_user_id_spent_at ON expenses(user_id, spent_at);
