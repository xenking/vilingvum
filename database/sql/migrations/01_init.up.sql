CREATE SCHEMA IF NOT EXISTS public;

CREATE TABLE IF NOT EXISTS users
(
    id           BIGINT PRIMARY KEY NOT NULL,
    name         TEXT               NOT NULL,
    username     TEXT               NOT NULL,
    state        TEXT               NOT NULL,
    is_admin     BOOLEAN            NOT NULL DEFAULT false,
    settings     JSONB              NOT NULL DEFAULT '{}',
    dictionary   JSONB              NOT NULL DEFAULT '{}',
    active_until TIMESTAMP          NOT NULL,
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
    user_id    BIGINT    NOT NULL REFERENCES users (id),
    topic_id   BIGSERIAL NOT NULL REFERENCES topics (id),
    response   JSONB     NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT user_answers_pk
        PRIMARY KEY (user_id, topic_id)
);
