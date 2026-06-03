"use client";

import Link from 'next/link';
import { Card, CardContent } from "@tormentnexus/ui";
import { Library, FileCode, Hammer, Brain, BookOpenText, Lightbulb, Layers, ScrollText, ExternalLink, Loader2 } from "lucide-react";
import { trpc } from '@/utils/trpc';

type ResourceSection = {
    title: string;
    description: string;
    href: string;
    icon: React.ComponentType<{ className?: string }>;
    accentClass: string;
    count?: number;
    countLabel?: string;
};

function ResourceCard({ item }: { item: ResourceSection }) {
    const Icon = item.icon;
    return (
        <Link href={item.href} className="group">
            <Card className="h-full bg-zinc-900 border-zinc-800 hover:border-zinc-600 transition-colors">
                <CardContent className="p-5 space-y-3">
                    <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                            <Icon className={`h-5 w-5 ${item.accentClass}`} />
                            <span className="text-sm font-semibold text-white">{item.title}</span>
                        </div>
                        {item.count != null && (
                            <span className="text-xs text-zinc-500 bg-zinc-800 px-2 py-0.5 rounded-full">
                                {item.count} {item.countLabel ?? 'items'}
                            </span>
                        )}
                    </div>
                    <p className="text-sm text-zinc-500 leading-relaxed">{item.description}</p>
                    <div className="flex items-center gap-1 text-xs text-zinc-500 group-hover:text-zinc-300 transition-colors">
                        <span>Open</span>
                        <ExternalLink className="h-3 w-3" />
                    </div>
                </CardContent>
            </Card>
        </Link>
    );
}

export default function LibraryDashboard() {
    const scriptsQuery = trpc.savedScripts.list.useQuery();
    const skillsQuery = trpc.skills.list.useQuery();

    // Normalize counts safely.
    const scriptCount = Array.isArray(scriptsQuery.data) ? scriptsQuery.data.length : undefined;
    const skillCount = Array.isArray(skillsQuery.data) ? (skillsQuery.data as unknown[]).length : undefined;

    const sections: ResourceSection[] = [
        {
            title: "Saved Scripts",
            description: "Reusable automation scripts for common TormentNexus operations and workflows.",
            href: "/dashboard/mcp/scripts",
            icon: FileCode,
            accentClass: "text-blue-400",
            count: scriptCount,
            countLabel: "scripts",
        },
        {
            title: "Skills",
            description: "Curated skill bundles that extend TormentNexus's reasoning and action capabilities.",
            href: "/dashboard/skills",
            icon: Hammer,
            accentClass: "text-orange-400",
            count: skillCount,
            countLabel: "skills",
        },
        {
            title: "Tool Sets",
            description: "Named collections of MCP tools pre-grouped for specific workflow contexts.",
            href: "/dashboard/mcp/tool-sets",
            icon: Layers,
            accentClass: "text-cyan-400",
        },
        {
            title: "Memory Bank",
            description: "Searchable observations, prompts, and session summaries persisted by TormentNexus.",
            href: "/dashboard/memory",
            icon: Brain,
            accentClass: "text-purple-400",
        },
        {
            title: "Plans",
            description: "Structured reasoning plans and goal decompositions generated or stored by TormentNexus.",
            href: "/dashboard/plans",
            icon: Lightbulb,
            accentClass: "text-yellow-400",
        },
        {
            title: "Manual",
            description: "Operator documentation, usage guides, and feature reference for TormentNexus.",
            href: "/dashboard/manual",
            icon: BookOpenText,
            accentClass: "text-emerald-400",
        },
        {
            title: "Chronicle",
            description: "Git commit history and working-tree status for the active TormentNexus workspace.",
            href: "/dashboard/chronicle",
            icon: ScrollText,
            accentClass: "text-violet-400",
        },
        {
            title: "Architecture",
            description: "System design diagrams, component topology, and architectural documentation.",
            href: "/dashboard/architecture",
            icon: Library,
            accentClass: "text-rose-400",
        },
    ];

    const isLoading = scriptsQuery.isLoading || skillsQuery.isLoading;

    return (
        <div className="p-8 space-y-8">
            <div className="flex items-start justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <Library className="h-8 w-8 text-amber-500" />
                        Resource Library
                    </h1>
                    <p className="text-zinc-500 mt-2 max-w-2xl">
                        Central hub for scripts, skills, tool sets, memory, plans, and documentation — all the reusable resources that power TormentNexus workflows.
                    </p>
                </div>
                {isLoading && (
                    <Loader2 className="h-5 w-5 animate-spin text-zinc-600 mt-1" />
                )}
            </div>

            <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {sections.map(section => (
                    <ResourceCard key={section.href} item={section} />
                ))}
            </div>
        </div>
    );
}
