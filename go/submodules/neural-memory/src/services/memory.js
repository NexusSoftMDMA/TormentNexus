/**
 * Memory Service v2.0 - Minimal API
 * Logica di business per la gestione della memoria neurale unificata
 *
 * 8 funzioni essenziali: addNode, searchNodes, deleteNode, registerSkill, suggestSkills, saveContextSnapshot, restoreContext, getMemoryReport
 */

import { getConnection, initDataDir } from '../database/connection.js';
import { initModels } from '../database/models/index.js';
import { runMigrations } from '../database/migrator.js';

class MemoryService {
    constructor() {
        // Cache del modello (unico DB)
        this.modelsCache = null;
        this.initialized = false;
    }

    /**
     * Helper: парсит keywords (могут быть array или JSON string)
     */
    parseKeywords(keywords) {
        if (!keywords) return [];
        if (Array.isArray(keywords)) return keywords;
        try {
            return JSON.parse(keywords);
        } catch {
            return [];
        }
    }

    /**
     * Inizializza il servizio (una volta sola)
     */
    async ensureInitialized() {
        if (!this.initialized) {
            initDataDir();
            const sequelize = getConnection();
            this.modelsCache = initModels(sequelize);
            await runMigrations(sequelize);
            this.initialized = true;
        }
    }

    /**
     * Ottiene i modelli (sempre dalla stessa cache)
     * @returns {Object} - { Session, Node, Link }
     */
    async getModels() {
        await this.ensureInitialized();
        return this.modelsCache;
    }

    // ==================== NODE MANAGEMENT ====================

    /**
     * Aggiunge un nodo alla memoria
     * @param {Object} params
     * @returns {Object} - { node_id, confidence_score }
     */
    async addNode({
        sessionId = null,
        keywords = [],
        content = '',
        type = 'generic',
        parentId = null,
        metadata = {},
        weight = 1.0
    }) {
        const { Node } = await this.getModels();

        // Calcola profondità
        let depth = 0;
        if (parentId) {
            const parent = await Node.findByPk(parentId);
            if (parent) {
                depth = parent.depth + 1;
            }
        }

        // Crea il nodo
        const node = await Node.create({
            sessionId,
            parentId,
            type,
            keywords,
            content,
            metadata,
            depth,
            weight,
            keywordCount: keywords.length
        });

        // Aggiorna FTS5
        await this.updateFTS5(node);

        return {
            node_id: node.id,
            type: node.type,
            depth: node.depth,
            session_id: node.sessionId,
            keywords_count: keywords.length,
            metadata: node.metadata,
            success: true
        };
    }

    /**
     * Aggiorna l'indice FTS5 per un nodo
     * @param {Object} node
     */
    async updateFTS5(node) {
        const sequelize = getConnection();

        await sequelize.query(`
            INSERT INTO nodes_fts(node_id, keywords, content)
            VALUES (?, ?, ?)
        `, {
            replacements: [
                node.id,
                JSON.stringify(node.keywords),
                node.content || ''
            ]
        });
    }

    /**
     * Cerca nodi per keywords
     * @param {Object} params
     * @returns {Array} - Risultati con confidence score
     */
    async searchNodes({
        keywords = [],
        maxResults = 10,
        minConfidence = 0.1,
        type = null,
        sessionId = null,
        tags = []
    }) {
        const { Node } = await this.getModels();

        if (keywords.length === 0) {
            return [];
        }

        const searchQuery = keywords.join(' OR ');

        let ftsQuery = [];
        try {
            const results = await getConnection().query(`
                SELECT node_id, bm25(nodes_fts) as bm25_score
                FROM nodes_fts
                WHERE nodes_fts MATCH ?
                ORDER BY bm25(nodes_fts)
                LIMIT ?
            `, {
                replacements: [searchQuery, maxResults * 2]
            });
            ftsQuery = results[0] || [];
        } catch (error) {
            ftsQuery = [];
        }

        const ftsMap = new Map();
        for (const row of ftsQuery) {
            ftsMap.set(row.node_id, row.bm25_score);
        }

        const whereClause = {};
        if (type) {
            whereClause.type = type;
        }
        if (sessionId) {
            whereClause.sessionId = sessionId;
        }

        const nodes = await Node.findAll({
            where: whereClause,
            order: [['created_at', 'DESC']],
            limit: maxResults * 3
        });

        const scoredNodes = nodes.map(node => {
            const confidence = this.calculateConfidence(node, keywords, ftsMap);
            const nodeKeywords = this.parseKeywords(node.keywords);
            return {
                node_id: node.id,
                type: node.type,
                keywords: nodeKeywords,
                content: node.content,
                depth: node.depth,
                session_id: node.sessionId,
                parent_id: node.parentId,
                confidence: Math.round(confidence * 100) / 100,
                created_at: node.created_at
            };
        });

        return scoredNodes
            .filter(n => n.confidence >= minConfidence)
            .sort((a, b) => b.confidence - a.confidence)
            .slice(0, maxResults);
    }

