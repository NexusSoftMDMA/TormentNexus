import { describe, expect, it } from 'vitest';
import { parseConfig } from '../src/config/loader.js';
import { joinVirtualPath, MountRegistry, normalizeVirtualPath } from '../src/context/mount-registry.js';

describe('MountRegistry', () => {
  it('resolves global paths to the matching mount and relative path', () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'notes', type: 'local_folder', path: '/notes', root: '/real/notes' },
      ],
    });
    const registry = new MountRegistry(config.mounts);

    expect(registry.resolve('/notes/projects/mvmt.md')).toMatchObject({
      mount: { config: { name: 'notes', path: '/notes' } },
      relativePath: 'projects/mvmt.md',
      realPath: '/real/notes/projects/mvmt.md',
      virtualPath: '/notes/projects/mvmt.md',
    });
  });

  it('uses longest-prefix matching for nested mounts', () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'desktop', type: 'local_folder', path: '/desktop', root: '/mnt/desktop' },
        { name: 'desktop-projects', type: 'local_folder', path: '/desktop/projects', root: '/mnt/projects' },
      ],
    });
    const registry = new MountRegistry(config.mounts);

    expect(registry.resolve('/desktop/projects/mvmt/README.md')).toMatchObject({
      mount: { config: { name: 'desktop-projects' } },
      relativePath: 'mvmt/README.md',
      realPath: '/mnt/projects/mvmt/README.md',
    });
  });

  it('does not match partial path segments', () => {
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: '/real/notes' }],
    });
    const registry = new MountRegistry(config.mounts);

    expect(() => registry.resolve('/notebook/todo.md')).toThrow('unknown mount for path: /notebook/todo.md');
  });

  it('normalizes virtual path separators without regex backtracking', () => {
    expect(normalizeVirtualPath('////notes\\projects\\mvmt.md')).toBe('/notes/projects/mvmt.md');
    expect(joinVirtualPath('/notes////', '\\projects\\mvmt.md')).toBe('/notes/projects/mvmt.md');
  });
});
