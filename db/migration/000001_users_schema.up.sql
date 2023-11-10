CREATE TABLE "users"
(
    "username"            text PRIMARY KEY,
    "hashed_password"     text        NOT NULL,
    "full_name"           text        NOT NULL,
    "email"               text UNIQUE NOT NULL,
    "password_changed_at" timestamptz NOT NULL DEFAULT (now()),
    "created_at"          timestamptz NOT NULL DEFAULT (now())
);