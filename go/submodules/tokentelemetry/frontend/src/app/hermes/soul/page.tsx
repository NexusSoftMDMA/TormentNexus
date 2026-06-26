"use client";

import { useResource } from "@/lib/api";
import { PageHeader, Card, EmptyState } from "@/components/ui";
import { Sparkles } from "lucide-react";

export default function SoulPage() {
  const { data } = useResource<{ content: string; exists: boolean }>("/hermes/soul");
  const isLoading = !data;

  return (
    <div className="px-8 py-8 max-w-[1200px] mx-auto space-y-6 pb-20">
      <PageHeader
        backHref="/hermes"
        icon={<div className="h-10 w-10 grid place-items-center rounded-[var(--tt-radius)] bg-fuchsia-500/10 border border-fuchsia-500/30"><Sparkles className="text-fuchsia-500" size={20} /></div>}
        eyebrow="Hermes Agent"
        title="Soul & Persona"
        description="The core persona and instructions defined in SOUL.md"
      />
      {isLoading ? (
        <div className="animate-pulse h-32 bg-[var(--tt-panel)] rounded-xl" />
      ) : !data?.exists ? (
        <EmptyState title="No SOUL.md found" description="The agent is running without a custom persona file." />
      ) : (
        <Card className="p-6 overflow-x-auto">
          <pre className="whitespace-pre-wrap font-mono text-[13px] text-[var(--tt-fg)]">
            {data.content}
          </pre>
        </Card>
      )}
    </div>
  );
}
