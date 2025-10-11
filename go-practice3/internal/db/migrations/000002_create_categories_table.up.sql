CREATE TABLE IF NOT EXISTS categories (
                                          id         BIGSERIAL PRIMARY KEY,
                                          name       TEXT NOT NULL,
                                          user_id    BIGINT NULL REFERENCES users(id) ON DELETE CASCADE,
                                          CONSTRAINT categories_user_name_uq UNIQUE(user_id, name)
);


CREATE UNIQUE INDEX IF NOT EXISTS categories_global_name_uq
    ON categories (name)
    WHERE user_id IS NULL;


CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
