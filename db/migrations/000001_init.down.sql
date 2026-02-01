-- Rollback initial schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_files_updated_at ON files;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS users;

-- Drop extension (optional - might be used by other schemas)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
