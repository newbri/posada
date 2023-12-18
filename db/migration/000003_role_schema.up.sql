CREATE TABLE "role"
(
    "id"            uuid PRIMARY KEY,
    "role"          text        NOT NULL,
    "description"   text        NOT NULL,
    "updated_at"    timestamptz NOT NULL,
    "created_at"    timestamptz NOT NULL DEFAULT (now())
);