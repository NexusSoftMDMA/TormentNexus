# Neural Memory v2.0 - Specifica Tecnica

## Panoramica

**Neural Memory** è un sistema MCP (Model Context Protocol) che implementa una rete neurale artificiale per la memorizzazione e il recupero di conoscenze contestuali. La versione 2.0 introduce un **database unificato** con **Session Management** avanzato e **Skills Framework**.

## Architettura v2.0

### Database Unificato

**Non più database separati per progetto.** Tutta la memoria risiede in un unico database SQLite:

```
data/
└── neural-memory-unified.sqlite
```

### Schema Database

#### Tabella `sessions` (NUOVA in v2.0)
| Campo | Tipo | Descrizione |
|-------|------|-------------|
| id | UUID | Identificatore unico |
| name | TEXT | Nome sessione |
| description | TEXT | Descrizione lavoro |
| started_at | DATETIME | Inizio sessione |
| ended_at | DATETIME | Fine sessione (null = attiva) |
| context | JSON | Snapshot contesto iniziale |
| project_path | TEXT | Riferimento percorso progetto |
| tags | JSON | Tag categorizzazione |
| stats | JSON | Statistiche: nodes, skills, duration |
| is_active | BOOLEAN | Sessione attualmente attiva |

#### Tabella `nodes`
| Campo | Tipo | Descrizione |
|-------|------|-------------|
| id | UUID | Identificatore unico |
| session_id | UUID | Sessione di appartenenza (nullable) |
| parent_id | UUID | Nodo padre per gerarchia |
| type | TEXT | Categoria: skill, error, task, etc. |
| keywords | JSON | Keywords per ricerca |
| content | TEXT | Contenuto long-text |
| metadata | JSON | Dati strutturati |
| depth | INTEGER | Livello gerarchia |
| weight | FLOAT | Peso per ranking (0.1-10.0) |
| keyword_count | INTEGER | Ottimizzazione ricerca |
| task_date | DATE | Data task associato |

#### Tabella `node_links`
| Campo | Tipo | Descrizione |
|-------|------|-------------|
| id | UUID | Identificatore |
| from_node_id | UUID | Nodo sorgente |
| to_node_id | UUID | Nodo destinazione |
| link_type | TEXT | child, parent, related, reference, trigger, caused |
| weight | FLOAT | Peso collegamento |
| metadata | JSON | Dati aggiuntivi |

#### Tabella `nodes_fts` (FTS5 Virtual Table)
Per full-text search ottimizzato.

---

## Tools MCP v2.0

### Node Management

#### `register_skill`
Registra skill con schema obbligatorio.
```json
{
  "name": "Fastify CRUD API",
  "framework": "fastify",
  "language": "javascript",
  "filePattern": "**/*.service.js",
  "learnSteps": [
    "1. Inizializzare progetto con npm",
    "2. Creare schema di validazione",
    "3. Implementare handler CRUD"
  ],
  "useCases": [
    "Creare API REST con validazione",
    "Gestire errori standardizzati"
  ],
  "implementation": "// Codice di esempio...",
  "examples": ["auth.service.js", "user.service.js"],
  "prerequisites": ["Node.js basics", "SQL fundamentals"]
}
```

#### `suggest_skills`
Suggerisce skills basate su contesto.
```json
{
  "currentKeywords": ["database", "postgres"],
  "domain": "javascript",
  "maxResults": 5
}
```

---

### Context Management

#### `save_context_snapshot`
Salva snapshot contesto dettagliato.
```json
{
  "summary": "Refactoring completato all'80%",
  "workDone": {
    "completed": ["Auth middleware", "User schema"],
    "inProgress": ["Order service"]
  },
  "pendingTasks": [
    "Completare Order service",
    "Scrivere test di integrazione"
  ],
  "keyDecisions": [
    "Usare Prisma come ORM",
    "Validazione con Zod"
  ],
  "blockers": [
    "Manca configurazione staging"
  ],
  "learnings": [
    "Prisma richiede migration explicite"
  ],
  "nextSteps": [
    "Testare endpoint orders",
    "Deploy su staging"
  ]
}
```

