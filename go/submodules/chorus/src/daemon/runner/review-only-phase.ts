import fs from 'fs';
import path from 'path';
import type { ReviewOnlyPhase, StandardPhase } from '../../lib/template-schema.js';
import type { ErrorDetector } from '../error-detector.js';
import type { TmuxManager } from '../tmux-types.js';
import { runReviewers } from './reviewer-driver.js';
import type { RunnerEvent } from './types.js';

/**
 * Run a review-only phase. The artifact (supplied at chat-create time) is
 * written to a synthetic doer answer slot; reviewers then critique it
 * just like a real doer's answer. Single pass — no iterate, no retry.
 * Reviewer agreement/disagreement is reported via outcome.allReviewersFailed;
 * the verdict itself doesn't gate further phases (review-only is the whole
 * point of the chat).
 *
 * Synthetic doer events make the cockpit + replay code paths Just Work
 * without special-casing the missing doer card.
 */
export async function runReviewOnlyPhase(args: {
  chatDir: string;
  chatId: string;
  phase: ReviewOnlyPhase;
  phaseIdx: number;
  artifact: string;
  work: string;
  filesBlock: string;
  tmuxMgr: TmuxManager;
  errorDetector: ErrorDetector;
  onEvent: (e: RunnerEvent) => void;
  abortSignal: AbortSignal;
  templateFallbackReviewer?: ReadonlyArray<{ lineage: string; models: string[] }>;
}): Promise<{
  completed: boolean;
  allReviewersFailed: boolean;
  /** True iff reviewer agreement met phase.reviewer.require threshold. */
  agreed: boolean;
  /** Human-readable summary line ('2/3 reviewers agreed' etc.). */
  summary: string;
}> {
  const {
    chatDir,
    chatId,
    phase,
    phaseIdx,
    artifact,
    work,
    filesBlock,
    tmuxMgr,
    errorDetector,
    onEvent,
    abortSignal,
  } = args;

  if (abortSignal.aborted) {
    return { completed: false, allReviewersFailed: false, agreed: false, summary: 'aborted' };
  }

  const round = 1; // review-only is always single-pass
  const roundDir = path.join(chatDir, `round-${round}`);
  if (!fs.existsSync(roundDir)) {
    fs.mkdirSync(roundDir, { recursive: true });
  }
  // Synthetic doer dir holds the artifact as answer.md so the cockpit's
  // existing replay code finds it under the same path shape it expects.
  const syntheticDoerDir = path.join(roundDir, 'doer-artifact');
  if (!fs.existsSync(syntheticDoerDir)) {
    fs.mkdirSync(syntheticDoerDir, { recursive: true });
  }
  const answerFile = path.join(syntheticDoerDir, 'answer.md');
  // Strip trailing whitespace before the sentinel check so an artifact
  // ending with "## DONE\n" or "## DONE  " doesn't produce a duplicate
  // sentinel after we append. Idempotent: artifacts without the sentinel
  // get a clean "\n\n## DONE\n" tail.
  const trimmed = artifact.replace(/\s+$/, '');
  const artifactWithSentinel = /##\s*DONE$/i.test(trimmed)
    ? `${trimmed}\n`
    : `${trimmed}\n\n## DONE\n`;
  fs.writeFileSync(answerFile, artifactWithSentinel);

  // Synthetic doer phase events. agent='artifact' is a sentinel value the
  // cockpit can render as "user-supplied" rather than as a real CLI run.
  onEvent({
    chatId,
    type: 'phase_start',
    payload: {
      phaseId: phase.id,
      phaseIdx,
      kind: phase.kind,
      round,
      role: 'doer',
      agent: 'artifact',
      synthetic: true,
    },
    ts: Date.now(),
  });
  onEvent({
    chatId,
    type: 'phase_progress',
    payload: {
      phaseId: phase.id,
      round,
      role: 'doer',
      agent: 'artifact',
      output: artifact.slice(0, 500),
      synthetic: true,
    },
    ts: Date.now(),
  });

  // The synthetic StandardPhase shape carries the same reviewer block
  // plus a no-op iterate config so runReviewers can consume it. Local
  // to this call — never escapes back into the template.
  const syntheticStandardPhase: StandardPhase = {
    id: phase.id,
    kind: 'review',
    title: phase.title,
    description: phase.description,
    doer: { lineage: 'any' },
    reviewer: phase.reviewer,
    inputs: phase.inputs,
    iterate: {
      maxRounds: 1,
      onDisagreement: 'continue',
      shareSessionAcrossRounds: false,
      shareSessionAcrossPhases: false,
    },
    timeoutMs: phase.timeoutMs,
  };

  const consensus = await runReviewers(
    chatDir,
    chatId,
    syntheticStandardPhase,
    phaseIdx,
    round,
    artifact,
    work,
    filesBlock,
    tmuxMgr,
    errorDetector,
    onEvent,
    abortSignal,
    args.templateFallbackReviewer,
  );

  return {
    completed: !abortSignal.aborted,
    allReviewersFailed: consensus.allFailed,
    agreed: consensus.agreed,
    summary: consensus.summary,
  };
}
