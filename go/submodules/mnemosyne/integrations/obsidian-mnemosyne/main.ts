import { App, Plugin, PluginSettingTab, Setting, Notice, TFile, TFolder, parseYaml, stringifyYaml } from 'obsidian';

interface MnemosyneSettings {
  dataDir: string;
  bank: string;
  syncFolder: string;
  autoSync: boolean;
  syncInterval: number;
  includeAssistant: boolean;
}

const DEFAULT_SETTINGS: MnemosyneSettings = {
  dataDir: '.hermes/mnemosyne/data',
  bank: 'default',
  syncFolder: 'Mnemosyne',
  autoSync: false,
  syncInterval: 30,
  includeAssistant: false,
};

export default class MnemosynePlugin extends Plugin {
  settings: MnemosyneSettings;
  private _syncTimer: number | null = null;

  async onload() {
    await this.loadSettings();

    this.addCommand({
      id: 'mnemosyne-sync',
      name: 'Sync memories to vault',
      callback: () => this.syncMemories(),
    });

    this.addCommand({
      id: 'mnemosyne-search',
      name: 'Search memories',
      callback: () => {
        new Notice('Use Obsidian search in the Mnemosyne folder');
      },
    });

    this.addCommand({
      id: 'mnemosyne-stats',
      name: 'Show memory stats',
      callback: () => this.showStats(),
    });

    this.addRibbonIcon('database', 'Mnemosyne Sync', () => this.syncMemories());

    this.addSettingTab(new MnemosyneSettingTab(this.app, this));

    if (this.settings.autoSync) {
      this.startAutoSync();
    }
  }

  onunload() {
    this.stopAutoSync();
  }

  async loadSettings() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
  }

  async saveSettings() {
    await this.saveData(this.settings);
    if (this.settings.autoSync) {
      this.startAutoSync();
    } else {
      this.stopAutoSync();
    }
  }

  startAutoSync() {
    this.stopAutoSync();
    this._syncTimer = window.setInterval(
      () => this.syncMemories(),
      this.settings.syncInterval * 60 * 1000
    );
  }

  stopAutoSync() {
    if (this._syncTimer !== null) {
      clearInterval(this._syncTimer);
      this._syncTimer = null;
    }
  }

  /**
   * Run the Python sync script that exports memories as markdown files.
   */
  async syncMemories() {
    new Notice('Mnemosyne: Syncing memories...');

    try {
      const { execSync } = require('child_process');
      const path = require('path');
      const fs = require('fs');

      const vaultPath = (this.app.vault as any).adapter.basePath;
      const dataDir = path.join(
        require('os').homedir(),
        this.settings.dataDir
      );
      const dbPath = path.join(dataDir, `${this.settings.bank}.db`);
      const outputDir = path.join(vaultPath, this.settings.syncFolder);

      // Ensure output directory exists
      if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true });
      }

      // Python script to export memories as markdown
      const pyScript = `
import json, sqlite3, os, sys
from datetime import datetime
from pathlib import Path

db_path = ${JSON.stringify(dbPath)}
output_dir = ${JSON.stringify(outputDir)}
include_assistant = ${this.settings.includeAssistant}

conn = sqlite3.connect(db_path)
conn.row_factory = sqlite3.Row
cursor = conn.cursor()

cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
tables = {r[0] for r in cursor.fetchall()}

count = 0
for table in ['working_memory', 'episodic_memory', 'memories']:
    if table not in tables: continue
    try:
        cursor.execute(f"SELECT * FROM {table} ORDER BY timestamp DESC")
        for row in cursor.fetchall():
            content = row['content'] or ''
            source = row.get('source', 'unknown')
            ts = str(row.get('timestamp', ''))[:19]
            imp = row.get('importance', 0)

            if source == 'openwebui:assistant' and not include_assistant:
                continue

            safe_id = str(row['id']).replace('/', '_').replace(':', '_')
            frontmatter = {
                'id': str(row['id']),
                'source': source,
                'tier': table,
                'timestamp': ts,
                'importance': imp,
                'synced': datetime.now().isoformat()[:19],
            }

            md = '---\\n'
            md += json.dumps(frontmatter, indent=2)
          md += '\\n---\\n\\n'
            md += f'**Source:** {source}  \\n'
            md += f'**Tier:** {table}  \\n'
            md += f'**Time:** {ts}  \\n'
            md += f'**Importance:** {imp}  \\n\\n'
            md += content

            filepath = os.path.join(output_dir, f'{safe_id}.md')
            # Avoid overwriting newer notes
            if os.path.exists(filepath):
                existing = open(filepath).read()
                if 'synced' in existing:
                    continue

            with open(filepath, 'w') as f:
                f.write(md)
            count += 1

    except Exception as e:
        print(f'Error in {table}: {e}', file=sys.stderr)

conn.close()
print(f'Synced {count} memories to {output_dir}')
`;
      const result = execSync(`python3 -c ${JSON.stringify(pyScript)}`, {
        timeout: 30000,
        encoding: 'utf-8',
      });

      // Refresh vault to show new files
      await this.app.vault.createFolder(this.settings.syncFolder).catch(() => {});

      new Notice(`Mnemosyne: Synced memories to "${this.settings.syncFolder}/"`);
      console.log('Mnemosyne sync result:', result.trim());

    } catch (err: any) {
      new Notice(`Mnemosyne sync failed: ${err.message || err}`);
      console.error('Mnemosyne sync error:', err);
    }
  }

  async showStats() {
    try {
      const { execSync } = require('child_process');
      const result = execSync(
        `python3 -c "
import json, sqlite3
from pathlib import Path
db = str(Path.home() / ${JSON.stringify(this.settings.dataDir)} / '${this.settings.bank}.db')
conn = sqlite3.connect(db)
c = conn.cursor()
c.execute(\\\"SELECT name FROM sqlite_master WHERE type='table'\\\")
tables = {r[0] for r in c.fetchall()}
stats = {}
for t in ['working_memory','episodic_memory','memories']:
    if t in tables:
        c.execute(f'SELECT COUNT(*) FROM {t}')
        stats[t] = c.fetchone()[0]
conn.close()
print(json.dumps(stats))
"`,
        { timeout: 10000, encoding: 'utf-8' }
      );
      const stats = JSON.parse(result.trim());
      const lines = Object.entries(stats).map(
        ([t, c]) => `- **${t.replace('_', ' ')}:** ${c}`
      );
      new Notice(
        `Mnemosyne Stats:\n${lines.join('\n')}`,
        8000
      );
    } catch (err) {
      new Notice('Failed to get Mnemosyne stats');
    }
  }
}

