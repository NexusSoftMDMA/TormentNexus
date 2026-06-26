/**
 * Test Automatico per Sistema Migration v2.0
 * Verifica che il sistema migration funzioni correttamente con schema unificato
 */

import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { getConnection, initDataDir, closeConnection } from '../src/database/connection.js';
import { runMigrations, getCurrentVersion, syncSchema } from '../src/database/migrator.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Colori console
const colors = {
    reset: '\x1b[0m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    red: '\x1b[31m',
    cyan: '\x1b[36m'
};

function log(message, color = 'reset') {
    console.log(`${colors[color]}${message}${colors.reset}`);
}

function assert(condition, message) {
    if (!condition) {
        throw new Error(`❌ Assertion failed: ${message}`);
    }
}

async function cleanTestDb() {
    const dataDir = path.join(__dirname, '../data');
    if (!fs.existsSync(dataDir)) {
        fs.mkdirSync(dataDir, { recursive: true });
    }
    
    // Rimuovi DB di test se esiste
    const testDbPath = path.join(dataDir, 'migration-test-v2.sqlite');
    if (fs.existsSync(testDbPath)) {
        fs.unlinkSync(testDbPath);
    }
    return testDbPath;
}

async function runTests() {
    log('\n🧪 Neural Memory v2.0 - Test Sistema Migration\n', 'cyan');
    log('═'.repeat(50), 'blue');

    let passedTests = 0;
    let failedTests = 0;

    try {
        // Clean up test DB
        const testDbPath = await cleanTestDb();
        log('\n🧹 Database di test pulito', 'yellow');

        // ============ TEST 1: Prima migration - DB vuoto ============
        log('\n📦 Test 1: Prima migration su DB vuoto', 'yellow');
        
        // Crea una connessione di test temporanea
        const { Sequelize } = await import('sequelize');
        const sequelize1 = new Sequelize({
            dialect: 'sqlite',
            storage: testDbPath,
            logging: false,
            define: {
                timestamps: true,
                underscored: true
            }
        });
        
        // Verifica che non ci siano tabelle
        const [tables1] = await sequelize1.query(`
            SELECT name FROM sqlite_master WHERE type='table'
        `);
        assert(tables1.length === 0, 'DB dovrebbe essere vuoto inizialmente');
        
        // Esegui migration
        const result1 = await runMigrations(sequelize1);
        log(`   Migration applicate: ${result1.applied}, totali: ${result1.total}`, 'reset');
        
        // Verifica che le tabelle siano state create
        const [tables2] = await sequelize1.query(`
            SELECT name FROM sqlite_master WHERE type='table'
        `);
        assert(tables2.length >= 4, `Dovrebbero esserci almeno 4 tabelle (sessions, nodes, node_links, schema_info), trovate: ${tables2.length}`);
        
        const tableNames = tables2.map(t => t.name);
        assert(tableNames.includes('sessions'), 'Tabella sessions dovrebbe esistere (v2.0)');
        assert(tableNames.includes('nodes'), 'Tabella nodes dovrebbe esistere');
        assert(tableNames.includes('node_links'), 'Tabella node_links dovrebbe esistere');
        assert(tableNames.includes('schema_info'), 'Tabella schema_info dovrebbe esistere');
        
        // Verifica versione
        const version1 = await getCurrentVersion(sequelize1);
        assert(version1 === 2, `Versione dovrebbe essere 2, trovata: ${version1}`);
        
        log('   ✓ Prima migration completata con successo', 'green');
        passedTests++;

        // ============ TEST 2: Seconda run - nessuna migration pendente ============
        log('\n📦 Test 2: Run ripetuto - nessuna migration pendente', 'yellow');
        
        const result2 = await runMigrations(sequelize1);
        assert(result2.applied === 0, 'Non dovrebbero esserci migration pendenti');
        
        log('   ✓ Nessuna migration pendente confermata', 'green');
        passedTests++;

        // ============ TEST 3: Verifica contenuto schema_info ============
        log('\n📦 Test 3: Verifica contenuto schema_info', 'yellow');
        
        const [versions] = await sequelize1.query(`
            SELECT version, description FROM schema_info ORDER BY version
        `);
        assert(versions.length === 2, `Dovrebbero esserci 2 versioni registrate, trovate: ${versions.length}`);
        assert(versions[0].version === 1, 'Prima versione dovrebbe essere 1');
        assert(versions[1].version === 2, 'Seconda versione dovrebbe essere 2');
        
        log(`   ✓ Versioni registrate: ${versions.map(v => v.version).join(', ')}`, 'green');
        passedTests++;

        // ============ TEST 4: Verifica che i dati non vengano cancellati ============
        log('\n📦 Test 4: Verifica persistenza dati tra migration', 'yellow');
        
        // Inserisci una sessione di test (v2.0 - senza project_id)
        const testSessionId = 'test-session-123';
        await sequelize1.query(`
            INSERT INTO sessions (id, name, description, tags, stats, is_active)
            VALUES (?, ?, ?, ?, ?, ?)
        `, {
            replacements: [testSessionId, 'Test Session', 'Test description', '["test"]', '{"nodesCreated":0}', 1]
        });
        
        // Inserisci un nodo di test (v2.0 - con session_id invece di project_id)
        const testNodeId = 'test-node-123';
        await sequelize1.query(`
            INSERT INTO nodes (id, session_id, type, keywords, content, depth, weight)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        `, {
            replacements: [testNodeId, testSessionId, 'task', '["test", "v2"]', 'Test content v2.0', 0, 1.0]
        });
        
        // Verifica che il nodo esista
        const [nodeCheck] = await sequelize1.query(`SELECT id, session_id FROM nodes WHERE id = ?`, {
            replacements: [testNodeId]
        });
        assert(nodeCheck.length === 1, 'Nodo di test dovrebbe esistere');
        assert(nodeCheck[0].session_id === testSessionId, 'Nodo dovrebbe avere session_id corretto');
        
        // Riesegui migration
        await runMigrations(sequelize1);
        
        // Verifica che il nodo esista ancora
        const [nodeCheck2] = await sequelize1.query(`SELECT id, session_id FROM nodes WHERE id = ?`, {
            replacements: [testNodeId]
        });
        assert(nodeCheck2.length === 1, 'Nodo di test dovrebbe esistere ancora dopo migration');
        
        log('   ✓ Dati preservati tra migration (schema v2.0)', 'green');
        passedTests++;

        // ============ TEST 5: syncSchema fallback ============
        log('\n📦 Test 5: Test syncSchema fallback', 'yellow');
        
        await syncSchema(sequelize1); // Non dovrebbe lanciare errori
        
        log('   ✓ syncSchema completato senza errori', 'green');
        passedTests++;

        // ============ TEST 6: Verifica schema sessions v2.0 ============
        log('\n📦 Test 6: Verifica schema sessions (nuovo in v2.0)', 'yellow');
        
        // Verifica che la sessione abbia i campi corretti
        const [sessionCheck] = await sequelize1.query(`SELECT id, name, is_active, stats FROM sessions WHERE id = ?`, {
            replacements: [testSessionId]
        });
        assert(sessionCheck.length === 1, 'Sessione dovrebbe esistere');
        assert(sessionCheck[0].is_active === 1, 'Sessione dovrebbe essere attiva');
        
        log('   ✓ Schema sessions v2.0 funziona correttamente', 'green');
        passedTests++;

        // ============ TEST 7: Verifica FTS5 ============
        log('\n📦 Test 7: Verifica tabella FTS5', 'yellow');
        
        const [ftsCheck] = await sequelize1.query(`
            SELECT name FROM sqlite_master WHERE type='table' AND name='nodes_fts'
        `);
        assert(ftsCheck.length === 1, 'Tabella FTS5 dovrebbe esistere');
        
        log('   ✓ FTS5 funziona correttamente', 'green');
        passedTests++;

        // ============ TEST 8: Verifica colonne node v2.0 ============
        log('\n📦 Test 8: Verifica schema nodes v2.0 (senza project_id)', 'yellow');
        
        // Verifica che nodes abbia session_id e NON project_id
        const [nodeSchema] = await sequelize1.query(`PRAGMA table_info(nodes)`);
        const nodeColumns = nodeSchema.map(c => c.name);
        
        assert(nodeColumns.includes('session_id'), 'nodes dovrebbe avere session_id');
        assert(!nodeColumns.includes('project_id'), 'nodes NON dovrebbe avere project_id (rimosso in v2.0)');
        
        log('   ✓ Schema nodes v2.0 corretto (session_id presente, project_id rimosso)', 'green');
        passedTests++;

        // ============ PULIZIA ============
        log('\n🧹 Pulizia database di test...', 'yellow');
        
        // Chiudi connessione
        await sequelize1.close();
        
        // Rimuovi file di test
        if (fs.existsSync(testDbPath)) {
            fs.unlinkSync(testDbPath);
        }

    } catch (error) {
        log(`\n❌ Test fallito: ${error.message}`, 'red');
        log(error.stack, 'red');
        failedTests++;
    }

    // ============ RIEPILOGO ============
    log('\n' + '═'.repeat(50), 'blue');
    log(`✅ Test passati: ${passedTests}`, 'green');
    log(`❌ Test falliti: ${failedTests}`, failedTests > 0 ? 'red' : 'green');
    log('═'.repeat(50), 'blue');

    if (failedTests > 0) {
        process.exit(1);
    }
}

// Esegui test
runTests().then(() => {
    log('\n🎉 Tutti i test sono passati! Il sistema migration v2.0 funziona correttamente nya~\n', 'cyan');
    process.exit(0);
}).catch((error) => {
    log(`\n💥 Errore durante i test: ${error.message}`, 'red');
    process.exit(1);
});
