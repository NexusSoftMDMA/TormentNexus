'use client';

import React, { useState, useEffect, useRef } from 'react';
import { PageHeader } from '@/components/PageHeader';
import { PageStatusBanner } from '@/components/PageStatusBanner';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Tabs, TabsContent, TabsList, TabsTrigger, Badge, ScrollArea } from '@tormentnexus/ui';
import { SwarmTranscript } from '@/components/swarm/SwarmTranscript';
import { DebateVisualizer } from "@/components/council/DebateVisualizer";
import { SquadsPanel } from '@tormentnexus/ui';
import { trpc } from '@/utils/trpc';
import {
    Users as UsersIcon,
    Scale as ScaleIcon,
    ArrowRightLeft as ArrowsRightLeftIcon,
    Play as PlayIcon,
    Radio as RadioIcon,
    Activity as ActivityIcon,
    Shield as ShieldIcon,
    Server as ServerIcon,
    BrainCircuit,
    GitBranch,
    Zap,
    Gavel,
    Loader2,
    MessageSquare,
    Target,
    Settings,
    Eye
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { normalizeDirectorAutonomyLevel, normalizeDirectorPlan } from './director-page-normalizers';
import { normalizeCouncilSessions, type CouncilSessionRow } from './council-page-normalizers';

interface SwarmMessage {
    id: string;
    sender: string;
    target?: string;
    type: string;
    payload: any;
    timestamp: number;
}

interface SwarmTask {
    id: string;
    description: string;
    status: 'pending' | 'running' | 'completed' | 'failed' | 'pending_approval' | 'awaiting_subtask' | 'healing' | 'throttled' | 'verifying';
    result?: string;
    priority: number;
    usage?: { tokens: number; estimatedMemory: number };
    subMissionId?: string;
    retryCount?: number;
    verifiedBy?: string;
    slashed?: boolean;
    deniedToolEvents?: Array<{
        tool: string;
        reason: string;
        timestamp: number;
    }>;
    isRedTeam?: boolean;
}

interface SwarmToolPolicy {
    allow?: string[];
    deny?: string[];
}

interface StartSwarmFeedback {
    missionId?: string;
    taskCount?: number;
    effectiveToolPolicy?: SwarmToolPolicy;
    policyWarnings?: string[];
}

interface SwarmMission {
    id: string;
    goal: string;
    status: 'active' | 'completed' | 'failed' | 'paused';
    tasks: SwarmTask[];
    parentId?: string;
    priority: number;
    usage: { tokens: number; estimatedMemory: number };
    context?: Record<string, any>;
    createdAt: string;
    updatedAt: string;
}

interface MissionRiskSummary {
    totalMissions: number;
    missionsWithDeniedEvents: number;
    totalDeniedEvents: number;
    topRiskMission: {
        missionId: string;
        deniedEventCount: number;
    } | null;
    severityScore: number;
    topDeniedTools: Array<{ tool: string; count: number }>;
    statusBreakdown: {
        active: number;
        completed: number;
        failed: number;
        paused: number;
    };
    deniedEventsLast24h: number;
    deniedEventsByHour24: Array<{ hourOffset: number; count: number }>;
}

type MissionStatusFilter = 'all' | SwarmMission['status'];

interface MissionRiskRow {
    mission: SwarmMission;
    deniedEventCount: number;
    deniedEventsLast24h: number;
    missionRiskScore: number;
}

interface MeshStatus {
    nodeId: string;
    peersCount: number;
}

interface RemoteMeshCapabilities {
    capabilities: string[];
    role?: string;
    load?: number;
    cachedAt: number;
}

interface MatchingMeshPeer {
    nodeId: string;
    capabilities: string[];
    role?: string;
    load?: number;
}

function parseCommaSeparatedList(input: string): string[] {
    return input
        .split(',')
        .map(value => value.trim())
        .filter(Boolean);
}

type DebateMode = 'standard' | 'adversarial';
type DebateTopicType = 'general' | 'mission-plan';

export default function SwarmDashboard() {
    const [activeTab, setActiveTab] = useState<'swarm' | 'squads' | 'director' | 'supervisor' | 'council' | 'missions' | 'telemetry' | 'transcript'>('swarm');

    // Telemetry State
    const [messages, setMessages] = useState<SwarmMessage[]>([]);
    const [streamStatus, setStreamStatus] = useState<'connecting' | 'online' | 'offline'>('connecting');

    // Swarm Queries
    const missionHistoryQuery = (trpc.swarm as any).getMissionHistory.useQuery(undefined, {
        refetchInterval: 5000
    });
    const missionRiskSummaryQuery = (trpc.swarm as any).getMissionRiskSummary.useQuery(undefined, {
        refetchInterval: 5000
    });
    const meshCapabilitiesQuery = (trpc.swarm as any).getMeshCapabilities.useQuery(undefined, {
        refetchInterval: 10000
    });
    const meshStatusQuery = (trpc.mesh as any).getStatus.useQuery(undefined, {
        refetchInterval: 10000
    });
    const meshPeersQuery = (trpc.mesh as any).getPeers.useQuery(undefined, {
        refetchInterval: 10000
    });

    const [masterPrompt, setMasterPrompt] = useState("");
    const [selectedModel, setSelectedModel] = useState("gpt-4o-mini");
    const [missionPriority, setMissionPriority] = useState(3);
    const [requestedTools, setRequestedTools] = useState("");
    const [policyAllowInput, setPolicyAllowInput] = useState("");
    const [policyDenyInput, setPolicyDenyInput] = useState("");
    const [selectedMeshNode, setSelectedMeshNode] = useState('');
    const [meshCapabilitySearchInput, setMeshCapabilitySearchInput] = useState('git');
    const [lastLaunchFeedback, setLastLaunchFeedback] = useState<StartSwarmFeedback | null>(null);
    const [sortMissionsByRisk, setSortMissionsByRisk] = useState(true);
    const [missionStatusFilter, setMissionStatusFilter] = useState<MissionStatusFilter>('all');
    const [showHighRiskOnly, setShowHighRiskOnly] = useState(false);
    const [riskThresholdInput, setRiskThresholdInput] = useState('50');
    const parsedRiskThreshold = Number.parseInt(riskThresholdInput, 10);
    const riskThreshold = Number.isFinite(parsedRiskThreshold) ? Math.max(0, Math.min(100, parsedRiskThreshold)) : 50;

    const missionRiskRowsQuery = (trpc.swarm as any).getMissionRiskRows.useQuery({
        statusFilter: missionStatusFilter,
        sortBy: sortMissionsByRisk ? 'risk' : 'recent',
        minRisk: showHighRiskOnly ? riskThreshold : undefined
    }, {
        refetchInterval: 5000
    });

    const requiredMeshCapabilities = parseCommaSeparatedList(meshCapabilitySearchInput);
    const remoteMeshCapabilitiesQuery = (trpc.mesh as any).queryCapabilities.useQuery({
        nodeId: selectedMeshNode,
        timeoutMs: 3000
    }, {
        enabled: !!selectedMeshNode,
        refetchInterval: selectedMeshNode ? 10000 : false
    });
    const meshCapabilityMatchQuery = (trpc.mesh as any).findPeerForCapabilities.useQuery({
        requiredCapabilities: requiredMeshCapabilities,
        timeoutMs: 3000
    }, {
        enabled: requiredMeshCapabilities.length > 0,
        refetchInterval: requiredMeshCapabilities.length > 0 ? 10000 : false
    });

    // SSE Connection
    useEffect(() => {
        const sseBase = process.env.NEXT_PUBLIC_CORE_SSE_URL || 'http://localhost:3001';
        const eventSource = new EventSource(`${sseBase}/api/mesh/stream`);

        eventSource.onopen = () => setStreamStatus('online');
        eventSource.onerror = () => setStreamStatus('offline');

        eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'CONNECTED') return;
                setMessages((prev) => [data, ...prev].slice(0, 50));
            } catch (err) {
                console.error('[Mesh] Parse Error', err);
            }
        };

        return () => eventSource.close();
    }, []);

    useEffect(() => {
        const peers = (meshPeersQuery.data ?? []) as string[];
        if (peers.length === 0) {
            if (selectedMeshNode) setSelectedMeshNode('');
            return;
        }
        if (!selectedMeshNode || !peers.includes(selectedMeshNode)) {
            setSelectedMeshNode(peers[0]);
        }
    }, [meshPeersQuery.data, selectedMeshNode]);

    const launchMutation = trpc.swarm.startSwarm.useMutation({
        onSuccess: (data: StartSwarmFeedback) => {
            setLastLaunchFeedback(data);
            missionHistoryQuery.refetch();
        }
    });

    // --- DIRECTOR PANEL STATE ---
    const { data: directorConfig } = trpc.directorConfig.get.useQuery();
    const { data: directorTaskStatus } = trpc.getTaskStatus.useQuery({});
    const { data: directorAutonomyLevel } = trpc.autonomy.getLevel.useQuery();
    const directorPlan = normalizeDirectorPlan(directorConfig, directorTaskStatus);
    const normalizedAutonomyLevel = normalizeDirectorAutonomyLevel(directorAutonomyLevel);

    // --- SUPERVISOR PANEL STATE ---
    const [supervisorGoal, setSupervisorGoal] = useState('');
    const [supervisorPlan, setSupervisorPlan] = useState<any[] | null>(null);
    const [supervisorLog, setSupervisorLog] = useState<string>('');
    const [supervisorExecuting, setSupervisorExecuting] = useState(false);

    const decomposeMutation = trpc.supervisor.decompose.useMutation();
    const superviseMutation = trpc.supervisor.supervise.useMutation();

    const handleDecompose = async () => {
        if (!supervisorGoal) return;
        try {
            const result = await decomposeMutation.mutateAsync({ goal: supervisorGoal });
            setSupervisorPlan(result);
        } catch (e: any) {
            setSupervisorLog(prev => prev + `\n[Error] Decomposition failed: ${e.message}`);
        }
    };

    const handleExecuteSupervisor = async () => {
        if (!supervisorGoal) return;
        setSupervisorExecuting(true);
        setSupervisorLog(prev => prev + `\n[System] Starting supervision of goal: "${supervisorGoal}"...`);
        try {
            const result = await superviseMutation.mutateAsync({ goal: supervisorGoal });
            setSupervisorLog(prev => prev + `\n${result}`);
        } catch (e: any) {
            setSupervisorLog(prev => prev + `\n[Error] Execution failed: ${e.message}`);
        } finally {
            setSupervisorExecuting(false);
        }
    };

    // --- COUNCIL PANEL STATE ---
    const [councilSessions, setCouncilSessions] = useState<CouncilSessionRow[]>([]);
    const [selectedCouncilSession, setSelectedCouncilSession] = useState<CouncilSessionRow | null>(null);
    const [newCouncilTopic, setNewCouncilTopic] = useState("");

    const councilListQuery = trpc.council.listSessions.useQuery(undefined, {
        refetchInterval: 5000
    });
    const runCouncilMutation = trpc.council.runSession.useMutation();

    useEffect(() => {
        if (councilListQuery.data) {
            const data = normalizeCouncilSessions(councilListQuery.data);
            setCouncilSessions(data);
            if (!selectedCouncilSession && data.length > 0) {
                setSelectedCouncilSession(data[0]);
            }
            if (selectedCouncilSession) {
                const updated = data.find(s => s.id === selectedCouncilSession.id);
                if (updated) setSelectedCouncilSession(updated);
            }
        }
    }, [councilListQuery.data, selectedCouncilSession]);

    const handleCreateCouncilSession = async () => {
        if (!newCouncilTopic) return;
        try {
            await runCouncilMutation.mutateAsync({ proposal: newCouncilTopic });
            setNewCouncilTopic("");
            councilListQuery.refetch();
        } catch (e) {
            console.error("Failed to start session:", e);
        }
    };

    const riskSummary = missionRiskSummaryQuery.data as MissionRiskSummary | undefined;
    const missionCards = (missionRiskRowsQuery.data ?? []) as MissionRiskRow[];
    const meshStatus = meshStatusQuery.data as MeshStatus | undefined;
    const meshCapabilityMap = (meshCapabilitiesQuery.data ?? {}) as Record<string, string[]>;
    const selectedPeerDetails = remoteMeshCapabilitiesQuery.data as RemoteMeshCapabilities | undefined;
    const matchingPeer = meshCapabilityMatchQuery.data as MatchingMeshPeer | null | undefined;

    return (
        <div className="flex flex-col h-full bg-slate-950 text-slate-100 p-6 space-y-6 overflow-hidden">
            <PageHeader
                title="Swarm & Agent Command Center"
                description="Consolidated interface for multi-model delegation, worktrees, debates, task planning, and orchestration."
            />
            <PageStatusBanner
                status="experimental"
                message="Autonomy surfaces, worktree spawners, and consensus debates are running live on the local TormentNexus mesh."
            />

            <div className="flex flex-wrap gap-2 border-b border-slate-800 pb-2">
                <Button
                    variant={activeTab === 'swarm' ? 'default' : 'ghost'}
                    className={activeTab === 'swarm' ? 'bg-indigo-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('swarm')}
                >
                    <ServerIcon className="w-4 h-4 mr-2" /> Swarm & Mesh
                </Button>
                <Button
                    variant={activeTab === 'squads' ? 'default' : 'ghost'}
                    className={activeTab === 'squads' ? 'bg-violet-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('squads')}
                >
                    <UsersIcon className="w-4 h-4 mr-2" /> Squad Worktrees
                </Button>
                <Button
                    variant={activeTab === 'director' ? 'default' : 'ghost'}
                    className={activeTab === 'director' ? 'bg-amber-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('director')}
                >
                    <BrainCircuit className="w-4 h-4 mr-2" /> Director Office
                </Button>
                <Button
                    variant={activeTab === 'supervisor' ? 'default' : 'ghost'}
                    className={activeTab === 'supervisor' ? 'bg-emerald-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('supervisor')}
                >
                    <ShieldIcon className="w-4 h-4 mr-2" /> Supervisor Control
                </Button>
                <Button
                    variant={activeTab === 'council' ? 'default' : 'ghost'}
                    className={activeTab === 'council' ? 'bg-indigo-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('council')}
                >
                    <Gavel className="w-4 h-4 mr-2" /> Council Debates
                </Button>
                <Button
                    variant={activeTab === 'missions' ? 'default' : 'ghost'}
                    className={activeTab === 'missions' ? 'bg-blue-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('missions')}
                >
                    <Target className="w-4 h-4 mr-2" /> Missions
                </Button>
                <Button
                    variant={activeTab === 'telemetry' ? 'default' : 'ghost'}
                    className={activeTab === 'telemetry' ? 'bg-cyan-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('telemetry')}
                >
                    <RadioIcon className={`w-4 h-4 mr-2 ${streamStatus === 'online' ? 'animate-pulse text-cyan-400' : ''}`} /> Telemetry
                </Button>
                <Button
                    variant={activeTab === 'transcript' ? 'default' : 'ghost'}
                    className={activeTab === 'transcript' ? 'bg-fuchsia-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('transcript')}
                >
                    <ActivityIcon className="w-4 h-4 mr-2" /> Neural Transcript
                </Button>
            </div>

            <div className="flex-1 min-h-0 overflow-y-auto">
                <AnimatePresence mode="wait">
                    {/* SWARM & MESH */}
                    {activeTab === 'swarm' && (
                        <motion.div
                            key="swarm"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="grid grid-cols-1 lg:grid-cols-3 gap-6"
                        >
                            <Card className="col-span-1 border-slate-800 bg-slate-900 shadow-2xl">
                                <CardHeader>
                                    <CardTitle className="text-indigo-400 font-bold uppercase tracking-tighter text-lg">Swarm Settings</CardTitle>
                                    <CardDescription>Initiate parallel multi-agent missions.</CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="space-y-2">
                                        <label className="text-[10px] uppercase tracking-widest text-slate-500 font-bold">Master Directive</label>
                                        <textarea
                                            value={masterPrompt}
                                            onChange={e => setMasterPrompt(e.target.value)}
                                            className="w-full bg-slate-950 border border-slate-800 rounded-md p-2 text-sm text-white min-h-[120px] focus:border-indigo-500 outline-none"
                                        />
                                    </div>
                                    <Button
                                        className="bg-indigo-600 hover:bg-indigo-500 text-white font-bold h-12 w-full"
                                        onClick={() => (launchMutation.mutate as any)({
                                            masterPrompt,
                                            model: selectedModel,
                                            priority: missionPriority,
                                            tools: requestedTools.split(',').map(t => t.trim()).filter(Boolean),
                                            toolPolicy: {
                                                allow: policyAllowInput.split(',').map(t => t.trim()).filter(Boolean),
                                                deny: policyDenyInput.split(',').map(t => t.trim()).filter(Boolean)
                                            }
                                        })}
                                        disabled={launchMutation.isPending || !masterPrompt}
                                    >
                                        {launchMutation.isPending ? 'DECOMPOSING...' : 'INITIATE SWARM'}
                                    </Button>
                                </CardContent>
                            </Card>

                            <Card className="col-span-2 border-slate-800 bg-slate-900">
                                <CardHeader>
                                    <CardTitle className="text-sm uppercase text-slate-500">Mesh Operator Registry</CardTitle>
                                    <CardDescription>Live node status, peer capability cache, and matching.</CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="rounded border border-slate-800 bg-slate-950 p-3 space-y-2">
                                        <div className="text-[10px] uppercase tracking-widest text-slate-500 font-bold">Local Mesh Node</div>
                                        <div className="text-sm text-slate-300">
                                            Node ID: <span className="font-mono text-cyan-400">{meshStatus?.nodeId ?? 'loading...'}</span>
                                        </div>
                                        <div className="text-sm text-slate-300">
                                            Peers Count: <span className="font-mono text-emerald-400">{meshStatus?.peersCount ?? 0}</span>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </motion.div>
                    )}

                    {/* SQUAD WORKTREES */}
                    {activeTab === 'squads' && (
                        <motion.div
                            key="squads"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                        >
                            <SquadsPanel />
                        </motion.div>
                    )}

                    {/* DIRECTOR OFFICE */}
                    {activeTab === 'director' && (
                        <motion.div
                            key="director"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="grid grid-cols-1 lg:grid-cols-3 gap-6"
                        >
                            <Card className="lg:col-span-2 border-amber-500/20 bg-amber-950/5">
                                <CardHeader>
                                    <CardTitle className="flex items-center justify-between">
                                        <span>Current Strategic Goal</span>
                                        <Badge variant="outline" className={`border-amber-500 ${directorPlan.status === 'IN_PROGRESS' ? 'text-amber-500 animate-pulse' : 'text-muted-foreground'}`}>
                                            {directorPlan.status}
                                        </Badge>
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-xl font-medium mb-6 p-4 bg-background/50 rounded border border-slate-800">
                                        "{directorPlan.goal}"
                                    </div>
                                    <div className="relative border-l-2 border-muted ml-4 space-y-8 pl-8 py-2">
                                        {directorPlan.steps.length > 0 ? directorPlan.steps.map((step) => (
                                            <div key={step.id} className="relative">
                                                <div className={`absolute -left-[41px] h-4 w-4 rounded-full border-2 ${step.status === 'DONE' ? 'bg-green-500 border-green-500' :
                                                    step.status === 'RUNNING' ? 'bg-amber-500 border-amber-500 animate-ping' :
                                                        'bg-background border-muted'
                                                    }`} />
                                                <div className="flex flex-col gap-1">
                                                    <div className="font-mono text-sm text-muted-foreground uppercase">{step.action}</div>
                                                    <div className="font-medium">{step.result}</div>
                                                </div>
                                            </div>
                                        )) : (
                                            <div className="text-muted-foreground italic">No active strategic tasks.</div>
                                        )}
                                    </div>
                                </CardContent>
                            </Card>

                            <div className="space-y-6">
                                <Card className="border-slate-800 bg-slate-900">
                                    <CardHeader>
                                        <CardTitle className="text-sm uppercase text-muted-foreground">Autonomy Level</CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <div className={`flex items-center gap-2 ${normalizedAutonomyLevel === 'high' ? 'text-green-400' : 'text-yellow-400'}`}>
                                            <ShieldIcon className="h-5 w-5" />
                                            <span className="font-bold uppercase">{normalizedAutonomyLevel}</span>
                                        </div>
                                        <p className="text-xs text-muted-foreground mt-2">
                                            The Director is authorized to recruit squads and perform deep research without explicit approval.
                                        </p>
                                    </CardContent>
                                </Card>
                            </div>
                        </motion.div>
                    )}

                    {/* SUPERVISOR CONTROL */}
                    {activeTab === 'supervisor' && (
                        <motion.div
                            key="supervisor"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="space-y-6"
                        >
                            <Card className="p-6 space-y-4 border-slate-800 bg-slate-900">
                                <h2 className="text-xl font-semibold">Mission Control</h2>
                                <div className="flex gap-4">
                                    <Input
                                        placeholder="Enter a high-level goal (e.g. 'Research React 19 features')"
                                        value={supervisorGoal}
                                        onChange={(e) => setSupervisorGoal(e.target.value)}
                                        className="flex-1 bg-slate-950 border-slate-800"
                                    />
                                    <Button onClick={handleDecompose} disabled={!supervisorGoal || decomposeMutation.isPending} className="bg-slate-800 hover:bg-slate-700">
                                        {decomposeMutation.isPending ? 'Planning...' : 'Plan'}
                                    </Button>
                                    <Button onClick={handleExecuteSupervisor} disabled={!supervisorGoal || supervisorExecuting} className="bg-emerald-600 hover:bg-emerald-500">
                                        {supervisorExecuting ? 'Supervising...' : 'Execute'}
                                    </Button>
                                </div>
                            </Card>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <Card className="p-6 flex flex-col gap-4 border-slate-800 bg-slate-900">
                                    <h3 className="font-semibold border-b border-slate-800 pb-2">Proposed Plan</h3>
                                    <div className="space-y-4 max-h-[300px] overflow-y-auto">
                                        {!supervisorPlan && <div className="text-muted-foreground italic">No plan generated yet.</div>}
                                        {supervisorPlan?.map((task: any) => (
                                            <div key={task.id} className="border border-slate-800 rounded p-3 bg-slate-950">
                                                <div className="flex justify-between items-start mb-2">
                                                    <span className="font-mono text-xs bg-indigo-500/10 text-indigo-400 px-2 py-1 rounded uppercase">
                                                        {task.assignedTo}
                                                    </span>
                                                    <span className="text-xs text-muted-foreground">{task.status}</span>
                                                </div>
                                                <p className="text-sm">{task.description}</p>
                                            </div>
                                        ))}
                                    </div>
                                </Card>

                                <Card className="p-6 flex flex-col gap-4 border-slate-800 bg-slate-900">
                                    <h3 className="font-semibold border-b border-slate-800 pb-2">Execution Log</h3>
                                    <div className="h-[300px] overflow-y-auto bg-black text-green-400 font-mono text-sm p-4 rounded-md whitespace-pre-wrap">
                                        {supervisorLog || "// Ready for orders..."}
                                    </div>
                                </Card>
                            </div>
                        </motion.div>
                    )}

                    {/* COUNCIL DEBATES */}
                    {activeTab === 'council' && (
                        <motion.div
                            key="council"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="space-y-6"
                        >
                            <div className="flex justify-between items-center border-b border-slate-800 pb-4">
                                <h3 className="text-lg font-bold">Consensus Debate Loop</h3>
                                <div className="flex gap-2">
                                    <Input
                                        placeholder="Propose a topic for debate..."
                                        value={newCouncilTopic}
                                        onChange={(e) => setNewCouncilTopic(e.target.value)}
                                        className="w-80 bg-slate-950 border-slate-800"
                                    />
                                    <Button onClick={handleCreateCouncilSession} className="bg-indigo-600 hover:bg-indigo-700">
                                        Convene Session
                                    </Button>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                                <div className="space-y-4">
                                    <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Active & Recent</h4>
                                    {councilListQuery.isPending ? (
                                        <div className="flex justify-center p-8">
                                            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                        </div>
                                    ) : councilSessions.length === 0 ? (
                                        <div className="text-center p-8 border border-dashed border-slate-800 rounded-lg">
                                            <p className="text-muted-foreground">No sessions found.</p>
                                        </div>
                                    ) : (
                                        <div className="space-y-2">
                                            {councilSessions.map(session => (
                                                <div
                                                    key={session.id}
                                                    onClick={() => setSelectedCouncilSession(session)}
                                                    className={`p-4 rounded-lg border cursor-pointer transition-colors ${selectedCouncilSession?.id === session.id
                                                        ? 'bg-indigo-500/10 border-indigo-500/50'
                                                        : 'bg-slate-900 hover:bg-slate-800 border-slate-800'
                                                        }`}
                                                >
                                                    <div className="flex justify-between items-start mb-2">
                                                        <Badge variant={session.status === 'active' ? 'default' : 'secondary'} className={session.status === 'active' ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30' : ''}>
                                                            {session.status}
                                                        </Badge>
                                                        <span className="text-xs text-muted-foreground">
                                                            {new Date(session.createdAt).toLocaleTimeString()}
                                                        </span>
                                                    </div>
                                                    <h4 className="font-medium text-sm line-clamp-2 mb-2">{session.topic}</h4>
                                                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                        <MessageSquare className="h-3 w-3" />
                                                        <span>{session.opinions.length} Opinions</span>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>

                                <div className="lg:col-span-2">
                                    {selectedCouncilSession ? (
                                        <DebateVisualizer
                                            topic={selectedCouncilSession.topic}
                                            transcripts={selectedCouncilSession.opinions.map(o => ({
                                                speaker: o.agentId,
                                                text: o.content,
                                                round: o.round,
                                                vote: undefined
                                            }))}
                                            config={{
                                                rounds: selectedCouncilSession.round,
                                                status: selectedCouncilSession.status,
                                                result: selectedCouncilSession.status === 'concluded' ? "Session Concluded" : undefined
                                            }}
                                        />
                                    ) : (
                                        <div className="h-[400px] border border-dashed border-slate-800 rounded-lg flex items-center justify-center text-muted-foreground">
                                            Select a session to view details
                                        </div>
                                    )}
                                </div>
                            </div>
                        </motion.div>
                    )}

                    {/* MISSIONS */}
                    {activeTab === 'missions' && (
                        <motion.div
                            key="missions"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="space-y-4"
                        >
                            <Card className="bg-slate-900 border-slate-800">
                                <CardHeader>
                                    <CardTitle className="text-white">Active Swarm Missions</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    {missionCards.length === 0 ? (
                                        <div className="text-center text-slate-600 italic">No missions found.</div>
                                    ) : (
                                        <div className="space-y-4">
                                            {missionCards.map(row => (
                                                <div key={row.mission.id} className="p-4 border border-slate-800 rounded-lg bg-slate-950">
                                                    <div className="flex justify-between items-start mb-2">
                                                        <div className="font-semibold text-amber-400">{row.mission.goal}</div>
                                                        <div className="text-xs text-slate-500 font-mono">{row.mission.id}</div>
                                                    </div>
                                                    <div className="text-sm text-slate-400">Risk Score: {row.missionRiskScore}</div>
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </motion.div>
                    )}

                    {/* TELEMETRY */}
                    {activeTab === 'telemetry' && (
                        <motion.div
                            key="telemetry"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="h-[500px]"
                        >
                            <Card className="h-full border-slate-800 bg-slate-900 flex flex-col overflow-hidden">
                                <CardHeader className="border-b border-slate-800">
                                    <CardTitle className="text-sm uppercase text-slate-500">Live Telemetry Feed</CardTitle>
                                </CardHeader>
                                <CardContent className="flex-1 overflow-y-auto font-mono text-xs p-4 bg-black text-cyan-400 space-y-2">
                                    {messages.map((msg) => (
                                        <div key={msg.id} className="border-b border-zinc-900 pb-1">
                                            <span className="text-purple-400">[{new Date(msg.timestamp).toLocaleTimeString()}]</span>{' '}
                                            <span className="text-green-400">{msg.sender}</span>
                                            {msg.target && <span> &rarr; <span className="text-yellow-400">{msg.target}</span></span>}:{' '}
                                            <span className="text-slate-300 font-bold">{msg.type}</span>
                                            <pre className="mt-1 text-slate-400 max-w-full overflow-x-auto whitespace-pre-wrap">{JSON.stringify(msg.payload, null, 2)}</pre>
                                        </div>
                                    ))}
                                    {messages.length === 0 && <p className="text-gray-600">// Awaiting events from the mesh stream...</p>}
                                </CardContent>
                            </Card>
                        </motion.div>
                    )}

                    {/* TRANSCRIPT */}
                    {activeTab === 'transcript' && (
                        <motion.div
                            key="transcript"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                        >
                            <SwarmTranscript />
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>
        </div>
    );
}
