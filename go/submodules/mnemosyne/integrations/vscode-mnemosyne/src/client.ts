import { exec, execSync } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export interface Memory {
  id: string;
  content: string;
  source: string;
  timestamp: string;
  importance: number;
  score: number;
  tier: string;
}

export interface Stats {
  total: number;
  tiers: Record<string, number>;
  error?: string;
}

export class MnemosyneClient {
  private pythonCmd: string;

  constructor() {
    this.pythonCmd = 'python3';
    // Try to find python
    try {
      execSync('python3 --version', { stdio: 'ignore' });
    } catch {
      try {
        execSync('python --version', { stdio: 'ignore' });
        this.pythonCmd = 'python';
      } catch {
        this.pythonCmd = 'python3';
      }
    }
  }

  /**
   * Run a Python snippet that queries Mnemosyne and returns JSON on stdout.
   */
  private async runPython(code: string): Promise<string> {
    const { stdout, stderr } = await execAsync(
      `${this.pythonCmd} -c ${JSON.stringify(code)}`,
      { timeout: 10000 }
    );
    if (stderr) console.error('Mnemosyne CLI stderr:', stderr);
    return stdout.trim();
  }

  /**
   * Search memories by query string.
   */
  async search(query: string, limit = 20): Promise<Memory[]> {
    const code = `
import json, sys
try:
  from mnemosyne import recall
  results = recall(${JSON.stringify(query)}, top_k=${limit})
  memories = []
  for r in (results or []):
    memories.append({
      'id': str(r.get('memory_id', r.get('id', ''))),
      'content': str(r.get('content', ''))[:300],
      'source': str(r.get('source', '')),
      'timestamp': str(r.get('timestamp', '')),
      'importance': r.get('importance', 0),
      'score': r.get('score', 0),
      'tier': str(r.get('tier', 'working')),
    })
  print(json.dumps(memories))
except Exception as e:
  print(json.dumps({'error': str(e)}))
`;
    const output = await this.runPython(code);
    try {
      const data = JSON.parse(output);
      if (data.error) {
        console.error('Search error:', data.error);
        return [];
      }
      return data;
    } catch {
      return [];
    }
  }

  /**
   * Get memory statistics.
   */
  async getStats(): Promise<Stats> {
    const code = `
import json, os, sqlite3
from pathlib import Path
try:
  data_dir = os.environ.get('MNEMOSYNE_DATA_DIR', str(Path.home() / '.hermes' / 'mnemosyne' / 'data'))
  db_path = str(Path(data_dir) / 'default.db')
  conn = sqlite3.connect(db_path)
  conn.row_factory = sqlite3.Row
  cursor = conn.cursor()
  cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
  tables = {r[0] for r in cursor.fetchall()}
  stats = {'total': 0, 'tiers': {}}
  for table in ['working_memory', 'episodic_memory', 'memories']:
    if table in tables:
      cursor.execute(f"SELECT COUNT(*) FROM {table}")
      c = cursor.fetchone()[0]
      stats['tiers'][table] = c
      stats['total'] += c
  conn.close()
  print(json.dumps(stats))
except Exception as e:
  print(json.dumps({'error': str(e), 'total': 0, 'tiers': {}}))
`;
    const output = await this.runPython(code);
    try {
      return JSON.parse(output);
    } catch {
      return { total: 0, tiers: {} };
    }
  }

  /**
   * Store a memory.
   */
  async remember(content: string, source = 'vscode', importance = 0.5): Promise<string> {
    const code = `
import json
try:
  from mnemosyne import remember
  mid = remember(${JSON.stringify(content)}, source=${JSON.stringify(source)}, importance=${importance})
  print(json.dumps({'id': str(mid)}))
except Exception as e:
  print(json.dumps({'error': str(e)}))
`;
    const output = await this.runPython(code);
    try {
      const data = JSON.parse(output);
      return data.id || '';
    } catch {
      return '';
    }
  }

  /**
   * List recent memories.
   */
  async listRecent(limit = 30): Promise<Memory[]> {
    const code = `
import json, os, sqlite3, sys
from pathlib import Path
try:
  data_dir = os.environ.get('MNEMOSYNE_DATA_DIR', str(Path.home() / '.hermes' / 'mnemosyne' / 'data'))
  db_path = str(Path(data_dir) / 'default.db')
  conn = sqlite3.connect(db_path)
  conn.row_factory = sqlite3.Row
  cursor = conn.cursor()
  cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
  tables = {r[0] for r in cursor.fetchall()}
  memories = []
  for table in ['working_memory', 'episodic_memory', 'memories']:
    if table not in tables: continue
    try:
      cursor.execute(f"SELECT * FROM {table} ORDER BY timestamp DESC LIMIT ${limit}")
      for row in cursor.fetchall():
        memories.append({
          'id': str(row['id']),
          'content': str(row['content'])[:300],
          'source': str(row.get('source', '')),
          'timestamp': str(row.get('timestamp', '')),
          'importance': row.get('importance', 0),
          'score': 0,
          'tier': table,
        })
    except: pass
  conn.close()
  memories.sort(key=lambda m: m.get('timestamp', ''), reverse=True)
  print(json.dumps(memories[:${limit}]))
except Exception as e:
  print(json.dumps([]))
`;
    const output = await this.runPython(code);
    try {
      return JSON.parse(output);
    } catch {
      return [];
    }
  }
}
