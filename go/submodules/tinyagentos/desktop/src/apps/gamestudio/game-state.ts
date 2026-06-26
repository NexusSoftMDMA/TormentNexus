/** Per-window create -> play state between CreateView and PlayView. */

import type { BuildStep, Template } from "./types";

export const GAME_CREATED_EVENT = "gamestudio:game-created";

type PendingGame = {
  template: Template;
  prompt: string | null;
  steps: BuildStep[];
};

const pendingByWindow = new Map<string, PendingGame>();

export function seedCreatedGame(
  windowId: string,
  template: Template,
  prompt: string,
  steps: BuildStep[],
): void {
  pendingByWindow.set(windowId, {
    template,
    prompt: prompt.trim() || null,
    steps,
  });
  window.dispatchEvent(
    new CustomEvent(GAME_CREATED_EVENT, {
      detail: { windowId, template, prompt, steps },
    }),
  );
}

export function takeCreatedGame(windowId: string): {
  template: Template | null;
  prompt: string | null;
  steps: BuildStep[] | null;
} {
  const pending = pendingByWindow.get(windowId);
  if (!pending) {
    return { template: null, prompt: null, steps: null };
  }
  pendingByWindow.delete(windowId);
  return {
    template: pending.template,
    prompt: pending.prompt,
    steps: pending.steps,
  };
}