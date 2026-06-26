
'use client';

import { PromptLibrary } from "@tormentnexus/ui";

export default function PromptsPage() {
    return (
        <div className="p-8 h-screen flex flex-col">
            <h1 className="text-3xl font-bold mb-6 bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent">
                Prompt Registry
            </h1>
            <PromptLibrary />
        </div>
    );
}
