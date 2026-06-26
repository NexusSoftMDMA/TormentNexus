"use client";

import React from "react";
import { trpc } from "@/utils/trpc";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@tormentnexus/ui";
import { Badge } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@tormentnexus/ui";
import { ScrollArea } from "@tormentnexus/ui";
import { Bot, Wrench, Server, Cpu, Activity, Zap, Loader2, RefreshCw, AlertTriangle } from "lucide-react";

interface ToolItem {
    uuid: string;
    name: string;
    description?: string;
    serverId?: string;
}

interface ServerItem {
    id: string;
    name: string;
    status?: string;
    transport?: string;
}

function normalizeArray<T>(value: unknown): T[] {
    if (!Array.isArray(value)) return [];
    return value as T[];
}

export default function SuperAssistantDashboardPage() {
    const toolsQuery = trpc.tools.list.useQuery();
    const serversQuery = trpc.mcpServers.list.useQuery();
    const skillsQuery = trpc.skills.list.useQuery();

    const rawTools = toolsQuery.data;
    const rawServers = serversQuery.data;
    const rawSkills = skillsQuery.data;

    const tools = normalizeArray<ToolItem>(rawTools);
    const servers = normalizeArray<ServerItem>(rawServers);
    const skills = normalizeArray<{ id: string; name: string; description: string }>(rawSkills);

    const activeServers = servers.filter((s) => s.status === "connected" || s.status === "active");
    const hasErrors = toolsQuery.isError || serversQuery.isError || skillsQuery.isError;
    const isLoading = toolsQuery.isLoading || serversQuery.isLoading || skillsQuery.isLoading;

    const handleRefresh = async () => {
        await Promise.all([
            toolsQuery.refetch(),
            serversQuery.refetch(),
            skillsQuery.refetch(),
        ]);
    };

    return (
        <div className="w-full h-full flex flex-col">
            <div className="p-4 border-b border-gray-800 flex justify-between items-center bg-gray-900">
                <div>
                    <h1 className="text-xl font-bold text-white flex items-center gap-2">
                        <Bot className="w-5 h-5 text-purple-400" /> MCP SuperAssistant
                    </h1>
                    <p className="text-gray-400 text-sm">
                        System-wide MCP capability overview — {tools.length} tools, {servers.length} servers, {skills.length} skills
                    </p>
                </div>
                <div className="flex gap-2">
                    {hasErrors ? (
                        <Badge variant="outline" className="border-rose-600 text-rose-400">
                            <AlertTriangle className="w-3 h-3 mr-1" /> Degraded
                        </Badge>
                    ) : null}
                    {isLoading ? (
                        <Badge variant="outline" className="border-blue-600 text-blue-400">
                            <Loader2 className="w-3 h-3 mr-1 animate-spin" /> Loading
                        </Badge>
                    ) : null}
                    <Badge variant="outline" className="border-green-600 text-green-400">
                        <Activity className="w-3 h-3 mr-1" /> {activeServers.length} Active
                    </Badge>
                    <Badge variant="outline" className="border-purple-600 text-purple-400">
                        <Zap className="w-3 h-3 mr-1" /> {tools.length} Tools
                    </Badge>
                    <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleRefresh}>
                        <RefreshCw className="w-3 h-3 mr-1" /> Refresh
                    </Button>
                </div>
            </div>

            <div className="flex-1 p-6 overflow-auto">
                {hasErrors ? (
                    <div className="mb-4 rounded-md border border-rose-500/30 bg-rose-950/20 px-3 py-2 text-xs text-rose-300 flex items-center justify-between gap-3">
                        <span>One or more data sources failed to load. Displaying partial results where available.</span>
                        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleRefresh}>
                            Retry fetch
                        </Button>
                    </div>
                ) : null}
                <Tabs defaultValue="overview" className="w-full">
                    <TabsList>
                        <TabsTrigger value="overview">Overview</TabsTrigger>
                        <TabsTrigger value="tools">Tools ({tools.length})</TabsTrigger>
                        <TabsTrigger value="servers">Servers ({servers.length})</TabsTrigger>
                        <TabsTrigger value="skills">Skills ({skills.length})</TabsTrigger>
                    </TabsList>

                    <TabsContent value="overview" className="mt-4">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2 text-sm">
                                        <Server className="w-4 h-4 text-blue-400" /> MCP Servers
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold text-white">{servers.length}</div>
                                    <p className="text-xs text-gray-400 mt-1">{activeServers.length} connected</p>
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2 text-sm">
                                        <Wrench className="w-4 h-4 text-amber-400" /> Available Tools
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold text-white">{tools.length}</div>
                                    <p className="text-xs text-gray-400 mt-1">across all servers</p>
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader>
                                    <CardTitle className="flex items-center gap-2 text-sm">
                                        <Cpu className="w-4 h-4 text-green-400" /> Skills
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-3xl font-bold text-white">{skills.length}</div>
                                    <p className="text-xs text-gray-400 mt-1">assimilated capabilities</p>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    <TabsContent value="tools" className="mt-4">
                        <Card>
                            <CardHeader>
                                <CardTitle>Tool Registry</CardTitle>
                                <CardDescription>All MCP tools available to the agent</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <ScrollArea className="h-[500px]">
                                    <div className="space-y-2">
                                        {tools.map((tool) => (
                                            <div key={tool.uuid} className="flex items-center justify-between p-3 rounded border border-zinc-800 bg-zinc-900/50">
                                                <div>
                                                    <div className="font-mono text-sm text-white">{tool.name}</div>
                                                    <div className="text-xs text-gray-400 mt-1">{tool.description || "No description"}</div>
                                                </div>
                                                {tool.serverId && <Badge variant="outline" className="text-xs">{tool.serverId}</Badge>}
                                            </div>
                                        ))}
                                        {tools.length === 0 && <p className="text-gray-500 italic">No tools registered yet.</p>}
                                    </div>
                                </ScrollArea>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    <TabsContent value="servers" className="mt-4">
                        <Card>
                            <CardHeader>
                                <CardTitle>MCP Server Pool</CardTitle>
                                <CardDescription>Connected and configured MCP servers</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    {servers.map((server) => (
                                        <div key={server.id} className="flex items-center justify-between p-3 rounded border border-zinc-800 bg-zinc-900/50">
                                            <div>
                                                <div className="font-medium text-white">{server.name}</div>
                                                <div className="text-xs text-gray-400">{server.transport || "stdio"}</div>
                                            </div>
                                            <Badge variant={server.status === "connected" ? "default" : "outline"}>
                                                {server.status || "unknown"}
                                            </Badge>
                                        </div>
                                    ))}
                                    {servers.length === 0 && <p className="text-gray-500 italic">No MCP servers configured.</p>}
                                </div>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    <TabsContent value="skills" className="mt-4">
                        <Card>
                            <CardHeader>
                                <CardTitle>Skill Library</CardTitle>
                                <CardDescription>Assimilated capabilities</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    {skills.map((skill) => (
                                        <div key={skill.id} className="flex items-center justify-between p-3 rounded border border-zinc-800 bg-zinc-900/50">
                                            <div>
                                                <div className="font-medium text-white">{skill.name}</div>
                                                <div className="text-xs text-gray-400">{skill.description}</div>
                                            </div>
                                            <Badge variant="outline" className="text-green-400">Active</Badge>
                                        </div>
                                    ))}
                                    {skills.length === 0 && <p className="text-gray-500 italic">No skills assimilated.</p>}
                                </div>
                            </CardContent>
                        </Card>
                    </TabsContent>
                </Tabs>
            </div>
        </div>
    );
}
