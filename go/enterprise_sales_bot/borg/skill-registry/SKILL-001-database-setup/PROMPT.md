# Task: SKILL-001 - Build Deduplicating Skill Database

## Goal
Design and implement the SQLite database schema in `.pi/db/skills.db` (or unified `tormentnexus.db`) to store, query, and deduplicate skills based on content.

## Key Requirements
- deduplicate based on content using a similarity threshold of 98%.
- merge content if they are revisions of each other.
- store as JSON or structured SQL schema.
