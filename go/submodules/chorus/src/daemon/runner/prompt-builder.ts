/**
 * Pure prompt-construction helpers.
 *
 * Three functions that take phase config + user inputs and return the
 * ask.md text the runner pastes into the doer/reviewer CLIs. No fs writes,
 * no subprocess — just string assembly + (for packAttachedFiles) read-only
 * filesystem inspection that's exercised through tests against tmp dirs.
 *
 * Extracted out of runner.ts so the streaming hot paths can be split later
 * without breaking these contracts.
 */
import * as fs from 'fs';
import * as path from 'path';
import type { Phase } from '../../lib/template-schema.js';

// Per-file cap and total cap when inlining attached files into a prompt.
// Numbers chosen to keep prompts comfortably within Anthropic / OpenAI / Google
// 1M-token budgets while still surfacing realistic source files. Hardcoded
// for now; if template authors need larger payloads we'd lift these into
// template config (template.inputs.maxFileBytes / maxTotalBytes).
const ATTACHED_FILE_MAX_BYTES = 64 * 1024;
const ATTACHED_FILES_TOTAL_BYTES = 256 * 1024;

/**
 * Inline the contents of user-attached files into a single markdown block
 * the doer/reviewer can read directly. Drops files that:
 *   - traverse out of repoPath/cwd via `..` (security)
 *   - are symlinks (TOCTOU defence)
 *   - aren't regular files (sockets, fifos, etc.)
 *   - don't exist
 *   - would blow past the 256KB total cap
 *
 * Each surviving file is fenced as a markdown code block with its
 * extension as the language hint.
 */
