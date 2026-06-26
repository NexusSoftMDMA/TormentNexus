"use client";

import { useResource } from "@/lib/api";
import { PageHeader, Card, EmptyState, Badge } from "@/components/ui";
import { Wrench } from "lucide-react";

export default function ToolsPage() {
  const toolsRes = useResource<{ enabled_tools: string[] }>("/hermes/tools");
  const loading = !toolsRes.data;
  const tools = toolsRes.data?.enabled_tools || [];

  return (
    <div className="px-8 py-8 max-w-[1200px] mx-auto space-y-6 pb-20">
      <PageHeader
        backHref="/hermes"
        icon={<div className="h-10 w-10 grid place-items-center rounded-[var(--tt-radius)] bg-orange-500/10 border border-orange-500/30"><Wrench className="text-orange-500" size={20} /></div>}
        eyebrow="Hermes Agent"
        title="Configured Tools"
        description="Core CLI toolsets enabled in config.yaml"
      />
      {loading ? (
        <div className="animate-pulse h-32 bg-[var(--tt-panel)] rounded-xl" />
      ) : tools.length === 0 ? (
        <EmptyState title="No tools configured" description="All default toolsets may be active if config.yaml is empty." />
      ) : (
        <Card className="p-6">
          <div className="flex flex-wrap gap-2">
            {tools.map(t => (
              <Badge key={t} variant="neutral" size="sm">{t}</Badge>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}
