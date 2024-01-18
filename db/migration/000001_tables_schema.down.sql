ALTER TABLE IF EXISTS "users" DROP CONSTRAINT IF EXISTS "fk_role";
DROP SEQUENCE IF EXISTS "role_sequence";
DROP INDEX IF EXISTS "role_external_id_index";
DROP INDEX IF EXISTS "fk_sessions_username";
DROP INDEX IF EXISTS "fk_role_id";
ALTER TABLE IF EXISTS "property_user" DROP CONSTRAINT IF EXISTS "fk_property";
ALTER TABLE IF EXISTS "property_user" DROP CONSTRAINT IF EXISTS "fk_user";

DROP TABLE IF EXISTS "sessions";
DROP TABLE IF EXISTS "users";
DROP TABLE IF EXISTS "role";
DROP TABLE IF EXISTS "property";
DROP TABLE IF EXISTS "property_user";