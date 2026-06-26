/**
 * TORMENTNEXUS MCP Router - CLI Integration
 *
 * Fixing webui imports to use correct relative paths to core services
 */

'use client';

import React from 'react';
import { useEffect, useState, useCallback } from 'react';
import { MCPRouterStats } from './status-cards';
import * as MCPCommands from './mcp-commands';

/**
 * MCPRouterPage Component
 *
 * Main dashboard page for Ultimate MCP Router
 */
export default function MCPRouterPage() {
    const [loading, setLoading] = useState(true);
    const [registryStats, setRegistryStats] = useState<{
        totalServers: number;
        installedServers: number;
        categories: number;
    }>({ totalServers: 0, installedServers: 0, categories: 0 });
    
    const [sessionStats, setSessionStats] = useState<{
        totalSessions: number;
        running: number;
        stopped: number;
        error: number;
        totalClients: number;
    }>({ totalSessions: 0, running: 0, stopped: 0, error: 0, totalClients: 0 });

    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState<any[]>([]);
    const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState<'registry' | 'sessions' | 'config'>('registry');
    const [commandOutput, setCommandOutput] = useState<string>('');

    // Data loading effect
    useEffect(() => {
        async function loadData() {
            try {
                setLoading(true);
                
                const statsRes = await fetch('/api/mcp-router/stats', { method: 'POST' });
                const statsData = await statsRes.json();
                
                const sessionsRes = await fetch('/api/mcp-router/session-stats', { method: 'POST' });
                const sessionsData = await sessionsRes.json();
                
                if (statsData.success) {
                    const stats = JSON.parse(statsData.result);
                    setRegistryStats({
                        totalServers: stats.totalServers || 0,
                        installedServers: stats.installedServers || 0,
                        categories: stats.categories || 0
                    });
                }
                
                if (sessionsData.success) {
                    const stats = JSON.parse(sessionsData.result);
                    setSessionStats({
                        totalSessions: stats.totalSessions || 0,
                        running: stats.running || 0,
                        stopped: stats.stopped || 0,
                        error: stats.error || 0,
                        totalClients: stats.totalClients || 0
                    });
                }
                
                setLoading(false);
            } catch (error) {
                console.error('Failed to load MCP Router data:', error);
                setLoading(false);
            }
        }
        
        loadData();
    }, []);

    // Search handler
    const handleSearch = async () => {
        if (!searchQuery.trim()) {
            setSearchResults([]);
            return;
        }
        
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/search', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: searchQuery, category: selectedCategory || undefined })
            });
            const data = await res.json();
            
            if (data.success) {
                const results = JSON.parse(data.result || '{}');
                setSearchResults(results);
            }
            setLoading(false);
        } catch (error) {
            console.error('Search failed:', error);
            setLoading(false);
        }
    };

    // Registry commands
    const handleDiscover = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/discover', { method: 'POST' });
            const data = await res.json();
            
            if (data.success) {
                const result = data.result;
                setCommandOutput(result);
                
                if (result) {
                    const stats = JSON.parse(result);
                    if (stats.totalServers !== undefined) {
                        setRegistryStats(prev => ({ 
                            ...prev, 
                            totalServers: stats.totalServers 
                        }));
                    }
                }
            }
            setLoading(false);
        } catch (error) {
            console.error('Discover failed:', error);
            setLoading(false);
        }
    };

    const handleInstall = async (serverName: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/install', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: serverName })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Install failed:', error);
            setLoading(false);
        }
    };

    const handleUninstall = async (serverId: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/uninstall', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ serverId })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Uninstall failed:', error);
            setLoading(false);
        }
    };

    const handleCheckUpdates = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/check-updates', { method: 'POST' });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Check updates failed:', error);
            setLoading(false);
        }
    };

    const handleUpdate = async (serverId: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/update', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ serverId })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Update failed:', error);
            setLoading(false);
        }
    };

    const handleHealthCheck = async (serverId: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/health', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ serverId })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Health check failed:', error);
            setLoading(false);
        }
    };

    // Configuration commands
    const handleDetectConfigs = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/detect-configs', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ recursive: false })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Detect configs failed:', error);
            setLoading(false);
        }
    };

    const handleImportConfigs = async (files: string[]) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/import-configs', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ files })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Import configs failed:', error);
            setLoading(false);
        }
    };

    const handleExportConfigs = async (format: 'tormentnexus' | 'claude' | 'openai' | 'google') => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/export-configs', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ format })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Export configs failed:', error);
            setLoading(false);
        }
    };

    // Session commands
    const handleInitializeSessions = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/init-sessions', { method: 'POST' });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Initialize sessions failed:', error);
            setLoading(false);
        }
    };

    const getSessionStats = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/session-stats', { method: 'POST' });
            const data = await res.json();
            
            if (data.success) {
                const stats = JSON.parse(data.result);
                setSessionStats({
                    totalSessions: stats.totalSessions || 0,
                    running: stats.running || 0,
                    stopped: stats.stopped || 0,
                    error: stats.error || 0,
                    totalClients: stats.totalClients || 0
                });
            }
            
            setLoading(false);
        } catch (error) {
            console.error('Get session stats failed:', error);
            setLoading(false);
        }
    };

    const listSessions = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/list-sessions', { method: 'POST' });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('List sessions failed:', error);
            setLoading(false);
        }
    };

    const startSession = async (serverId: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/start-session', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ serverId })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Start session failed:', error);
            setLoading(false);
        }
    };

    const stopSession = async (serverId: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/stop-session', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ serverId })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Stop session failed:', error);
            setLoading(false);
        }
    };

    const restartSession = async (serverId: string) => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/restart-session', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ serverId })
            });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Restart session failed:', error);
            setLoading(false);
        }
    };

    const shutdownSessions = async () => {
        try {
            setLoading(true);
            const res = await fetch('/api/mcp-router/shutdown-sessions', { method: 'POST' });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Shutdown sessions failed:', error);
            setLoading(false);
        }
    };

    const getSessionMetrics = async (serverName: string) => {
        try {
            setLoading(true);
            const res = await fetch(`/api/mcp-router/session-metrics/${serverName}`, { method: 'GET' });
            const data = await res.json();
            setCommandOutput(data);
            setLoading(false);
        } catch (error) {
            console.error('Get session metrics failed:', error);
            setLoading(false);
        }
    };

    const categories = ['file-system', 'database', 'development', 'api', 'ai-ml', 'utility', 'productivity'];

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-50 via-gray-900 text-gray-900">
            <div className="container mx-auto px-4 py-8">
                <h1 className="text-4xl font-bold text-white mb-2">
                    Ultimate MCP Router
                </h1>
                <p className="text-gray-400 text-lg mb-6">
                    Discover, install, and manage MCP servers from 100+ registries with real-time updates
                </p>

                <MCPRouterStats
                    registryStats={registryStats}
                    sessionStats={sessionStats}
                />

                <div className="mb-6 border-b border-gray-700">
                    <div className="flex space-x-1" role="tablist">
                        <button
                            onClick={() => setActiveTab('registry')}
                            className={`px-4 py-3 text-sm font-medium transition-colors ${
                                activeTab === 'registry'
                                    ? 'text-blue-400 border-b-2 border-blue-400'
                                    : 'text-gray-400 border-b-2 border-transparent hover:text-gray-300'
                            }`}
                        >
                            Registry
                        </button>
                        <button
                            onClick={() => setActiveTab('sessions')}
                            className={`px-4 py-3 text-sm font-medium transition-colors ${
                                activeTab === 'sessions'
                                    ? 'text-blue-400 border-b-2 border-blue-400'
                                    : 'text-gray-400 border-b-2 border-transparent hover:text-gray-300'
                            }`}
                        >
                            Sessions
                        </button>
                        <button
                            onClick={() => setActiveTab('config')}
                            className={`px-4 py-3 text-sm font-medium transition-colors ${
                                activeTab === 'config'
                                    ? 'text-blue-400 border-b-2 border-blue-400'
                                    : 'text-gray-400 border-b-2 border-transparent hover:text-gray-300'
                            }`}
                        >
                            Configuration
                        </button>
                    </div>
                </div>

                {activeTab === 'registry' && (
                    <>
                        <div className="flex gap-4 mb-6">
                            <input
                                type="text"
                                placeholder="Search MCP servers..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                                className="flex-1 bg-gray-800 text-white rounded-lg px-4 py-3 border border-gray-700 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none"
                            />
                            <button
                                onClick={handleDiscover}
                                disabled={loading}
                                className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg font-medium transition-colors"
                            >
                                {loading ? 'Discovering...' : 'Discover All'}
                            </button>
                        </div>

                        <div className="flex gap-2 mb-6">
                            <button
                                onClick={() => setSelectedCategory(null)}
                                className="px-3 py-2 rounded-md text-sm font-medium bg-blue-600 text-white"
                            >
                                All
                            </button>
                            {categories.map((cat) => (
                                <button
                                    key={cat}
                                    onClick={() => setSelectedCategory(selectedCategory === cat ? null : cat)}
                                    className={`px-3 py-2 rounded-md text-sm font-medium bg-gray-700 text-gray-300 hover:bg-gray-600`}
                                >
                                    {cat}
                                </button>
                            ))}
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {searchResults.length > 0 ? (
                            searchResults.map((server: any) => (
                                <div key={server.name || server.serverId} className="bg-gray-800 rounded-lg p-6 border border-gray-700 hover:border-gray-600 transition-colors">
                                    <h3 className="text-lg font-semibold text-white">
                                        {server.name}
                                    </h3>
                                    <p className="text-gray-400 text-sm mb-2">
                                        {server.description}
                                    </p>
                                    <div className="flex items-center gap-2 mb-3">
                                        <span className="text-gray-400">
                                            Category: {server.category}
                                        </span>
                                        {server.rating && (
                                            <span className="ml-2 text-yellow-400">
                                                {server.rating}
                                            </span>
                                        )}
                                        <span className="text-gray-400">
                                            Source: {server.source}
                                        </span>
                                    </div>
                                    <button
                                        onClick={() => server.installed ? handleUninstall(server.serverId) : handleInstall(server.name)}
                                        disabled={loading}
                                        className="px-3 py-2 rounded-md text-sm font-medium bg-blue-600 text-white"
                                    >
                                        {loading ? 'Processing...' : server.installed ? 'Uninstall' : 'Install'}
                                    </button>
                                    {server.installed && (
                                        <button
                                            onClick={() => handleHealthCheck(server.serverId)}
                                            disabled={loading}
                                            className="px-3 py-2 rounded-md text-sm font-medium bg-gray-700 text-gray-300 hover:bg-gray-600 ml-2"
                                        >
                                            Health
                                        </button>
                                    )}
                                </div>
                            ))
                        ) : (
                            <div className="col-span-3 p-8 text-center text-gray-500">
                                {searchQuery ? 'No servers found matching your search.' : 'Search for servers or click Discover All to get started.'}
                            </div>
                        )}
                    </div>
                    </>
                )}

                {activeTab === 'sessions' && (
                    <div className="space-y-6">
                        <div className="flex gap-4 mb-6">
                            <button
                                onClick={handleInitializeSessions}
                                disabled={loading}
                                className="flex-1 bg-green-600 hover:bg-green-700 text-white px-6 py-3 rounded-lg font-medium transition-colors"
                            >
                                {loading ? 'Initializing...' : 'Initialize Sessions'}
                            </button>
                            <button
                                onClick={shutdownSessions}
                                disabled={loading}
                                className="flex-1 bg-red-600 hover:bg-red-700 text-white px-6 py-3 rounded-lg font-medium transition-colors"
                            >
                                {loading ? 'Shutting down...' : 'Shutdown All Sessions'}
                            </button>
                        </div>

                        <div className="bg-gray-800 rounded-lg overflow-hidden">
                            <div className="px-6 py-4 border-b border-gray-700">
                                <div className="grid grid-cols-6 gap-4 text-sm text-gray-400">
                                    <div>Name</div>
                                    <div>Status</div>
                                    <div>Clients</div>
                                    <div>Uptime</div>
                                    <div>Latency</div>
                                    <div>Actions</div>
                                </div>
                            </div>
                            <div className="divide-y divide-gray-700">
                                {sessionStats.running > 0 && (
                                    <div className="p-4 bg-green-900/20 border-b border-green-700/30">
                                        <div className="text-sm text-green-400">
                                            <strong>Running: {sessionStats.running}</strong>
                                        </div>
                                    </div>
                                )}
                                {sessionStats.stopped > 0 && (
                                    <div className="p-4 bg-gray-700/20 border-b border-gray-700/30">
                                        <div className="text-sm text-gray-400">
                                            <strong>Stopped: {sessionStats.stopped}</strong>
                                        </div>
                                    </div>
                                )}
                                {sessionStats.error > 0 && (
                                    <div className="p-4 bg-red-900/20 border-b border-red-700/30">
                                        <div className="text-sm text-red-400">
                                            <strong>Error: {sessionStats.error}</strong>
                                        </div>
                                    </div>
                                )}
                                {sessionStats.totalSessions === 0 && (
                                    <div className="p-4 text-center text-gray-500">
                                        No sessions. Initialize to get started.
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                )}

                {activeTab === 'config' && (
                    <div className="space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="bg-gray-800 rounded-lg p-6 border-gray-700">
                                <h3 className="text-lg font-semibold text-white mb-4">
                                    Auto-Detect Configurations
                                </h3>
                                <p className="text-gray-400 mb-4">
                                    Scan common configuration paths for MCP server configs
                                </p>
                                <button
                                    onClick={handleDetectConfigs}
                                    disabled={loading}
                                    className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-3 rounded-lg font-medium transition-colors"
                                    >
                                    {loading ? 'Scanning...' : 'Detect Configs'}
                                </button>
                            </div>

                            <div className="bg-gray-800 rounded-lg p-6 border-gray-700">
                                <h3 className="text-lg font-semibold text-white mb-4">
                                    Import Configurations
                                </h3>
                                <p className="text-gray-400 mb-4">
                                    Import MCP configurations from JSON files
                                </p>
                                <div className="mt-4 p-4 bg-gray-900/50 rounded border border-gray-700">
                                    <input
                                        type="file"
                                        multiple
                                        accept=".json"
                                        className="w-full bg-gray-700 text-white rounded px-3 py-2 border border-gray-600"
                                    />
                                </div>
                            </div>

                            <div className="bg-gray-800 rounded-lg p-6 border-gray-700">
                                <h3 className="text-lg font-semibold text-white mb-4">
                                    Export Configurations
                                </h3>
                                <p className="text-gray-400 mb-4">
                                    Export current configurations to various formats
                                </p>
                                <div className="grid grid-cols-2 gap-4 mt-4">
                                    <button
                                        onClick={() => console.log('Export: TORMENTNEXUS format')}
                                        disabled={loading}
                                        className="bg-green-600 hover:bg-green-700 text-white px-4 py-3 rounded-lg font-medium transition-colors"
                                    >
                                        {loading ? 'Exporting...' : 'TORMENTNEXUS'}
                                    </button>
                                    <button
                                        onClick={() => console.log('Export: Claude format')}
                                        disabled={loading}
                                        className="bg-purple-600 hover:bg-purple-700 text-white px-4 py-3 rounded-lg font-medium transition-colors"
                                    >
                                        {loading ? 'Exporting...' : 'Claude'}
                                    </button>
                                    <button
                                        onClick={() => console.log('Export: OpenAI format')}
                                        disabled={loading}
                                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-3 rounded-lg font-medium transition-colors"
                                    >
                                        {loading ? 'Exporting...' : 'OpenAI'}
                                    </button>
                                    <button
                                        onClick={() => console.log('Export: Google format')}
                                        disabled={loading}
                                        className="bg-red-600 hover:bg-red-700 text-white px-4 py-3 rounded-lg font-medium transition-colors"
                                    >
                                        {loading ? 'Exporting...' : 'Google'}
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {commandOutput && (
                    <div className="mt-6 bg-gray-800 rounded-lg p-6 border-gray-700">
                        <div className="flex justify-between items-center mb-4">
                            <h3 className="text-lg font-semibold text-white">
                                Command Output
                            </h3>
                            <button
                                onClick={() => setCommandOutput('')}
                                className="text-gray-400 hover:text-white text-sm font-medium"
                            >
                                Clear
                            </button>
                        </div>
                        <div className="mt-4 p-4 bg-gray-900 p-6 rounded-lg overflow-x-auto text-gray-100">
                            {commandOutput}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
