import { matchTemplate, isChessPrompt } from "./match-template";
import type { BuildStep, Template } from "./types";

export interface CreateGameResult {
  template: Template;
  steps: BuildStep[];
}

type StepUpdater = (steps: BuildStep[]) => void;

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function setStepState(steps: BuildStep[], index: number, state: BuildStep["state"]): BuildStep[] {
  return steps.map((s, i) => (i === index ? { ...s, state } : s));
}

async function tryChessAgentNote(
  prompt: string,
  onUpdate: StepUpdater,
  steps: BuildStep[],
): Promise<BuildStep[]> {
  if (!isChessPrompt(prompt)) return steps;

  let agents: string[] = [];
  try {
    const res = await fetch("/api/agents");
    if (res.ok) {
      const data = (await res.json()) as { name?: string }[];
      agents = Array.isArray(data)
        ? data.map((a) => a?.name).filter((n): n is string => !!n)
        : [];
    }
  } catch {
    return steps;
  }
  if (agents.length === 0) return steps;

  const agent = agents[0]!;
  try {
    const res = await fetch("/api/games/chess/move", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        agent_name: agent,
        fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        legal_moves: ["e2e4", "d2d4", "g1f3", "b1c3"],
        history: [],
      }),
    });
    const data = (await res.json()) as { commentary?: string; move?: string };
    const note = data.commentary?.trim();
    if (!note) return steps;

    const next = steps.map((s) =>
      s.tag === "routing" && s.director
        ? {
            ...s,
            what: `${s.what} Chess agent "${agent}" suggested opening ${data.move ?? "e2e4"}: ${note.slice(0, 120)}`,
          }
        : s,
    );
    onUpdate(next);
    return next;
  } catch {
    return steps;
  }
}

/** Run the local create pipeline: match a template, animate build steps, optional games.py chess consult. */
export async function runCreateGame(
  prompt: string,
  genres: Set<string>,
  onUpdate: StepUpdater,
): Promise<CreateGameResult> {
  const template = matchTemplate(prompt, genres);

  let steps: BuildStep[] = [
    {
      who: "Director",
      what: `Matched "${prompt.trim().slice(0, 80) || "your idea"}" to the ${template.title} template.`,
      tag: "routing",
      state: "run",
      director: true,
    },
    {
      who: "Gameplay",
      what: `Wiring the ${template.genre.toLowerCase()} loop and controls for the preview stage.`,
      tag: "gameplay",
      state: "queue",
    },
    {
      who: "Graphics",
      what: `Loading the ${template.scene} three.js scene into Play & Test.`,
      tag: "graphics",
      state: "queue",
    },
    {
      who: "QA",
      what: "Checking WebGL, fullscreen exit, and WASD input on the preview stage.",
      tag: "qa",
      state: "queue",
    },
  ];
  onUpdate(steps);

  steps = await tryChessAgentNote(prompt, onUpdate, steps);
  await delay(400);
  steps = setStepState(steps, 0, "done");
  steps = setStepState(steps, 1, "run");
  onUpdate(steps);

  await delay(500);
  steps = setStepState(steps, 1, "done");
  steps = setStepState(steps, 2, "run");
  onUpdate(steps);

  await delay(500);
  steps = setStepState(steps, 2, "done");
  steps = setStepState(steps, 3, "run");
  onUpdate(steps);

  await delay(350);
  steps = setStepState(steps, 3, "done");
  onUpdate(steps);

  return { template, steps };
}