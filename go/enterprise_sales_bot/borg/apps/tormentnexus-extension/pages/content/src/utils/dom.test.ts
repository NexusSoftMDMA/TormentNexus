import { describe, expect, it } from 'vitest';

import { findBestActionButton } from './dom';

describe('findBestActionButton', () => {
  it('prefers an exact send button over a nearby dropdown trigger', () => {
    document.body.innerHTML = `
      <form>
        <textarea id="chat"></textarea>
        <button type="button" aria-label="Always Run Ask every time">Always Run</button>
        <button type="submit" aria-label="Send message">Send</button>
      </form>
    `;

    const chatInput = document.getElementById('chat') as HTMLTextAreaElement;
    const button = findBestActionButton({
      actionLabels: ['Send message', 'Send', 'Submit'],
      preferredSelectors: ['button[aria-label="Send message"]', 'button[type="submit"]'],
      near: chatInput,
    });

    expect(button?.getAttribute('aria-label')).toBe('Send message');
  });

  it('rejects menu-style buttons even when they contain the target word', () => {
    document.body.innerHTML = `
      <div>
        <button type="button" aria-label="Run options" aria-haspopup="menu">Run</button>
        <button type="button" aria-label="Run">Run</button>
      </div>
    `;

    const button = findBestActionButton({
      actionLabels: ['Run'],
      preferredSelectors: ['button[aria-label="Run"]'],
    });

    expect(button?.getAttribute('aria-label')).toBe('Run');
    expect(button?.getAttribute('aria-haspopup')).toBeNull();
  });
});
