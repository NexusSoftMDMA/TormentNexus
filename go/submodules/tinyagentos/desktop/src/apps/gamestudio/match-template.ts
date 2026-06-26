import { TEMPLATES } from "./templates";
import type { Template } from "./types";

const RUNNER_HINTS = [
  "runner",
  "run",
  "endless",
  "platform",
  "jump",
  "racing",
  "race",
  "drift",
  "fps",
  "shooter",
  "corridor",
];

const ORBIT_HINTS = [
  "orbit",
  "tower",
  "defense",
  "defence",
  "puzzle",
  "top-down",
  "top down",
  "arena",
  "vr",
  "xr",
  "headset",
];

function scoreTemplate(template: Template, text: string): number {
  let score = 0;
  const genre = template.genre.toLowerCase();
  const title = template.title.toLowerCase();

  if (text.includes(genre)) score += 4;
  if (text.includes(title)) score += 3;
  for (const idPart of template.id.split("-")) {
    if (idPart.length > 3 && text.includes(idPart)) score += 2;
  }
  for (const word of template.desc.toLowerCase().split(/\W+/)) {
    if (word.length > 4 && text.includes(word)) score += 1;
  }

  if (RUNNER_HINTS.some((h) => text.includes(h)) && template.scene === "runner") {
    score += 3;
  }
  if (ORBIT_HINTS.some((h) => text.includes(h)) && template.scene === "orbit") {
    score += 3;
  }

  return score;
}

/** Pick the closest starter template for a prompt + genre chips. */
export function matchTemplate(prompt: string, genres: Iterable<string>): Template {
  const text = `${prompt} ${[...genres].join(" ")}`.toLowerCase().trim();
  if (!text) return TEMPLATES[0]!;

  let best = TEMPLATES[0]!;
  let bestScore = -1;
  for (const template of TEMPLATES) {
    const score = scoreTemplate(template, text);
    if (score > bestScore) {
      bestScore = score;
      best = template;
    }
  }
  return best;
}

export function isChessPrompt(prompt: string): boolean {
  return /\bchess\b/i.test(prompt);
}