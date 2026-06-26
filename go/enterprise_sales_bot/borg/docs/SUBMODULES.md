# Global Submodule & Reference Mapping

This document provides a comprehensive mapping of all git submodules and nested gitlinks within the Omni-Workspace. It is intended as a universal reference for AI agents and human operators to understand the project's dependency tree.

## 1. Root Workspace Submodules
These are the top-level projects integrated into the monorepo.

| Path | Description | Rationale |
| :--- | :--- | :--- |
| `submodules/hyperharness/` | TormentNexus CLI Harness | Experimental primary CLI harness assimilation track for TormentNexus. |
| `apps/maestro/` | electron-orchestrator desktop shell | Core desktop surface for multi-agent coordination. |
| `tormentnexus/` | Autonomous IDE | The central intelligence and supervision engine. |
| `submodules/tormentnexus/` | tormentnexus CLI Harness | Experimental primary CLI harness assimilation track for tormentnexus. |
| `apps/maestro/` | electron-orchestrator desktop shell | Core desktop surface for multi-agent coordination. |
| `tormentnexus/` | Autonomous IDE | The central intelligence and supervision engine. |
| `antigravity-autopilot/` | UI/Mobile Bridge | Autonomous UI testing and mobile control plane. |
| `bg/` | Game Engine Cluster | Core game logic and legacy references. |
| `bobmani/` | Rhythm Ecosystem | Unified platform for IIDX, DDR, and chart tools. |
| `topaz-ffmpeg/` | Video Processing | High-performance AI upscaling core. |
| `computer-use-preview/` | Gemini Computer Use | Autonomous UI interaction and desktop automation. |

## 2. Nested Libraries & References (`bg/bobsgameonlinejava`)
The Java-based legacy engine includes an exhaustive list of reference editors and libraries.

- **Core Libs:** `lwjgl3`, `micromod`, `lz4-java`, `jinput`.
- **Reference Editors:**
    - `aseprite`: Industry-standard pixel art editor.
    - `Pixelorama`: GDScript-based sprite editor.
    - `Tiled`: Map editor for 2D games.
    - `Blockbench`: 3D modeling for low-poly assets.
    - `OgmoEditor3-CE`: Generic level editor for procedural generation.

## 3. Low-Level C++ Dependencies (`bg/okgame/lib`)
The high-performance game engine relies on industrial-grade C++ infrastructure.

- **Standard Abstractions:** `boost` (full ecosystem), `poco`.
- **Media Core:** `SDL2`, `ffmpeg` (via topaz), `libwebp`, `libpng`.
- **Rhythm Components:** `MilkDrop3`, `paulxstretch`, `libxmp`, `SoundTouch`.
- **Graphics UI:** `imgui`, `raylib`, `Nuklear`, `nanogui`.

## 4. Automation & AI Utilities (`bobbybookmarks/submodules`)
Utilities for personal knowledge management and AI skill sets.

- `tormentnexus`: Meta-server for Model Context Protocol.
- `anthropics-skills`: Official skill sets for Claude.
- `openai-skills`: Official skill sets for GPT.
- `ykdojo-claude-code-tips`: Optimization patterns for agent interaction.

## 5. Directory Legend
- **`src/`**: Primary source code for the project.
- **`libs/`**: Internal or closely-coupled library dependencies.
- **`references/`**: Third-party projects used for architectural study or feature parity.
- **`vendor/`**: Third-party binaries or large static assets.
- **`research/`**: Experimental scratchpads and untracked projects.

---
*Note: This document is automatically synchronized during v1.6.3+ stabilization passes.*
