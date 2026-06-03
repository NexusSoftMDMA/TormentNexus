'use client';

import { Card, CardHeader, CardTitle, CardContent, CardDescription, Tabs, TabsContent, TabsList, TabsTrigger } from "@tormentnexus/ui";
import { useEffect, useState } from "react";
import { fetchSubmodulesAction, healSubmodulesAction, fetchUserLinksAction } from "./actions";
import { Button } from "@tormentnexus/ui";
import { Loader2, RefreshCw, GitCommit, Calendar, ExternalLink, Copy, Check } from "lucide-react";
import {
    normalizeSubmodules,
    normalizeUserLinks,
    summarizeSubmoduleCounts,
    type NormalizedLinkCategory,
    type NormalizedSubmoduleInfo,
} from './submodules-page-normalizers';

export default function SubmodulesPage() {
    const [submodules, setSubmodules] = useState<NormalizedSubmoduleInfo[]>([]);
    const [userLinks, setUserLinks] = useState<NormalizedLinkCategory[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        Promise.all([
            fetchSubmodulesAction(),
            fetchUserLinksAction()
        ]).then(([subs, links]) => {
            setSubmodules(normalizeSubmodules(subs));
            setUserLinks(normalizeUserLinks(links));
            setLoading(false);
        });
    }, []);

    const summaryCounts = summarizeSubmoduleCounts(submodules, userLinks);

    return (
        <div className="p-8 space-y-8 max-w-7xl mx-auto">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">System Knowledge & Modules</h1>
                    <p className="text-muted-foreground mt-1">Manage git submodules, project structure, and external resources.</p>
                </div>
                <HealButton />
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <StatusCard title="Clean" value={summaryCounts.clean} color="text-green-500" />
                <StatusCard title="Dirty" value={summaryCounts.dirty} color="text-yellow-500" />
                <StatusCard title="Missing" value={summaryCounts.missing} color="text-red-500" />
                <StatusCard title="Resources" value={summaryCounts.resources} color="text-blue-500" />
            </div>

            <Tabs defaultValue="modules" className="space-y-4">
                <TabsList>
                    <TabsTrigger value="modules">Git Submodules</TabsTrigger>
                    <TabsTrigger value="resources">User Resources</TabsTrigger>
                    <TabsTrigger value="structure">Project Structure</TabsTrigger>
                </TabsList>

                <TabsContent value="modules" className="space-y-4">
                    <Card>
                        <CardHeader>
                            <CardTitle>Repository Map ({submodules.length})</CardTitle>
                            <CardDescription>
                                Active git submodules tracked in .gitmodules
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {loading ? (
                                <div className="flex items-center justify-center p-8 text-muted-foreground gap-2">
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                    Scanning repository...
                                </div>
                            ) : (
                                <div className="rounded-md border overflow-hidden">
                                    <table className="w-full text-sm text-left">
                                        <thead className="bg-muted/50 text-muted-foreground font-medium">
                                            <tr>
                                                <th className="p-4">Module Name</th>
                                                <th className="p-4">Package</th>
                                                <th className="p-4">Version</th>
                                                <th className="p-4">Status</th>
                                                <th className="p-4">HEAD Commit</th>
                                                <th className="p-4">Last Update</th>
                                            </tr>
                                        </thead>
                                        <tbody className="divide-y">
                                            {submodules.map((sub) => (
                                                <tr key={sub.path} className="hover:bg-muted/20 transition-colors">
                                                    <td className="p-4">
                                                        <div className="font-medium text-base">{sub.path.split('/').pop()}</div>
                                                        <div className="text-muted-foreground text-xs font-mono mt-1">{sub.path}</div>
                                                    </td>
                                                    <td className="p-4">
                                                        <div className="font-mono text-xs text-blue-400">{sub.pkgName || '-'}</div>
                                                    </td>
                                                    <td className="p-4">
                                                        <div className="font-mono text-xs bg-zinc-800 px-2 py-1 rounded w-fit">{sub.version || '-'}</div>
                                                    </td>
                                                    <td className="p-4">
                                                        <StatusBadge status={sub.status} />
                                                    </td>
                                                    <td className="p-4">
                                                        <div className="flex flex-col gap-1">
                                                            <div className="flex items-center gap-2 font-mono text-xs bg-zinc-100 dark:bg-zinc-800 w-fit px-2 py-1 rounded">
                                                                <GitCommit className="h-3 w-3" />
                                                                {sub.lastCommit?.substring(0, 7) || 'N/A'}
                                                            </div>
                                                            <div className="text-xs text-muted-foreground truncate max-w-[300px]" title={sub.lastCommitMessage}>
                                                                {sub.lastCommitMessage || 'No commit message'}
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td className="p-4">
                                                        <div className="flex items-center gap-2 text-muted-foreground text-xs">
                                                            <Calendar className="h-3 w-3" />
                                                            {sub.lastCommitDate || 'Unknown'}
                                                        </div>
                                                    </td>
                                                </tr>
                                            ))}
                                            {submodules.length === 0 && (
                                                <tr>
                                                    <td colSpan={4} className="p-8 text-center text-muted-foreground">
                                                        No submodules found in .gitmodules
                                                    </td>
                                                </tr>
                                            )}
                                        </tbody>
                                    </table>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="resources" className="space-y-4">
                    <Card>
                        <CardHeader>
                            <CardTitle>User Provided Resources</CardTitle>
                            <CardDescription>
                                Archived links and tools for assimilation. Source: `docs/USER_LINKS_ARCHIVE.md`
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                                {userLinks.map((cat, idx) => (
                                    <Card key={idx} className="h-fit">
                                        <CardHeader className="pb-3">
                                            <CardTitle className="text-base font-semibold">{cat.name}</CardTitle>
                                        </CardHeader>
                                        <CardContent className="grid gap-2 text-sm">
                                            {cat.links.map((link, i) => (
                                                <ResourceLink key={i} url={link} />
                                            ))}
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="structure">
                    <Card>
                        <CardHeader>
                            <CardTitle>TormentNexus Project Structure</CardTitle>
                            <CardDescription>
                                Architectural overview of the monorepo.
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="text-sm text-muted-foreground space-y-1 font-mono">
                                <StructureItem icon="📂" name="apps/" description="Application Entry Points">
                                    <StructureItem icon="📦" name="web" description="Next.js Dashboard (Mission Control, 31+ pages)" />
                                    <StructureItem icon="📦" name="extension" description="Browser Agent (Chrome/Edge Bridge)" />
                                    <StructureItem icon="📦" name="vscode" description="VS Code Extension (Observer)" />
                                </StructureItem>

                                <StructureItem icon="📂" name="packages/" description="Shared Logic">
                                    <StructureItem icon="📦" name="core" description="Backend: Express + tRPC + WebSocket + MCP Server" />
                                    <StructureItem icon="📦" name="ai" description="LLMService, ModelSelector, provider management" />
                                    <StructureItem icon="📦" name="agents" description="Director, Council, Supervisor, orchestration" />
                                    <StructureItem icon="📦" name="tools" description="File, terminal, browser, chain executor tools" />
                                    <StructureItem icon="📦" name="memory" description="VectorStore, MemoryManager, graph memory" />
                                    <StructureItem icon="📦" name="search" description="SearchService (semantic, ripgrep, AST)" />
                                    <StructureItem icon="📦" name="ui" description="Shared React components (Tailwind)" />
                                    <StructureItem icon="📦" name="cli" description="Commander.js CLI with 11 command groups" />
                                </StructureItem>

                                <StructureItem icon="📂" name="references/" description="Submodule reference implementations (200+)" />
                                <StructureItem icon="📂" name="docs/" description="Project documentation & architecture records" />
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}

function StructureItem({ icon, name, description, children }: { icon: string, name: string, description: string, children?: React.ReactNode }) {
    return (
        <div className="ml-2">
            <div className="flex items-center gap-2 py-1">
                <span className="text-blue-500">{icon}</span>
                <span className="font-semibold text-foreground">{name}</span>
                <span className="text-xs text-muted-foreground">({description})</span>
            </div>
            {children && (
                <div className="pl-6 border-l ml-2 border-zinc-200 dark:border-zinc-800">
                    {children}
                </div>
            )}
        </div>
    )
}

function ResourceLink({ url }: { url: string }) {
    const [copied, setCopied] = useState(false);

    const copy = () => {
        navigator.clipboard.writeText(url);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    // Extract domain for display
    let display = url;
    try {
        const u = new URL(url);
        display = u.hostname + (u.pathname.length > 1 ? u.pathname : '');
        if (display.length > 40) display = display.substring(0, 37) + '...';
    } catch { }

    return (
        <div className="flex items-center justify-between group rounded p-2 hover:bg-muted/50 transition-colors">
            <a href={url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 hover:underline truncate mr-2" title={url}>
                <ExternalLink className="h-3 w-3 text-muted-foreground flex-shrink-0" />
                <span className="truncate">{display}</span>
            </a>
            <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity" onClick={copy}>
                {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
            </Button>
        </div>
    )
}

function StatusCard({ title, value, color }: { title: string, value: number, color: string }) {
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{title}</CardTitle>
            </CardHeader>
            <CardContent>
                <div className={`text-2xl font-bold ${color}`}>{value}</div>
            </CardContent>
        </Card>
    )
}

function StatusBadge({ status }: { status: string }) {
    const styles = {
        clean: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
        dirty: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
        missing: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
        error: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300"
    };
    return (
        <span className={`px-2 py-1 rounded-full text-xs font-semibold ${styles[status as keyof typeof styles] || styles.error}`}>
            {status.toUpperCase()}
        </span>
    );
}

function HealButton() {
    const [healing, setHealing] = useState(false);

    const handleHeal = async () => {
        setHealing(true);
        try {
            const res = await healSubmodulesAction();
            if (res.success) {
                alert("Submodules Healed! Refreshing page...");
                window.location.reload();
            } else {
                alert("Heal Failed: " + res.message);
            }
        } catch (e) {
            alert("Error: " + e);
        } finally {
            setHealing(false);
        }
    };

    return (
        <Button onClick={handleHeal} disabled={healing} variant="default">
            {healing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
            {healing ? "Healing..." : "Heal Submodules"}
        </Button>
    );
}

