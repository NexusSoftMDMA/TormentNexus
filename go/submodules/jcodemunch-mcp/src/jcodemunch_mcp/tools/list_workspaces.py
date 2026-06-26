"""list_workspaces tool: detect monorepo workspaces inside an indexed repo.

Modern repos increasingly ship as monorepos — one git root, many independent
packages with their own tech stacks. Agents asking about a single package
need to see *just that package*, not repo-wide aggregates. This tool
enumerates the workspace members so the agent can scope follow-up queries.

Detected layouts:
  * **pnpm**     — pnpm-workspace.yaml `packages:` globs
  * **yarn/npm** — package.json `workspaces:` (list or {packages: [...]} form)
  * **turborepo** — turbo.json piggybacks on the npm/yarn/pnpm workspace list
  * **lerna**    — lerna.json `packages:` globs (typically same as npm form)
  * **rush**     — rush.json `projects: [{packageName, projectFolder}, ...]`
  * **go**       — go.work `use ( ... )` directive
  * **cargo**    — Cargo.toml `[workspace] members = [...]` globs

Each member is reported as
``{path, package_name, manager}`` where ``path`` is repo-root-relative
(forward-slash) and ``package_name`` is the value from the member's
manifest (package.json `name`, Cargo.toml `[package] name`, module name
from `go.mod`). When multiple managers cover the same path (turborepo +
pnpm is common), the more-specific manager wins.

Read-only. No filesystem mutations. No network. Bounded by a per-manager
glob cap (default 2,000 matches) so a pathological monorepo can't OOM.
"""

from __future__ import annotations

import fnmatch
import json
import logging
import os
import re
import time
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)

_MAX_MEMBERS_PER_MANAGER = 2_000


# --------------------------------------------------------------------------- #
# Glob expansion (subset — no `**` recursion, just leading/trailing wildcards) #
# --------------------------------------------------------------------------- #

def _expand_glob(root: Path, pattern: str) -> list[Path]:
    """Expand a workspace glob relative to ``root``.

    Supports the patterns workspace managers actually emit:
      * ``packages/*`` — one level of children
      * ``packages/**`` — recursive (treated like `packages/**/`)
      * ``packages/foo`` — literal path
      * ``apps/web``    — literal path
      * negated patterns starting with `!` are filtered by caller

    Returns directory matches only.
    """
    if not pattern or pattern.startswith("!"):
        return []

    # Treat "packages/**" same as "packages/**/" → recursive directory walk.
    p = pattern.rstrip("/")
    out: list[Path] = []

    if "**" in p:
        prefix = p.split("**", 1)[0].rstrip("/")
        base = (root / prefix) if prefix else root
        if base.is_dir():
            for sub in base.rglob("*"):
                if sub.is_dir():
                    out.append(sub)
        return out

    if "*" in p or "?" in p:
        parts = p.split("/")
        # Resolve the literal prefix
        bases: list[Path] = [root]
        for part in parts:
            if "*" in part or "?" in part:
                next_bases: list[Path] = []
                for b in bases:
                    if not b.is_dir():
                        continue
                    try:
                        children = list(b.iterdir())
                    except OSError:
                        continue
                    for c in children:
                        if c.is_dir() and fnmatch.fnmatch(c.name, part):
                            next_bases.append(c)
                bases = next_bases
            else:
                bases = [b / part for b in bases if (b / part).exists()]
        out = [b for b in bases if b.is_dir()]
        return out

    candidate = (root / p)
    if candidate.is_dir():
        return [candidate]
    return []


def _apply_glob_set(root: Path, patterns: list[str]) -> list[Path]:
    """Apply a list of include / negate patterns. Returns deduplicated dirs."""
    include: list[Path] = []
    negate_patterns: list[str] = []
    for pat in patterns:
        if isinstance(pat, str) and pat.startswith("!"):
            negate_patterns.append(pat[1:])
        elif isinstance(pat, str):
            include.extend(_expand_glob(root, pat))

    if not negate_patterns:
        dirs = include
    else:
        dirs = []
        for d in include:
            rel = d.relative_to(root).as_posix()
            if not any(fnmatch.fnmatch(rel, np) for np in negate_patterns):
                dirs.append(d)

    seen: set = set()
    deduped: list[Path] = []
    for d in dirs:
        key = d.resolve()
        if key not in seen:
            seen.add(key)
            deduped.append(d)
        if len(deduped) >= _MAX_MEMBERS_PER_MANAGER:
            break
    return deduped


