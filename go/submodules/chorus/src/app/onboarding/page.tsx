"use client";

import { ArrowRight, Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState, useTransition } from "react";
import { TriadLogo } from "@/components/triad-logo";
import { Button } from "@/components/ui/button";
import {
  detectInstalledClis,
  saveCliPath,
  validateCliPath,
  type CliDetection,
  type DetectableCliId,
  type SandboxProfile,
} from "@/lib/api/settings";
import {
  listOpencodeModels,
  type OpencodeModelsResult,
} from "@/lib/api/orchestrators";
import { listVoices, updateVoice, type Voice } from "@/lib/api/voices";
import { ApiKeysSection } from "./api-keys-section";
import { CliSection } from "./cli-section";
import { PermissionsSection } from "./permissions-section";
import { describeError, submitOnboarding } from "./submit";

export default function OnboardingPage() {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);

  const [selectedClis, setSelectedClis] = useState<Set<string>>(new Set());
  const [cliVoices, setCliVoices] = useState<Voice[]>([]);
  // Set rather than scalar so concurrent toggles don't fight over the
  // single "in-flight" id — mashing 5 toggles used to clear the lock the
  // moment the first request finished, briefly re-enabling the other 4
  // buttons while their requests were still pending.
  const [savingVoiceIds, setSavingVoiceIds] = useState<Set<string>>(new Set());
  const [voiceSaveError, setVoiceSaveError] = useState<string | null>(null);
  const [apiKeys, setApiKeys] = useState<Record<string, string>>({});
  const [sandboxProfile, setSandboxProfile] =
    useState<SandboxProfile>("workspace");
  const [autoApprovePrompts, setAutoApprovePrompts] = useState<boolean>(true);
  const [networkAccess, setNetworkAccess] = useState<boolean>(false);
  const [detection, setDetection] = useState<Record<string, CliDetection>>({});
  // Tracks the initial detect-clis call so the UI can show "Searching…"
  // instead of falsely rendering every CLI as "Not found" before the
  // probe completes. Flips false on both success + failure paths.
  const [detecting, setDetecting] = useState(true);
  const [manualOpen, setManualOpen] = useState<Set<string>>(new Set());
  const [manualPath, setManualPath] = useState<Record<string, string>>({});
  const [manualError, setManualError] = useState<Record<string, string>>({});
  const [manualBusy, setManualBusy] = useState<Set<string>>(new Set());

  // OpenCode model picker: lazily fetched the first time the user ticks
  // OpenCode AND the binary is installed. The user picks which
  // subscription models chorus should expose as voices; persisted in
  // submit.ts.
  const [opencodeModels, setOpencodeModels] = useState<OpencodeModelsResult | null>(
    null,
  );
  const [opencodeModelsError, setOpencodeModelsError] = useState<string | null>(null);
  const [opencodeModelsLoading, setOpencodeModelsLoading] = useState(false);
  const [selectedOpencodeModels, setSelectedOpencodeModels] = useState<Set<string>>(
    new Set(),
  );

  useEffect(() => {
    // `cancelled` guards every setter against unmount-during-fetch
    // (StrictMode double-mount + nav-away mid-flight). Without it,
    // React logs warnings + we can land stale data on a dead component.
    let cancelled = false;
    detectInstalledClis()
      .then((rows) => {
        if (cancelled) return;
        const map: Record<string, CliDetection> = {};
        const preTick = new Set<string>();
        for (const row of rows) {
          map[row.id] = row;
          if (row.found) preTick.add(row.id);
        }
        setDetection(map);
        if (preTick.size > 0) setSelectedClis(preTick);
      })
      .catch(() => {
        // Detection is best-effort; if the daemon probe fails the user
        // can still tick boxes manually. No need to surface an error.
      })
      .finally(() => {
        if (!cancelled) setDetecting(false);
      });
    // Voices are seeded server-side on daemon boot — fetch the current
    // state so each CLI card can render its actual model list with
    // toggles, not just "default model auto-enabled."
    listVoices({ source: "cli" })
      .then((v) => {
        if (!cancelled) setCliVoices(v);
      })
      .catch(() => {
        /* best-effort — seeding may not have completed yet */
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const toggleVoice = async (v: Voice) => {
    setSavingVoiceIds((prev) => new Set(prev).add(v.id));
    setVoiceSaveError(null);
    try {
      const next = await updateVoice(v.id, { enabled: !v.enabled });
      setCliVoices((prev) => prev.map((p) => (p.id === next.id ? next : p)));
    } catch (err) {
      setVoiceSaveError(
        err instanceof Error ? err.message : "Couldn't save voice toggle.",
      );
    } finally {
      setSavingVoiceIds((prev) => {
        const next = new Set(prev);
        next.delete(v.id);
        return next;
      });
    }
  };

  const toggleCli = (id: string) => {
    setSelectedClis((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  // Lazy-fetch the OpenCode model list the first time the user ticks
  // OpenCode AND the binary is installed. Runs `opencode models` on the
  // daemon side and groups by gateway prefix.
  useEffect(() => {
    if (!selectedClis.has("opencode-cli")) return;
    if (!detection["opencode-cli"]?.found) return;
    // Skip when we already succeeded, are mid-flight, OR previously
    // errored. Without the error check, finally() flipping loading=false
    // re-runs this effect (loading is in deps), which retries forever.
    if (opencodeModels || opencodeModelsLoading || opencodeModelsError) return;
    setOpencodeModelsLoading(true);
    setOpencodeModelsError(null);
    listOpencodeModels()
      .then((res) => {
        setOpencodeModels(res);
        // Pre-select the fleet defaults (kimi + deepseek when present).
        setSelectedOpencodeModels((prev) => {
          if (prev.size > 0) return prev;
          return new Set(res.defaultPicks);
        });
      })
      .catch((err) => {
        setOpencodeModelsError(
          err instanceof Error
            ? err.message
            : "Couldn't list OpenCode models. Is the CLI authed (run `opencode auth login`)?",
        );
      })
      .finally(() => setOpencodeModelsLoading(false));
  }, [
    selectedClis,
    detection,
    opencodeModels,
    opencodeModelsLoading,
    opencodeModelsError,
  ]);

  const toggleOpencodeModel = (m: string) => {
    setSelectedOpencodeModels((prev) => {
      const next = new Set(prev);
      if (next.has(m)) next.delete(m);
      else next.add(m);
      return next;
    });
  };

  const toggleManual = (id: string) => {
    setManualOpen((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
    setManualError((prev) => ({ ...prev, [id]: "" }));
  };

  const submitManualPath = async (id: DetectableCliId) => {
    const value = (manualPath[id] || "").trim();
    if (!value) {
      setManualError((prev) => ({
        ...prev,
        [id]:
          "Enter the full path to the CLI program (e.g. /usr/local/bin/claude).",
      }));
      return;
    }
    setManualBusy((prev) => new Set(prev).add(id));
    setManualError((prev) => ({ ...prev, [id]: "" }));
    try {
      const result = await validateCliPath(id, value);
      if (result.found) {
        // Persist server-side BEFORE flipping UI state. If the save call
        // fails, surface the error and don't pretend the CLI is wired —
        // pre-fix this whole save step was missing, which is what made
        // the path forgotten across daemon restarts.
        try {
          await saveCliPath(id, result.path ?? value);
        } catch (err) {
          setManualError((prev) => ({
            ...prev,
            [id]: `Validation passed but save failed: ${
              err instanceof Error ? err.message : "unknown error"
            }. Try again, or paste the path again.`,
          }));
          return;
        }
        setDetection((prev) => ({ ...prev, [id]: result }));
        setSelectedClis((prev) => new Set(prev).add(id));
        // Refresh voices so the model list populates for the just-validated CLI.
        listVoices({ source: "cli" })
          .then(setCliVoices)
          .catch(() => {
            /* best-effort */
          });
        setManualOpen((prev) => {
          const next = new Set(prev);
          next.delete(id);
          return next;
        });
      } else {
        // Daemon now returns `reason` on validation failures — surface it.
        // Falls back to the generic message for older daemons.
        const reason = result.reason ?? "couldn't run that path";
        setManualError((prev) => ({
          ...prev,
          [id]: `Validation failed: ${reason}.`,
        }));
      }
    } catch {
      setManualError((prev) => ({
        ...prev,
        [id]: "Validation failed. Is the daemon running?",
      }));
    } finally {
      setManualBusy((prev) => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  };

  const updateApiKey = (provider: string, value: string) => {
    setApiKeys((prev) => ({ ...prev, [provider]: value }));
  };

  const filledCount =
    selectedClis.size +
    Object.values(apiKeys).filter((v) => v.trim().length > 0).length;

  const handleSubmit = () => {
    setError(null);
    if (filledCount === 0) {
      setError("Pick at least one CLI or paste at least one API key to continue.");
      return;
    }
    startTransition(async () => {
      try {
        await submitOnboarding({
          selectedClis,
          apiKeys,
          sandboxProfile,
          autoApprovePrompts,
          networkAccess,
          opencodeModels,
          selectedOpencodeModels,
        });
        router.push("/");
        router.refresh();
      } catch (err) {
        setError(describeError(err));
      }
    });
  };

  return (
    <main className="min-h-screen bg-background text-foreground">
      <div className="mx-auto w-full max-w-5xl px-4 py-10 sm:px-6 sm:py-14">
        <div className="mb-8 flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-md bg-primary/15 text-primary">
            <TriadLogo className="h-6 w-6" />
          </div>
          <div>
            <p className="text-xs uppercase tracking-wider text-muted-foreground">
              Welcome to Chorus
            </p>
            <h1 className="text-2xl font-semibold tracking-tight">
              Connect at least one model to begin
            </h1>
          </div>
        </div>

        <p className="mb-8 text-sm leading-relaxed text-muted-foreground">
          Chorus runs your prompt past 2–4 LLMs of different lineages and
          synthesises consensus. Pick the CLI subscriptions you already have,
          or paste API keys. You can change these later in Settings.
        </p>

        <CliSection
          selectedClis={selectedClis}
          toggleCli={toggleCli}
          detection={detection}
          detecting={detecting}
          cliVoices={cliVoices}
          savingVoiceIds={savingVoiceIds}
          voiceSaveError={voiceSaveError}
          toggleVoice={toggleVoice}
          manualOpen={manualOpen}
          toggleManual={toggleManual}
          manualPath={manualPath}
          setManualPath={setManualPath}
          manualError={manualError}
          manualBusy={manualBusy}
          submitManualPath={submitManualPath}
          opencodeModels={opencodeModels}
          opencodeModelsError={opencodeModelsError}
          opencodeModelsLoading={opencodeModelsLoading}
          selectedOpencodeModels={selectedOpencodeModels}
          toggleOpencodeModel={toggleOpencodeModel}
        />

        <ApiKeysSection apiKeys={apiKeys} updateApiKey={updateApiKey} />

        <PermissionsSection
          sandboxProfile={sandboxProfile}
          setSandboxProfile={setSandboxProfile}
          autoApprovePrompts={autoApprovePrompts}
          setAutoApprovePrompts={setAutoApprovePrompts}
          networkAccess={networkAccess}
          setNetworkAccess={setNetworkAccess}
        />

        {error && (
          <div className="mb-6 rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-xs text-muted-foreground">
            {filledCount === 0
              ? "Pick at least one to continue."
              : `${filledCount} ${filledCount === 1 ? "credential" : "credentials"} ready to save.`}
          </p>
          <Button
            onClick={handleSubmit}
            disabled={isPending || filledCount === 0}
            className="w-full sm:w-auto"
          >
            {isPending ? (
              <>
                <Loader2 className="mr-1 h-4 w-4 animate-spin" />
                Saving…
              </>
            ) : (
              <>
                Get started
                <ArrowRight className="ml-1 h-4 w-4" />
              </>
            )}
          </Button>
        </div>
      </div>
    </main>
  );
}
