-- Add password change tracking fields to users table
-- This migration adds support for forcing password change on first login

-- Add must_change_password field (default true for new users)
ALTER TABLE users ADD COLUMN must_change_password BOOLEAN DEFAULT true;

-- Add last_password_change timestamp
ALTER TABLE users ADD COLUMN last_password_change TIMESTAMPTZ;

-- Update existing users to not require password change (they are already active)
UPDATE users SET must_change_password = false WHERE is_active = true;

-- Create index for quick lookups of users who must change password
CREATE INDEX idx_users_must_change_password ON users(must_change_password) WHERE must_change_password = true;

-- Add comment for documentation
COMMENT ON COLUMN users.must_change_password IS 'Forces user to change password on first login (invite-only system)';
COMMENT ON COLUMN users.last_password_change IS 'Timestamp of last password change';