# --------------------------------------------------------------------------- #
# Manifest readers                                                            #
# --------------------------------------------------------------------------- #

def _read_json(path: Path) -> Optional[dict]:
    try:
        with path.open("r", encoding="utf-8", errors="replace") as f:
            return json.load(f)
    except (OSError, json.JSONDecodeError):
        logger.debug("Could not parse JSON: %s", path, exc_info=True)
        return None


_RE_TOML_TABLE = re.compile(r"^\[([^\]]+)\]\s*$")
_RE_TOML_KEY = re.compile(r"^([A-Za-z_][\w-]*)\s*=\s*(.+?)\s*$")
_RE_TOML_INLINE_LIST = re.compile(r"\[(.*?)\]", re.DOTALL)


def _read_cargo_workspace_members(path: Path) -> tuple[list[str], Optional[str]]:
    """Pull ``[workspace] members = [...]`` and ``[package] name`` from a Cargo.toml.

    Tiny hand-rolled TOML reader — Cargo.toml workspace files are simple and
    avoiding an optional ``tomli`` import keeps this tool dependency-free.
    """
    members: list[str] = []
    pkg_name: Optional[str] = None
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return members, pkg_name

    # Find `members = [...]` within [workspace]
    in_workspace = False
    in_package = False
    buf: list[str] = []
    members_open = False
    name_open = False
    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("#"):
            continue
        m = _RE_TOML_TABLE.match(stripped)
        if m:
            tbl = m.group(1).strip()
            in_workspace = tbl == "workspace"
            in_package = tbl == "package"
            members_open = False
            name_open = False
            continue

        if in_workspace and (stripped.startswith("members") or members_open):
            # Accumulate until we close the bracket
            buf.append(stripped)
            joined = " ".join(buf)
            if "[" in joined and "]" in joined:
                inside = _RE_TOML_INLINE_LIST.search(joined)
                if inside:
                    raw_items = inside.group(1)
                    for item in raw_items.split(","):
                        item = item.strip().strip('"').strip("'")
                        if item:
                            members.append(item)
                buf = []
                members_open = False
            else:
                members_open = "[" in joined
            continue

        if in_package and stripped.startswith("name"):
            km = _RE_TOML_KEY.match(stripped)
            if km:
                val = km.group(2).strip()
                if val.startswith('"') and val.endswith('"'):
                    pkg_name = val[1:-1]
                elif val.startswith("'") and val.endswith("'"):
                    pkg_name = val[1:-1]
                else:
                    pkg_name = val
    return members, pkg_name


def _read_go_work_use(path: Path) -> list[str]:
    """Pull paths from a go.work `use ( ... )` block (or a single `use ./foo`)."""
    out: list[str] = []
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return out

    # Multi-line block: `use (` … `)`
    block = re.search(r"use\s*\(\s*([\s\S]*?)\s*\)", text)
    if block:
        for line in block.group(1).splitlines():
            line = line.strip()
            if not line or line.startswith("//"):
                continue
            # Lines may be quoted or bare
            if line.startswith('"') and line.endswith('"'):
                line = line[1:-1]
            out.append(line)
    # Single-line `use ./module`
    for m in re.finditer(r"^\s*use\s+([^\s(].*)$", text, re.MULTILINE):
        val = m.group(1).strip()
        if val.startswith('"') and val.endswith('"'):
            val = val[1:-1]
        out.append(val)
    return out


def _read_go_mod_name(path: Path) -> Optional[str]:
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return None
    m = re.search(r"^\s*module\s+(\S+)\s*$", text, re.MULTILINE)
    return m.group(1).strip() if m else None


