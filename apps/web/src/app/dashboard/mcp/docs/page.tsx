"use client";

import { useEffect, useState } from "react";
import { Card, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { FileText, Loader2, RefreshCw } from "lucide-react";
import { trpc } from '@/utils/trpc';

type DocRef = {
    title: string;
    path: string;
    description: string;
};

const DOCS: DocRef[] = [
    { title: 'STATUS', path: 'STATUS.md', description: 'Current execution and wiring status.' },
    { title: 'ROADMAP', path: 'ROADMAP.md', description: 'Upcoming implementation plan.' },
    { title: 'DETAILED_BACKLOG', path: 'docs/DETAILED_BACKLOG.md', description: 'Granular backlog and reconciliation notes.' },
    { title: 'CHANGELOG', path: 'CHANGELOG.md', description: 'Recent shipped deltas.' },
    { title: 'HANDOFF', path: 'handoff.md', description: 'Operational continuity notes.' },
];

export default function DocsDashboard() {
    const executeTool = trpc.executeTool.useMutation();
    const [activeDoc, setActiveDoc] = useState<DocRef>(DOCS[0]);
    const [content, setContent] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    const loadDoc = async (doc: DocRef) => {
        setActiveDoc(doc);
        setLoading(true);
        setError(null);

        try {
            let output: unknown;
            try {
                output = await executeTool.mutateAsync({
                    name: 'read_file',
                    args: { filePath: doc.path }
                });
            } catch {
                output = await executeTool.mutateAsync({
                    name: 'read_file',
                    args: { path: doc.path }
                });
            }

            setContent(typeof output === 'string' ? output : JSON.stringify(output, null, 2));
        } catch (e) {
            const message = e instanceof Error ? e.message : 'Unknown read error';
            setError(`Unable to load ${doc.path}: ${message}`);
            setContent('');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void loadDoc(DOCS[0]);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (
        <div className="p-8 space-y-8 h-full overflow-y-auto">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Documentation</h1>
                    <p className="text-zinc-500">
                        Live project docs loaded from workspace files
                    </p>
                </div>
                <Button variant="outline" size="sm" className="h-8" onClick={() => loadDoc(activeDoc)} disabled={loading}>
                    {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
                    <span className="ml-2">Refresh</span>
                </Button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                {/* Navigation Sidebar */}
                <div className="space-y-2">
                    <div className="font-semibold text-white px-3 py-2">Project Docs</div>
                    {DOCS.map((doc) => (
                        <NavKey
                            key={doc.path}
                            title={doc.title}
                            subtitle={doc.description}
                            active={activeDoc.path === doc.path}
                            onClick={() => loadDoc(doc)}
                        />
                    ))}
                </div>

                {/* Content Area */}
                <div className="col-span-2 space-y-8 max-w-4xl">
                    <Section title={activeDoc.title}>
                        <div className="text-xs text-zinc-500 mb-3">{activeDoc.path}</div>
                        {loading ? (
                            <div className="flex items-center gap-2 text-zinc-400">
                                <Loader2 className="h-4 w-4 animate-spin" /> Loading document...
                            </div>
                        ) : error ? (
                            <div className="rounded-md border border-rose-500/30 bg-rose-950/20 px-3 py-2 text-sm text-rose-300">
                                {error}
                            </div>
                        ) : (
                            <pre className="whitespace-pre-wrap font-mono text-xs text-zinc-300 bg-black/40 border border-zinc-800 rounded-md p-4 max-h-[70vh] overflow-auto">
                                {content || 'No content returned.'}
                            </pre>
                        )}
                    </Section>
                </div>
            </div>
        </div>
    );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
    return (
        <section>
            <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
                <FileText className="h-6 w-6 text-blue-500" />
                {title}
            </h2>
            <div className="text-zinc-300 leading-relaxed space-y-4">
                {children}
            </div>
        </section>
    );
}

function NavKey({
    title,
    subtitle,
    active,
    onClick,
}: {
    title: string;
    subtitle?: string;
    active?: boolean;
    onClick?: () => void;
}) {
    return (
        <button
            onClick={onClick}
            className={`w-full text-left px-3 py-2 rounded text-sm cursor-pointer transition-colors ${active ? 'bg-blue-600/10 text-blue-400 border-l-2 border-blue-500' : 'text-zinc-400 hover:text-white hover:bg-zinc-800'
                }`}
        >
            <div>{title}</div>
            {subtitle ? <div className="text-[10px] opacity-70 mt-0.5">{subtitle}</div> : null}
        </button>
    );
}
