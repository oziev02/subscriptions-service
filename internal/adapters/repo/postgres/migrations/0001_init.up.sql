CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
                                             id UUID PRIMARY KEY,
                                             service_name TEXT NOT NULL CHECK (char_length(service_name) > 0),
    price INTEGER NOT NULL CHECK (price >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_date IS NULL OR start_date <= end_date)
    );

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service ON subscriptions (service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start ON subscriptions (start_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end ON subscriptions (end_date);
