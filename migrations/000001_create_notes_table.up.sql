CREATE TABLE notes (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL,
    title      VARCHAR(255) NOT NULL,
    content    TEXT,
    color      VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    is_pinned  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_user_pinned ON notes(user_id, is_pinned DESC, updated_at DESC);
