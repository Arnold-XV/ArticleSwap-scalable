-- Schema awal ArticleSwap.
-- File ini dijalankan otomatis oleh container PostgreSQL saat volume database masih kosong.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES users(id),
    receiver_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_processing_results (
    article_id UUID PRIMARY KEY REFERENCES articles(id) ON DELETE CASCADE,
    stemmed_content TEXT,
    word_frequencies_json JSONB,
    stemming_status TEXT NOT NULL DEFAULT 'queued',
    wordcloud_status TEXT NOT NULL DEFAULT 'queued',
    processing_started_at TIMESTAMPTZ,
    processing_finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key TEXT PRIMARY KEY,
    request_hash TEXT NOT NULL,
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pipeline_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    service_name TEXT NOT NULL,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_receiver_created ON articles(receiver_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_content_hash ON articles(content_hash);
CREATE INDEX IF NOT EXISTS idx_pipeline_events_article ON pipeline_events(article_id, created_at ASC);

INSERT INTO users (name, email) VALUES
    ('Maulana Faris Al Ghifari', 'maulana.faris@example.local'),
    ('Raditya Nathaniel Nugroho', 'raditya.nathaniel@example.local'),
    ('Ajie Armansyah Sunaryo', 'ajie.armansyah@example.local'),
    ('Arnoldus Dharma Wasesa M.', 'arnoldus.dharma@example.local'),
    ('Aliya Khairun Nisa', 'aliya.khairun@example.local')
ON CONFLICT (email) DO NOTHING;

