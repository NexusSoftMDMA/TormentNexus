# 🧠 Neural Memory v2 - AI Usage Guidelines

Questo documento definisce le regole per l'uso ottimale di Neural Memory v2.

## 🎯 Golden Rule
**MEMORIZZA SEMPRE** - Dopo ogni operazione significativa, salva un nodo o uno snapshot. Meglio eccesso di contesto che perdita di memoria.

---

## 🔄 Ciclo di Vita della Sessione

1. **Inizio**: Esegui `start_session` all'avvio di ogni nuovo task o sessione di lavoro.
2. **Durante**: 
   - Salva nodi atomici con `add_node` (task, error, file, entity).
   - Registra procedure riutilizzabili con `register_skill`.
   - Crea checkpoint con `save_context_snapshot` ogni 3-5 passi o dopo modifiche strutturali.
3. **Fine**: Esegui `end_session` per generare le statistiche e consolidare la memoria.

---

## 🏗️ Tipi di Nodi e Priorità

Usa il `type` corretto per ottimizzare il recupero (Ranking Score):

| Tipo | Priorità | Quando usarlo |
|:---|:---|:---|
| `error` | ⭐⭐⭐ | Errori risolti, bug critici, fix complessi. |
| `skill` | ⭐⭐⭐ | Pattern di codice, setup framework, procedure "how-to". |
| `convention` | ⭐⭐ | Regole di naming, stile, architettura di progetto. |
| `task` | ⭐⭐ | Task completati o milestone raggiunte. |
| `operation` | ⭐⭐ | Comandi CLI, workflow di deployment/migrazione. |
| `edge_case` | ⭐ | Comportamenti anomali o casi limite scoperti. |
| `generic` | ⭐ | Tutto ciò che non rientra sopra. |

---

## 📝 Esempi Rapidi (v2)

### 1. Salvare un Fix (Errore)
```javascript
add_node({
  sessionId: "current",
  type: "error",
  keywords: ["null-pointer", "async", "auth"],
  content: "Risolto crash in auth middleware: la variabile user era undefined prima del caricamento.",
  metadata: { file: "src/auth.js", fix: "added optional chaining" }
})
```

### 2. Registrare una Procedura (Skill)
Usa `register_skill` per tutto ciò che vorresti saper rifare istantaneamente in futuro.
```javascript
register_skill({
  name: "Setup Fastify con Prisma",
  framework: "fastify",
  language: "typescript",
  learnSteps: ["Installa deps", "Configura schema.prisma", "Genera client"],
  implementation: "export const db = new PrismaClient()..."
})
```

### 3. Snapshot del Contesto
Usa `save_context_snapshot` per "salvare la partita".
```javascript
save_context_snapshot({
  summary: "Auth system completato",
  learnings: ["JWT richiede secret in env", "Prisma accelera query mapping"],
  nextSteps: ["Implementare RBAC", "Testare logout"]
})
```

---

## 💡 Best Practices

- **Keywords**: Usa sempre 3-5 parole chiave (tecnologia, modulo, azione).
- **Metadata**: Includi percorsi file, nomi di funzioni o endpoint.
- **Gerarchia**: Usa `parentId` per collegare task piccoli a task grandi.
- **Relazioni**: Usa `link_nodes` per collegare fix a errori o skill a pattern.
- **Search**: Prima di iniziare un task, cerca sempre se esiste una `skill` o un `error` correlato.

---
*Memory is Power. Più scrivi, più diventi intelligente.*
