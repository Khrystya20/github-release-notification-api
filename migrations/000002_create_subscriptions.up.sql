CREATE TABLE IF NOT EXISTS subscriptions (
                                             id BIGSERIAL PRIMARY KEY,
                                             email TEXT NOT NULL,
                                             repository_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    confirm_token TEXT NOT NULL UNIQUE,
    unsubscribe_token TEXT NOT NULL UNIQUE,
    confirmed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT subscriptions_email_repository_unique UNIQUE (email, repository_id)
    );

CREATE INDEX IF NOT EXISTS idx_subscriptions_email
    ON subscriptions(email);

CREATE INDEX IF NOT EXISTS idx_subscriptions_confirmed_active
    ON subscriptions(confirmed, active);

CREATE INDEX IF NOT EXISTS idx_subscriptions_repository_id
    ON subscriptions(repository_id);