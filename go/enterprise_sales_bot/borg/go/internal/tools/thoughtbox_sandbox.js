/**
 * Thoughtbox Sandbox Runner
 * Spawns vm or mocks capabilities to execute Code Mode scripts (search / execute).
 */

const vmMod = require("vm");

const action = process.argv[2];
const code = process.argv[3];

if (!action || !code) {
  console.log(JSON.stringify({ error: "Missing action or code arguments" }));
  process.exit(1);
}

// 1. Define standard mock catalog for thoughtbox_search
const catalog = {
  publicTools: [
    { name: "thoughtbox_search", description: "Discover Thoughtbox operations by writing JS", operations: [] },
    { name: "thoughtbox_execute", description: "Run JS using the tb SDK to chain operations", operations: [] },
    { name: "thoughtbox_peer_notebook", description: "Brokered MCP peer notebook pilot", operations: [] }
  ],
  operations: {
    session: {
      session_list: { title: "List sessions", description: "List all sessions", category: "session" },
      session_get: { title: "Get session", description: "Get details of a session", category: "session" },
      session_search: { title: "Search sessions", description: "Search for sessions by keyword", category: "session" }
    },
    thought: {
      thought: { title: "Log thought", description: "Log a reasoning step", category: "thought" }
    },
    knowledge: {
      knowledge_create_entity: { title: "Create entity", description: "Create a knowledge entity", category: "knowledge" }
    }
  },
  prompts: [
    { name: "list_mcp_assets", description: "List all MCP assets", args: [] },
    { name: "interleaved-thinking", description: "Run interleaved thinking sessions", args: ["task"] }
  ],
  resources: [],
  resourceTemplates: []
};

// 2. Mock SDK for thoughtbox_execute
const mockTb = {
  thought: async (args) => {
    return { success: true, thoughtNumber: args.thoughtNumber || 1, sessionId: "session_mock_123" };
  },
  session: {
    list: async () => [{ sessionId: "session_mock_123", createdAt: new Date().toISOString() }],
    get: async (id) => ({ sessionId: id, thoughts: [] }),
    search: async (q) => [],
    resume: async (id) => ({ sessionId: id }),
    export: async (id) => "mock exported markdown session",
    analyze: async (id) => ({ analysis: "mock analysis" }),
    extractLearnings: async (id) => []
  },
  knowledge: {
    createEntity: async (args) => ({ id: "entity_mock", name: args.name }),
    getEntity: async (id) => ({ id, name: "mock entity" }),
    listEntities: async () => [],
    addObservation: async () => ({ success: true }),
    createRelation: async () => ({ success: true }),
    queryGraph: async () => ({ nodes: [], edges: [] }),
    stats: async () => ({ entityCount: 0, relationCount: 0 })
  }
};

const logs = [];
const cappedConsole = {
  log: (...args) => logs.push(args.map(String).join(" ")),
  warn: (...args) => logs.push(`[warn] ${args.map(String).join(" ")}`),
  error: (...args) => logs.push(`[error] ${args.map(String).join(" ")}`)
};

async function run() {
  const start = Date.now();
  let contextObj = {};

  if (action === "search") {
    contextObj = {
      __catalogJson: JSON.stringify(catalog),
      console: cappedConsole,
      setTimeout,
      clearTimeout
    };
  } else if (action === "execute") {
    contextObj = {
      tb: mockTb,
      console: cappedConsole,
      setTimeout,
      clearTimeout
    };
  }

  const context = vmMod.createContext(contextObj);

  try {
    let scriptContent = "";
    if (action === "search") {
      scriptContent = `
        const catalog = Object.freeze(JSON.parse(__catalogJson));
        Promise.resolve((${code})()).then(
          r => JSON.stringify(r),
          e => { throw e; }
        )
      `;
    } else {
      scriptContent = `
        Promise.resolve((${code})()).then(
          r => JSON.stringify(r),
          e => { throw e; }
        )
      `;
    }

    const script = new vmMod.Script(scriptContent, { filename: "thoughtbox_sandbox.js" });
    const rawResult = script.runInContext(context, { timeout: 10000 });
    const serialized = await Promise.race([
      rawResult,
      new Promise((_, reject) => setTimeout(() => reject(new Error("Execution timed out")), 10000))
    ]);

    const durationMs = Date.now() - start;
    console.log(JSON.stringify({
      result: serialized ? JSON.parse(serialized) : null,
      logs,
      durationMs
    }, null, 2));

  } catch (err) {
    console.log(JSON.stringify({
      result: null,
      logs,
      error: err.message || String(err),
      durationMs: Date.now() - start
    }, null, 2));
  }
}

run();
