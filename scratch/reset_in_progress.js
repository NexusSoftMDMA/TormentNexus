const db = require('better-sqlite3')('catalog.db');
db.prepare("UPDATE published_mcp_servers SET status = 'discovered' WHERE status = 'in_progress'").run();
console.log("Reset in_progress to discovered.");
