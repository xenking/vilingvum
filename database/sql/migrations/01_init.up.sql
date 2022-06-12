CREATE SCHEMA IF NOT EXISTS public;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id           BIGINT PRIMARY KEY NOT NULL,
    name         TEXT               NOT NULL,
    username     TEXT               NOT NULL,
    email        TEXT,
    phone_number TEXT,
    state        TEXT               NOT NULL,
    is_admin     BOOLEAN            NOT NULL DEFAULT false,
    settings     JSONB              NOT NULL DEFAULT '{}',
    active_until TIMESTAMP,
    created_at   TIMESTAMP          NOT NULL DEFAULT now()
);


CREATE TABLE IF NOT EXISTS topics
(
    id            BIGSERIAL PRIMARY KEY,
    next_topic_id BIGSERIAL REFERENCES topics (id),
    type          TEXT      NOT NULL,
    content       JSONB     NOT NULL,
    updated_at    TIMESTAMP NOT NULL DEFAULT now(),
    created_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_answers
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT    NOT NULL REFERENCES users (id),
    topic_id   BIGSERIAL NOT NULL REFERENCES topics (id),
    response   JSONB     NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS user_answers_idx ON user_answers (user_id, topic_id);
CREATE INDEX IF NOT EXISTS user_answers_users ON user_answers (user_id);

CREATE TABLE IF NOT EXISTS dictionary
(
    id       BIGSERIAL PRIMARY KEY,
    topic_id BIGSERIAL NOT NULL REFERENCES topics (id),
    word     TEXT      NOT NULL,
    meaning  TEXT      NOT NULL
);

CREATE TABLE IF NOT EXISTS invoices
(
    uuid       UUID PRIMARY KEY,
    user_id    BIGINT    NOT NULL REFERENCES users (id),
    payload    JSONB     NOT NULL,
    email      TEXT,
    charge_id  TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at TIMESTAMP NOT NULL DEFAULT now()
);