    /**
     * Calcola score di confidence per un nodo
     * @param {Object} node
     * @param {Array} searchKeywords
     * @param {Map} ftsMap
     * @returns {number} - Confidence 0.0 - 1.0
     */
    calculateConfidence(node, searchKeywords, ftsMap) {
        let score = 0;

        // 1. BM25 score normalizzato (0-0.35)
        const bm25 = ftsMap.get(node.id) || 0;
        const normalizedBm25 = Math.max(0, Math.min(0.35, -bm25 / 10));

        // 2. Keyword match (0-0.25)
        let keywords = node.keywords;
        if (typeof keywords === 'string') {
            try {
                keywords = JSON.parse(keywords);
            } catch (e) {
                keywords = [];
            }
        }
        if (!Array.isArray(keywords)) {
            keywords = [];
        }
        const matchedKeywords = keywords.filter(k =>
            searchKeywords.some(sk =>
                k.toLowerCase().includes(sk.toLowerCase()) ||
                sk.toLowerCase().includes(k.toLowerCase())
            )
        ).length;
        const keywordScore = searchKeywords.length > 0
            ? (matchedKeywords / searchKeywords.length) * 0.25
            : 0;

        // 3. Recency bonus (0-0.10)
        const ageInDays = (Date.now() - new Date(node.created_at)) / (1000 * 60 * 60 * 24);
        const recencyScore = ageInDays < 1 ? 0.10 :
            ageInDays < 7 ? 0.08 :
                ageInDays < 30 ? 0.05 : 0;

        // 4. Type bonus rafforzato per categorie semantiche (0-0.15)
        const typeScores = {
            'error': 0.15,
            'skill': 0.15,        // Skills sono importanti!
            'operation': 0.14,
            'convention': 0.13,
            'edge_case': 0.12,
            'pattern': 0.11,
            'task': 0.10,
            'action': 0.09,
            'entity': 0.07,
            'file': 0.06,
            'concept': 0.06,
            'summary': 0.05,
            'generic': 0.04
        };
        const typeScore = typeScores[node.type] || 0.05;

        // 5. Weight manuale (0-0.15)
        const weightScore = Math.min(0.15, (node.weight || 1.0) * 0.03);

        return Math.min(1.0, normalizedBm25 + keywordScore + recencyScore + typeScore + weightScore);
    }

    /**
     * Elimina un nodo
     * @param {Object} params
     */
    async deleteNode({ nodeId, cascade = false }) {
        const { Node, Link } = await this.getModels();

        if (cascade) {
            await this.deleteChildren(nodeId);
        } else {
            const node = await Node.findByPk(nodeId);
            if (node) {
                await Node.update(
                    { parentId: node.parentId },
                    { where: { parentId: nodeId } }
                );
            }
        }

        await Link.destroy({ where: { fromNodeId: nodeId } });
        await Link.destroy({ where: { toNodeId: nodeId } });
        await Node.destroy({ where: { id: nodeId } });

        return { deleted: true, nodeId };
    }

    async deleteChildren(parentId) {
        const { Node } = await this.getModels();
        const children = await Node.findAll({ where: { parentId } });

        for (const child of children) {
            await this.deleteChildren(child.id);
            await child.destroy();
        }
    }

    // ==================== SKILLS MANAGEMENT ====================

