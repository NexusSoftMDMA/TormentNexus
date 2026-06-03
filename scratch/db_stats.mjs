import Database from 'better-sqlite3';

// Stats from tormentnexus.db
const db = new Database('tormentnexus.db');
const servers = db.prepare('SELECT COUNT(*) as cnt FROM mcp_servers').get();
const tools = db.prepare('SELECT COUNT(*) as cnt FROM tools').get();
console.log(`=== tormentnexus.db ===`);
console.log(`mcp_servers (registered): ${servers.cnt}`);
console.log(`tools (registered):       ${tools.cnt}`);
db.close();

// Stats from catalog.db
const cat = new Database('catalog.db');
const pubTotal = cat.prepare('SELECT COUNT(*) as cnt FROM published_mcp_servers').get();
const allStatuses = cat.prepare("SELECT status, COUNT(*) as cnt FROM published_mcp_servers GROUP BY status ORDER BY cnt DESC").all();
const validRuns = cat.prepare('SELECT COUNT(*) as cnt FROM published_mcp_validation_runs').get();
const successRuns = cat.prepare("SELECT COUNT(*) as cnt FROM published_mcp_validation_runs WHERE outcome='success'").get();
const failRuns = cat.prepare("SELECT COUNT(*) as cnt FROM published_mcp_validation_runs WHERE outcome='failure'").get();

console.log(`\n=== catalog.db ===`);
console.log(`published_mcp_servers total: ${pubTotal.cnt}`);
console.log(`Status breakdown:`);
allStatuses.forEach(s => console.log(`  ${s.status || 'null'}: ${s.cnt}`));
console.log(`\nValidation runs total:       ${validRuns.cnt}`);
console.log(`  Success:                   ${successRuns.cnt}`);
console.log(`  Failure:                   ${failRuns.cnt}`);
cat.close();
