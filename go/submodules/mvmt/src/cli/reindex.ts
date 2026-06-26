import chalk from 'chalk';
import { defaultTextIndexPath, TextContextIndex } from '../context/text-index.js';
import { configExists, loadConfig, resolveConfigPath } from '../config/loader.js';

export interface ReindexOptions {
  config?: string;
}

export async function reindex(options: ReindexOptions = {}): Promise<void> {
  const configPath = resolveConfigPath(options.config);
  if (!configExists(configPath)) {
    console.error(`Config not found at ${configPath}`);
    console.error('Run `mvmt config setup` to create one.');
    process.exitCode = 1;
    return;
  }

  const config = loadConfig(configPath);
  if (config.mounts.length === 0) {
    console.error('No mounts configured.');
    console.error('Run `mvmt mounts add <name> <folder>` or `mvmt config setup` first.');
    process.exitCode = 1;
    return;
  }

  const index = new TextContextIndex({
    mounts: config.mounts,
    indexPath: defaultTextIndexPath(configPath),
  });
  const stats = await index.rebuild();
  const truncated = stats.truncated ? ' Index was truncated; narrow mounts or add exclude rules.' : '';
  console.log(chalk.green(`Indexed ${stats.files} text files into ${stats.chunks} chunks.${truncated}`));
  console.log(chalk.dim(`Index: ${defaultTextIndexPath(configPath)}`));
}
