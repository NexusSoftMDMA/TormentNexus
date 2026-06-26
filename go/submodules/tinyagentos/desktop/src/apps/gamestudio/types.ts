/* ------------------------------------------------------------------ */
/*  Game Studio — shared types                                         */
/*                                                                     */
/*  Phase 1 ships the app shell + a real three.js preview. AI          */
/*  generation, the skill pack and asset backends arrive in later      */
/*  phases; the types here describe the shell, the demo scenes and     */
/*  the honest "later phase" affordances, not a generation pipeline.   */
/* ------------------------------------------------------------------ */

export type StudioView = "create" | "play" | "share";

/** Device preview the Play stage emulates. XR is a labelled affordance in
 *  phase 1 (real WebXR is a later phase); Desktop/Mobile resize the stage. */
export type DevicePreview = "desktop" | "mobile" | "xr";

/** A demo scene the three.js preview can load. Each template maps to one.
 *  Phase 1 ships two genuinely interactive scenes (runner, orbit); the rest
 *  fall back to the runner scene with their own label, so "Use template"
 *  always loads something real rather than faking a bespoke build. */
export type SceneKind = "runner" | "orbit";

/** Genre chips shown on the Create view. Multi-select, illustrative; they do
 *  not drive generation in phase 1. */
export const GENRES = [
  "Platformer",
  "Endless Runner",
  "Tower Defense",
  "Top-down Shooter",
  "Racing",
  "Puzzle",
  "FPS",
  "XR / VR Experience",
] as const;

export type Genre = (typeof GENRES)[number];

/** Offline model selector options. Labelled-but-inert in phase 1: offline
 *  generation is a later phase, so picking a model changes the label only. */
export const OFFLINE_MODELS = [
  "Gemma 4 E4B (offline)",
  "Qwen3 4B (offline)",
  "Llama 3.2 3B (offline)",
] as const;

export const ART_STYLES = [
  "Low-poly 3D",
  "Pixel art",
  "Flat vector",
  "Hand-drawn",
] as const;

export const DIFFICULTIES = ["Casual", "Normal", "Hard"] as const;

/** A starter template. Selecting one loads its demo scene into the Play view. */
export interface Template {
  id: string;
  title: string;
  genre: string;
  desc: string;
  /** CSS background for the card cover + the publish cover. */
  cover: string;
  /** Which real three.js scene the Play view loads for this template. */
  scene: SceneKind;
}

/** A build-log entry. Static/illustrative in phase 1: it represents the
 *  future agent build trace (a director routes a prompt across specialist
 *  blocks), not a live pipeline. */
export interface BuildStep {
  who: string;
  what: string;
  tag: string;
  state: "done" | "run" | "queue";
  director?: boolean;
}

/** Visibility for the Share view publish card. */
export type Visibility = "private" | "community";
