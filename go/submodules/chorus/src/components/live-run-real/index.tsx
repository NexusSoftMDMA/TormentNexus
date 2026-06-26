"use client";

/**
 * Real-data run view for the /runs/<id> page. Renders doer + reviewer
 * cards with content read from disk on the server and live progress
 * streamed via SSE.
 *
 * Visual structure mirrors the prototype demo: status header, phase
 * progress, grid of reviewer cards. Mock simulation effects from the
 * v0.6 LiveRunView are gone — every value comes from a real source.
 *
 * Header actions live in `header-actions.tsx`, the secondary stepper
 * in `phase-progress.tsx`, and the placeholder-slot synthesis in
 * `enrich-rounds.ts`.
 */

import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { isReviewOnlyTemplate, type Template } from "@/lib/types";
import { BriefHeading } from "../run-viewer/brief-heading";
import { RoundView } from "../run-viewer/round-view";
import type {
  FallbackSwap,
  ParticipantSnapshot,
  ParticipantWarning,
  RoundSnapshot,
} from "../run-viewer/types";
import { enrichRounds } from "./enrich-rounds";
import { HeaderActions } from "./header-actions";
import {
  deriveStatusMeta,
  participantKey,
  STATUS_DOT_COLOR,
  TERMINAL_STATUSES,
  type SSEEvent,
} from "./helpers";
import { PhaseProgress } from "./phase-progress";

/**
 * Demo hook for the unlinked /demo/[scenario] route. When set, the
 * component substitutes the live SSE source with the provided factory
 * and skips every artifact-polling fetch (the demo's mock stream is
 * the single source of truth — no daemon round-trips). Default is
 * `undefined`, which preserves real production behaviour.
 */
export interface DemoDataSource {
  /** Factory that returns an EventSource-compatible object. */
  createEventSource: () => EventSource;
  /** Mock for /api/run-artifacts. Called when the SSE handler would
   *  otherwise fetch artifacts (participant_done, chat_done). Returns
   *  the rounds snapshot for the current scripted moment. */
  fetchArtifacts: () => { rounds: RoundSnapshot[]; swaps?: FallbackSwap[] };
}

interface Props {
  chatId: string;
  initialStatus: string;
  initialRounds: RoundSnapshot[];
  template: Template | null;
  /** Raw template id from the chat row. Used as a header fallback when
   * `template` resolved to null (template deleted after the chat was
   * created). Optional for forward-compat with callers that don't have it. */
  templateId?: string;
  work: string;
  projectName?: string;
  /** PR URL when ship phase succeeded (chat status=merged). */
  initialPrUrl?: string;
  /** Ship phase failure detail when status=blocked. */
  initialShipError?: string;
  /** Reviewer-level outcome (separate from system-level status). When
   * status='approved' but verdict='request_changes', the run finished
   * but reviewers said no — header must reflect that, not green-stamp it. */
  initialVerdict?: string;
  /** Demo-only — see DemoDataSource. */
  demoDataSource?: DemoDataSource;
}

