'use client';

import React, { useEffect, useRef, useState } from 'react';
import mermaid from 'mermaid';

export default function Mermaid({ chart }: { chart: string }) {
    const ref = useRef<HTMLDivElement>(null);
    const [svg, setSvg] = useState<string>('');
    const [error, setError] = useState<string>('');

    useEffect(() => {
        let mounted = true;

        const render = async () => {
            if (!chart) return;
            try {
                mermaid.initialize({
                    startOnLoad: false,
                    theme: 'dark',
                    securityLevel: 'loose'
                });

                const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`;
                const { svg } = await mermaid.render(id, chart);

                if (mounted) {
                    setSvg(svg);
                    setError('');
                }
            } catch (e: unknown) {
                console.error("Mermaid Render Error:", e);
                if (mounted) {
                    const message = e instanceof Error ? e.message : String(e);
                    setError(message);
                }
            }
        };

        render();

        return () => { mounted = false; };
    }, [chart]);

    if (error) {
        return (
            <div className="text-red-400 text-xs border border-red-900/50 p-2 rounded bg-red-900/10">
                Render Failed: {error}
                <pre className="mt-2 text-zinc-600">{chart}</pre>
            </div>
        );
    }

    if (!svg) {
        return <div className="text-zinc-500 animate-pulse text-xs">Rendering Chart...</div>;
    }

    return (
        <div
            ref={ref}
            className="mermaid-container overflow-x-auto p-4 bg-zinc-950 rounded border border-zinc-800"
            dangerouslySetInnerHTML={{ __html: svg }}
        />
    );
}
