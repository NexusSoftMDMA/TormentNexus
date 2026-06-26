# 🧠 Neural Memory v2.0

> ⚠️ **IMPORTANT NOTICE**: This is an **experimental MCP (Model Context Protocol)** tool. While it has shown promising results when used locally, it is still under active development and may evolve over time. **Feedback and contributions are highly appreciated!** Please report any issues, suggestions, or feature requests on our [GitHub Issues](https://github.com/drakonkat/neural-memory/issues) page. Your input helps make this tool better for everyone! nya~

---

**Model Context Protocol (MCP) system for knowledge memorization and contextual retrieval.**

---

## 🚀 Try Neural Memory NOW!

Want to try it right away in your Cline? Add this configuration to your `cline_mcp_settings.json`:

```json
{
  "mcpServers": {
    "neural-memory": {
      "timeout": 120,
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@drakonkat/neural-memory@latest"
      ]
    }
  }
}
```

### For Local Development

```json
{
  "mcpServers": {
    "neural-memory": {
      "timeout": 120,
      "type": "stdio",
      "command": "node",
      "args": [
        "./neural-memory/src/index.js"
      ]
    }
  }
}
```

> ⚠️ **Note:** Replace the path with your local project path!

---

## What's New in v2.0

### 🎯 Unified Database
- **No more separate databases per project!** All memory resides in a single SQLite database
- Faster, simpler, zero configuration

### 🎓 Skills Framework
- **Rigid schema** for registering skills with consistent structure
- `framework`, `language`, `filePattern`, `learnSteps`, `useCases`
- Automatic context-based suggestions

### 📊 HTML Reports
- Generate visual reports of your memory
- Statistics, most used skills, recent work

## Quick Start

```javascript
import { initialize, handleMcpRequest } from 'neural-memory';

// Initialize the service
await initialize();

// Add a node
await handleMcpRequest('add_node', {
  keywords: ['fastify', 'middleware', 'auth'],
  content: 'Implementation of JWT middleware...',
  type: 'generic'
});

// Register a skill
await handleMcpRequest('register_skill', {
  name: 'Fastify CRUD API',
  framework: 'fastify',
  language: 'javascript',
  filePattern: '**/*.service.js',
  learnSteps: [
    '1. Initialize project',
    '2. Create validation schema',
    '3. Implement handler'
  ],
  useCases: [
    'Create REST API',
    'Handle standardized errors'
  ]
});

// Search for skills
const skills = await handleMcpRequest('suggest_skills', {
  currentKeywords: ['fastify', 'api'],
  domain: 'javascript'
});

// Save context
await handleMcpRequest('save_context_snapshot', {
  summary: 'Refactoring 80% complete',
  pendingTasks: ['Integration testing', 'Deploy to staging'],
  learnings: ['Prisma requires explicit migrations']
});

// Generate report
const report = await handleMcpRequest('get_memory_report', {
  format: 'html'
});
```

## Available Tools

### Node Management
| Tool | Description |
|------|-------------|
| `add_node` | Add node to memory |
| `search_nodes` | Search nodes (with FTS5 + confidence) |
| `delete_node` | Delete node |

### Skills Framework
| Tool | Description |
|------|-------------|
| `register_skill` | Register skill (rigid schema) |
| `suggest_skills` | Suggest skills |

### Context Management
| Tool | Description |
|------|-------------|
| `save_context_snapshot` | Save context snapshot |
| `restore_context` | Restore context |

### Reports
| Tool | Description |
|------|-------------|
| `get_memory_report` | JSON/HTML report |

## Node Types

| Type | Priority | Use Case |
|------|----------|----------|
| `skill` | ⭐⭐⭐ | Knowledge skill |
| `error` | ⭐⭐⭐ | Resolved error |
| `operation` | ⭐⭐ | How-to |
| `convention` | ⭐⭐ | Naming rules |
| `edge_case` | ⭐⭐ | Edge case |
| `pattern` | ⭐ | Design pattern |
| `task` | ⭐ | Completed task |
| `generic` | ⭐ | Generic node |

## Confidence Scoring

The system calculates a **confidence score (0.0-1.0)** for each result based on:

- **BM25 Score** (35%) - Full-text search ranking
- **Keyword Match** (25%) - Keywords overlap
- **Recency Bonus** (10%) - Recent nodes preferred
- **Type Score** (15%) - Skills/errors are more important
- **Weight** (15%) - Manual weight

## Rigid Skill Schema

```json
{
  "name": "Skill Name",
  "framework": "fastify|react|prisma|...",
  "language": "javascript|typescript|python|...",
  "filePattern": "**/*.service.js",
  "learnSteps": ["Step 1", "Step 2"],
  "useCases": ["Case 1", "Case 2"],
  "implementation": "// Code...",
  "examples": ["file1.js"],
  "prerequisites": ["Prerequisite 1"]
}
```

## Installation

```bash
npm install neural-memory
```

## Database

The unified database is located at:
```
data/neural-memory-unified.sqlite
```

## License

MIT