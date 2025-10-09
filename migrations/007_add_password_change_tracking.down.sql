-- Rollback password change tracking migration

-- Drop index
DROP INDEX IF EXISTS idx_users_must_change_password;

-- Drop columns
ALTER TABLE users DROP COLUMN IF EXISTS last_password_change;
ALTER TABLE users DROP COLUMN IF EXISTS must_change_password;

