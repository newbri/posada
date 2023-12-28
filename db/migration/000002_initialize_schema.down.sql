ALTER TABLE "users" DROP CONSTRAINT IF EXISTS "fk_role";
DROP SEQUENCE IF EXISTS "role_sequence";
DROP INDEX IF EXISTS "role_external_id_index";
DROP INDEX IF EXISTS "fk_sessions_username";
DROP INDEX IF EXISTS "fk_role_id";