/**
 * Requests ground-truth generator (sverklo-bench#1 — Python dataset).
 *
 * `psf/requests` 2.32.3 is the first Python codebase on the bench.
 * Modular package layout under `src/requests/` with ~6K LOC, stable
 * public API across patch releases. Closes the long-standing gap in
 * the bench's language coverage (express, lodash, sverklo were all
 * JavaScript/TypeScript).
 *
 * Methodology:
 *
 *   P1 (definition lookup): 10 well-known public symbols with hand-
 *     verified line numbers at v2.32.3:
 *       8 top-level functions in src/requests/api.py (the public HTTP
 *         verbs: get, post, put, delete, patch, head, options,
 *         request) plus
 *       2 classes (Session in sessions.py, Response in models.py)
 *     If upstream renames or moves these we'll need to re-verify.
 *     Verification command:
 *       grep -nE "^def <name>\(|^class <name>\b" src/requests/<file>
 *
 *   P2 (reference finding): 10 of the same names, references resolved
 *     at run-time via grep against `src/requests/`. Each symbol is
 *     called/used internally many times — stable expected sets.
 *
 *   P4 (file dependencies): 5 files exercising real Python import
 *     patterns (relative `from .X import` and `from .X.Y import`).
 *     Picked to span the layered architecture: api → sessions →
 *     adapters → models, plus exceptions as a leaf module.
 *
 *   P5 (dead code): empty-expected pattern, same as express + lodash.
 *     Score punishes false positives. Tests baselines' ability to
 *     correctly identify that public API symbols are NOT dead.
 *
 * Adding a Python dataset exercises:
 *   - sverklo's parsePython() (regex parser; tree-sitter Python
 *     fallback if SVERKLO_PARSER=tree-sitter is set)
 *   - jcodemunch-mcp's Python tree-sitter symbol indexing
 *   - GitNexus's KuzuDB schema for Python imports
 *   - smart-grep's Python-specific definition patterns (def NAME(,
 *     class NAME)
 */

import { execSync } from "node:child_process";
import { readFileSync, existsSync } from "node:fs";
import { join } from "node:path";
import type { Task, Location } from "../../types.ts";

// Hand-verified line numbers in src/requests/ for v2.32.3.
// Verified 2026-05-07 against
// https://github.com/psf/requests/tree/v2.32.3 — re-run grep if
// upstream bumps the patch version (these have been stable for years
// but `Session.send` reorganization in 2.x lines did move neighbors).
const P1_DEFINITIONS: { name: string; file: string; line: number }[] = [
  { name: "get", file: "src/requests/api.py", line: 62 },
  { name: "options", file: "src/requests/api.py", line: 76 },
  { name: "head", file: "src/requests/api.py", line: 88 },
  { name: "post", file: "src/requests/api.py", line: 103 },
  { name: "put", file: "src/requests/api.py", line: 118 },
  { name: "patch", file: "src/requests/api.py", line: 133 },
  { name: "delete", file: "src/requests/api.py", line: 148 },
  { name: "request", file: "src/requests/api.py", line: 14 },
  { name: "Session", file: "src/requests/sessions.py", line: 356 },
  { name: "Response", file: "src/requests/models.py", line: 640 },
];

const P2_SYMBOLS: string[] = [
  "get",
  "options",
  "head",
  "post",
  "put",
  "patch",
  "delete",
  "request",
  "Session",
  "Response",
];

// 5 files exercising different shapes of the requests package.
const P4_FILES: string[] = [
  "src/requests/api.py",        // top of the call stack: imports sessions only
  "src/requests/sessions.py",   // imports many: adapters, auth, cookies, exceptions, hooks, models, status_codes
  "src/requests/models.py",     // imports many; importers: __init__, adapters, sessions
  "src/requests/adapters.py",   // imports models, exceptions, structures, utils
  "src/requests/exceptions.py", // leaf module; importers: __init__, adapters, sessions, models, utils
];

export function generateRequestsTasks(rootPath: string): Task[] {
  if (!existsSync(rootPath)) {
    throw new Error(`requests checkout missing: ${rootPath}`);
  }
  const tasks: Task[] = [];

  // ───── P1 ─────
  for (let i = 0; i < P1_DEFINITIONS.length; i++) {
    const { name, file, line } = P1_DEFINITIONS[i];
    tasks.push({
      id: `rq-p1-${pad(i + 1)}`,
      category: "P1",
      dataset: "requests",
      query: name,
      expected: { kind: "locations", locations: [{ file, line }] },
    });
  }

  // ───── P2 ─────
  for (let i = 0; i < P2_SYMBOLS.length; i++) {
    const name = P2_SYMBOLS[i];
    const refs = findReferencesInRequests(rootPath, name);
    tasks.push({
      id: `rq-p2-${pad(i + 1)}`,
      category: "P2",
      dataset: "requests",
      query: name,
      expected: { kind: "locations", locations: refs },
    });
  }

  // ───── P4 ─────
  for (let i = 0; i < P4_FILES.length; i++) {
    const file = P4_FILES[i];
    const abs = join(rootPath, file);
    let imports: string[] = [];
    let importers: string[] = [];
    try {
      const content = readFileSync(abs, "utf-8");
      imports = extractPythonImports(content, file);
    } catch {}
    importers = findImportersInRequests(rootPath, file);
    tasks.push({
      id: `rq-p4-${pad(i + 1)}`,
      category: "P4",
      dataset: "requests",
      query: file,
      expected: { kind: "deps", imports, importers },
    });
  }

  // ───── P5 ─────
  // Empty-expected — score punishes false positives. Tests whether
  // baselines correctly identify that public API symbols (`get`,
  // `Session`, etc.) are NOT dead.
  for (let i = 0; i < 5; i++) {
    tasks.push({
      id: `rq-p5-${pad(i + 1)}`,
      category: "P5",
      dataset: "requests",
      query: "",
      expected: { kind: "names", names: [] },
    });
  }

  return tasks;
}

