import { describe, it, expect } from 'vitest';
import { chatDisplayTitle } from '@/lib/chat-title';

describe('chatDisplayTitle', () => {
  it('extracts and formats persona-wrapped work', () => {
    const work = `# Persona: Sentinel

Security perspective

---

# User request

Review auth code`;
    const result = chatDisplayTitle(work);
    expect(result).toBe('[Sentinel] Review auth code');
  });

  it('returns plain work without persona wrapper', () => {
    const work = 'Plain work request without persona';
    const result = chatDisplayTitle(work);
    expect(result).toBe('Plain work request without persona');
  });

  it('returns empty string for empty work', () => {
    const result = chatDisplayTitle('');
    expect(result).toBe('');
  });

  it('handles whitespace around persona name and request body', () => {
    const work = `# Persona:   Cartographer

Cross-platform expert

---

# User request

   Refactor component structure   `;
    const result = chatDisplayTitle(work);
    expect(result).toBe('[Cartographer] Refactor component structure');
  });

  it('completes adversarial 1MB input with no separator in <50ms', () => {
    // Create a 1MB input with no `---` separator to test regex backtracking guard
    const largeInput = 'x'.repeat(1024 * 1024 - 100) + 'no separator here';

    const start = performance.now();
    const result = chatDisplayTitle(largeInput);
    const elapsed = performance.now() - start;

    expect(result).toBe(largeInput); // Should fall back to raw string
    expect(elapsed).toBeLessThan(50);
  });

  it('handles undefined/null gracefully', () => {
    const result = chatDisplayTitle(undefined as unknown as Parameters<typeof chatDisplayTitle>[0]);
    expect(result).toBe('');
  });
});
