/**
 * Database Connection Manager v2.0
 * Gestisce connessione Sequelize per il DB UNIFICATO
 * 
 * CAMBIAMENTI v2.0:
 * - UNICO database per tutta la memoria
 * - Rimosso: gestione DB multipli per progetto
 */

import { Sequelize } from 'sequelize';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

const connections = new Map();

// Directory per il database unificato
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const DATA_DIR = path.join(__dirname, '../../data');
const UNIFIED_DB_NAME = 'neural-memory-unified.sqlite';

/**
 * Inizializza la directory dei dati se non esiste
 */
function initDataDir() {
  if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
  }
}

/**
 * Ottiene o crea una connessione Sequelize per il DB unificato
 * @returns {Sequelize} - Istanza Sequelize per il DB unificato
 */
function getConnection() {
  const dbKey = 'unified';
  
  // Se esiste già la connessione, restituiscila
  if (connections.has(dbKey)) {
    return connections.get(dbKey);
  }

  initDataDir();

  // Percorso DB unificato
  const databasePath = path.join(DATA_DIR, UNIFIED_DB_NAME);
  
  const sequelize = new Sequelize({
    dialect: 'sqlite',
    storage: databasePath,
    logging: false, // Imposta true per debug
    define: {
      timestamps: true,
      underscored: true
    }
  });

  connections.set(dbKey, sequelize);
  return sequelize;
}
/**
 * Chiude la connessione
 */
async function closeConnection() {
  const dbKey = 'unified';
  if (connections.has(dbKey)) {
    const sequelize = connections.get(dbKey);
    await sequelize.close();
    connections.delete(dbKey);
  }
}




export {
  getConnection,
  closeConnection,
  initDataDir,
  DATA_DIR,
  UNIFIED_DB_NAME
};
