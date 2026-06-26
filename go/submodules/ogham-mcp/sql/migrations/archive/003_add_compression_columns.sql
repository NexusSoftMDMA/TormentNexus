-- Ogham MCP migration 003: Add compression and scoring columns
-- Safe to run multiple times (ADD COLUMN IF NOT EXISTS)
-- Run BEFORE upgrading to v0.3.1
--
-- These columns prepare for future features:
--   importance: regex-scored content importance (decisions, errors, architecture)
--   surprise: novelty score at store time (1 - max similarity to existing memories)
--   compression_level: 0=full text, 1=gist, 2=tags only
--   original_content: preserved full text when a memory is compressed
--
-- The current MCP server ignores these columns. They have safe defaults
-- and do not affect existing behaviour.
--
-- Neon users: run this against the DIRECT endpoint, not the pooler.
-- Supabase users: paste into SQL Editor.
-- Self-hosted: run via psql or your preferred client.

ALTER TABLE memories ADD COLUMN IF NOT EXISTS importance float NOT NULL DEFAULT 0.5;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS surprise float NOT NULL DEFAULT 0.5;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS compression_level integer NOT NULL DEFAULT 0;
ALTER TABLE memories ADD COLUMN IF NOT EXISTS original_content text;
