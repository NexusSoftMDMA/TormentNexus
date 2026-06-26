import { db, schema, sqliteInstance } from '../../../db/index.js';

class DatabaseService {
  constructor() {}

  getDb(): import('better-sqlite3').Database {
    // In TormentNexus, we use Drizzle 'db' for ORM and 'sqliteInstance' for raw sqlite.
    return sqliteInstance;
  }

  getDrizzle() {
    return db;
  }

  getSchema() {
    return schema;
  }

  close() {
    // Managed by central DB index
  }
}

export const dbService = new DatabaseService();
