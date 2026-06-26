/**
 * Database Migrator v2.0
 * Sistema automatico di migration per DB UNIFICATO
 * 
 * CAMBIAMENTI v2.0:
 * - UNICO database per tutta la memoria
 * - Tabelle: sessions, nodes, node_links, nodes_fts
 * - RIMOSSO: projects (non più necessario)
 */

import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Definizione delle migration v2.0
 * Ogni migration ha: version, description, up()
 */
const migrations = [
  {
    version: 1,
    description: 'Initial schema v2.0 - unified memory with sessions',
    async up(sequelize) {
      // Leggi SQL raw per compatibilità
      await sequelize.query(`
        CREATE TABLE IF NOT EXISTS schema_info (
          version INTEGER PRIMARY KEY,
          applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          description TEXT
        )
      `);

      // ===== SESSION TABLE (NUOVO in v2.0) =====
      await sequelize.query(`
        CREATE TABLE IF NOT EXISTS sessions (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          description TEXT DEFAULT '',
          started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          ended_at DATETIME,
          context TEXT DEFAULT '{}',
          project_path TEXT,
          tags TEXT DEFAULT '[]',
          stats TEXT DEFAULT '{"nodesCreated":0,"skillsRegistered":0,"skillsUsed":0,"durationMinutes":0}',
          is_active INTEGER DEFAULT 1,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
      `);

      // ===== NODE TABLE v2.0 =====
      // CAMBIAMENTO: RIMOSSO project_id, AGGIUNTO session_id
      await sequelize.query(`
        CREATE TABLE IF NOT EXISTS nodes (
          id TEXT PRIMARY KEY,
          session_id TEXT,
          parent_id TEXT,
          type TEXT DEFAULT 'generic',
          keywords TEXT DEFAULT '[]',
          content TEXT,
          metadata TEXT DEFAULT '{}',
          depth INTEGER DEFAULT 0,
          weight REAL DEFAULT 1.0,
          keyword_count INTEGER DEFAULT 0,
          task_date DATE,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
      `);

      // ===== NODE LINKS TABLE =====
      await sequelize.query(`
        CREATE TABLE IF NOT EXISTS node_links (
          id TEXT PRIMARY KEY,
          from_node_id TEXT NOT NULL,
          to_node_id TEXT NOT NULL,
          link_type TEXT DEFAULT 'related',
          weight REAL DEFAULT 1.0,
          metadata TEXT DEFAULT '{}',
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
      `);

      // ===== INDICES FOR SESSION =====
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(is_active)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions(started_at)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_path)`);

      // ===== INDICES FOR NODE =====
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_nodes_session ON nodes(session_id)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_nodes_parent ON nodes(parent_id)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_nodes_created ON nodes(created_at)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_nodes_session_type ON nodes(session_id, type)`);

      // ===== INDICES FOR LINK =====
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_links_from ON node_links(from_node_id)`);
      await sequelize.query(`CREATE INDEX IF NOT EXISTS idx_links_to ON node_links(to_node_id)`);

      // Inserisci versione iniziale
      await sequelize.query(`
        INSERT OR IGNORE INTO schema_info (version, description) VALUES (1, 'Initial schema v2.0 - unified memory with sessions')
      `);
    }
  },
  {
    version: 2,
    description: 'Add FTS5 full-text search support',
    async up(sequelize) {
      // Crea tabella FTS5 se non esiste (SQLite virtual table)
      await sequelize.query(`
        CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
          node_id UNINDEXED,
          keywords,
          content,
          content_rowid='rowid'
        )
      `);

      // Inserisci versione
      await sequelize.query(`
        INSERT OR IGNORE INTO schema_info (version, description) VALUES (2, 'Add FTS5 full-text search support')
      `);
    }
  }
];

/**
 * Ottiene la versione corrente dello schema
 * @param {Sequelize} sequelize
 * @returns {Promise<number>}
 */
async function getCurrentVersion(sequelize) {
  try {
    // Verifica se la tabella schema_info esiste
    const [tables] = await sequelize.query(`
      SELECT name FROM sqlite_master WHERE type='table' AND name='schema_info'
    `);

    if (!tables || tables.length === 0) {
      return 0; // Nessuna migration applicata
    }

    // Ottieni la versione massima
    const [result] = await sequelize.query(`
      SELECT MAX(version) as version FROM schema_info
    `);

    return result[0]?.version || 0;
  } catch (error) {
    // Se c'è errore (tabella non esiste), restituisci 0
    return 0;
  }
}

/**
 * Esegue le migration pendenti
 * @param {Sequelize} sequelize
 * @returns {Promise<{applied: number, total: number}>}
 */
async function runMigrations(sequelize) {
  const currentVersion = await getCurrentVersion(sequelize);
  const pendingMigrations = migrations.filter(m => m.version > currentVersion);

  if (pendingMigrations.length === 0) {
    return { applied: 0, total: 0 };
  }

  console.log(`[Migrator v2.0] Found ${pendingMigrations.length} pending migrations`);

  for (const migration of pendingMigrations) {
    try {
      console.log(`[Migrator v2.0] Applying migration v${migration.version}: ${migration.description}`);
      await migration.up(sequelize);
      console.log(`[Migrator v2.0] ✓ Migration v${migration.version} applied successfully`);
    } catch (error) {
      console.error(`[Migrator v2.0] ✗ Migration v${migration.version} failed:`, error.message);
      // Non fermare le altre migration, ma logga l'errore
    }
  }

  return { 
    applied: pendingMigrations.length, 
    total: migrations.length 
  };
}

/**
 * Verifica se una tabella esiste
 * @param {Sequelize} sequelize
 * @param {string} tableName
 * @returns {Promise<boolean>}
 */
async function tableExists(sequelize, tableName) {
  const [tables] = await sequelize.query(`
    SELECT name FROM sqlite_master WHERE type='table' AND name=?
  `, {
    replacements: [tableName]
  });
  return tables && tables.length > 0;
}

/**
 * Sincronizza schema senza cancellare dati (usato come fallback)
 * Crea tabelle mancanti ma non rimuove dati esistenti
 * @param {Sequelize} sequelize
 * @returns {Promise<void>}
 */
async function syncSchema(sequelize) {
  // Verifica e crea tabella sessions se non esiste
  if (!(await tableExists(sequelize, 'sessions'))) {
    await sequelize.query(`
      CREATE TABLE IF NOT EXISTS sessions (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT DEFAULT '',
        started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        ended_at DATETIME,
        context TEXT DEFAULT '{}',
        project_path TEXT,
        tags TEXT DEFAULT '[]',
        stats TEXT DEFAULT '{}',
        is_active INTEGER DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);
  }

  // Verifica e crea tabella nodes se non esiste (v2.0 senza project_id)
  if (!(await tableExists(sequelize, 'nodes'))) {
    await sequelize.query(`
      CREATE TABLE IF NOT EXISTS nodes (
        id TEXT PRIMARY KEY,
        session_id TEXT,
        parent_id TEXT,
        type TEXT DEFAULT 'generic',
        keywords TEXT DEFAULT '[]',
        content TEXT,
        metadata TEXT DEFAULT '{}',
        depth INTEGER DEFAULT 0,
        weight REAL DEFAULT 1.0,
        keyword_count INTEGER DEFAULT 0,
        task_date DATE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);
  }

  // Verifica e crea tabella node_links se non esiste
  if (!(await tableExists(sequelize, 'node_links'))) {
    await sequelize.query(`
      CREATE TABLE IF NOT EXISTS node_links (
        id TEXT PRIMARY KEY,
        from_node_id TEXT NOT NULL,
        to_node_id TEXT NOT NULL,
        link_type TEXT DEFAULT 'related',
        weight REAL DEFAULT 1.0,
        metadata TEXT DEFAULT '{}',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);
  }

  // Verifica e crea FTS5 se non esiste
  if (!(await tableExists(sequelize, 'nodes_fts'))) {
    await sequelize.query(`
      CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
        node_id UNINDEXED,
        keywords,
        content,
        content_rowid='rowid'
      )
    `);
  }
}

export {
  runMigrations,
  getCurrentVersion,
  syncSchema,
  tableExists,
  migrations
};