    /**
     * Registra una skill con schema rigido
     * @param {Object} params
     * @returns {Object}
     */
    async registerSkill({
        name,
        framework,
        language,
        filePattern,
        learnSteps = [],
        useCases = [],
        implementation = '',
        examples = [],
        prerequisites = [],
        keywords = [],
        content = ''
    }) {
        const { Node } = await this.getModels();

        // Costruisci il contenuto strutturato
        const structuredContent = content || `SKILL: ${name}\n\n` +
            `FRAMEWORK: ${framework}\n` +
            `LANGUAGE: ${language}\n` +
            `FILE PATTERN: ${filePattern}\n\n` +
            `LEARN STEPS:\n${learnSteps.map((s, i) => `${i + 1}. ${s}`).join('\n')}\n\n` +
            `USE CASES:\n${useCases.map(u => `- ${u}`).join('\n')}\n\n` +
            (implementation ? `IMPLEMENTATION:\n${implementation}` : '');

        // Costruisci keywords combinate
        const allKeywords = [
            name.toLowerCase(),
            framework.toLowerCase(),
            language.toLowerCase(),
            ...filePattern.toLowerCase().split(/[\/\-\.]/).filter(Boolean),
            ...keywords
        ];

        // Metadati strutturati per la skill
        const metadata = {
            framework,
            language,
            filePattern,
            learnSteps,
            useCases,
            implementation,
            examples,
            prerequisites,
            confidence: 0.5,
            usageCount: 0
        };

        const node = await Node.create({
            sessionId: null,
            type: 'skill',
            keywords: allKeywords,
            content: structuredContent,
            metadata,
            weight: 1.5,
            keywordCount: allKeywords.length
        });

        await this.updateFTS5(node);

        return {
            skill_id: node.id,
            name,
            framework,
            language,
            file_pattern: filePattern,
            learn_steps: learnSteps,
            use_cases: useCases,
            confidence: 0.5,
            success: true
        };
    }

    /**
     * Suggerisci skills basate su contesto
     * @param {Object} params
     * @returns {Array}
     */
    async suggestSkills({ currentKeywords = [], maxResults = 5, domain = null }) {
        const results = await this.searchNodes({
            keywords: currentKeywords,
            maxResults: maxResults * 2,
            minConfidence: 0.2,
            type: 'skill'
        });

        let skills = results;
        if (domain) {
            skills = results.filter(r => {
                // Note: searchNodes returns content which contains language info
                const content = r.content || '';
                const langLower = content.toLowerCase();
                return langLower.includes(`language: ${domain.toLowerCase()}`) ||
                       langLower.includes(`framework: ${domain.toLowerCase()}`);
            });
        }

        return skills.slice(0, maxResults).map(r => ({
            skill_id: r.node_id,
            type: r.type,
            keywords: r.keywords,
            framework: r.metadata?.framework,
            language: r.metadata?.language,
            confidence: r.confidence,
            reason: `Condivide ${r.keywords.filter(k =>
                currentKeywords.some(ck => ck.toLowerCase() === k.toLowerCase())
            ).length} keyword(s)`
        }));
    }

    // ==================== CONTEXT MANAGEMENT ====================

    /**
     * Salva snapshot contesto dettagliato
     * @param {Object} params
     * @returns {Object}
     */
    async saveContextSnapshot({
        summary = '',
        workDone = {},
        pendingTasks = [],
        keyDecisions = [],
        blockers = [],
        learnings = [],
        nextSteps = []
    }) {
        const { Node } = await this.getModels();

        const content = JSON.stringify({
            summary,
            workDone,
            pendingTasks,
            keyDecisions,
            blockers,
            learnings,
            nextSteps,
            savedAt: new Date().toISOString()
        }, null, 2);

        const node = await Node.create({
            sessionId: null,
            type: 'context_snapshot',
            keywords: ['context-snapshot', 'summary'],
            content,
            metadata: {
                summary,
                pendingTasksCount: pendingTasks.length,
                blockersCount: blockers.length,
                learningsCount: learnings.length
            },
            weight: 2.0
        });

        await this.updateFTS5(node);

        return {
            snapshot_id: node.id,
            session_id: null,
            summary,
            pending_tasks_count: pendingTasks.length,
            success: true
        };
    }

    /**
     * Recupera contesto compresso
     * @param {string} snapshotId
     * @returns {Object}
     */
    async restoreContext(snapshotId) {
        const { Node } = await this.getModels();

        const node = await Node.findByPk(snapshotId);
        if (!node) {
            throw new Error('Snapshot not found');
        }

        try {
            const context = JSON.parse(node.content);
            return {
                snapshot_id: node.id,
                session_id: node.sessionId,
                ...context,
                success: true
            };
        } catch (e) {
            return {
                snapshot_id: node.id,
                session_id: node.sessionId,
                raw_content: node.content,
                success: false,
                error: 'Failed to parse context'
            };
        }
    }

    // ==================== REPORTS ====================

