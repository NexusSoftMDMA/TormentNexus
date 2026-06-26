/**
 * Test Manuale per Neural Memory MCP v2.0 - Minimal API
 * 8 funzioni essenziali
 */

import memoryService from '../src/services/memory.js';

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

async function testDirect() {
    log('\nNeural Memory MCP v2.0 - Test Manuale (API Minimalista)\n', 'cyan');
    log('='.repeat(50), 'blue');

    try {
        log('\nTest 1: add_node', 'yellow');
        const node1 = await memoryService.addNode({
            keywords: ['entity', 'User', 'database'],
            content: 'Ho creato l\'entità User con campi: id, name, email',
            type: 'entity',
            weight: 1.5
        });
        log(`OK Nodo 1 creato: ${node1.node_id}`, 'green');

        const node2 = await memoryService.addNode({
            keywords: ['task', 'API', 'CRUD'],
            content: 'Implementato CRUD completo per l\'entità User',
            type: 'task',
            parentId: node1.node_id,
            weight: 1.2
        });
        log(`OK Nodo 2 creato: ${node2.node_id}`, 'green');

        log('\nTest 2: search_nodes', 'yellow');
        const searchResults = await memoryService.searchNodes({
            keywords: ['User', 'entity'],
            maxResults: 10
        });
        log(`OK Trovati ${searchResults.length} risultati`, 'green');

        log('\nTest 3: register_skill', 'yellow');
        const skill = await memoryService.registerSkill({
            name: 'Express.js API Development',
            framework: 'express',
            language: 'javascript',
            filePattern: 'routes/*.js',
            learnSteps: ['1. Setup Express app', '2. Definisci router', '3. Implementa CRUD'],
            useCases: ['REST API', 'Microservices']
        });
        log(`OK Skill registrata: ${skill.skill_id}`, 'green');

        log('\nTest 4: suggest_skills', 'yellow');
        const suggestions = await memoryService.suggestSkills({
            currentKeywords: ['api', 'rest'],
            maxResults: 3
        });
        log(`OK Suggerite ${suggestions.length} skills`, 'green');

        log('\nTest 5: save_context_snapshot', 'yellow');
        const snapshot = await memoryService.saveContextSnapshot({
            summary: 'Testing minimal API v2.0',
            workDone: { files: ['src/index.js'] },
            pendingTasks: ['Add documentation'],
            learnings: ['Minimal API works']
        });
        log(`OK Snapshot salvato: ${snapshot.snapshot_id}`, 'green');

        log('\nTest 6: restore_context', 'yellow');
        const restored = await memoryService.restoreContext(snapshot.snapshot_id);
        log(`OK Contesto ripristinato: ${restored.success}`, 'green');

        log('\nTest 7: get_memory_report (HTML)', 'yellow');
        const report = await memoryService.getMemoryReport({ format: 'html' });
        log(`OK Report HTML generato (${report.length} chars)`, 'green');

        log('\nTest 8: delete_node', 'yellow');
        const deleted = await memoryService.deleteNode({ nodeId: node1.node_id });
        log(`OK Nodo eliminato: ${deleted.deleted}`, 'green');

        log('\n' + '='.repeat(50), 'blue');
        log('Tutti i test API minimalista completati!', 'green');
        log('='.repeat(50), 'blue');

    } catch (error) {
        log(`\nErrore: ${error.message}`, 'red');
        log(error.stack, 'red');
        process.exit(1);
    }
}

testDirect().then(() => {
    process.exit(0);
}).catch((error) => {
    log(`\nTest falliti: ${error.message}`, 'red');
    process.exit(1);
});