export function packAttachedFiles(
  paths: string[] | undefined,
  repoPath: string | undefined,
): string {
  if (!paths || paths.length === 0) return '';

  const chunks: string[] = [];
  let totalBytes = 0;

  const cwdRoot = path.resolve(repoPath ?? process.cwd());

  for (const rel of paths) {
    if (path.isAbsolute(rel)) {
      chunks.push(`### \`${rel}\` — _absolute path rejected, skipping_`);
      continue;
    }

    const abs = path.resolve(path.join(cwdRoot, rel));
    const display = rel;

    if (!abs.startsWith(cwdRoot + path.sep) && abs !== cwdRoot) {
      chunks.push(`### \`${display}\` — _path traversal rejected, skipping_`);
      continue;
    }

    if (!fs.existsSync(abs)) {
      chunks.push(`### \`${display}\` — _file not found, skipping_`);
      continue;
    }

    let body: string;
    try {
      let fd = -1;
      try {
        // O_NOFOLLOW on Linux/macOS fails with ELOOP if path is a symlink.
        // On Windows, O_NOFOLLOW is unsupported; fall back to lstat+read.
        if (process.platform !== 'win32') {
          try {
            fd = fs.openSync(abs, fs.constants.O_RDONLY | fs.constants.O_NOFOLLOW);
          } catch (openErr) {
            // ELOOP = symlink detected via O_NOFOLLOW
            if (openErr instanceof Error && openErr.message.includes('ELOOP')) {
              chunks.push(`### \`${display}\` — _symlink rejected, skipping_`);
              continue;
            }
            throw openErr;
          }
          const stat = fs.fstatSync(fd);
          if (!stat.isFile()) {
            chunks.push(`### \`${display}\` — _not a regular file, skipping_`);
            continue;
          }
          body = fs.readFileSync(abs, 'utf-8');
        } else {
          // Windows fallback: lstat + read (not race-proof but best effort)
          const lstat = fs.lstatSync(abs);
          if (lstat.isSymbolicLink()) {
            chunks.push(`### \`${display}\` — _symlink rejected, skipping_`);
            continue;
          }
          if (!lstat.isFile()) {
            chunks.push(`### \`${display}\` — _not a regular file, skipping_`);
            continue;
          }
          body = fs.readFileSync(abs, 'utf-8');
        }
      } finally {
        if (fd >= 0) fs.closeSync(fd);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      chunks.push(`### \`${display}\` — _read error: ${msg}_`);
      continue;
    }

    const truncated = body.length > ATTACHED_FILE_MAX_BYTES;
    const slice = truncated ? body.slice(0, ATTACHED_FILE_MAX_BYTES) : body;
    const remainingBudget = ATTACHED_FILES_TOTAL_BYTES - totalBytes;

    if (slice.length > remainingBudget) {
      chunks.push(
        `### \`${display}\` — _skipped: would exceed ${ATTACHED_FILES_TOTAL_BYTES}-byte total cap_`,
      );
      continue;
    }

    totalBytes += slice.length;
    const ext = path.extname(display).slice(1) || '';
    chunks.push(
      `### \`${display}\`${truncated ? ` (truncated to ${ATTACHED_FILE_MAX_BYTES} bytes)` : ''}\n\`\`\`${ext}\n${slice}\n\`\`\``,
    );
  }

  if (chunks.length === 0) return '';
  return ['## Attached files', '', ...chunks, ''].join('\n');
}

/**
 * Wrap a persona system_prompt in a fenced block for ask.md. Returns
 * empty string when no persona prompt is provided so call sites can do
 * `[personaBlock(...), ...rest].join('\n')` without conditionals.
 *
 * **Fence rationale (retroactive PR #17 review, all 3 reviewers flagged):**
 * persona.system_prompt is user-editable text from the personas table.
 * Without delimiters a malicious or just typo'd persona can ship markdown
 * structure (`# heading`, `---` HR, ``` code fences) that bleeds into the
 * task framing — at best confusing the LLM about heading hierarchy, at
 * worst overriding "Your role" / "How to respond" sections downstream.
 *
 * We use HTML/XML-style tags rather than markdown fences because:
 *   - Markdown fences would force the LLM to read the persona AS code,
 *     not instructions.
 *   - HTML tags are widely understood by current LLMs as semantic
 *     boundaries (Anthropic + OpenAI both document this pattern).
 *   - A persona that contains the literal `</persona_instructions>` to
 *     try to break out is the only escape vector, and we strip it.
 */
function personaPromptBlock(systemPrompt: string | undefined): string {
  if (!systemPrompt || systemPrompt.trim().length === 0) return '';
  // Defensive escape: strip any closing tag that would break out of our
  // fence. Keeps the worst case (a malicious persona) from rewriting the
  // task framing. Open tags are harmless; only the closer matters.
  const sanitized = systemPrompt.trim().replace(/<\/persona_instructions>/gi, '');
  return [
    '<persona_instructions>',
    sanitized,
    '</persona_instructions>',
    '',
  ].join('\n');
}

/** Build the doer ask.md prompt for one phase iteration. */
export function buildAsk(
  phase: Phase,
  _phaseIdx: number,
  round: number,
  work: string,
  inputs: Phase['inputs'],
  filesBlock: string,
  personaSystemPrompt?: string,
  priorRoundFeedback?: string,
): string {
  const lines: string[] = [];

  const personaBlock = personaPromptBlock(personaSystemPrompt);
  if (personaBlock) {
    lines.push(personaBlock);
  }
  lines.push(`# Chorus task — round ${round}, phase ${phase.id}`);
  lines.push('');
  lines.push('## Your role');
  lines.push('doer');
  lines.push('');
  lines.push('## What to do');
  lines.push(phase.title);
  if (phase.description) {
    lines.push('');
    lines.push(phase.description);
  }
  lines.push('');
  lines.push("## The user's request");
  lines.push(work);
  lines.push('');

  if (filesBlock) {
    lines.push(filesBlock);
  }

  if (inputs.include && inputs.include.length > 0) {
    lines.push('## Inputs (from prior phases)');
    for (const includePhaseId of inputs.include) {
      lines.push(`- Phase ${includePhaseId}: (link to answer.md)`);
    }
    lines.push('');
  }

  if (inputs.exclude && inputs.exclude.length > 0) {
    lines.push('## Excluded (do NOT read)');
    for (const excludePhaseId of inputs.exclude) {
      lines.push(`- Phase ${excludePhaseId}: explicitly blocked`);
    }
    lines.push('');
  }

  if (priorRoundFeedback && priorRoundFeedback.trim().length > 0) {
    lines.push(priorRoundFeedback);
  }

  lines.push('## How to respond');
  lines.push('Write your full answer and end with: ## DONE');

  return lines.join('\n');
}

/** Build the reviewer ask.md prompt for one phase iteration. */
export function buildReviewerAsk(
  phase: Phase,
  _phaseIdx: number,
  round: number,
  work: string,
  doerOutput: string,
  filesBlock: string,
  personaSystemPrompt?: string,
): string {
  const lines: string[] = [];

  const personaBlock = personaPromptBlock(personaSystemPrompt);
  if (personaBlock) {
    lines.push(personaBlock);
  }
  lines.push(`# Chorus review — round ${round}, phase ${phase.id}`);
  lines.push('');
  lines.push('## Your role');
  lines.push('reviewer');
  lines.push('');
  lines.push('## What to review');
  lines.push(phase.title);
  if (phase.description) {
    lines.push('');
    lines.push(phase.description);
  }
  lines.push('');
  lines.push("## The user's request");
  lines.push(work);
  lines.push('');

  if (filesBlock) {
    lines.push(filesBlock);
  }

  lines.push('## Artifact to review');
  lines.push('```');
  // Truncation cap: 256 KB matches MAX_PHASE_OUTPUT_BYTES in lib/db. The
  // prior 2000-char cap silently amputated any diff or draft over ~50
  // lines, which made review-only mode useless and degraded standard
  // review mode whenever the doer wrote a real implementation. 256 KB
  // covers ~5000 lines of typical code; bigger artifacts truncate with a
  // visible marker so reviewers can still flag the gap.
  const ARTIFACT_PROMPT_CAP_BYTES = 256 * 1024;
  const byteLen = Buffer.byteLength(doerOutput, 'utf-8');
  if (byteLen <= ARTIFACT_PROMPT_CAP_BYTES) {
    lines.push(doerOutput);
  } else {
    // Slice on bytes, then walk back to the last valid UTF-8 start byte so
    // we don't hand the LLM a U+FFFD-laden tail. UTF-8 continuation bytes
    // start with 0b10xxxxxx — walk left while the cut byte is a
    // continuation byte; landing on a start byte (or ASCII) is safe.
    const buf = Buffer.from(doerOutput, 'utf-8');
    let cut = ARTIFACT_PROMPT_CAP_BYTES;
    while (cut > 0 && (buf[cut] & 0b1100_0000) === 0b1000_0000) cut--;
    lines.push(buf.subarray(0, cut).toString('utf-8'));
    lines.push(`... (truncated — full artifact was ${byteLen} bytes, cap is ${ARTIFACT_PROMPT_CAP_BYTES} bytes)`);
  }
  lines.push('```');
  lines.push('');
  lines.push('## Your verdict');
  lines.push(
    'Do you approve? Answer: approve or request changes, end with: ## DONE',
  );

  return lines.join('\n');
}
