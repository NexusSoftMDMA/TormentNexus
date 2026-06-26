/** Shared prompt seeding between TemplatesView and BuildView. */

export const PROMPT_SEEDED_EVENT = "appstudio:prompt-seeded";
export const SHOW_BUILD_VIEW_EVENT = "appstudio:show-build";

let pendingPrompt: string | null = null;

export function seedBuildPrompt(text: string): void {
  const trimmed = text.trim();
  if (!trimmed) return;
  pendingPrompt = trimmed;
  window.dispatchEvent(new CustomEvent(PROMPT_SEEDED_EVENT, { detail: trimmed }));
  window.dispatchEvent(new CustomEvent(SHOW_BUILD_VIEW_EVENT));
}

export function takePendingPrompt(): string | null {
  const prompt = pendingPrompt;
  pendingPrompt = null;
  return prompt;
}