'use client';

import React from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { PromptLibrary } from './PromptLibrary';
import { SkillLibrary } from './SkillLibrary';
import { Book, Lightbulb } from 'lucide-react';

export function LibraryPage() {
    return (
        <div className="flex flex-col h-full bg-black text-white p-6">
            <h1 className="text-2xl font-bold mb-6 text-zinc-200 tracking-tight">System Library</h1>

            <Tabs defaultValue="prompts" className="flex-1 flex flex-col overflow-hidden">
                <TabsList className="bg-zinc-900 border-b border-zinc-800 w-full justify-start rounded-none p-0 h-10">
                    <TabsTrigger
                        value="prompts"
                        className="data-[state=active]:bg-zinc-800 data-[state=active]:text-blue-400 data-[state=active]:border-b-2 data-[state=active]:border-blue-500 rounded-none h-full px-6 gap-2"
                    >
                        <Lightbulb className="w-4 h-4" />
                        Prompt Library
                    </TabsTrigger>
                    <TabsTrigger
                        value="skills"
                        className="data-[state=active]:bg-zinc-800 data-[state=active]:text-emerald-400 data-[state=active]:border-b-2 data-[state=active]:border-emerald-500 rounded-none h-full px-6 gap-2"
                    >
                        <Book className="w-4 h-4" />
                        Skill Registry
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="prompts" className="flex-1 overflow-hidden mt-0 border-0 outline-none">
                    <PromptLibrary />
                </TabsContent>

                <TabsContent value="skills" className="flex-1 overflow-hidden mt-0 border-0 outline-none">
                    <SkillLibrary />
                </TabsContent>
            </Tabs>
        </div>
    );
}
