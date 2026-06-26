
"use client";
import { trpc } from "../utils/trpc";
import { useState } from "react";

type SkillListItem = {
    id: string;
    name: string;
    description: string;
    content: string;
    path: string;
};

function normalizeSkillList(value: unknown): SkillListItem[] {
    if (!Array.isArray(value)) {
        return [];
    }

    return value.filter((entry): entry is SkillListItem => {
        if (!entry || typeof entry !== "object") {
            return false;
        }

        const skill = entry as Partial<SkillListItem>;
        return (
            typeof skill.id === "string" &&
            typeof skill.name === "string" &&
            typeof skill.description === "string" &&
            typeof skill.content === "string" &&
            typeof skill.path === "string"
        );
    });
}

function extractSkillContent(value: unknown): string {
    if (!value || typeof value !== "object") {
        return "No content.";
    }

    const record = value as Record<string, unknown>;
    if (!Array.isArray(record.content) || record.content.length === 0) {
        return "No content.";
    }

    const first = record.content[0];
    if (!first || typeof first !== "object") {
        return "No content.";
    }

    const firstRecord = first as Record<string, unknown>;
    return typeof firstRecord.text === "string" ? firstRecord.text : "No content.";
}

export function SkillsViewer() {
    const { data: rawSkills, isLoading } = trpc.skills.list.useQuery();
    const skills = normalizeSkillList(rawSkills);
    const [selectedSkill, setSelectedSkill] = useState<string | null>(null);

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 shadow-xl w-full">
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                    📚 Skills Library
                </h2>
                <div className="text-xs text-zinc-500 uppercase font-bold tracking-wider">
                    {skills.length} Available
                </div>
            </div>

            {isLoading && <div className="text-zinc-500 animate-pulse">Loading skills...</div>}

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {skills.map((skill) => (
                    <div
                        key={skill.name}
                        onClick={() => setSelectedSkill(skill.name === selectedSkill ? null : skill.name)}
                        className={`p-4 rounded-lg border cursor-pointer transition-all hover:scale-[1.02] ${selectedSkill === skill.name
                            ? 'bg-blue-900/20 border-blue-500/50 shadow-[0_0_15px_rgba(59,130,246,0.2)]'
                            : 'bg-zinc-800/50 border-zinc-700/50 hover:bg-zinc-800 hover:border-zinc-600'
                            }`}
                    >
                        <div className="font-bold text-zinc-200 mb-1 flex items-center gap-2">
                            <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                            {skill.name}
                        </div>
                        <div className="text-xs text-zinc-400 line-clamp-2">
                            {skill.description}
                        </div>
                    </div>
                ))}
            </div>

            {selectedSkill && (
                <div className="mt-6 p-4 bg-black/50 border border-zinc-800 rounded-lg text-sm font-mono text-zinc-300">
                    <SkillDetails name={selectedSkill} />
                </div>
            )}
        </div>
    );
}

function SkillDetails({ name }: { name: string }) {
    const details = trpc.skills.read.useQuery({ name });

    if (details.isLoading) return <div className="text-zinc-500">Loading definition...</div>;

    const content = extractSkillContent(details.data);

    return (
        <div className="whitespace-pre-wrap">
            {content}
        </div>
    );
}
