"""
Borg Tiered Memory System — L1 (Hot) / L2 (Warm) / L3 (Cold)

Heat-based memory promotion/demotion inspired by:
- LangMem 3-tier (Semantic/Episodic/Procedural)
- MemoryOS heat-based promotion
- Letta cross-session persistence

L1 Hot:    Active session facts, working context (in-memory dict)
L2 Warm:   Recent extracted facts, success patterns (SQLite table)
L3 Cold:   Compressed archive of past extractions (SQLite table, compressed)

Memory flows: L1 -> L2 on session end | L2 -> L3 on decay | L3 -> L2 on recall
"""

import os
import json
import time
import sqlite3
import hashlib
import logging
from datetime import datetime, timezone
from collections import defaultdict

logger = logging.getLogger(__name__)


class TieredMemory:
    """Three-layer memory with heat-based promotion for Borg Intelligence."""

    DB_PATH = 'bookmarks.db'

    # Heat thresholds
    PROMOTE_THRESHOLD = 5.0   # accesses needed to promote L2 -> L1
    DEMOTE_THRESHOLD = 0.1    # heat decay below this -> L2 -> L3
    HEAT_DECAY = 0.85         # multiply heat by this each decay cycle
    HEAT_BOOST = 1.0          # add this on each access

    def __init__(self, db_path=None):
        self.db_path = db_path or self.DB_PATH
        self._init_tables()
        # L1: in-memory working context
        self.l1_cache = {}  # key -> {value, heat, last_access}
        self.l1_max = 100   # max items in L1

    def _get_conn(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn

    def _init_tables(self):
        conn = self._get_conn()
        c = conn.cursor()

        # L2: Warm memory — recent facts, patterns, success templates
        c.execute("""
            CREATE TABLE IF NOT EXISTS memory_l2_warm (
                key TEXT PRIMARY KEY,
                value TEXT NOT NULL,
                category TEXT DEFAULT 'general',
                source_url TEXT,
                heat REAL DEFAULT 1.0,
                access_count INTEGER DEFAULT 0,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                last_accessed DATETIME DEFAULT CURRENT_TIMESTAMP,
                expires_at DATETIME
            )
        """)

        # L3: Cold memory — compressed archive
        c.execute("""
            CREATE TABLE IF NOT EXISTS memory_l3_cold (
                key TEXT PRIMARY KEY,
                compressed_value TEXT NOT NULL,
                original_size INTEGER,
                category TEXT DEFAULT 'archive',
                archived_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                recall_count INTEGER DEFAULT 0
            )
        """)

        # Memory access log for analytics
        c.execute("""
            CREATE TABLE IF NOT EXISTS memory_access_log (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                memory_key TEXT,
                tier TEXT,
                action TEXT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        """)

        # Skill registry — learned extraction patterns
        c.execute("""
            CREATE TABLE IF NOT EXISTS borg_skills (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                skill_name TEXT UNIQUE NOT NULL,
                skill_type TEXT DEFAULT 'extraction',
                pattern TEXT NOT NULL,
                success_count INTEGER DEFAULT 0,
                fail_count INTEGER DEFAULT 0,
                last_used DATETIME,
                evolved_from TEXT,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                metadata JSON
            )
        """)

        # Tool registry — discovered MCP tools and capabilities
        c.execute("""
            CREATE TABLE IF NOT EXISTS borg_tool_registry (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                tool_name TEXT UNIQUE NOT NULL,
                tool_type TEXT DEFAULT 'mcp',
                schema_json TEXT,
                endpoint TEXT,
                capability_tags TEXT,
                call_count INTEGER DEFAULT 0,
                success_rate REAL DEFAULT 0.0,
                last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
                discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                metadata JSON
            )
        """)

        conn.commit()
        conn.close()
        logger.info("Tiered memory tables initialized")

    # ------------------------------------------------------------------
    # L1: HOT MEMORY (in-process working context)
    # ------------------------------------------------------------------
    def l1_set(self, key, value, category='general'):
        """Store in L1 hot cache. Evicts coldest item if full."""
        if len(self.l1_cache) >= self.l1_max:
            self._evict_coldest_l1()
        self.l1_cache[key] = {
            'value': value,
            'heat': 1.0,
            'category': category,
            'last_access': time.time(),
        }

    def l1_get(self, key):
        """Retrieve from L1, boosting heat on access."""
        if key in self.l1_cache:
            entry = self.l1_cache[key]
            entry['heat'] += self.HEAT_BOOST
            entry['last_access'] = time.time()
            return entry['value']
        return None

    def l1_get_all(self, category=None):
        """Get all L1 entries, optionally filtered by category."""
        results = {}
        for key, entry in self.l1_cache.items():
            if category is None or entry.get('category') == category:
                results[key] = entry['value']
        return results

    def _evict_coldest_l1(self):
        """Evict the coldest L1 item to L2."""
        if not self.l1_cache:
            return
        coldest_key = min(self.l1_cache, key=lambda k: self.l1_cache[k]['heat'])
        entry = self.l1_cache.pop(coldest_key)
        self.l2_set(coldest_key, entry['value'], entry.get('category', 'general'))
        logger.debug("Evicted L1->L2: %s (heat=%.2f)", coldest_key, entry['heat'])

    def flush_l1_to_l2(self):
        """Flush all L1 entries to L2 (call on session end)."""
        count = 0
        for key, entry in list(self.l1_cache.items()):
            self.l2_set(key, entry['value'], entry.get('category', 'general'))
            count += 1
        self.l1_cache.clear()
        logger.info("Flushed %d L1 entries to L2", count)
        return count

    # ------------------------------------------------------------------
    # L2: WARM MEMORY (SQLite — recent facts, patterns)
    # ------------------------------------------------------------------
    def l2_set(self, key, value, category='general', source_url=None):
        """Store in L2 warm memory. Upsert pattern."""
        conn = self._get_conn()
        try:
            conn.execute("""
                INSERT INTO memory_l2_warm (key, value, category, source_url, heat, access_count, last_accessed)
                VALUES (?, ?, ?, ?, 1.0, 0, ?)
                ON CONFLICT(key) DO UPDATE SET
                    value=excluded.value,
                    category=excluded.category,
                    source_url=COALESCE(excluded.source_url, memory_l2_warm.source_url),
                    heat=MIN(memory_l2_warm.heat + 0.5, 10.0),
                    last_accessed=excluded.last_accessed
            """, (key, json.dumps(value) if not isinstance(value, str) else value,
                  category, source_url, datetime.now(timezone.utc).isoformat()))
            conn.commit()
        finally:
            conn.close()

    def l2_get(self, key):
        """Retrieve from L2, boosting heat."""
        conn = self._get_conn()
        try:
            row = conn.execute(
                "SELECT value, heat FROM memory_l2_warm WHERE key = ?", (key,)
            ).fetchone()
            if row:
                conn.execute(
                    "UPDATE memory_l2_warm SET heat = heat + ?, access_count = access_count + 1, last_accessed = ? WHERE key = ?",
                    (self.HEAT_BOOST, datetime.now(timezone.utc).isoformat(), key)
                )
                conn.commit()
                try:
                    return json.loads(row['value'])
                except (json.JSONDecodeError, TypeError):
                    return row['value']
        finally:
            conn.close()
        return None

    def l2_search(self, query, category=None, limit=20):
        """Search L2 by text match (simple LIKE). Returns list of (key, value, heat)."""
        conn = self._get_conn()
        try:
            sql = "SELECT key, value, heat FROM memory_l2_warm WHERE value LIKE ?"
            params = [f"%{query}%"]
            if category:
                sql += " AND category = ?"
                params.append(category)
            sql += " ORDER BY heat DESC LIMIT ?"
            params.append(limit)
            rows = conn.execute(sql, params).fetchall()
            results = []
            for r in rows:
                try:
                    val = json.loads(r['value'])
                except (json.JSONDecodeError, TypeError):
                    val = r['value']
                results.append((r['key'], val, r['heat']))
            return results
        finally:
            conn.close()

    def l2_decay(self):
        """Decay all L2 heat values. Demote cold items to L3."""
        conn = self._get_conn()
        try:
            # Decay heat
            conn.execute(
                f"UPDATE memory_l2_warm SET heat = heat * {self.HEAT_DECAY}"
            )
            # Find items to demote
            cold = conn.execute(
                "SELECT key, value, category FROM memory_l2_warm WHERE heat < ?",
                (self.DEMOTE_THRESHOLD,)
            ).fetchall()
            for row in cold:
                self.l3_archive(row['key'], row['value'], row['category'])
                conn.execute("DELETE FROM memory_l2_warm WHERE key = ?", (row['key'],))
            conn.commit()
            if cold:
                logger.info("L2 decay: demoted %d items to L3", len(cold))
            return len(cold)
        finally:
            conn.close()

    def l2_stats(self):
        """Get L2 memory statistics."""
        conn = self._get_conn()
        try:
            total = conn.execute("SELECT COUNT(*) FROM memory_l2_warm").fetchone()[0]
            by_cat = conn.execute(
                "SELECT category, COUNT(*), ROUND(AVG(heat), 2), SUM(access_count) FROM memory_l2_warm GROUP BY category ORDER BY COUNT(*) DESC"
            ).fetchall()
            return {'total': total, 'by_category': [dict(zip(['category', 'count', 'avg_heat', 'total_accesses'], r)) for r in by_cat]}
        finally:
            conn.close()

    # ------------------------------------------------------------------
    # L3: COLD MEMORY (Compressed archive)
    # ------------------------------------------------------------------
    def l3_archive(self, key, value, category='archive'):
        """Archive a value to L3 cold storage."""
        val_str = json.dumps(value) if not isinstance(value, str) else value
        original_size = len(val_str)
        # Simple compression: truncate if too large, otherwise store as-is
        # (In production, use zlib; keeping it readable for SQLite inspection)
        compressed = val_str
        if len(compressed) > 2000:
            compressed = val_str[:2000] + "...[compressed]"

        conn = self._get_conn()
        try:
            conn.execute("""
                INSERT INTO memory_l3_cold (key, compressed_value, original_size, category, archived_at)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(key) DO UPDATE SET
                    compressed_value=excluded.compressed_value,
                    original_size=excluded.original_size,
                    recall_count=memory_l3_cold.recall_count + 1
            """, (key, compressed, original_size, category,
                  datetime.now(timezone.utc).isoformat()))
            conn.commit()
        finally:
            conn.close()

    def l3_recall(self, key):
        """Recall from L3, promoting back to L2 on access."""
        conn = self._get_conn()
        try:
            row = conn.execute(
                "SELECT compressed_value FROM memory_l3_cold WHERE key = ?", (key,)
            ).fetchone()
            if row:
                conn.execute(
                    "UPDATE memory_l3_cold SET recall_count = recall_count + 1 WHERE key = ?",
                    (key,)
                )
                conn.commit()
                try:
                    return json.loads(row['compressed_value'])
                except (json.JSONDecodeError, TypeError):
                    return row['compressed_value']
        finally:
            conn.close()
        return None

    # ------------------------------------------------------------------
    # UNIFIED ACCESS: Try L1 -> L2 -> L3
    # ------------------------------------------------------------------
    def recall(self, key):
        """Recall from any tier: L1 -> L2 -> L3."""
        # Try L1
        val = self.l1_get(key)
        if val is not None:
            return val
        # Try L2
        val = self.l2_get(key)
        if val is not None:
            # Promote hot L2 items to L1
            self.l1_set(key, val)
            return val
        # Try L3
        val = self.l3_recall(key)
        if val is not None:
            # Promote recalled L3 to L2
            self.l2_set(key, val)
            return val
        return None

    def memorize(self, key, value, category='general', source_url=None):
        """Store in L1 (hot). Will flow to L2 on eviction or session end."""
        self.l1_set(key, value, category)
        # Also persist to L2 for durability
        self.l2_set(key, value, category, source_url)

    # ------------------------------------------------------------------
    # SKILLS: Learned extraction patterns
    # ------------------------------------------------------------------
    def record_skill_outcome(self, skill_name, success, pattern=None, metadata=None):
        """Record whether a skill/pattern succeeded or failed."""
        conn = self._get_conn()
        try:
            existing = conn.execute(
                "SELECT success_count, fail_count FROM borg_skills WHERE skill_name = ?",
                (skill_name,)
            ).fetchone()

            if existing:
                if success:
                    conn.execute(
                        "UPDATE borg_skills SET success_count = success_count + 1, last_used = ? WHERE skill_name = ?",
                        (datetime.now(timezone.utc).isoformat(), skill_name))
                else:
                    conn.execute(
                        "UPDATE borg_skills SET fail_count = fail_count + 1, last_used = ? WHERE skill_name = ?",
                        (datetime.now(timezone.utc).isoformat(), skill_name))
            else:
                conn.execute("""
                    INSERT INTO borg_skills (skill_name, pattern, success_count, fail_count, last_used, metadata)
                    VALUES (?, ?, ?, ?, ?, ?)
                """, (skill_name, pattern or '',
                     1 if success else 0, 0 if success else 1,
                     datetime.now(timezone.utc).isoformat(),
                     json.dumps(metadata) if metadata else None))
            conn.commit()
        finally:
            conn.close()

    def get_best_skills(self, skill_type='extraction', limit=10):
        """Get skills ranked by success rate (minimum 3 attempts)."""
        conn = self._get_conn()
        try:
            rows = conn.execute("""
                SELECT skill_name, pattern, success_count, fail_count,
                       ROUND(CAST(success_count AS REAL) / (success_count + fail_count), 3) as win_rate,
                       metadata
                FROM borg_skills
                WHERE (success_count + fail_count) >= 3
                AND (? IS NULL OR skill_type = ?)
                ORDER BY win_rate DESC, success_count DESC
                LIMIT ?
            """, (skill_type, skill_type, limit)).fetchall()

            return [dict(r) for r in rows]
        finally:
            conn.close()

    def evolve_skill(self, skill_name, new_pattern, reason=''):
        """Evolve a skill with a new pattern, keeping lineage."""
        conn = self._get_conn()
        try:
            conn.execute("""
                UPDATE borg_skills SET
                    pattern = ?,
                    evolved_from = skill_name,
                    success_count = 0,
                    fail_count = 0,
                    metadata = json_set(COALESCE(metadata, '{}'), '$.evolved_reason', ?)
                WHERE skill_name = ?
            """, (new_pattern, reason, skill_name))
            conn.commit()
        finally:
            conn.close()
        logger.info("Skill evolved: %s -> %s", skill_name, reason)

    # ------------------------------------------------------------------
    # TOOL REGISTRY: Discovered capabilities
    # ------------------------------------------------------------------
    def register_tool(self, tool_name, tool_type='mcp', schema_json=None,
                      endpoint=None, capability_tags=None, metadata=None):
        """Register a discovered tool/capability."""
        conn = self._get_conn()
        try:
            conn.execute("""
                INSERT INTO borg_tool_registry (tool_name, tool_type, schema_json, endpoint,
                    capability_tags, metadata)
                VALUES (?, ?, ?, ?, ?, ?)
                ON CONFLICT(tool_name) DO UPDATE SET
                    schema_json=COALESCE(excluded.schema_json, borg_tool_registry.schema_json),
                    endpoint=COALESCE(excluded.endpoint, borg_tool_registry.endpoint),
                    capability_tags=COALESCE(excluded.capability_tags, borg_tool_registry.capability_tags),
                    last_seen=CURRENT_TIMESTAMP
            """, (tool_name, tool_type,
                  json.dumps(schema_json) if isinstance(schema_json, dict) else schema_json,
                  endpoint,
                  ','.join(capability_tags) if isinstance(capability_tags, list) else capability_tags,
                  json.dumps(metadata) if metadata else None))
            conn.commit()
        finally:
            conn.close()

    def search_tools(self, query, limit=10):
        """Search registered tools by name or capability tags."""
        conn = self._get_conn()
        try:
            rows = conn.execute("""
                SELECT tool_name, tool_type, schema_json, endpoint, capability_tags, success_rate
                FROM borg_tool_registry
                WHERE tool_name LIKE ? OR capability_tags LIKE ?
                ORDER BY success_rate DESC, call_count DESC
                LIMIT ?
            """, (f"%{query}%", f"%{query}%", limit)).fetchall()
            return [dict(r) for r in rows]
        finally:
            conn.close()

    def record_tool_call(self, tool_name, success):
        """Record a tool call outcome for success rate tracking."""
        conn = self._get_conn()
        try:
            conn.execute("""
                UPDATE borg_tool_registry SET
                    call_count = call_count + 1,
                    success_rate = CASE
                        WHEN call_count = 0 THEN CASE WHEN ? THEN 1.0 ELSE 0.0 END
                        ELSE success_rate * 0.9 + CASE WHEN ? THEN 0.1 ELSE 0.0 END
                    END
                WHERE tool_name = ?
            """, (1 if success else 0, 1 if success else 0, tool_name))
            conn.commit()
        finally:
            conn.close()

    # ------------------------------------------------------------------
    # ANALYTICS
    # ------------------------------------------------------------------
    def get_memory_stats(self):
        """Full memory system statistics."""
        conn = self._get_conn()
        try:
            l2_count = conn.execute("SELECT COUNT(*) FROM memory_l2_warm").fetchone()[0]
            l3_count = conn.execute("SELECT COUNT(*) FROM memory_l3_cold").fetchone()[0]
            skills_count = conn.execute("SELECT COUNT(*) FROM borg_skills").fetchone()[0]
            tools_count = conn.execute("SELECT COUNT(*) FROM borg_tool_registry").fetchone()[0]

            top_skills = conn.execute("""
                SELECT skill_name, success_count, fail_count,
                       ROUND(CAST(success_count AS REAL) / MAX(success_count + fail_count, 1), 3) as win_rate
                FROM borg_skills ORDER BY success_count DESC LIMIT 5
            """).fetchall()

            return {
                'l1_count': len(self.l1_cache),
                'l2_count': l2_count,
                'l3_count': l3_count,
                'skills_count': skills_count,
                'tools_count': tools_count,
                'top_skills': [dict(r) for r in top_skills],
            }
        finally:
            conn.close()