export function LiveRunReal({
  chatId,
  initialStatus,
  initialRounds,
  templateId,
  template,
  work,
  projectName,
  initialPrUrl,
  initialShipError,
  initialVerdict,
  demoDataSource,
}: Props) {
  const [status, setStatus] = useState(initialStatus);
  const [verdict, setVerdict] = useState<string | undefined>(initialVerdict);
  const [rounds, setRounds] = useState<RoundSnapshot[]>(initialRounds);
  const [activeParticipants, setActiveParticipants] = useState<Set<string>>(
    new Set(),
  );
  const [prUrl, setPrUrl] = useState<string | undefined>(initialPrUrl);
  const [shipError, setShipError] = useState<string | undefined>(initialShipError);

  // Live tail per participant (`<role>-<agentName>` → most recent ~500
  // chars). When headless transport is in use, runner emits
  // phase_progress events with payload.output containing the latest
  // accumulated tail. Render this immediately for instant feedback,
  // falling back to disk-polled content when the SSE event hasn't
  // arrived yet.
  const [liveTails, setLiveTails] = useState<Record<string, string>>({});

  // Live phase-completion counter, driven from phase_done SSE events.
  // The status-only `completedPhaseCount` derivation stays at 0 until
  // the chat reaches a terminal state, which made multi-phase chats
  // look frozen for their entire duration. Tracking phase_done gives
  // the stepper the signal it needs to advance during the run. Persist
  // max-seen instead of last-seen because phase_done events carry an
  // explicit phaseIdx and out-of-order arrival is rare-but-possible
  // after a reattach replay.
  const [livePhaseDoneIdx, setLivePhaseDoneIdx] = useState<number>(-1);

  // Warnings keyed by participant dir name (same key the on-disk
  // artifacts route returns). The runner emits cli_warning events with
  // payload.agent === participant identifier (e.g. "claude-code" for
  // doer, "codex-cli-0" for reviewer). Multiple warnings stack on the
  // card; cleared at session end when SSE closes.
  const [participantWarnings, setParticipantWarnings] = useState<
    Record<string, ParticipantWarning[]>
  >({});

  // Cross-lineage / cross-model fallback swaps, keyed nowhere — rendered
  // as their own cards on the run page so the user sees "codex hit
  // quota → claude took over" without having to read the warnings
  // banner on the failed card. Sources merged into one array:
  //   - SSE cli_warning events (live, while chat is in flight)
  //   - _swaps.json sidecars from /api/run-artifacts (post-reload, when
  //     the SSE is closed because the chat went terminal)
  const [fallbackSwaps, setFallbackSwaps] = useState<FallbackSwap[]>([]);
  // Dedup key includes phaseId + role + agent so a future multi-phase
  // template can't collapse two distinct swaps that happen to share the
  // (round, agent, fromLineage, fromModel) tuple. Today's review-only
  // template has one phase + one role-per-agent, so this is purely
  // future-proofing — but the sidecar already stores the full identity.
  const swapKey = (s: FallbackSwap) =>
    `${s.round}:${s.phaseId}:${s.role}:${s.agent}:${s.fromLineage}:${s.fromModel}`;
  const mergeSwapsFromArtifacts = (incoming: FallbackSwap[]) => {
    setFallbackSwaps((prev) => {
      const seen = new Set(prev.map(swapKey));
      const merged = [...prev];
      for (const s of incoming) {
        const key = swapKey(s);
        if (seen.has(key)) continue;
        seen.add(key);
        merged.push(s);
      }
      return merged;
    });
  };

  const isTerminal = (TERMINAL_STATUSES as readonly string[]).includes(status);

  // Periodic refresh of artifacts from disk (cheap server fetch). The
  // SSE stream tells us *when* something changed; this fetches the new
  // content. 8s instead of 4s because each refresh is a same-origin
  // proxy + filesystem read of every artifact in the chat dir; at 4s a
  // 10-minute run did 150 round-trips, most of them unchanged. SSE
  // deltas drive the live ticker.
  useEffect(() => {
    if (isTerminal) return;
    // Demo mode: the mock SSE stream is the single source of truth.
    // Skipping the artifact poll here means the run-page renders only
    // what the scenario script emits, with no surprise late writes
    // from a real /api/run-artifacts response.
    if (demoDataSource) return;
    const refresh = async () => {
      try {
        const res = await fetch(`/api/run-artifacts/${chatId}`);
        if (!res.ok) return;
        const data = (await res.json()) as {
          rounds: RoundSnapshot[];
          swaps?: FallbackSwap[];
        };
        setRounds(data.rounds);
        if (Array.isArray(data.swaps) && data.swaps.length > 0) {
          mergeSwapsFromArtifacts(data.swaps);
        }
      } catch {
        // ignore — next tick retries
      }
    };
    const id = setInterval(refresh, 8000);
    return () => clearInterval(id);
  }, [chatId, isTerminal, demoDataSource]);

  useEffect(() => {
    if (isTerminal) return;
    // In demo mode the scripted scenario provides every event we
    // render; the real /api/daemon SSE is bypassed entirely.
    const es = demoDataSource
      ? demoDataSource.createEventSource()
      : new EventSource(`/api/daemon/chats/${chatId}/stream`);
    es.onmessage = (msg) => {
      try {
        const e = JSON.parse(msg.data) as SSEEvent;
        const role = e.payload.role as string | undefined;
        const agent = e.payload.agent as string | undefined;
        const phaseId = e.payload.phaseId as string | undefined;

        if (e.type === "phase_start" && role && agent && phaseId) {
          // Format mirrors directory naming: "<role>-<agentName>" plus
          // phaseId so renderer maps back to dir-name participants by
          // matching role + lineage.
          setActiveParticipants((prev) => {
            const next = new Set(prev);
            next.add(`${role}-${agent}-${phaseId}`);
            return next;
          });
          // Demo mode — swap rounds when a new phase begins so a
          // multi-phase scenario can rotate the participant grid as the
          // stepper advances. Production has /api/run-artifacts polling
          // for that; we shortcut here.
          if (demoDataSource) {
            const snapshot = demoDataSource.fetchArtifacts();
            setRounds(snapshot.rounds);
            if (Array.isArray(snapshot.swaps)) mergeSwapsFromArtifacts(snapshot.swaps);
          }
        }

        if (e.type === "phase_progress" && role && agent) {
          const output = e.payload.output as string | undefined;
          if (typeof output === "string" && output.length > 0) {
            // Keying must match the on-disk directory name format. The
            // payload's `agent` already includes the index suffix for
            // reviewers (`opencode-cli-1`, `opencode-cli-2`), so two
            // reviewers of the same lineage land in distinct liveTails
            // entries instead of clobbering each other.
            const key = participantKey(role, agent);
            setLiveTails((prev) => ({ ...prev, [key]: output }));
          }
        }

        if (e.type === "phase_done" || e.type === "phase_failed") {
          // Clear actives — next phase_start re-populates. Don't clear
          // liveTails — let the disk-poll update them with the final
          // answer instead of flashing empty.
          setActiveParticipants(new Set());
          if (e.type === "phase_done") {
            const idx = (e.payload?.phaseIdx as number | undefined) ?? -1;
            if (Number.isInteger(idx) && idx >= 0) {
              setLivePhaseDoneIdx((prev) => (idx > prev ? idx : prev));
            }
          }
        }

        if (e.type === "cli_warning" && agent && role) {
          // doer → "doer-<agent>"; reviewer → already includes the
          // index in agent (e.g. "codex-cli-0").
          const key = role === "doer" ? `doer-${agent}` : `reviewer-${agent}`;
          const reason = (e.payload.reason as string | undefined) ?? undefined;
          // Older runner code emitted `kind`; current runner emits
          // `reason`. Accept either so reattach against an in-flight
          // chat from a daemon-restart edge doesn't drop the banner.
          const kind =
            reason ?? (e.payload.kind as string | undefined) ?? "warning";
          const message =
            (e.payload.message as string | undefined) ?? "(no detail)";
          setParticipantWarnings((prev) => {
            const next = { ...prev };
            const existing = next[key] ?? [];
            // Suppress duplicates (same kind + message). Repeated
            // emissions from a retried runner shouldn't pile up
            // identical banners.
            if (existing.some((w) => w.kind === kind && w.message === message)) {
              return prev;
            }
            next[key] = [...existing, { kind, message, ts: e.ts ?? Date.now() }];
            return next;
          });

          // Fallback swap signal — runner emits this with reason
          // 'lineage_fallback' (cross-lineage) or 'model_fallback'
          // (same-lineage, different model). Render as its own card on
          // the round so the user sees voice-X-failed → voice-Y-active
          // without having to expand a banner on the failed card.
          if (
            (reason === "lineage_fallback" || reason === "model_fallback") &&
            typeof e.payload.fromLineage === "string" &&
            typeof e.payload.toLineage === "string"
          ) {
            const round = (e.payload.round as number | undefined) ?? 1;
            const phaseId = (e.payload.phaseId as string | undefined) ?? "";
            const fromModel =
              (e.payload.fromModel as string | undefined) ?? "(default)";
            const toModel =
              (e.payload.toModel as string | undefined) ?? "(default)";
            const fallbackIdx =
              (e.payload.fallbackIdx as number | undefined) ?? 0;
            const candidate: FallbackSwap = {
              round,
              phaseId,
              role,
              agent,
              reason,
              fromLineage: e.payload.fromLineage as string,
              toLineage: e.payload.toLineage as string,
              fromModel,
              toModel,
              fallbackIdx,
              ts: e.ts ?? Date.now(),
            };
            setFallbackSwaps((prev) => {
              // Dedup with the same key as merge-from-artifacts so live
              // SSE + post-reload sidecar reads can't double-count.
              const candidateKey = swapKey(candidate);
              if (prev.some((s) => swapKey(s) === candidateKey)) return prev;
              return [...prev, candidate];
            });
          }
        }

        if (e.type === "participant_done") {
          // The runner has just written `## DONE` to this participant's
          // answer.md. Pull artifacts immediately so the card flips
          // from WORKING to DONE without waiting for the 8s polling
          // tick. Demo mode hands the scripted snapshot in via
          // demoDataSource.fetchArtifacts() instead of /api/run-artifacts.
          if (demoDataSource) {
            const snapshot = demoDataSource.fetchArtifacts();
            setRounds(snapshot.rounds);
            if (Array.isArray(snapshot.swaps)) mergeSwapsFromArtifacts(snapshot.swaps);
          } else {
            fetch(`/api/run-artifacts/${chatId}`)
              .then((r) => (r.ok ? r.json() : null))
              .then((data) => {
                if (!data) return;
                setRounds(data.rounds);
                if (Array.isArray(data.swaps)) mergeSwapsFromArtifacts(data.swaps);
              })
              .catch(() => {});
          }
        }

        if (e.type === "chat_done") {
          // Runner emits chat_done with payload.status as the canonical
          // terminal state ('completed' / 'merged' / 'blocked' /
          // 'no_review'). Prefer that over verdict for the UI.
          const finalStatus = e.payload.status as string | undefined;
          if (finalStatus === "merged") setStatus("merged");
          else if (finalStatus === "blocked") setStatus("blocked");
          else if (finalStatus === "no_review") setStatus("no_review");
          else if (finalStatus === "failed") setStatus("failed");
          else if (finalStatus === "cancelled") setStatus("cancelled");
          else setStatus("approved");

          const finalVerdict = e.payload.verdict as string | undefined;
          if (typeof finalVerdict === "string" && finalVerdict.length > 0) {
            setVerdict(finalVerdict);
          }

          const payloadPrUrl = e.payload.prUrl as string | undefined;
          if (typeof payloadPrUrl === "string" && payloadPrUrl.length > 0) {
            setPrUrl(payloadPrUrl);
          }
          const payloadShipError = e.payload.shipError as string | undefined;
          if (typeof payloadShipError === "string" && payloadShipError.length > 0) {
            setShipError(payloadShipError);
          }

          es.close();
          if (demoDataSource) {
            const snapshot = demoDataSource.fetchArtifacts();
            setRounds(snapshot.rounds);
            if (Array.isArray(snapshot.swaps)) mergeSwapsFromArtifacts(snapshot.swaps);
          } else {
            fetch(`/api/run-artifacts/${chatId}`)
              .then((r) => (r.ok ? r.json() : null))
              .then((data) => {
                if (!data) return;
                setRounds(data.rounds);
                if (Array.isArray(data.swaps)) mergeSwapsFromArtifacts(data.swaps);
              })
              .catch(() => {});
          }
        }
      } catch {
        // skip malformed
      }
    };
    return () => es.close();
  }, [chatId, isTerminal, demoDataSource]);

  // One-shot fetch on mount (incl. for terminal chats where the SSE
  // useEffect early-returns). Without this, navigating to a completed
  // chat would never load the swap sidecars — the periodic refresh and
  // SSE branches both skip when isTerminal is true. Demo mode skips —
  // scenario events drive every fallback render directly.
  useEffect(() => {
    if (demoDataSource) return;
    let cancelled = false;
    fetch(`/api/run-artifacts/${chatId}`)
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (cancelled || !data) return;
        if (Array.isArray(data.swaps)) mergeSwapsFromArtifacts(data.swaps);
      })
      .catch(() => {
        /* best-effort */
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chatId, demoDataSource]);

  /** Active keys are built as `${role}-${agent}-${phaseId}` in the
   * phase_start handler (where `agent` includes the per-slot index for
   * reviewers — e.g. `opencode-cli-1`). The participant's
   * `p.participant` is the on-disk dir name (`reviewer-opencode-cli-1`).
   * Match by dir name as a strict prefix so two same-lineage reviewers
   * don't both light up when only one is streaming. Earlier code
   * matched on `${p.role}-${p.lineage}-` which is identical for
   * `opencode-cli-1` / `opencode-cli-2`, so the active glow leaked. */
  const lineageMatchActive = (p: ParticipantSnapshot): boolean => {
    const prefix = `${p.participant}-`;
    for (const k of activeParticipants) {
      if (k.startsWith(prefix)) return true;
    }
    return false;
  };

  const meta = deriveStatusMeta(status, verdict);
  const totalPhases = template?.phases?.length ?? 1;

  // Phase completion is driven by terminal status (not disk snapshots).
  // The previous "any participant has an answer → phase done" heuristic
  // flipped the stepper to DONE the moment the doer wrote its first
  // byte, even though reviewers were still running and consensus wasn't
  // reached. With status-driven logic the phase only goes "done" when
  // the chat itself is in an approved-equivalent terminal state.
  //
  // While drafting/reviewing, livePhaseDoneIdx (the highest phaseIdx
  // seen with phase_done; +1 converts to a count, clamped to
  // totalPhases for safety in case a stale replay sends an out-of-range
  // index) provides the live signal so multi-phase chats don't sit at
  // "0/N done" the entire run.
  const completedPhaseCount = useMemo(() => {
    if (status === "approved" || status === "merged") return totalPhases;
    if (status === "no_review" || status === "blocked") return totalPhases;
    if (status === "failed" || status === "cancelled") return 0;
    return Math.min(Math.max(0, livePhaseDoneIdx + 1), totalPhases);
  }, [status, totalPhases, livePhaseDoneIdx]);

  const reviewOnly = useMemo(() => isReviewOnlyTemplate(template), [template]);
  const enrichedRounds = useMemo<RoundSnapshot[]>(
    () => enrichRounds(rounds, template, participantWarnings),
    [rounds, template, participantWarnings],
  );

  const latestRound = enrichedRounds[enrichedRounds.length - 1];
  const olderRounds = enrichedRounds.slice(0, -1);

  return (
    <div className="flex h-full flex-col">
      <div className="sticky top-0 z-20 border-b border-border bg-card/80 backdrop-blur-sm px-4 py-3 sm:px-8">
        {/* Two-row header: meta-bar on top (status pill + template badge
            on the left, action buttons on the right, fixed-height row that
            never shifts), then the title/brief block below — gives both
            rows independent layout so a long title or expanded brief
            never pushes the action buttons around. */}
        <div className="mb-2 flex w-full items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <Link
              href="/runs"
              className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground transition hover:text-foreground"
            >
              <ArrowLeft className="h-3 w-3" />
              <span>{projectName ?? "Runs"}</span>
            </Link>

            <span className="text-muted-foreground/40">·</span>

            {/* Status pill — dot + label together; reads cleanly in
                isolation rather than the orphan dot floating next to the
                title. */}
            <span
              title={meta.text}
              className="inline-flex shrink-0 items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground"
            >
              <span
                className={`h-2 w-2 shrink-0 rounded-full ${STATUS_DOT_COLOR[meta.color]} ${
                  isTerminal ? "" : "animate-pulse-soft"
                }`}
              />
              {status}
            </span>

            {/* Template badge. Falls back to the raw templateId when the
                template row was deleted out from under the chat. */}
            {(template || templateId) && (
              <>
                <span className="text-muted-foreground/40">·</span>
                <Link
                  href={`/templates${template ? `#${encodeURIComponent(template.id)}` : ""}`}
                  title={template ? `Template: ${template.name}` : `Template (deleted): ${templateId}`}
                  className="inline-flex min-w-0 shrink items-center gap-1.5 text-[11px] text-muted-foreground transition hover:text-primary"
                >
                  <span className="font-mono uppercase tracking-wider">tpl</span>
                  <span className="truncate font-medium text-foreground">
                    {template?.name ?? templateId}
                  </span>
                </Link>
              </>
            )}
          </div>

          <HeaderActions
            chatId={chatId}
            status={status}
            isTerminal={isTerminal}
            template={template}
            onCancel={() => setStatus("cancelled")}
          />
        </div>

        {/* Title row — full width, BriefHeading owns its own truncation
            and "Show full brief" expander without competing with the
            action buttons for vertical space. */}
        <div className="min-w-0">
          <BriefHeading work={work} />
        </div>
      </div>

      <PhaseProgress
        template={template}
        status={status}
        totalPhases={totalPhases}
        completedPhaseCount={completedPhaseCount}
        rounds={rounds}
        enrichedRounds={enrichedRounds}
        prUrl={prUrl}
        shipError={shipError}
      />

      {/* Body — full-width container. Reviewer outputs are text-heavy
          and benefit from the extra horizontal space. The 6xl cap was
          inherited from a marketing-style layout that doesn't fit a
          tool surface. */}
      <div className="flex-1 overflow-y-auto px-4 py-6 sm:px-8">
        <div className="mx-auto w-full space-y-8">
          {rounds.length === 0 && (
            <div className="rounded-lg border border-dashed border-border p-12 text-center text-sm text-muted-foreground">
              Waiting for first phase to start…
            </div>
          )}

          {latestRound && (
            <RoundView
              round={latestRound}
              isLatest
              activeFor={lineageMatchActive}
              liveTails={liveTails}
              chatTerminal={isTerminal}
              reviewOnly={reviewOnly}
              chatId={chatId}
              swaps={fallbackSwaps}
            />
          )}

          {/* Review-only chats are single-pass by design — there is
              never an "earlier rounds" panel because there's exactly
              one round. */}
          {!reviewOnly && olderRounds.length > 0 && (
            <details className="rounded-lg border border-border bg-card">
              <summary className="cursor-pointer px-4 py-3 text-xs uppercase tracking-wider text-muted-foreground hover:text-foreground">
                Earlier rounds ({olderRounds.length})
              </summary>
              <div className="space-y-6 border-t border-border p-4">
                {olderRounds
                  .slice()
                  .reverse()
                  .map((r) => (
                    <RoundView
                      key={r.round}
                      round={r}
                      activeFor={() => false}
                      liveTails={{}}
                      chatTerminal={isTerminal}
                      swaps={fallbackSwaps}
                    />
                  ))}
              </div>
            </details>
          )}
        </div>
      </div>
    </div>
  );
}
