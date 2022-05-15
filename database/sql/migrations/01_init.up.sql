CREATE SCHEMA IF NOT EXISTS public;

CREATE TABLE IF NOT EXISTS users
(
    id           BIGINT PRIMARY KEY NOT NULL,
    name         TEXT               NOT NULL,
    username     TEXT               NOT NULL,
    state        TEXT               NOT NULL,
    is_admin     BOOLEAN            NOT NULL DEFAULT false,
    settings     JSONB              NOT NULL DEFAULT '{}',
    invite_code  TEXT               NOT NULL,
    active_until TIMESTAMP          NOT NULL,
    created_at   TIMESTAMP          NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invite_codes
(
    code       TEXT UNIQUE PRIMARY KEY NOT NULL,
    created_by BIGINT REFERENCES users (id),
    used_by    BIGINT REFERENCES users (id),
    used_at    TIMESTAMP,
    created_at TIMESTAMP               NOT NULL DEFAULT now()
);

ALTER TABLE users
    ADD CONSTRAINT users_invite_codes_fk
        FOREIGN KEY (invite_code) REFERENCES invite_codes (code);

CREATE TABLE IF NOT EXISTS posts
(
    id           BIGSERIAL PRIMARY KEY,
    next_post_id BIGSERIAL REFERENCES posts (id),
    content      JSONB                 NOT NULL,
    updated_at   TIMESTAMP             NOT NULL DEFAULT now(),
    created_at   TIMESTAMP             NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_entries
(
    post_id    BIGSERIAL NOT NULL REFERENCES posts (id),
    user_id    BIGINT    NOT NULL REFERENCES users (id),
    status     TEXT      NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT post_entries_pk
        PRIMARY KEY (post_id, user_id)
);

