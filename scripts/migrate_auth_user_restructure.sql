-- One-time migration for existing Docker volumes after auth/user restructure.
-- Run while services are STOPPED, from host with psql access to all DBs.
--
-- Example (adjust ports if needed):
--   psql -h localhost -p 15433 -U auth_user -d auth_db -f scripts/migrate_auth_user_restructure_auth.sql
-- This file documents the full sequence; run each section against the correct database.

-- ─── 1. user_db: copy profiles from auth_db.users ───────────────────
-- Connect to user_db (port 5434) and run:
/*
INSERT INTO users (id, name, email, student_id, is_active, created_at, updated_at)
SELECT id, name, email, student_id, is_active, created_at, updated_at
FROM dblink('host=auth-db port=5432 dbname=auth_db user=auth_user password=auth_pass',
  'SELECT id, name, email, student_id, is_active, created_at, updated_at FROM users')
  AS t(id uuid, name text, email text, student_id uuid, is_active bool, created_at timestamptz, updated_at timestamptz)
ON CONFLICT (id) DO NOTHING;
*/

-- Simpler approach without dblink (run from host with two exports):
--   docker exec auth-db pg_dump -U auth_user -d auth_db -t users --data-only --column-inserts > /tmp/auth_users.sql
--   Edit table name if needed, then load into user-db.

-- ─── 2. auth_db: credentials from legacy users.password ─────────────
/*
INSERT INTO user_credentials (user_id, password_hash, created_at, updated_at)
SELECT id, password, created_at, updated_at FROM users
ON CONFLICT (user_id) DO NOTHING;
*/

-- ─── 3. auth_db: copy RBAC from user_db ─────────────────────────────
-- Export roles, permissions, role_permissions from user_db and import into auth_db.
-- GORM auto-migrate on auth-service startup creates empty tables; then:
/*
INSERT INTO permissions SELECT * FROM user_db.permissions ON CONFLICT DO NOTHING;
INSERT INTO roles SELECT * FROM user_db.roles ON CONFLICT DO NOTHING;
INSERT INTO role_permissions SELECT * FROM user_db.role_permissions ON CONFLICT DO NOTHING;
*/

-- ─── 4. auth_db: user_roles from school_db.user_schools.role_id ─────
/*
INSERT INTO user_roles (user_id, school_id, role_id, created_at, updated_at)
SELECT user_id, school_id, role_id, created_at, updated_at FROM school_db.user_schools
ON CONFLICT DO NOTHING;
*/

-- ─── 5. school_db: drop role_id from memberships ────────────────────
/*
ALTER TABLE user_schools DROP COLUMN IF EXISTS role_id;
*/

-- ─── 6. auth_db: drop legacy users table ────────────────────────────
/*
DROP TABLE IF EXISTS users;
*/

-- ─── 7. user_db: drop old RBAC tables (optional cleanup) ────────────
/*
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
*/

-- Fresh installs: skip this script — services auto-migrate on startup.
