import { dbService } from './db.js';
import { wsManager } from './ws-manager.js';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../../../db/sqliteAvailability.js';

export interface Fact {
  id: string;
  key: string;
  value: string;
  sourceSession: string;
  confidence: number;
  timestamp: number;
  tags: string[];
}

class CollectiveMemoryService {
  private tableInitialized = false;

  constructor() {}

  private initializeTable(): void {
    if (this.tableInitialized) {
      return;
    }

    const db = dbService.getDb();
    db.exec(`
      CREATE TABLE IF NOT EXISTS facts (
        id TEXT PRIMARY KEY,
        key TEXT,
        value TEXT,
        sourceSession TEXT,
        confidence REAL,
        timestamp INTEGER,
        tags TEXT
      )
    `);
    db.exec('CREATE INDEX IF NOT EXISTS idx_facts_key ON facts(key);');
    this.tableInitialized = true;
  }

  private getDb(): import('better-sqlite3').Database {
    try {
      this.initializeTable();
      return dbService.getDb();
    } catch (error) {
      if (isSqliteUnavailableError(error)) {
        throw new Error(formatOptionalSqliteFailure('Collective memory is unavailable', error));
      }

      throw error;
    }
  }

  async storeFact(fact: Omit<Fact, 'id' | 'timestamp'>): Promise<Fact> {
    const db = this.getDb();
    const id = `fact-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const timestamp = Date.now();
    
    const newFact: Fact = { ...fact, id, timestamp };

    db.prepare(`
      INSERT INTO facts (id, key, value, sourceSession, confidence, timestamp, tags)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `).run(
      id,
      fact.key,
      fact.value,
      fact.sourceSession,
      fact.confidence,
      timestamp,
      JSON.stringify(fact.tags)
    );

    wsManager.broadcast({
      type: 'log',
      payload: { 
        level: 'info', 
        message: `[CollectiveMemory] New Fact Learned: ${fact.key} = ${fact.value.slice(0, 30)}...`,
        timestamp,
        source: 'tormentnexus-memory'
      },
      timestamp
    });

    return newFact;
  }

  async recallFact(key: string): Promise<Fact[]> {
    const db = this.getDb();
    const rows = db.prepare('SELECT * FROM facts WHERE key = ? ORDER BY confidence DESC').all(key);
    
    return rows.map((r: any) => ({
      ...r,
      tags: JSON.parse(r.tags)
    }));
  }

  async searchFacts(query: string): Promise<Fact[]> {
    const db = this.getDb();
    const rows = db.prepare('SELECT * FROM facts WHERE value LIKE ? OR key LIKE ?').all(`%${query}%`, `%${query}%`);
    
    return rows.map((r: any) => ({
      ...r,
      tags: JSON.parse(r.tags)
    }));
  }

  async getAllFacts(): Promise<Fact[]> {
    const db = this.getDb();
    const rows = db.prepare('SELECT * FROM facts ORDER BY timestamp DESC').all();
    return rows.map((r: any) => ({
      ...r,
      tags: JSON.parse(r.tags)
    }));
  }
}

export const collectiveMemory = new CollectiveMemoryService();
