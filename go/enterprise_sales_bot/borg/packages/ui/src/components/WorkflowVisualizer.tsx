'use client';

import React, { useCallback, useMemo } from 'react';
import ReactFlow, {
    Background,
    Controls,
    Edge,
    Node,
    useNodesState,
    useEdgesState,
    Position,
    MarkerType,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Card } from './ui/card';

interface WorkflowVisualizerProps {
    data: { nodes: any[]; edges: any[] };
    activeNodeId?: string;
    className?: string;
}

export function WorkflowVisualizer({ data, activeNodeId, className }: WorkflowVisualizerProps) {
    // Transform data to ReactFlow format
    const initialNodes: Node[] = useMemo(() => {
        return data.nodes.map((node, index) => ({
            id: node.id,
            position: { x: 100 + (index % 3) * 200, y: 100 + Math.floor(index / 3) * 150 }, // Simple auto-layout fallback
            data: { label: node.label || node.id },
            style: {
                background: node.id === activeNodeId ? '#ecfdf5' : '#fff',
                border: node.id === activeNodeId ? '2px solid #10b981' : '1px solid #777',
                fontWeight: node.id === activeNodeId ? 'bold' : 'normal',
                borderRadius: '8px',
                padding: '10px',
                width: 150,
            },
            type: node.type === 'checkpoint' ? 'output' : 'default' // Just mapping for shape
        }));
    }, [data.nodes, activeNodeId]);

    const initialEdges: Edge[] = useMemo(() => {
        return data.edges.map((edge) => ({
            id: edge.id,
            source: edge.source,
            target: edge.target === 'dynamic' ? 'unknown' : edge.target, // Handle dynamic targets
            animated: edge.animated || edge.target === 'dynamic',
            label: edge.label,
            markerEnd: {
                type: MarkerType.ArrowClosed,
            },
            style: { stroke: '#888' },
        }));
    }, [data.edges]);

    const [nodes, , onNodesChange] = useNodesState(initialNodes);
    const [edges, , onEdgesChange] = useEdgesState(initialEdges);

    // Sync with props if they change (simple effect for now, could be better)
    React.useEffect(() => {
        // Force layout update or similar if needed. 
        // For now initialNodes handles it on mount/prop change.
    }, [data, activeNodeId]);

    return (
        <div className={`w-full h-[500px] border rounded-md bg-white dark:bg-zinc-950 ${className}`}>
            <ReactFlow
                nodes={initialNodes} // Controlled-ish
                edges={initialEdges}
                fitView
            >
                <Background />
                <Controls />
            </ReactFlow>
        </div>
    );
}
