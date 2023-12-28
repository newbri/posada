CREATE TABLE "users"
(
    "username"            text PRIMARY KEY,
    "hashed_password"     text        NOT NULL,
    "full_name"           text        NOT NULL,
    "email"               text UNIQUE NOT NULL,
    "role_id"             uuid        NOT NULL,
    "password_changed_at" timestamptz NOT NULL DEFAULT (now()),
    "created_at"          timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions"
(
    "id"            uuid PRIMARY KEY,
    "username"      text        NOT NULL,
    "refresh_token" text        NOT NULL,
    "user_agent"    text        NOT NULL,
    "client_ip"     text        NOT NULL,
    "is_blocked"    boolean     NOT NULL DEFAULT false,
    "expired_at"    timestamptz NOT NULL,
    "created_at"    timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS "role"
(
    "internal_id" uuid PRIMARY KEY,
    "name"        text        NOT NULL,
    "description" text        NOT NULL,
    "external_id" text        NOT NULL UNIQUE,
    "updated_at"  timestamptz NOT NULL DEFAULT (now()),
    "created_at"  timestamptz NOT NULL DEFAULT (now())
);