function pad(n: number): string {
  return n.toString().padStart(2, "0");
}

/**
 * Extract relative imports from a Python file. Conservative: handles
 * `from .X import Y` and `from .X.Y import Z` (the two forms requests
 * uses). Resolves to a path string under src/requests/ for the bench's
 * file-dep contract. Drops `from .` (package-root re-imports) and
 * `from typing` / external imports.
 */
function extractPythonImports(content: string, file: string): string[] {
  const out = new Set<string>();
  // Compute the package directory (src/requests/) for resolution.
  const fileDir = file.replace(/\/[^/]+$/, "");
  const re = /^from\s+(\.[\w.]*)\s+import/gm;
  let m;
  while ((m = re.exec(content)) !== null) {
    const dotted = m[1]; // e.g., ".adapters", ".compat"
    if (dotted === ".") continue; // bare `from . import X` — package re-export
    // Strip leading dot, convert dots to slashes, append .py
    const rel = dotted.slice(1).replace(/\./g, "/");
    if (!rel) continue;
    const resolved = `${fileDir}/${rel}.py`;
    out.add(resolved);
  }
  return [...out];
}

function findReferencesInRequests(root: string, name: string): Location[] {
  // Word-grep within src/requests/ + tests/, drop the definition line
  // itself (def NAME, class NAME). Mirrors lodash.gen.ts pattern.
  const out = grep(root, `\\b${escapeRe(name)}\\b`, ["src/requests"]);
  const defRe = new RegExp(
    `^\\s*(def|class)\\s+${escapeRe(name)}\\b`,
  );
  return out
    .filter((h) => !defRe.test(h.snippet ?? ""))
    .map((h) => ({ file: h.file, line: h.line }));
}

function findImportersInRequests(root: string, file: string): string[] {
  // Python: find `from .module import ...` or `from .pkg.module import ...`
  // referencing this file. Module name is filename without .py.
  const baseName = file.replace(/^.*\//, "").replace(/\.py$/, "");
  if (!baseName) return [];
  const out = new Set<string>();
  try {
    // Match `from .baseName ` and `from ..baseName ` (relative imports
    // that name this module). Doesn't catch `from . import baseName`
    // (package-root re-imports) — those are package wiring, not
    // load-bearing dependencies, and would inflate counts unhelpfully.
    const cmd = `grep -rln --include='*.py' --exclude-dir=__pycache__ ${shellQuote(`from \\.\\.\\?${baseName} import\\|from \\.\\.\\?[a-zA-Z_]\\+\\.${baseName} import`)} src/requests/ 2>/dev/null || true`;
    const result = execSync(cmd, {
      cwd: root,
      encoding: "utf-8",
      timeout: 30000,
      maxBuffer: 5 * 1024 * 1024,
      shell: "/bin/bash",
    });
    for (const line of result.split("\n")) {
      const p = line.trim();
      if (!p || p === file) continue;
      out.add(p);
    }
  } catch {}
  return [...out];
}

interface Hit {
  file: string;
  line: number;
  snippet?: string;
}

function grep(root: string, pattern: string, paths: string[]): Hit[] {
  const hits: Hit[] = [];
  for (const p of paths) {
    const target = join(root, p);
    if (!existsSync(target)) continue;
    try {
      const out = execSync(
        `grep -rnE --include='*.py' --exclude-dir=__pycache__ ${shellQuote(pattern)} ${shellQuote(p)} 2>/dev/null || true`,
        {
          cwd: root,
          encoding: "utf-8",
          timeout: 60000,
          maxBuffer: 20 * 1024 * 1024,
          shell: "/bin/bash",
        },
      );
      for (const line of out.split("\n")) {
        if (!line) continue;
        const m = line.match(/^(.+?):(\d+):(.*)$/);
        if (m) hits.push({ file: m[1], line: parseInt(m[2], 10), snippet: m[3] });
      }
    } catch {}
  }
  return hits;
}

function shellQuote(s: string): string {
  return `'${s.replace(/'/g, `'\\''`)}'`;
}
function escapeRe(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