def _read_pnpm_workspace(path: Path) -> list[str]:
    """Minimal YAML reader for pnpm-workspace.yaml `packages:` list."""
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return []
    packages: list[str] = []
    in_packages = False
    for line in text.splitlines():
        stripped = line.rstrip()
        if not stripped or stripped.lstrip().startswith("#"):
            continue
        # Top-level `packages:`
        if stripped == "packages:" or stripped.startswith("packages:"):
            in_packages = True
            # Inline-form `packages: [a, b]` is rare; handle just in case
            after = stripped.split(":", 1)[1].strip()
            if after.startswith("[") and after.endswith("]"):
                for item in after[1:-1].split(","):
                    item = item.strip().strip('"').strip("'")
                    if item:
                        packages.append(item)
                in_packages = False
            continue
        if in_packages:
            ls = line.lstrip()
            if line and not line[0].isspace():
                # New top-level key — packages list ended
                in_packages = False
                continue
            if ls.startswith("- "):
                val = ls[2:].strip().strip('"').strip("'")
                if val:
                    packages.append(val)
    return packages


# --------------------------------------------------------------------------- #
# Per-manager discovery                                                       #
# --------------------------------------------------------------------------- #

def _discover_pnpm(root: Path) -> list[dict]:
    cfg = root / "pnpm-workspace.yaml"
    if not cfg.exists():
        return []
    patterns = _read_pnpm_workspace(cfg)
    members: list[dict] = []
    for d in _apply_glob_set(root, patterns):
        pj = d / "package.json"
        if pj.exists():
            data = _read_json(pj) or {}
            members.append({
                "path": d.relative_to(root).as_posix(),
                "package_name": data.get("name") or "",
                "manager": "pnpm",
            })
    return members


def _discover_npm_yarn(root: Path) -> list[dict]:
    pj = root / "package.json"
    if not pj.exists():
        return []
    data = _read_json(pj) or {}
    ws = data.get("workspaces")
    patterns: list[str] = []
    if isinstance(ws, list):
        patterns = [p for p in ws if isinstance(p, str)]
    elif isinstance(ws, dict):
        if isinstance(ws.get("packages"), list):
            patterns = [p for p in ws["packages"] if isinstance(p, str)]
    if not patterns:
        return []

    # Pick a manager label — lockfile sniff with sensible fallback
    if (root / "yarn.lock").exists():
        manager = "yarn"
    elif (root / "package-lock.json").exists():
        manager = "npm"
    else:
        manager = "npm"

    members: list[dict] = []
    for d in _apply_glob_set(root, patterns):
        sub_pj = d / "package.json"
        if sub_pj.exists():
            sub = _read_json(sub_pj) or {}
            members.append({
                "path": d.relative_to(root).as_posix(),
                "package_name": sub.get("name") or "",
                "manager": manager,
            })
    return members


def _discover_lerna(root: Path) -> list[dict]:
    cfg = root / "lerna.json"
    if not cfg.exists():
        return []
    data = _read_json(cfg) or {}
    patterns = data.get("packages") or ["packages/*"]
    members: list[dict] = []
    for d in _apply_glob_set(root, patterns):
        pj = d / "package.json"
        if pj.exists():
            sub = _read_json(pj) or {}
            members.append({
                "path": d.relative_to(root).as_posix(),
                "package_name": sub.get("name") or "",
                "manager": "lerna",
            })
    return members


def _discover_rush(root: Path) -> list[dict]:
    cfg = root / "rush.json"
    if not cfg.exists():
        return []
    data = _read_json(cfg) or {}
    members: list[dict] = []
    for proj in data.get("projects") or []:
        if not isinstance(proj, dict):
            continue
        folder = proj.get("projectFolder") or proj.get("project_folder")
        name = proj.get("packageName") or proj.get("package_name") or ""
        if folder and (root / folder).is_dir():
            members.append({
                "path": Path(folder).as_posix(),
                "package_name": name,
                "manager": "rush",
            })
    return members[:_MAX_MEMBERS_PER_MANAGER]


def _discover_turborepo(root: Path) -> bool:
    """Just signals turbo presence — the actual members come from the underlying
    pnpm/yarn/npm workspace config. Return True when turbo.json is present."""
    return (root / "turbo.json").exists()


