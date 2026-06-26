'use client';

import React, { useMemo, useState } from 'react';

interface Node {
    id: string;
    label: string;
    type: string;
    val: number;
    x?: number;
    y?: number;
    vx?: number;
    vy?: number;
}

interface Edge {
    source: string;
    target: string;
    value: number;
}

interface GraphData {
    nodes: Node[];
    edges: Edge[];
}

export function MemoryGraph({ data }: { data: GraphData }) {
    const [dimensions] = useState({ width: 800, height: 600 });

    // Simple state-based simulation (not real D3 force for now to keep deps low, just random scatter or circular)
    // We will improve this to a real force graph later or import a library.
    // For now, let's just arrange them in a circle or grid if small count.

    const processedNodes = useMemo(() => {
        const nodes = [...data.nodes];
        const count = nodes.length;
        const centerX = dimensions.width / 2;
        const centerY = dimensions.height / 2;
        const radius = Math.min(centerX, centerY) - 50;

        return nodes.map((node, i) => {
            const angle = (i / count) * 2 * Math.PI;
            return {
                ...node,
                x: centerX + radius * Math.cos(angle),
                y: centerY + radius * Math.sin(angle)
            };
        });

    }, [data, dimensions]);

    return (
        <div className="w-full h-full flex items-center justify-center bg-zinc-950 text-white relative">
            <svg width="100%" height="100%" viewBox={`0 0 ${dimensions.width} ${dimensions.height}`}>
                {/* Edges */}
                {data.edges.map((edge, i) => {
                    const sourceNode = processedNodes.find(n => n.id === edge.source);
                    const targetNode = processedNodes.find(n => n.id === edge.target);
                    if (!sourceNode || !targetNode) return null;

                    return (
                        <line
                            key={i}
                            x1={sourceNode.x}
                            y1={sourceNode.y}
                            x2={targetNode.x}
                            y2={targetNode.y}
                            stroke="#3f3f46"
                            strokeWidth={1}
                            opacity={0.5}
                        />
                    );
                })}

                {/* Nodes */}
                {processedNodes.map((node) => (
                    <g key={node.id}>
                        <circle
                            cx={node.x}
                            cy={node.y}
                            r={node.val * 2 + 5}
                            fill={node.type === 'topic' ? '#3b82f6' : '#10b981'}
                            className="cursor-pointer hover:fill-white transition-colors"
                        />
                        <text
                            x={node.x}
                            y={(node.y || 0) + 20}
                            textAnchor="middle"
                            fill="#a1a1aa"
                            fontSize="12px"
                            className="pointer-events-none"
                        >
                            {node.label}
                        </text>
                    </g>
                ))}
            </svg>

            {data.nodes.length === 0 && (
                <div className="absolute inset-0 flex items-center justify-center text-zinc-500">
                    No memories found. Try searching or adding some!
                </div>
            )}
        </div>
    );
}
