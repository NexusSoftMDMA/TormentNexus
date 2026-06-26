'use client';

import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Card, CardHeader, CardTitle, CardContent } from './ui/card';
import { ScrollArea } from './ui/scroll-area';
import { Map } from 'lucide-react';

interface RoadmapWidgetProps {
    content: string;
}

export function RoadmapWidget({ content }: RoadmapWidgetProps) {
    return (
        <Card className="bg-zinc-900/50 border-white/10 h-full">
            <CardHeader className="pb-2">
                <CardTitle className="text-white flex items-center gap-2">
                    <Map className="h-5 w-5 text-purple-400" />
                    Strategic Roadmap
                </CardTitle>
            </CardHeader>
            <CardContent>
                <ScrollArea className="h-[300px] w-full pr-4">
                    <div className="prose prose-invert prose-sm max-w-none">
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                            {content}
                        </ReactMarkdown>
                    </div>
                </ScrollArea>
            </CardContent>
        </Card>
    );
}
