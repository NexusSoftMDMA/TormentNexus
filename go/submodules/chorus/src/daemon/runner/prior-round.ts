/**
 * Prior-round feedback reader.
 *
 * On round ≥ 2 the doer was rerunning the SAME prompt as round 1 — only
 * the round number in the heading changed. Reviewer findings from round
 * N-1 were never threaded back, so disagreement → retry was effectively
 * "ask again, hope the LLM rolls a better die" instead of a real revision
 * loop.
 *
 * This helper assembles a "## Prior round feedback" markdown block from
 * the previous round's reviewer answer.md files. The doer prompt builder
 * inlines it before the "## How to respond" footer.
 *
 * Caps:
 *   - per-reviewer slice: 16KB (truncated answers still mark the cut)
 *   - total block:        64KB (additional reviewers replaced with marker)
 *
 * Numbers chosen to leave the 256KB doer prompt budget mostly intact for
 * attached files + persona + task framing, while still surfacing enough
 * reviewer text to be actionable.
 */
import * as fs from 'fs';
import * as path from 'path';

const PER_REVIEWER_MAX_BYTES = 16 * 1024;
const TOTAL_FEEDBACK_MAX_BYTES = 64 * 1024;

/**
 * Returns a markdown block summarizing the previous round's reviewer
 * findings, or empty string when there's nothing to feed back.
 *
 * Empty cases:
 *   - round <= 1 (no prior round exists)
 *   - prior round dir missing (e.g. crash mid-round)
 *   - no reviewer-* dirs found
 *   - all reviewer answer.md files unreadable / empty
 */
export function readPriorRoundFeedback(chatDir: string, round: number): string {
  if (round <= 1) return '';

  const priorRoundDir = path.join(chatDir, `round-${round - 1}`);
  if (!fs.existsSync(priorRoundDir)) return '';

  let entries: string[];
  try {
    entries = fs.readdirSync(priorRoundDir);
  } catch {
    return '';
  }

  // Stable sort by dir name so reviewer order in the prompt matches the
  // run page rendering. dir name embeds the reviewer index already.
  const reviewerDirs = entries
    .filter((n) => n.startsWith('reviewer-'))
    .sort();

  if (reviewerDirs.length === 0) return '';

  const chunks: string[] = [];
  let totalBytes = 0;

  for (const dir of reviewerDirs) {
    const answerPath = path.join(priorRoundDir, dir, 'answer.md');
    if (!fs.existsSync(answerPath)) continue;

    let body: string;
    try {
      body = fs.readFileSync(answerPath, 'utf-8');
    } catch {
      continue;
    }

    if (body.trim().length === 0) continue;

    // Reviewer dir name shape: `reviewer-<agentName>-<idx>`. Strip the
    // prefix + trailing index so the heading reads cleanly. Falls back
    // to the raw dir name if the regex misses (defensive — agent names
    // with internal dashes still match because we anchor on -<digits>$).
    const labelMatch = dir.match(/^reviewer-(.+)-(\d+)$/);
    const label = labelMatch
      ? `${labelMatch[1]} (#${labelMatch[2]})`
      : dir;

    const truncated = body.length > PER_REVIEWER_MAX_BYTES;
    const slice = truncated ? body.slice(0, PER_REVIEWER_MAX_BYTES) : body;

    const block =
      `### Reviewer: ${label}\n` +
      slice +
      (truncated ? `\n\n_(truncated — full answer was ${body.length} bytes)_\n` : '');

    const blockBytes = Buffer.byteLength(block, 'utf-8');
    if (totalBytes + blockBytes > TOTAL_FEEDBACK_MAX_BYTES) {
      const remaining = reviewerDirs.length - chunks.length;
      chunks.push(
        `### _(${remaining} more reviewer${remaining === 1 ? '' : 's'} omitted — feedback block exceeded ${TOTAL_FEEDBACK_MAX_BYTES}-byte cap)_`,
      );
      break;
    }

    chunks.push(block);
    totalBytes += blockBytes;
  }

  if (chunks.length === 0) return '';

  return [
    '## Prior round feedback',
    '',
    'The previous round did not reach reviewer consensus. Read each reviewer\'s findings below and revise your answer to address them — do not just repeat what you wrote before. Where reviewers disagree with each other, use your judgment.',
    '',
    ...chunks,
    '',
  ].join('\n');
}
