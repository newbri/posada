CREATE TABLE IF NOT EXISTS "role"
(
    "id"            uuid PRIMARY KEY,
    "name"          text        NOT NULL,
    "description"   text        NOT NULL,
    "external_id"   text        NOT NULL UNIQUE,
    "updated_at"    timestamptz NOT NULL DEFAULT (now()),
    "created_at"    timestamptz NOT NULL DEFAULT (now())
);

CREATE SEQUENCE IF NOT EXISTS role_sequence START 101;