    /**
     * Genera report memoria
     * @param {Object} params
     * @returns {Object}
     */
    async getMemoryReport({
        format = 'json',
        keywords = [],
        includeStats = true,
        includeRecentWork = true,
        includeTopSkills = true
    }) {
        const { Node } = await this.getModels();

        const report = {};

        if (includeStats) {
            const totalNodes = await Node.count();

            const byType = await Node.findAll({
                attributes: ['type'],
                group: ['type']
            });

            const typeStats = {};
            for (const t of byType) {
                typeStats[t.type] = await Node.count({ where: { type: t.type } });
            }

            report.stats = {
                total_nodes: totalNodes,
                by_type: typeStats
            };
        }

        if (includeTopSkills) {
            const skills = await this.searchNodes({
                keywords: keywords.length > 0 ? keywords : ['skill'],
                maxResults: 10,
                minConfidence: 0.1,
                type: 'skill'
            });
            report.top_skills = skills.map(s => ({
                type: s.type,
                keywords: s.keywords,
                confidence: s.confidence,
                node_id: s.node_id
            }));
        }

        if (includeRecentWork) {
            const recentNodes = await Node.findAll({
                order: [['created_at', 'DESC']],
                limit: 20
            });
            report.recent_work = recentNodes.map(n => ({
                id: n.id,
                type: n.type,
                keywords: this.parseKeywords(n.keywords),
                session_id: n.sessionId,
                created_at: n.created_at
            }));
        }

        if (keywords.length > 0) {
            const searchResults = await this.searchNodes({
                keywords,
                maxResults: 50
            });
            report.search_results = searchResults;
        }

        if (format === 'html') {
            return this.generateHtmlReport(report);
        }

        return {
            report,
            generated_at: new Date().toISOString(),
            format
        };
    }

    /**
     * Genera report HTML
     * @param {Object} report
     * @returns {string}
     */
    generateHtmlReport(report) {
        return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Neural Memory Report</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { color: #333; border-bottom: 2px solid #6366f1; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
        .stat-card { background: #f8f9fa; padding: 20px; border-radius: 8px; border-left: 4px solid #6366f1; }
        .stat-value { font-size: 2em; font-weight: bold; color: #6366f1; }
        .stat-label { color: #666; margin-top: 5px; }
        .node-list { list-style: none; padding: 0; }
        .node-item { background: #fff; border: 1px solid #e5e7eb; padding: 15px; margin: 10px 0; border-radius: 6px; }
        .node-type { display: inline-block; padding: 3px 8px; border-radius: 4px; font-size: 0.8em; font-weight: bold; }
        .type-skill { background: #dbeafe; color: #1e40af; }
        .type-task { background: #dcfce7; color: #166534; }
        .type-error { background: #fee2e2; color: #991b1b; }
        .type-generic { background: #f3f4f6; color: #374151; }
        .keywords { margin-top: 8px; }
        .keyword-tag { display: inline-block; background: #e5e7eb; padding: 2px 8px; border-radius: 12px; font-size: 0.8em; margin-right: 5px; }
        .timestamp { color: #888; font-size: 0.85em; }
    </style>
</head>
<body>
    <h1>🧠 Neural Memory Report</h1>
    <p class="timestamp">Generato il: ${new Date().toLocaleString()}</p>
    
    <h2>📊 Statistiche</h2>
    <div class="stats">
        <div class="stat-card">
            <div class="stat-value">${report.stats?.total_nodes || 0}</div>
            <div class="stat-label">Nodi Totali</div>
        </div>
    </div>
    
    <h2>📈 Nodi per Tipo</h2>
    <div class="stats">
        ${Object.entries(report.stats?.by_type || {}).map(([type, count]) => `
            <div class="stat-card">
                <div class="stat-value">${count}</div>
                <div class="stat-label">${type}</div>
            </div>
        `).join('')}
    </div>
    
    <h2>🏆 Top Skills</h2>
    <ul class="node-list">
        ${(report.top_skills || []).map(s => `
            <li class="node-item">
                <span class="node-type type-${s.type}">${s.type}</span>
                <strong>${s.keywords && s.keywords[0] ? s.keywords[0] : 'Unknown'}</strong>
                <div class="keywords">
                    ${(s.keywords || []).slice(0, 5).map(k => `<span class="keyword-tag">${k}</span>`).join('')}
                </div>
                <div class="timestamp">Confidence: ${(s.confidence * 100).toFixed(0)}%</div>
            </li>
        `).join('')}
    </ul>
    
    <h2>📝 Lavoro Recente</h2>
    <ul class="node-list">
        ${(report.recent_work || []).map(n => `
            <li class="node-item">
                <span class="node-type type-${n.type}">${n.type}</span>
                <div class="keywords">
                    ${(n.keywords || []).slice(0, 5).map(k => `<span class="keyword-tag">${k}</span>`).join('')}
                </div>
                <div class="timestamp">${new Date(n.created_at).toLocaleString()}</div>
            </li>
        `).join('')}
    </ul>
</body>
</html>`;
    }
}

export default new MemoryService();
