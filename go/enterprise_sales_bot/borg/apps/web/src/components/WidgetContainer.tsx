'use client';

import React, { useState } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical, ChevronDown, ChevronRight, X, HelpCircle, ExternalLink } from 'lucide-react';
import { cn } from '@/lib/utils';

// Widget metadata for tooltips and documentation
const WIDGET_HELP: Record<string, { tooltip: string; docLink?: string }> = {
    'help': { tooltip: 'Quick reference guide for all dashboard features', docLink: '#help' },
    'suggestions': { tooltip: 'AI-generated proactive recommendations based on your context', docLink: '#suggestions' },
    'connection': { tooltip: 'Real-time connection status to the TormentNexus orchestrator service', docLink: '#connection' },
    'indexing': { tooltip: 'Progress of deep code intelligence indexing across your codebase', docLink: '#indexing' },
    'ingestion': { tooltip: 'Data ingestion pipeline status and memory statistics', docLink: '#ingestion' },
    'healer': { tooltip: 'Self-healing events: automatic error detection and repair log', docLink: '#healer' },
    'autonomy': { tooltip: 'Control agent permission levels (Low/Medium/High autonomy)', docLink: '#autonomy' },
    'director_chat': { tooltip: 'Direct communication interface with the Director AI agent', docLink: '#director' },
    'council': { tooltip: 'Multi-agent debate system with Architect, Product, and Critic personas', docLink: '#council' },
    'audit': { tooltip: 'Complete audit trail of all agent actions for compliance', docLink: '#audit' },
    'context': { tooltip: 'Manage pinned files and active context for AI conversations', docLink: '#context' },
    'cheatsheet': { tooltip: 'Quick reference for available slash commands', docLink: '#commands' },
    'shell': { tooltip: 'Browse and search your PowerShell/bash command history', docLink: '#shell' },
    'skills': { tooltip: 'Installed MCPM skills and available tool extensions', docLink: '#skills' },
    'squad': { tooltip: 'Parallel agent workers in isolated git worktrees', docLink: '#squad' },
    'tests': { tooltip: 'Auto-test watcher: runs tests on file changes', docLink: '#tests' },
    'sandbox': { tooltip: 'Execute code safely in Docker containers', docLink: '#sandbox' },
    'graph_1': { tooltip: 'Interactive visualization of codebase structure. Click nodes to open files.', docLink: '#graph' },
    'graph_2': { tooltip: 'Secondary knowledge graph view', docLink: '#graph' },
    'config': { tooltip: 'Real-time system configuration and Director timing settings', docLink: '#config' },
    'runner': { tooltip: 'Execute shell commands directly from the dashboard', docLink: '#runner' },
    'trace': { tooltip: 'Trace viewer for debugging agent execution flow', docLink: '#trace' },
    'traffic': { tooltip: 'Inspect MCP traffic between agents and tools', docLink: '#traffic' },
    'activity_pulse': { tooltip: 'Live timeline of system events and agent activity', docLink: '#metrics' },
    'system_health': { tooltip: 'CPU, memory, and system resource monitoring', docLink: '#metrics' },
    'latency': { tooltip: 'Response time tracking across services', docLink: '#metrics' },
    'security': { tooltip: 'Policy-based security controls and action restrictions', docLink: '#security' }
};

interface WidgetContainerProps {
    id: string;
    title: string;
    children: React.ReactNode;
    onRemove?: () => void;
    className?: string;
    defaultCollapsed?: boolean;
}

export function WidgetContainer({ id, title, children, onRemove, className, defaultCollapsed = false }: WidgetContainerProps) {
    const [isCollapsed, setIsCollapsed] = useState(defaultCollapsed);
    const [showTooltip, setShowTooltip] = useState(false);

    const help = WIDGET_HELP[id] || { tooltip: 'Dashboard widget' };

    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
        isDragging
    } = useSortable({ id });

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        zIndex: isDragging ? 50 : 'auto',
        opacity: isDragging ? 0.8 : 1,
    };

    return (
        <div
            ref={setNodeRef}
            style={style}
            className={cn(
                "flex flex-col bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm overflow-hidden transition-shadow hover:shadow-md",
                className
            )}
        >
            {/* Header / Drag Handle */}
            <div className="flex items-center justify-between px-3 py-2 bg-zinc-50 dark:bg-zinc-950 border-b border-zinc-200 dark:border-zinc-800 h-10 select-none">
                <div className="flex items-center gap-2 overflow-hidden">
                    {/* Drag Handle */}
                    <div {...attributes} {...listeners} className="cursor-grab active:cursor-grabbing text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200">
                        <GripVertical size={16} />
                    </div>
                    {/* Title */}
                    <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300 truncate">
                        {title}
                    </span>
                    {/* Help Icon with Tooltip */}
                    <div className="relative">
                        <button
                            onMouseEnter={() => setShowTooltip(true)}
                            onMouseLeave={() => setShowTooltip(false)}
                            onClick={() => setShowTooltip(!showTooltip)}
                            className="p-0.5 text-zinc-400 hover:text-blue-500 rounded transition-colors"
                            aria-label="Show widget help"
                        >
                            <HelpCircle size={14} />
                        </button>
                        {showTooltip && (
                            <div className="absolute left-0 top-6 z-50 w-64 p-3 bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl text-xs">
                                <p className="text-zinc-300 mb-2">{help.tooltip}</p>
                                {help.docLink && (
                                    <a
                                        href={help.docLink}
                                        className="flex items-center gap-1 text-blue-400 hover:text-blue-300 text-[10px]"
                                    >
                                        <ExternalLink size={10} />
                                        View documentation
                                    </a>
                                )}
                            </div>
                        )}
                    </div>
                </div>

                <div className="flex items-center gap-1">
                    <button
                        onClick={() => setIsCollapsed(!isCollapsed)}
                        className="p-1 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200 rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
                        title={isCollapsed ? "Expand widget" : "Collapse widget"}
                    >
                        {isCollapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
                    </button>
                    {onRemove && (
                        <button
                            onClick={onRemove}
                            className="p-1 text-zinc-400 hover:text-red-500 rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
                            title="Remove widget from dashboard"
                        >
                            <X size={14} />
                        </button>
                    )}
                </div>
            </div>

            {/* Content Body */}
            {!isCollapsed && (
                <div className="p-4 flex-1 overflow-auto min-h-[100px]">
                    {children}
                </div>
            )}
        </div>
    );
}
