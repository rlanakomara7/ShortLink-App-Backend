CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_url TEXT NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_slug_format
        CHECK (slug ~ '^[A-Za-z0-9-]{3,50}$'),

    CONSTRAINT chk_slug_reserved
        CHECK (LOWER(slug) NOT IN (
            'api',
            'login',
            'register',
            'dashboard'
        ))
);

CREATE INDEX idx_links_user_id
ON links(user_id);
