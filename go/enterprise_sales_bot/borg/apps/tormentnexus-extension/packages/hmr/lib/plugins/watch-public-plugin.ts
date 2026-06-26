import type { Plugin, PluginContext } from 'rollup';
import fg from 'fast-glob';

export function watchPublicPlugin(): Plugin {
  return {
    name: 'watch-public-plugin',
    async buildStart(this: PluginContext) {
      const files = await fg(['public/**/*']);

      for (const file of files) {
        this.addWatchFile(file);
      }
    },
  };
}
