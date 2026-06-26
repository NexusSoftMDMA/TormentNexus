"use client";

import { useResource } from "@/lib/api";
import { PageHeader, Card, EmptyState } from "@/components/ui";
import { Users } from "lucide-react";

export default function ProfilesPage() {
  const res = useResource<{ profiles: { name: string }[] }>("/hermes/profiles");
  const loading = !res.data;
  const profiles = res.data?.profiles || [];

  return (
    <div className="px-8 py-8 max-w-[1200px] mx-auto space-y-6 pb-20">
      <PageHeader
        backHref="/hermes"
        icon={<div className="h-10 w-10 grid place-items-center rounded-[var(--tt-radius)] bg-blue-500/10 border border-blue-500/30"><Users className="text-blue-500" size={20} /></div>}
        eyebrow="Hermes Agent"
        title="Profiles"
        description="Local agent profiles available in ~/.hermes/profiles"
      />
      {loading ? (
        <div className="animate-pulse h-32 bg-[var(--tt-panel)] rounded-xl" />
      ) : profiles.length === 0 ? (
        <EmptyState title="No profiles found" description="The agent is using the default global profile." />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {profiles.map(p => (
            <Card key={p.name} className="p-4 flex items-center gap-3">
              <div className="h-10 w-10 bg-[var(--tt-border)] text-[var(--tt-fg)] rounded-full flex items-center justify-center font-semibold text-lg uppercase">
                {p.name.charAt(0)}
              </div>
              <div className="font-mono text-sm text-[var(--tt-fg)] truncate">{p.name}</div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
