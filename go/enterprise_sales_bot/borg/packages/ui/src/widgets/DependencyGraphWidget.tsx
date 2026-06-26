"use client";

import React, { useEffect, useMemo } from 'react';
import ReactFlow, {
    Node,
    Edge,
    Background,
    Controls,
    MiniMap,
    useNodesState,
    useEdgesState,
    Panel
} from 'reactflow';
import 'reactflow/dist/style.css';

interface GraphData {
    nodes: Array<{ id: string; name: string; group?: string }>;
    links: Array<{ source: string; target: string }>;
}

interface DependencyGraphWidgetProps {
    data: GraphData;
    onNodeClick?: (nodeId: string) => void;
}

export function DependencyGraphWidget({ data, onNodeClick }: DependencyGraphWidgetProps) {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    useEffect(() => {
        if (!data) return;

        // Simple Random Layout (placeholder for Force Layout)
        // ideally we use dagre or d3-force, but for now we scatter them
        // or organize by group

        const newNodes: Node[] = data.nodes.map((n, i) => ({
            id: n.id,
            data: { label: n.name },
            position: { x: Math.random() * 500, y: Math.random() * 500 },
            type: 'default', // or custom
            style: {
                background: '#1a1a1a',
                color: '#fff',
                border: '1px solid #333',
                fontSize: '10px',
                width: 150
            }
        }));

        const newEdges: Edge[] = data.links.map((l, i) => ({
            id: `e-${i}`,
            source: l.source,
            target: l.target,
            animated: true,
            style: { stroke: '#555' }
        }));

        setNodes(newNodes);
        setEdges(newEdges);
    }, [data, setNodes, setEdges]);

    return (
        <div className="w-full h-[600px] border border-zinc-800 rounded-lg overflow-hidden bg-zinc-950">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                fitView
                onNodeClick={(_, node) => onNodeClick?.(node.id)}
            >
                <Background color="#222" gap={16} />
                <Controls />
                <MiniMap
                    nodeColor={() => '#333'}
                    maskColor="rgba(0, 0, 0, 0.6)"
                    className="bg-zinc-900 border border-zinc-800"
                />
                <Panel position="top-right" className="bg-black/50 p-2 rounded text-xs text-white">
                    Nodes: {nodes.length} | Edges: {edges.length}
                </Panel>
            </ReactFlow>
        </div>
    );
}
