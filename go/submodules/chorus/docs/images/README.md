# README image assets

This directory holds the visual assets referenced from the top-level [`README.md`](../../README.md). Drop the files listed below in here and they'll render automatically on GitHub.

## Required assets

| Filename | Type | Recommended size | What it shows |
|---|---|---|---|
| `logo.png` | static | 240×240 (square, transparent bg) | Chorus mark, top of README |
| `hero-demo.gif` | animated | 1280×720, ≤8 MB, ≤15s loop | Full run: submit task → 3 reviewers stream → verdict |
| `run-page.gif` | animated | 800×500, ≤4 MB, ≤10s | Live cockpit run page with three reviewers streaming |
| `verdict.gif` | animated | 800×500, ≤4 MB, ≤6s | Final verdict + diff + cost chips |
| `templates.gif` | animated | 800×500, ≤4 MB, ≤8s | Drag-and-drop template editor with voice picker |
| `mcp.gif` | animated | 800×500, ≤4 MB, ≤8s | External CLI (e.g. Claude Code) calling `mcp__chorus__create_chat` |

## Static fallbacks (optional but recommended)

Some renderers (npmjs.com, RSS, archive.org) don't autoplay GIFs. Provide a `.png` companion for each `.gif` so they degrade cleanly:

- `hero-demo.png` (first frame of `hero-demo.gif`)
- `run-page.png`, `verdict.png`, `templates.png`, `mcp.png`

The README links point to `.gif` only — npm renders the static first frame as a still, which is fine.

## Capture conventions

- **Source resolution**: capture at 2× target size (e.g. 2560×1440 for a 1280×720 GIF) then downsample for sharp text.
- **Frame rate**: 12–15 fps is plenty for terminal/UI captures; keeps file size down.
- **Loop**: infinite, no fade-in/fade-out.
- **Theme**: dark mode (cockpit default — Linear/Raycast aesthetic).
- **Cursor**: hide the OS cursor unless it's load-bearing.
- **Tooling**: [Kap](https://getkap.co), [LICEcap](https://www.cockos.com/licecap/), or `ffmpeg` (`ffmpeg -i input.mov -vf "fps=15,scale=1280:-1:flags=lanczos" -loop 0 output.gif`).

## Optimisation

Run every GIF through [`gifsicle`](https://www.lcdf.org/gifsicle/) before committing:

```bash
gifsicle -O3 --colors 128 --lossy=80 input.gif -o optimized.gif
```

Target: under 4 MB per asset (8 MB hard ceiling for the hero). GitHub renders larger files but mobile + slow connections suffer.

## Adding new assets

If you add an image, also update this table and reference it from the README. Don't leave orphan files in this directory.

## Placeholder status

Until real captures are added, GitHub will show broken-image icons for each `<img>` tag in the README. That's intentional — they're load-bearing reminders to record the GIF. Don't commit transparent 1×1 PNGs as placeholders; the broken icon is more useful than a silent fake.
