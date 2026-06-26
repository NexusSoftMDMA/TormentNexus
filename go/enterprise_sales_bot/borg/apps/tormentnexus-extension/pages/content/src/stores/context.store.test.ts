import { beforeEach, describe, expect, it } from 'vitest';
import { useContextStore } from './context.store';

describe('useContextStore', () => {
  beforeEach(() => {
    localStorage.clear();
    useContextStore.setState({ contexts: [] });
  });

  it('captures browser context with source metadata', () => {
    const result = useContextStore.getState().captureContext({
      content: '  Selected browser text  ',
      name: 'Selection',
      source: 'context-menu',
      sourceUrl: 'https://example.com/docs',
      sourceTitle: 'Example Docs',
    });

    expect(result).not.toBeNull();
    expect(result?.duplicate).toBe(false);
    expect(result?.item.content).toBe('Selected browser text');
    expect(result?.item.name).toBe('Selection');
    expect(result?.item.source).toBe('context-menu');
    expect(result?.item.sourceUrl).toBe('https://example.com/docs');
    expect(result?.item.sourceTitle).toBe('Example Docs');
    expect(useContextStore.getState().contexts).toHaveLength(1);
  });

  it('deduplicates repeated captures and refreshes metadata', () => {
    const store = useContextStore.getState();
    const initial = store.captureContext({
      content: 'Repeated selection',
      source: 'context-menu',
      sourceTitle: 'First page',
    });

    const duplicate = useContextStore.getState().captureContext({
      content: 'Repeated selection',
      source: 'context-menu',
      sourceTitle: 'Updated page',
    });

    expect(initial).not.toBeNull();
    expect(duplicate).not.toBeNull();
    expect(duplicate?.duplicate).toBe(true);
    expect(useContextStore.getState().contexts).toHaveLength(1);
    expect(useContextStore.getState().contexts[0]?.sourceTitle).toBe('Updated page');
    expect(useContextStore.getState().contexts[0]?.id).toBe(initial?.item.id);
  });
});
