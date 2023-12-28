ALTER TABLE IF EXISTS "sessions"
    ADD CONSTRAINT "fk_sessions_username"
        FOREIGN KEY ("username") REFERENCES "users" ("username");

CREATE SEQUENCE IF NOT EXISTS "role_sequence" START 101;
CREATE UNIQUE INDEX IF NOT EXISTS "role_external_id_index" ON "role" ("external_id");
ALTER TABLE IF EXISTS "users"
    ADD CONSTRAINT "fk_role_id" FOREIGN KEY ("role_id") REFERENCES "role" ("internal_id");