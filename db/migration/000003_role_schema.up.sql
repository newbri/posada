CREATE TABLE "role"
(
    "id"            uuid PRIMARY KEY,
    "name"          text        NOT NULL,
    "description"   text        NOT NULL,
    "updated_at"    timestamptz NOT NULL DEFAULT (now()),
    "created_at"    timestamptz NOT NULL DEFAULT (now())
);