def _discover_cargo(root: Path) -> list[dict]:
    cargo = root / "Cargo.toml"
    if not cargo.exists():
        return []
    members_patterns, _ = _read_cargo_workspace_members(cargo)
    if not members_patterns:
        return []
    members: list[dict] = []
    for d in _apply_glob_set(root, members_patterns):
        sub_cargo = d / "Cargo.toml"
        if sub_cargo.exists():
            _, name = _read_cargo_workspace_members(sub_cargo)
            members.append({
                "path": d.relative_to(root).as_posix(),
                "package_name": name or "",
                "manager": "cargo",
            })
    return members


def _discover_go_work(root: Path) -> list[dict]:
    gw = root / "go.work"
    if not gw.exists():
        return []
    members: list[dict] = []
    for entry in _read_go_work_use(gw):
        # Entries are paths relative to root (typically `./pkg/foo`)
        rel = entry.lstrip("./").rstrip("/") or "."
        candidate = (root / rel).resolve()
        if not candidate.is_dir():
            continue
        try:
            candidate.relative_to(root.resolve())
        except ValueError:
            continue
        gm = candidate / "go.mod"
        mod_name = _read_go_mod_name(gm) if gm.exists() else None
        members.append({
            "path": candidate.relative_to(root).as_posix(),
            "package_name": mod_name or "",
            "manager": "go",
        })
        if len(members) >= _MAX_MEMBERS_PER_MANAGER:
            break
    return members


# --------------------------------------------------------------------------- #
# Public entry point                                                          #
# --------------------------------------------------------------------------- #

def list_workspaces(
    repo: str,
    storage_path: Optional[str] = None,
) -> dict:
    """Enumerate monorepo workspace members for an indexed repo.

    Returns ``{"result": {"workspaces": [{path, package_name, manager}, ...],
    "managers": [...], "is_monorepo": bool, "source_root": "..."}, "_meta": {...}}``.

    When the same path is claimed by multiple managers (turborepo over pnpm
    is the common case), the deeper / package-flavored manager wins. The
    ``managers`` field surfaces every manager that contributed, so agents can
    detect e.g. "turborepo + pnpm" stacks without parsing the member list.
    """
    t0 = time.monotonic()

    from ._utils import resolve_repo as _resolve_repo
    from ..storage import IndexStore

    try:
        owner, name = _resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if index is None:
        return {"error": f"No index found for {repo}"}

    source_root = getattr(index, "source_root", "")
    if not source_root or not os.path.isdir(source_root):
        return {
            "error": "list_workspaces requires a local index with source_root. "
                     "Use index_folder (not index_repo) to create one."
        }

    root = Path(source_root).resolve()

    # Detect each manager in turn. Order matters for the dedup pass:
    # more-specific managers should overwrite generic ones. We surface every
    # manager that contributed via `managers`, so the choice of "winning"
    # manager per path doesn't lose information.
    discovery = [
        ("rush", _discover_rush(root)),
        ("cargo", _discover_cargo(root)),
        ("go", _discover_go_work(root)),
        ("lerna", _discover_lerna(root)),
        ("pnpm", _discover_pnpm(root)),
        ("npm_yarn", _discover_npm_yarn(root)),
    ]

    seen_paths: dict = {}
    managers_seen: list[str] = []
    for label, members in discovery:
        if not members:
            continue
        managers_seen.append(label)
        for m in members:
            key = m["path"]
            # If a more-specific manager already claimed this path, keep theirs.
            if key not in seen_paths:
                seen_paths[key] = m

    if _discover_turborepo(root):
        managers_seen.append("turborepo")

    workspaces = sorted(seen_paths.values(), key=lambda m: m["path"])

    return {
        "result": {
            "workspaces": workspaces,
            "managers": sorted(set(managers_seen)),
            "is_monorepo": bool(workspaces) or bool(managers_seen),
            "source_root": str(root),
            "member_count": len(workspaces),
        },
        "_meta": {
            "timing_ms": round((time.monotonic() - t0) * 1000, 1),
        },
    }
