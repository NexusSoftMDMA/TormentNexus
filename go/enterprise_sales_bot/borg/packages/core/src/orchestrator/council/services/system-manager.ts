import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';

const execAsync = promisify(exec);

export interface SubmoduleInfo {
  path: string;
  commit: string;
  version?: string;
  url?: string;
  status: 'clean' | 'dirty' | 'unknown';
}

class SystemManagerService {
  async getSubmodules(): Promise<SubmoduleInfo[]> {
    try {
      // Execute from two levels up (repo root) relative to this package
      const rootDir = path.resolve(process.cwd(), '../../');
      const { stdout } = await execAsync('git submodule status --recursive', { cwd: rootDir });
      const lines = stdout.trim().split('\n');

      const submodules: SubmoduleInfo[] = [];

      for (const line of lines) {
        if (!line.trim()) continue;

        // Format: -d2e1... path/to/submodule (version)
        // or:  d2e1... path/to/submodule (version)
        // or: +d2e1... path/to/submodule (version)

        const match = line.match(/^([-\+ ])([0-9a-f]+)\s+(\S+)(?:\s+\((.*)\))?/);
        if (match) {
          const [, indicator, commit, path, version] = match;
          let status: SubmoduleInfo['status'] = 'clean';
          if (indicator === '+') status = 'dirty';
          if (indicator === '-') status = 'unknown'; // Not initialized

          submodules.push({
            path,
            commit,
            version,
            status
          });
        }
      }

      return submodules;
    } catch (error) {
      console.error('Failed to get submodules:', error);
      return [];
    }
  }

  async getProjectVersion(): Promise<string> {
    try {
      const rootDir = path.resolve(process.cwd(), '../../');
      const { stdout } = await execAsync('git describe --tags --always', { cwd: rootDir });
      return stdout.trim();
    } catch {
      return 'unknown';
    }
  }
}

export const systemManager = new SystemManagerService();
