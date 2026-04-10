CREATE TABLE IF NOT EXISTS repositories (
                                            id BIGSERIAL PRIMARY KEY,
                                            full_name TEXT NOT NULL UNIQUE,
                                            owner TEXT NOT NULL,
                                            name TEXT NOT NULL,
                                            last_seen_tag TEXT NULL,
                                            last_checked_at TIMESTAMPTZ NULL,
                                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );