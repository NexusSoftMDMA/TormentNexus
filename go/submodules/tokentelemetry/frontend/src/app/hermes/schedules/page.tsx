"use client";

// Read-only schedules listing.
//
// The full create / edit / pause / resume / run / delete UI was wired up on
// this branch (see git history + `frontend/src/components/schedules/`
// + `frontend/src/lib/hermesCron.ts`) but is intentionally disabled while
// the backend mutation endpoints are commented out. To re-enable: uncomment
// the DISABLED-MUTATIONS block in `backend/main.py` and restore the action
// buttons / modal wiring below.

import { useResource } from "@/lib/api";
import {
  PageHeader, Card, Table, THead, TBody, TR, TH, TD, Badge, EmptyState,
} from "@/components/ui";
import { Timer, AlertTriangle, CheckCircle2, Pause } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { CronJob } from "@/lib/hermesCron";

// formatDistanceToNow throws on invalid dates — defensive wrapper to keep one
// malformed timestamp from blanking the whole table.
function safeRel(iso: string | null | undefined, fallback = "—"): string {
  if (!iso) return fallback;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return fallback;
  return formatDistanceToNow(d, { addSuffix: true });
}

export default function SchedulesPage() {
  const res = useResource<{ cron_jobs?: CronJob[] }>("/hermes/overview", { pollMs: 15000 });
  const jobs: CronJob[] = res.data?.cron_jobs || [];

  return (
    <div className="px-8 py-8 max-w-[1200px] mx-auto space-y-6 pb-20">
      <PageHeader
        backHref="/hermes"
        icon={
          <div className="h-10 w-10 grid place-items-center rounded-[var(--tt-radius)] bg-indigo-500/10 border border-indigo-500/30">
            <Timer className="text-indigo-500" size={20} />
          </div>
        }
        eyebrow="Hermes Agent"
        title="Schedules"
        description="Cron jobs and periodic background tasks"
      />

      {res.error ? (
        <EmptyState
          title="Couldn't load schedules"
          description={res.error.message || "The /hermes/overview endpoint failed."}
        />
      ) : res.loading && !res.data ? (
        <Card>
          <div className="p-8 text-[13px] text-[var(--tt-fg-muted)]">Loading…</div>
        </Card>
      ) : jobs.length === 0 ? (
        <EmptyState
          title="No schedules"
          description="Hermes has no active cron jobs. Create one with `hermes cron create` in your terminal."
        />
      ) : (
        <Card>
          <Table>
            <THead>
              <TR>
                <TH className="pl-5">Job</TH>
                <TH>Schedule</TH>
                <TH>Deliver</TH>
                <TH>Last run</TH>
                <TH>Next run</TH>
                <TH className="pr-5">Status</TH>
              </TR>
            </THead>
            <TBody>
              {jobs.map((j, idx) => {
                const paused = j.state === "paused" || !j.enabled;
                return (
                  <TR key={j.id ?? j.name ?? idx}>
                    <TD className="pl-5 align-top py-3">
                      <div className="font-mono text-[12.5px] text-[var(--tt-fg)]">{j.name}</div>
                      {j.prompt && (
                        <div
                          className="text-[11.5px] text-[var(--tt-fg-muted)] mt-1 max-w-[260px] truncate"
                          title={j.prompt}
                        >
                          {j.prompt}
                        </div>
                      )}
                      {j.last_error && (
                        <div
                          className="text-[11px] text-rose-300 mt-1 max-w-[260px] truncate"
                          title={j.last_error}
                        >
                          {j.last_error}
                        </div>
                      )}
                    </TD>
                    <TD className="font-mono text-[11.5px] text-[var(--tt-fg-muted)] align-top py-3">
                      {j.schedule_display}
                    </TD>
                    <TD className="align-top py-3">
                      <div className="flex flex-wrap gap-1">
                        {j.deliver.map((d) => (
                          <Badge key={d} variant="neutral" size="xs">{d}</Badge>
                        ))}
                      </div>
                    </TD>
                    <TD className="text-[11.5px] text-[var(--tt-fg-muted)] tabular-nums align-top py-3">
                      {safeRel(j.last_run_at, "never")}
                    </TD>
                    <TD className="text-[11.5px] tabular-nums align-top py-3">
                      {paused ? (
                        <span className="text-[var(--tt-fg-dim)]">paused</span>
                      ) : j.next_run_at ? (
                        <span
                          className={
                            j.at_risk ? "text-[var(--tt-warn-fg)]" : "text-[var(--tt-fg-muted)]"
                          }
                        >
                          {safeRel(j.next_run_at)}
                        </span>
                      ) : "—"}
                    </TD>
                    <TD className="pr-5 align-top py-3">
                      {paused ? (
                        <Badge variant="neutral" size="xs">
                          <Pause size={10} className="mr-1" /> paused
                        </Badge>
                      ) : j.at_risk ? (
                        <Badge variant="warn" size="xs">
                          <AlertTriangle size={10} className="mr-1" /> at risk
                        </Badge>
                      ) : j.last_status === "ok" ? (
                        <Badge variant="success" size="xs">
                          <CheckCircle2 size={10} className="mr-1" /> ok
                        </Badge>
                      ) : j.last_status === "error" ? (
                        <Badge variant="danger" size="xs">
                          <AlertTriangle size={10} className="mr-1" /> error
                        </Badge>
                      ) : (
                        <Badge variant="neutral" size="xs">{j.last_status || "—"}</Badge>
                      )}
                    </TD>
                  </TR>
                );
              })}
            </TBody>
          </Table>
        </Card>
      )}
    </div>
  );
}
