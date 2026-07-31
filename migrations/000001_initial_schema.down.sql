-- 000001_initial_schema.down.sql
-- SPEC-02: Database Schema & Migrations
-- Reverse of up migration: DROP TABLE in reverse FK dependency order.
-- Dev only -- production is forward-only (Constitution principle VI).

DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS prospections CASCADE;
DROP TABLE IF EXISTS clients CASCADE;
DROP TABLE IF EXISTS properties CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
