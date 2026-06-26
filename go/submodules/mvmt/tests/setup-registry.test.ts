import { describe, expect, it } from 'vitest';
import { getConnectorSetupDefinition, getSetupRegistry } from '../src/connectors/setup-registry.js';

describe('setup registry', () => {
  it('returns connector setups in guided setup order', () => {
    expect(getSetupRegistry().map((definition) => definition.id)).toEqual([
      'filesystem',
    ]);
  });

  it('returns undefined for an unknown connector setup id', () => {
    expect(getConnectorSetupDefinition('unknown')).toBeUndefined();
  });
});
