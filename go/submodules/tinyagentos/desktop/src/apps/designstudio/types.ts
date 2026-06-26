/* ------------------------------------------------------------------ */
/*  Design Studio -- shared types                                      */
/* ------------------------------------------------------------------ */

export type DesignStudioView = "design" | "templates" | "elements" | "magic";

export interface GeneratedImage {
  id: string;
  url: string;
  prompt: string;
}

/** An image layer placed on the design canvas artboard. */
export interface CanvasImageElement {
  id: string;
  type: "image";
  url: string;
  prompt: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export type CanvasElement = CanvasImageElement;

export const MAGIC_STYLE_CHIPS = [
  "Bold",
  "Minimal",
  "Editorial",
  "Playful",
  "Dark",
  "Corporate",
] as const;