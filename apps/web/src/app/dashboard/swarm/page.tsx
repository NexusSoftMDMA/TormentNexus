'use client';

import React, { useState, useEffect } from 'react';
import { PageHeader } from '@/components/PageHeader';
import { PageStatusBanner } from '@/components/PageStatusBanner';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input } from '@borg/ui';
import { SwarmTranscript } from '@/components/swarm/SwarmTranscript';
import { trpc } from '@/utils/trpc';
import {
    Users as UsersIcon,
    Scale as ScaleIcon,
    ArrowRightLeft as ArrowsRightLeftIcon,
    Play as PlayIcon,
    Radio as RadioIcon,
    Activity as ActivityIcon,
    Shield as ShieldIcon,
    Server as ServerIcon
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

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

interface MissionRiskFacets {
    missionCount: number;
    averageRisk: number;
    maxRisk: number;
    minObservedRisk: number;
    dominantBand: 'low' | 'medium' | 'high';
    health: {
        severity: 'good' | 'warn' | 'critical';
        score: number;
        reasons: string[];
        recommendedAction: string;
        confidence: {
            score: number;
            level: 'high' | 'medium' | 'low';
            drivers: string[];
            inputs: {
                missionCount: number;
                healthReasonCount: number;
                freshnessBucket: 'fresh' | 'recent' | 'stale' | 'unknown';
                evaluatedAt: number;
            };
            components: {
                sampleSizePenalty: number;
                freshnessPenalty: number;
                signalCongestionPenalty: number;
                totalPenalty: number;
            };
            uncertaintyMargin: number;
            scoreRange: {
                min: number;
                max: number;
            };
            stability: 'stable' | 'watch' | 'volatile';
            advice: string;
            alertLevel: 'none' | 'warn' | 'critical';
            alertCount: number;
            hasCriticalAlert: boolean;
            alerts: string[];
        };
    };
    activity: {
        deniedLast24h: number;
        deniedPrev24h: number;
        deniedDelta: number;
        deniedDeltaPct: number;
        deniedTrend: 'up' | 'down' | 'flat';
    };
    freshness: {
        generatedAt: number;
        latestMissionUpdatedAt: number | null;
        latestUpdateAgeSeconds: number | null;
        freshnessBucket: 'fresh' | 'recent' | 'stale' | 'unknown';
    };
    statusDistribution: {
        counts: {
            active: number;
            completed: number;
            failed: number;
            paused: number;
        };
        percentages: {
            active: number;
            completed: number;
            failed: number;
            paused: number;
        };
    };
    bands: {
        low: number;
        medium: number;
        high: number;
    };
    bandPercentages: {
        low: number;
        medium: number;
        high: number;
    };
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
    const [activeTab, setActiveTab] = useState<'swarm' | 'debate' | 'consensus' | 'telemetry' | 'missions' | 'transcript'>('swarm');

    // Telemetry State
    const [messages, setMessages] = useState<SwarmMessage[]>([]);
    const [streamStatus, setStreamStatus] = useState<'connecting' | 'online' | 'offline'>('connecting');
    const scrollRef = React.useRef<HTMLDivElement>(null);

    // Persistence & Capabilities
    const missionHistoryQuery = (trpc.swarm as any).getMissionHistory.useQuery(undefined, {
        refetchInterval: 5000 // Poll for updates
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
    const [showDeniedOnly, setShowDeniedOnly] = useState(false);
    const [sortMissionsByRisk, setSortMissionsByRisk] = useState(true);
    const [missionStatusFilter, setMissionStatusFilter] = useState<MissionStatusFilter>('all');
    const [showHighRiskOnly, setShowHighRiskOnly] = useState(false);
    const [riskThresholdInput, setRiskThresholdInput] = useState('50');
    const parsedRiskThreshold = Number.parseInt(riskThresholdInput, 10);
    const riskThreshold = Number.isFinite(parsedRiskThreshold)
        ? Math.max(0, Math.min(100, parsedRiskThreshold))
        : 50;

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
    const missionRiskFacetsQuery = (trpc.swarm as any).getMissionRiskFacets.useQuery({
        statusFilter: missionStatusFilter,
        minRisk: showHighRiskOnly ? riskThreshold : undefined
    }, {
        refetchInterval: 5000
    });

    // Initial SSE Connection
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
            if (selectedMeshNode) {
                setSelectedMeshNode('');
            }
            return;
        }

        if (!selectedMeshNode || !peers.includes(selectedMeshNode)) {
            setSelectedMeshNode(peers[0]);
        }
    }, [meshPeersQuery.data, selectedMeshNode]);

    // Swarm State
    const [swarmPrompt, setSwarmPrompt] = useState('Build a Next.js landing page with Stripe integration and a dark mode toggle.');
    const launchMutation = trpc.swarm.startSwarm.useMutation({
        onSuccess: (data: StartSwarmFeedback) => {
            setLastLaunchFeedback(data);
            missionHistoryQuery.refetch();
        }
    });

    // Debate State
    const [debateTopic, setDebateTopic] = useState('Monorepo vs Polyrepo for enterprise scalability');
    const [debateMode, setDebateMode] = useState<DebateMode>('adversarial');
    const [debateTopicType, setDebateTopicType] = useState<DebateTopicType>('mission-plan');
    const executeDebateMutation = trpc.swarm.executeDebate.useMutation();

    // Consensus State
    const [consensusPrompt, setConsensusPrompt] = useState('What is the single most critical security vulnerability in standard JWT implementations?');
    const seekConsensusMutation = trpc.swarm.seekConsensus.useMutation();

    const handleStartDebate = async () => {
        await executeDebateMutation.mutateAsync({
            topic: debateTopic,
            proponentModel: 'claude-3-5-sonnet-20241022',
            opponentModel: 'gpt-4o',
            judgeModel: 'gemini-2.5-pro',
            rounds: 2,
            mode: debateMode,
            topicType: debateTopicType
        });
    };

    const handleSeekConsensus = async () => {
        await seekConsensusMutation.mutateAsync({
            prompt: consensusPrompt,
            models: ['claude-3-5-sonnet-20241022', 'gpt-4o', 'gemini-2.5-pro'],
            requiredAgreement: 0.66
        });
    };

    // Direct Message State
    const [dmTarget, setDmTarget] = useState('');
    const [dmPayload, setDmPayload] = useState('{"hello": "world"}');
    const sendDirectMessageMutation = trpc.swarm.sendDirectMessage.useMutation();

    const handleSendDirectMessage = async () => {
        if (!dmTarget || !dmPayload) return;
        try {
            const parsed = JSON.parse(dmPayload);
            await sendDirectMessageMutation.mutateAsync({ targetNodeId: dmTarget, payload: parsed });
            setDmPayload('{"hello": "world"}');
        } catch (e) {
            alert('Payload must be valid JSON');
        }
    };

    const riskSummary = missionRiskSummaryQuery.data as MissionRiskSummary | undefined;
    const missionCards = (missionRiskRowsQuery.data ?? []) as MissionRiskRow[];
    const meshStatus = meshStatusQuery.data as MeshStatus | undefined;
    const meshCapabilityMap = (meshCapabilitiesQuery.data ?? {}) as Record<string, string[]>;
    const selectedPeerDetails = remoteMeshCapabilitiesQuery.data as RemoteMeshCapabilities | undefined;
    const matchingPeer = meshCapabilityMatchQuery.data as MatchingMeshPeer | null | undefined;
    const missionDataLoading =
        missionRiskSummaryQuery.isLoading ||
        missionRiskRowsQuery.isLoading ||
        missionRiskFacetsQuery.isLoading;

    return (
        <div className="flex flex-col h-full bg-slate-950 text-slate-100 p-6 space-y-6 overflow-hidden">
            <PageHeader
                title="Swarm Orchestration"
                description="Horizontal multi-model delegation, adversarial debates, and consensus voting."
            />
            <PageStatusBanner
                status="experimental"
                message="Swarm multi-agent orchestration is experimental. Consensus, slashing, and adversarial debate features are under active development."
            />

            <div className="flex flex-wrap gap-2 border-b border-slate-800 pb-2">
                <Button
                    variant={activeTab === 'swarm' ? 'default' : 'ghost'}
                    className={activeTab === 'swarm' ? 'bg-indigo-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('swarm')}
                >
                    <UsersIcon className="w-4 h-4 mr-2" /> Swarm
                </Button>
                <Button
                    variant={activeTab === 'missions' ? 'default' : 'ghost'}
                    className={activeTab === 'missions' ? 'bg-amber-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('missions')}
                >
                    <ActivityIcon className="w-4 h-4 mr-2" /> Missions
                </Button>
                <Button
                    variant={activeTab === 'debate' ? 'default' : 'ghost'}
                    className={activeTab === 'debate' ? 'bg-rose-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('debate')}
                >
                    <ArrowsRightLeftIcon className="w-4 h-4 mr-2" /> Debate
                </Button>
                <Button
                    variant={activeTab === 'consensus' ? 'default' : 'ghost'}
                    className={activeTab === 'consensus' ? 'bg-emerald-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('consensus')}
                >
                    <ScaleIcon className="w-4 h-4 mr-2" /> Consensus
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
                    className={activeTab === 'transcript' ? 'bg-cyan-600' : 'text-slate-400'}
                    onClick={() => setActiveTab('transcript')}
                >
                    <ActivityIcon className="w-4 h-4 mr-2" /> Neural Transcript
                </Button>
            </div>

            <div className="flex-1 min-h-0">
                <AnimatePresence mode="wait">
                    {/* SWARM PANEL */}
                    {activeTab === 'swarm' && (
                        <motion.div
                            key="swarm"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -10 }}
                            className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-full"
                        >
                            <Card className="col-span-1 border-slate-800 bg-slate-900 shadow-2xl">
                                <CardHeader>
                                    <CardTitle className="text-indigo-400 font-bold uppercase tracking-tighter text-lg">Swarm Settings</CardTitle>
                                    <CardDescription>Split a prompt into parallel workers.</CardDescription>
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
                                    {/* ... policy inputs ... */}
                                    <Button
                                        className="bg-amber-600 hover:bg-amber-500 text-black font-bold h-12 w-full"
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
                                    <CardDescription>Live node health, peer capability cache, and capability matching.</CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="rounded border border-slate-800 bg-slate-950 p-3 space-y-2">
                                        <div className="text-[10px] uppercase tracking-widest text-slate-500 font-bold">Local Mesh Node</div>
                                        <div className="text-[10px] text-slate-300">
                                            Node: <span className="font-mono text-cyan-400">{meshStatus?.nodeId ?? 'loading...'}</span>
                                        </div>
                                        <div className="text-[10px] text-slate-300">
                                            Peers: <span className="font-mono text-emerald-400">{meshStatus?.peersCount ?? 0}</span>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </motion.div>
                    )}

                    {/* TRANSCRIPT PANEL */}
                    {activeTab === 'transcript' && (
                        <motion.div
                            key="transcript"
                            initial={{ opacity: 0, scale: 0.98 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.98 }}
                            className="h-full"
                        >
                            <SwarmTranscript />
                        </motion.div>
                    )}

                    {/* MISSIONS PANEL */}
                    {activeTab === 'missions' && (
                        <motion.div
                            key="missions"
                            initial={{ opacity: 0, x: 20 }}
                            animate={{ opacity: 1, x: 0 }}
                            exit={{ opacity: 0, x: -20 }}
                            className="flex flex-col h-full space-y-4 overflow-y-auto"
                        >
                             <Card className="bg-slate-900 border-slate-800">
                                <CardHeader>
                                    <CardTitle className="text-white">Active Missions</CardTitle>
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
                </AnimatePresence>
            </div>
        </div>
    );
}
