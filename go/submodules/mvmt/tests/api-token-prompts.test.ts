import { beforeEach, describe, expect, it, vi } from 'vitest';
import { parseConfig } from '../src/config/loader.js';

const promptMocks = vi.hoisted(() => ({
  confirm: vi.fn(),
  input: vi.fn(),
  select: vi.fn(),
}));

vi.mock('@inquirer/prompts', () => promptMocks);

const { promptAndEditApiToken } = await import('../src/cli/api-tokens.js');

const EXISTING_TOKEN_VERIFIER = 'scrypt:v1:AAAAAAAAAAAAAAAAAAAAAA:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

describe('API token prompts', () => {
  beforeEach(() => {
    promptMocks.confirm.mockReset();
    promptMocks.input.mockReset();
    promptMocks.select.mockReset();
  });

  it('does not silently downgrade read-only mounts when editing write scopes', async () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'documents', type: 'local_folder', path: '/documents', root: '/tmp/documents' },
        { name: 'workspace', type: 'local_folder', path: '/workspace', root: '/tmp/workspace', writeAccess: true },
      ],
      clients: [
        {
          id: 'claude',
          name: 'Claude',
          auth: { type: 'token', tokenHash: EXISTING_TOKEN_VERIFIER },
          permissions: [{ path: '/documents/**', actions: ['search', 'read'] }],
        },
      ],
    });

    promptMocks.select
      .mockResolvedValueOnce('claude')
      .mockResolvedValueOnce('scope')
      .mockResolvedValueOnce('specific')
      .mockResolvedValueOnce('write')
      .mockImplementationOnce(async (prompt: { choices: Array<{ value: { id: string } }> }) => (
        prompt.choices.find((choice) => choice.value.id === 'workspace')!.value
      ));
    promptMocks.confirm.mockResolvedValueOnce(false);

    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    let result: Awaited<ReturnType<typeof promptAndEditApiToken>> | undefined;
    try {
      result = await promptAndEditApiToken(config);
    } finally {
      output.mockRestore();
    }

    const sourcePrompt = promptMocks.select.mock.calls.find(([prompt]) => (
      prompt.message === 'Grant access to which connector?'
    ))?.[0] as { choices: Array<{ value: { id: string }; disabled?: string }> };
    expect(sourcePrompt.choices.find((choice) => choice.value.id === 'documents')?.disabled)
      .toContain('read-only');
    expect(result?.client.permissions).toEqual([
      { path: '/workspace/**', actions: ['search', 'read', 'write'] },
    ]);
  });
});
