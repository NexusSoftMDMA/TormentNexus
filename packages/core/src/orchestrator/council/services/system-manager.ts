import { spawnAsync } from '../../../utils/exec.js';
import path from 'path';

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
      const rootDir = path.resolve(process.cwd(), '../../');
      const result = await spawnAsync('git', ['submodule', 'status', '--recursive'], { cwd: rootDir });
      const lines = result.stdout.trim().split('\n');

      const submodules: SubmoduleInfo[] = [];

      for (const line of lines) {
        if (!line.trim()) continue;

        const match = line.match(/^([-\+ ])([0-9a-f]+)\s+(\S+)(?:\s+\((.*)\))?/);
        if (match) {
          const [, indicator, commit, subPath, version] = match;
          let status: SubmoduleInfo['status'] = 'clean';
          if (indicator === '+') status = 'dirty';
          if (indicator === '-') status = 'unknown';

          submodules.push({
            path: subPath,
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
      const result = await spawnAsync('git', ['describe', '--tags', '--always'], { cwd: rootDir });
      return result.stdout.trim();
    } catch {
      return 'unknown';
    }
  }
}

export const systemManager = new SystemManagerService();