**Nota:** `pendingTasks`, `keyDecisions`, `blockers`, `learnings`, e `nextSteps` devono essere array di `string[]`, non oggetti. La validazione Zod richiede stringhe semplici.

#### `restore_context`
Recupera contesto salvato.
```json
{
  "snapshotId": "uuid-snapshot"
}
```

---

### Node Management

#### `add_node`
Aggiunge nodo alla memoria.
```json
{
  "sessionId": "uuid-sessione",
  "keywords": ["fastify", "middleware", "auth"],
  "content": "Implementazione middleware JWT...",
  "type": "generic",
  "parentId": "uuid-parent",
  "metadata": {},
  "weight": 1.0
}
```

#### `search_nodes`
Cerca nodi con confidence scoring.
```json
{
  "keywords": ["fastify", "crud"],
  "maxResults": 10,
  "minConfidence": 0.3,
  "type": "skill",
  "sessionId": null
}
```

#### `delete_node`
Elimina nodo.
```json
{
  "nodeId": "uuid-nodo",
  "cascade": false
}
```

---

### Reports

#### `get_memory_report`
Genera report in formato JSON o HTML.
```json
{
  "format": "html",
  "keywords": ["fastify"],
  "includeStats": true,
  "includeRecentWork": true,
  "includeTopSkills": true
}
```

---

## Confidence Scoring

Il sistema calcola un **confidence score** (0.0-1.0) basato su:

| Componente | Peso Max | Descrizione |
|------------|----------|-------------|
| BM25 Score | 0.35 | Full-text search ranking |
| Keyword Match | 0.25 | Sovrapposizione keywords |
| Recency Bonus | 0.10 | Nodi recenti preferiti |
| Type Score | 0.15 | Skills/errors più importanti |
| Weight | 0.15 | Peso manuale |

### Type Scores
```javascript
const typeScores = {
  'error': 0.15,        // Priorità MASSIMA
  'skill': 0.15,        // Skills molto importanti
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
```

---

## Node Types

| Type | Descrizione | Priority |
|------|-------------|----------|
| `skill` | Knowledge skill registrata | ALTA |
| `error` | Errore risolto | ALTA |
| `operation` | How-to documentato | MEDIA-ALTA |
| `convention` | Regola di stile/naming | MEDIA-ALTA |
| `edge_case` | Caso limite scoperto | MEDIA |
| `pattern` | Design pattern | MEDIA |
| `task` | Task completato/pendente | MEDIA |
| `action` | Azione eseguita | MEDIA-BASSA |
| `entity` | Entità di sistema | BASSA |
| `file` | Riferimento file | BASSA |
| `concept` | Concetto astratto | BASSA |
| `summary` | Riassunto | BASSA |
| `generic` | Nodo generico | MINIMA |
| `context_snapshot` | Snapshot contesto | SPECIALE |

---

## Link Types

| Type | Descrizione |
|------|-------------|
| `child` | Figlio gerarchico |
| `parent` | Padre gerarchico |
| `related` | Collegamento generico |
| `reference` | Riferimento a docs |
| `trigger` | Causa-trigger |
| `caused` | Effetto-causato |

---

## Metriche Sessione

Ogni sessione traccia:
- `nodesCreated`: Nodi creati durante la sessione
- `skillsRegistered`: Skills apprese
- `skillsUsed`: Skills utilizzate/applicate
- `durationMinutes`: Durata totale

---

## Changelog v2.0

### Breaking Changes
- **RIMOSSO**: `projectId` dai nodi (memoria unificata)
- **RIMOSSO**: Database separati per progetto
- **MODIFICATO**: Schema Node senza project_id

### Nuove Funzionalità
- **AGGIUNTO**: Session Management completo
- **AGGIUNTO**: Skills Framework con schema rigido
- **AGGIUNTO**: Context Compression
- **AGGIUNTO**: Reports HTML
- **AGGIUNTO**: Confidence scoring rafforzato per Skills/Errors

### Miglioramenti
- Ricerca full-text ottimizzata
- Type-based scoring migliorato
- Gerarchia nodi più robusta
