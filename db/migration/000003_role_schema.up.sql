CREATE TABLE IF NOT EXISTS "role"
(
    "internal_id" uuid PRIMARY KEY,
    "name"        text        NOT NULL,
    "description" text        NOT NULL,
    "external_id" text        NOT NULL UNIQUE,
    "updated_at"  timestamptz NOT NULL DEFAULT (now()),
    "created_at"  timestamptz NOT NULL DEFAULT (now())
);

CREATE SEQUENCE IF NOT EXISTS "role_sequence" START 101;
CREATE UNIQUE INDEX IF NOT EXISTS "role_external_id_index" ON "role" ("external_id");