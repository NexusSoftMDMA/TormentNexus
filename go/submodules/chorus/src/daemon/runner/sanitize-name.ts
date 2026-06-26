/** tmux session names accept [a-zA-Z0-9_-]; drop everything else and clamp length. */
export function sanitizeName(name: string): string {
  return name.replace(/[^a-zA-Z0-9_-]+/g, '_').slice(0, 80);
}
