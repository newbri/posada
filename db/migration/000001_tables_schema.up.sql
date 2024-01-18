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

CREATE TABLE IF NOT EXISTS "property"
(
    "internal_id" uuid PRIMARY KEY,
    "external_id" text        NOT NULL UNIQUE,
    "name"        text        NOT NULL,
    "address"     text        NOT NULL,
    "state"       text        NOT NULL,
    "country"     text        NOT NULL,
    "postal_code" text        NOT NULL,
    "phone"       text        NOT NULL,
    "email"       text        NOT NULL,
    "expired_at"  timestamptz NOT NULL,
    "created_at"  timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS "property_user"
(
    "internal_id"          uuid PRIMARY KEY,
    "external_id"          text NOT NULL UNIQUE,
    "property_internal_id" uuid NOT NULL,
    "username"             text NOT NULL
);

ALTER TABLE IF EXISTS "sessions"
    ADD CONSTRAINT "fk_sessions_username"
        FOREIGN KEY ("username") REFERENCES "users" ("username");

CREATE UNIQUE INDEX IF NOT EXISTS "role_external_id_index" ON "role" ("external_id");
ALTER TABLE IF EXISTS "users"
    ADD CONSTRAINT "fk_role_id" FOREIGN KEY ("role_id") REFERENCES "role" ("internal_id");

ALTER TABLE IF EXISTS "property_user"
    ADD CONSTRAINT "fk_property"
        FOREIGN KEY (property_internal_id)
            REFERENCES "property" (internal_id)
            ON DELETE CASCADE;

ALTER TABLE IF EXISTS "property_user"
    ADD CONSTRAINT "fk_user"
        FOREIGN KEY (username)
            REFERENCES users (username)
            ON DELETE CASCADE;

CREATE SEQUENCE IF NOT EXISTS "role_sequence" START 101;

INSERT INTO "role" (internal_id, name, description, external_id)
VALUES (gen_random_uuid(), 'admin', E'Administrator\'s role', CONCAT('URE',nextval('role_sequence')));

INSERT INTO "role" (internal_id, name, description, external_id)
VALUES (gen_random_uuid(), 'customer', E'Customer\'s role', CONCAT('URE',nextval('role_sequence')));

INSERT INTO "role" (internal_id, name, description, external_id)
VALUES (gen_random_uuid(), 'visitor', E'Visitor\'s role', CONCAT('URE',nextval('role_sequence')));