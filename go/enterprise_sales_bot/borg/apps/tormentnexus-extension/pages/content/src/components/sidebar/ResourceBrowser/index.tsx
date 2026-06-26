import React, { useEffect, useState } from 'react';
import { useResourceStore } from '@src/stores';
import { useAppStore } from '@src/stores';
import { Typography, Icon, Button } from '../ui';
import { cn } from '@src/lib/utils';
import { createLogger } from '@extension/shared/lib/logger';

// Simple Card components since they might not be in ../ui
const Card = ({ children, className }: { children: React.ReactNode, className?: string }) => (
    <div className={cn("rounded-lg border bg-card text-card-foreground shadow-sm", className)}>
        {children}
    </div>
);

const CardContent = ({ children, className }: { children: React.ReactNode, className?: string }) => (
    <div className={cn("p-6 pt-0", className)}>
        {children}
    </div>
);

const logger = createLogger('ResourceBrowser');

export const ResourceBrowser: React.FC = () => {
    const {
        resources,
        templates,
        isLoading,
        error,
        selectedResourceUri,
        resourceContent,
        setResources,
        setLoading,
        setError,
        selectResource,
        setResourceContent,
    } = useResourceStore();

    const [activeTab, setActiveTab] = useState<'resources' | 'templates'>('resources');
    const [isRefreshing, setIsRefreshing] = useState(false);

    const fetchResources = async () => {
        setIsRefreshing(true);
        setLoading(true);
        setError(null);
        try {
            // Simulate API call for now since adapter get_resources isn't mapped directly in UI yet
            const simulatedResources = [
                { uri: 'file:///example/log.txt', name: 'Application Log', mimeType: 'text/plain' },
                { uri: 'github://repo/readme.md', name: 'Repository Readme', mimeType: 'text/markdown' }
            ];
            setResources(simulatedResources, []);

            // Wait, let's try calling it properly if adapter supports it via an internal RPC or message
            logger.debug('[ResourceBrowser] Fetched resources');
        } catch (err: any) {
            setError(err.message || 'Failed to fetch resources');
            logger.error('[ResourceBrowser] Fetch error', err);
        } finally {
            setIsRefreshing(false);
            setLoading(false);
        }
    };

    useEffect(() => {
        // Implement initial fetch when tab opens
        if (resources.length === 0 && !isLoading) {
            fetchResources();
        }
    }, []);

    const handleResourceClick = async (uri: string) => {
        if (selectedResourceUri === uri) {
            selectResource(null); // Toggle off
            return;
        }

        selectResource(uri);
        setLoading(true);

        try {
            // Simulate reading resource
            // Replace with actual adapter call: adapter.callTool('read_resource', { uri }) equivalent
            const mockContent = `Mock content for ${uri}\n\n# Header\nThis is a simulation of reading an MCP resource.`;
            setResourceContent(uri, mockContent);
        } catch (err: any) {
            setError(err.message || `Failed to read ${uri}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col h-full space-y-4">
            {/* Header and Controls */}
            <div className="flex items-center justify-between">
                <Typography variant="h4" className="text-slate-800 dark:text-slate-100 font-semibold">
                    MCP Resources
                </Typography>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={fetchResources}
                    disabled={isRefreshing || isLoading}
                    className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700"
                >
                    <Icon
                        name="refresh"
                        size="sm"
                        className={cn('mr-2', isRefreshing && 'animate-spin')}
                    />
                    Refresh
                </Button>
            </div>

            {/* Error display */}
            {error && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-md text-sm flex items-start gap-2 border border-red-200 dark:border-red-800/50">
                    <Icon name="alert-triangle" size="sm" className="mt-0.5 flex-shrink-0" />
                    <span>{error}</span>
                </div>
            )}

            {/* Tabs */}
            <div className="flex border-b border-slate-200 dark:border-slate-700">
                <button
                    className={cn(
                        'px-4 py-2 text-sm font-medium transition-colors border-b-2',
                        activeTab === 'resources'
                            ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                            : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300'
                    )}
                    onClick={() => setActiveTab('resources')}
                >
                    Active Resources ({resources.length})
                </button>
                <button
                    className={cn(
                        'px-4 py-2 text-sm font-medium transition-colors border-b-2',
                        activeTab === 'templates'
                            ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                            : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300'
                    )}
                    onClick={() => setActiveTab('templates')}
                >
                    Templates ({templates.length})
                </button>
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-y-auto pr-1 space-y-3 custom-scrollbar">
                {isLoading && !isRefreshing && resources.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-32 text-slate-400">
                        <Icon name="refresh" className="animate-spin mb-2" size="lg" />
                        <Typography variant="body" className="text-sm">Loading resources...</Typography>
                    </div>
                ) : activeTab === 'resources' && resources.length === 0 ? (
                    <div className="flex flex-col items-center justify-center p-8 text-center text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800/50 rounded-lg border border-dashed border-slate-200 dark:border-slate-700">
                        <Icon name="database" size="lg" className="mb-3 opacity-50" />
                        <Typography variant="subtitle" className="font-medium text-slate-700 dark:text-slate-300 mb-1">
                            No Resources Available
                        </Typography>
                        <Typography variant="body" className="text-xs">
                            The connected MCP server did not expose any static resources.
                        </Typography>
                    </div>
                ) : activeTab === 'resources' ? (
                    <div className="space-y-2">
                        {resources.map((resource) => (
                            <Card
                                key={resource.uri}
                                className={cn(
                                    "border overflow-hidden transition-all duration-200",
                                    selectedResourceUri === resource.uri
                                        ? "border-primary-500 dark:border-primary-500 ring-1 ring-primary-500/20"
                                        : "border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600"
                                )}
                            >
                                <div
                                    className={cn(
                                        "p-3 flex items-start gap-3 cursor-pointer",
                                        selectedResourceUri === resource.uri ? "bg-primary-50/50 dark:bg-primary-900/10" : "bg-white dark:bg-slate-800"
                                    )}
                                    onClick={() => handleResourceClick(resource.uri)}
                                >
                                    <div className="mt-0.5 text-slate-400 dark:text-slate-500">
                                        <Icon name={resource.mimeType === 'application/pdf' ? 'file-text' : 'database'} size="sm" />
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <Typography variant="subtitle" className="font-medium text-sm text-slate-800 dark:text-slate-200 truncate pr-2">
                                            {resource.name || resource.uri.split('/').pop() || 'Unnamed Resource'}
                                        </Typography>
                                        <Typography variant="caption" className="text-xs text-slate-500 dark:text-slate-400 font-mono truncate block mt-0.5">
                                            {resource.uri}
                                        </Typography>
                                        {resource.description && (
                                            <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-300 mt-1 line-clamp-2">
                                                {resource.description}
                                            </Typography>
                                        )}
                                    </div>
                                    <Icon
                                        name={selectedResourceUri === resource.uri ? 'chevron-up' : 'chevron-down'}
                                        size="xs"
                                        className="text-slate-400 mt-1 flex-shrink-0"
                                    />
                                </div>

                                {/* Resource Content Viewer */}
                                {selectedResourceUri === resource.uri && (
                                    <div className="border-t border-slate-100 dark:border-slate-700/50 bg-slate-50 dark:bg-slate-900/50 p-0">
                                        {isLoading && resourceContent?.uri !== resource.uri ? (
                                            <div className="p-4 flex items-center justify-center text-slate-400 text-xs">
                                                <Icon name="refresh" className="animate-spin mr-2" size="xs" />
                                                Loading content...
                                            </div>
                                        ) : resourceContent?.uri === resource.uri ? (
                                            <div className="p-4 max-h-96 overflow-y-auto custom-scrollbar text-sm text-slate-700 dark:text-slate-300">
                                                {/* If text/markdown or plain text, render it. Might need more complex renderer later. */}
                                                {resource.mimeType === 'text/markdown' || (typeof resourceContent.content === 'string' && resourceContent.content.includes('#')) ? (
                                                    <pre className="whitespace-pre-wrap font-sans text-sm bg-white dark:bg-slate-800 p-4 rounded border border-slate-200 dark:border-slate-700 max-w-none">
                                                        {String(resourceContent.content)}
                                                    </pre>
                                                ) : (
                                                    <pre className="whitespace-pre-wrap font-mono text-xs bg-white dark:bg-slate-800 p-3 rounded border border-slate-200 dark:border-slate-700">
                                                        {typeof resourceContent.content === 'object'
                                                            ? JSON.stringify(resourceContent.content, null, 2)
                                                            : String(resourceContent.content)}
                                                    </pre>
                                                )}
                                            </div>
                                        ) : null}
                                    </div>
                                )}
                            </Card>
                        ))}
                    </div>
                ) : (
                    <div className="flex flex-col items-center justify-center p-8 text-center text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800/50 rounded-lg border border-dashed border-slate-200 dark:border-slate-700">
                        <Icon name="box" size="lg" className="mb-3 opacity-50" />
                        <Typography variant="subtitle" className="font-medium text-slate-700 dark:text-slate-300 mb-1">
                            No Templates Available
                        </Typography>
                    </div>
                )}
            </div>
        </div>
    );
};
