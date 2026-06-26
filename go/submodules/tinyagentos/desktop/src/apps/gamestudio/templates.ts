import type { Template, BuildStep } from "./types";

/* ------------------------------------------------------------------ */
/*  Starter templates + the illustrative build log                     */
/*                                                                     */
/*  Templates mirror the approved mock. Each carries a `scene` that the */
/*  Play view renders with real three.js. Phase 1 ships two interactive */
/*  scenes (runner, orbit); the others reuse the runner scene under     */
/*  their own title, so "Use template" always loads a live 3D preview.  */
/* ------------------------------------------------------------------ */

export const TEMPLATES: Template[] = [
  {
    id: "pixel-platformer",
    title: "Pixel Platformer",
    genre: "Platformer",
    desc: "Run, jump and stomp across handcrafted levels with springs and moving platforms.",
    cover:
      "radial-gradient(120% 120% at 30% 20%, #1f5a3a, transparent 60%), linear-gradient(140deg,#12261b,#0c1712)",
    scene: "runner",
  },
  {
    id: "neon-runner",
    title: "Neon Runner",
    genre: "Endless Runner",
    desc: "Auto-run through a synthwave city. Dodge, slide and chase the high score.",
    cover:
      "radial-gradient(120% 120% at 70% 25%, #3a4a7a, transparent 60%), linear-gradient(140deg,#16213a,#0c1120)",
    scene: "runner",
  },
  {
    id: "keep-defense",
    title: "Keep Defense",
    genre: "Tower Defense",
    desc: "Place towers along the path and hold the line against waves of raiders.",
    cover:
      "radial-gradient(120% 120% at 40% 30%, #16607a, transparent 60%), linear-gradient(140deg,#0e2230,#0a1620)",
    scene: "orbit",
  },
  {
    id: "arena-top-down",
    title: "Arena Top-Down",
    genre: "Top-down Shooter",
    desc: "Twin-stick survival in a tight arena with pickups and rising difficulty.",
    cover:
      "radial-gradient(120% 120% at 60% 25%, #2a3f7a, transparent 60%), linear-gradient(140deg,#141a2b,#0d1119)",
    scene: "orbit",
  },
  {
    id: "drift-circuit",
    title: "Drift Circuit",
    genre: "Racing",
    desc: "Three low-poly tracks, drift boost and ghost laps to beat.",
    cover:
      "radial-gradient(120% 120% at 35% 25%, #1f4d63, transparent 60%), linear-gradient(140deg,#10222a,#0b161b)",
    scene: "runner",
  },
  {
    id: "block-puzzle",
    title: "Block Puzzle",
    genre: "Puzzle",
    desc: "Rotate and drop pieces to clear lines. Endless and daily modes.",
    cover:
      "radial-gradient(120% 120% at 60% 25%, #5a3a1f, transparent 60%), linear-gradient(140deg,#231811,#16100a)",
    scene: "orbit",
  },
  {
    id: "corridor-fps",
    title: "Corridor FPS",
    genre: "FPS",
    desc: "A short first-person shooting gallery with three rooms and a boss.",
    cover:
      "radial-gradient(120% 120% at 40% 30%, #1f5a3a, transparent 60%), linear-gradient(140deg,#10261b,#0a1712)",
    scene: "runner",
  },
  {
    id: "orbit-tap",
    title: "Orbit Tap",
    genre: "XR / VR",
    desc: "A room-scale tap-the-orb warm-up built for headsets and hand tracking.",
    cover:
      "radial-gradient(120% 120% at 65% 25%, #2a3f7a, transparent 60%), linear-gradient(140deg,#141a2b,#0d1119)",
    scene: "orbit",
  },
];

/** The template the Play stage loads before the user picks one, so the
 *  3D preview always has a real scene to render. */
export const DEFAULT_TEMPLATE: Template = TEMPLATES[0]!;

/** Illustrative build trace shown beside the Play stage. Static in phase 1:
 *  it pictures the future agent pipeline (a director splitting a prompt across
 *  specialist blocks), and is clearly labelled as a preview, not a live run. */
export const BUILD_LOG: BuildStep[] = [
  {
    who: "Director",
    what: "Reads the prompt, picks a template and splits the work across five specialist blocks.",
    tag: "routing",
    state: "done",
    director: true,
  },
  {
    who: "Gameplay",
    what: "Wires the run loop, jump and lane swap, following the movement guide for the genre.",
    tag: "gameplay",
    state: "done",
  },
  {
    who: "Graphics",
    what: "Places low-poly geometry, materials and lighting in the scene you can preview here.",
    tag: "graphics",
    state: "run",
  },
  {
    who: "UI",
    what: "Score, combo meter, a pause sheet and a game-over card.",
    tag: "ui",
    state: "queue",
  },
  {
    who: "Audio",
    what: "Footsteps, pickups and a loop from the free pack.",
    tag: "audio",
    state: "queue",
  },
  {
    who: "QA",
    what: "Runs the game on desktop, mobile and fullscreen and checks it holds frame rate.",
    tag: "qa",
    state: "queue",
  },
];