class MnemosyneSettingTab extends PluginSettingTab {
  plugin: MnemosynePlugin;

  constructor(app: App, plugin: MnemosynePlugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display(): void {
    const { containerEl } = this;
    containerEl.empty();

    containerEl.createEl('h2', { text: 'Mnemosyne Sync Settings' });

    new Setting(containerEl)
      .setName('Data directory')
      .setDesc('Path relative to home (~) where Mnemosyne stores its database')
      .addText(text => text
        .setPlaceholder('.hermes/mnemosyne/data')
        .setValue(this.plugin.settings.dataDir)
        .onChange(async val => {
          this.plugin.settings.dataDir = val;
          await this.plugin.saveSettings();
        }));

    new Setting(containerEl)
      .setName('Memory bank')
      .setDesc('Which Mnemosyne memory bank to sync')
      .addText(text => text
        .setPlaceholder('default')
        .setValue(this.plugin.settings.bank)
        .onChange(async val => {
          this.plugin.settings.bank = val;
          await this.plugin.saveSettings();
        }));

    new Setting(containerEl)
      .setName('Sync folder')
      .setDesc('Folder in your vault to store synced memories')
      .addText(text => text
        .setPlaceholder('Mnemosyne')
        .setValue(this.plugin.settings.syncFolder)
        .onChange(async val => {
          this.plugin.settings.syncFolder = val;
          await this.plugin.saveSettings();
        }));

    new Setting(containerEl)
      .setName('Auto-sync')
      .setDesc('Automatically sync memories on an interval')
      .addToggle(toggle => toggle
        .setValue(this.plugin.settings.autoSync)
        .onChange(async val => {
          this.plugin.settings.autoSync = val;
          await this.plugin.saveSettings();
        }));

    new Setting(containerEl)
      .setName('Sync interval')
      .setDesc('Minutes between auto-syncs')
      .addSlider(slider => slider
        .setLimits(5, 120, 5)
        .setValue(this.plugin.settings.syncInterval)
        .setDynamicTooltip()
        .onChange(async val => {
          this.plugin.settings.syncInterval = val;
          await this.plugin.saveSettings();
        }));

    new Setting(containerEl)
      .setName('Include assistant messages')
      .setDesc('Also sync AI assistant responses (not just user messages)')
      .addToggle(toggle => toggle
        .setValue(this.plugin.settings.includeAssistant)
        .onChange(async val => {
          this.plugin.settings.includeAssistant = val;
          await this.plugin.saveSettings();
        }));

    containerEl.createEl('hr');
    containerEl.createEl('p', {
      text: 'Use the ribbon icon or command palette to sync memories manually.',
      attr: { style: 'opacity: 0.7; font-size: 0.9em;' }
    });
  }
}
