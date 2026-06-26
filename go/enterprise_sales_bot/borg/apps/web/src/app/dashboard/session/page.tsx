"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Card, CardHeader, CardTitle, CardContent, Button, Badge } from "@tormentnexus/ui";
import { Loader2, Activity, Play, Square, Target, Crosshair, HelpCircle, ActivitySquare, RotateCcw } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

import { SessionCreateDialog } from './session-create-dialog';
import { SessionDetailsDialog, type SessionDetailsDialogSession } from './session-details-dialog';
import { formatRelativeTimestamp, formatRestartCountdown, getSessionTone } from './session-dashboard-utils';
import {
    normalizeSessionCatalog,
    normalizeSessionList,
    normalizeSessionState,
} from './session-page-normalizers';

export default function SessionDashboard() {
    const utils = trpc.useUtils();
    const { data: sessionState, isLoading, refetch } = trpc.session.getState.useQuery(undefined, { refetchInterval: 3000 });
    const sessionsQuery = trpc.session.list.useQuery(undefined, { refetchInterval: 3000 });
    const catalogQuery = trpc.session.catalog.useQuery(undefined, { refetchInterval: 15000 });
    const updateMutation = trpc.session.updateState.useMutation({
        onSuccess: () => {
            toast.success("Session state updated");
            refetch();
        },
        onError: (err) => {
            toast.error(`Update failed: ${err.message}`);
        }
    });

    const clearMutation = trpc.session.clear.useMutation({
        onSuccess: () => {
            toast.success("Session state cleared");
            refetch();
        }
    });

    const [pendingSessionActionId, setPendingSessionActionId] = useState<string | null>(null);
    const [currentTimestamp, setCurrentTimestamp] = useState(() => Date.now());

    useEffect(() => {
        const interval = window.setInterval(() => setCurrentTimestamp(Date.now()), 30_000);
        return () => window.clearInterval(interval);
    }, []);

    const refreshSessions = async () => {
        await Promise.all([
            sessionsQuery.refetch(),
            catalogQuery.refetch(),
            utils.session.list.invalidate(),
            utils.session.catalog.invalidate(),
        ]);
    };

    const startSessionMutation = trpc.session.start.useMutation({
        onSuccess: async () => {
            toast.success('Session started');
            setPendingSessionActionId(null);
            await refreshSessions();
        },
        onError: (error) => {
            toast.error(`Start failed: ${error.message}`);
            setPendingSessionActionId(null);
        },
    });

    const stopSessionMutation = trpc.session.stop.useMutation({
        onSuccess: async () => {
            toast.success('Session stopped');
            setPendingSessionActionId(null);
            await refreshSessions();
        },
        onError: (error) => {
            toast.error(`Stop failed: ${error.message}`);
            setPendingSessionActionId(null);
        },
    });

    const restartSessionMutation = trpc.session.restart.useMutation({
        onSuccess: async () => {
            toast.success('Session restarted');
            setPendingSessionActionId(null);
            await refreshSessions();
        },
        onError: (error) => {
            toast.error(`Restart failed: ${error.message}`);
            setPendingSessionActionId(null);
        },
    });

    const [goalInput, setGoalInput] = useState("");
    const [objectiveInput, setObjectiveInput] = useState("");

    const normalizedSessionState = normalizeSessionState(sessionState);

    // Keep inputs synced with external state changes only if user hasn't typed
    useEffect(() => {
        if (sessionState) {
            if (!goalInput) setGoalInput(normalizedSessionState.activeGoal);
            if (!objectiveInput) setObjectiveInput(normalizedSessionState.lastObjective);
        }
    }, [sessionState, normalizedSessionState.activeGoal, normalizedSessionState.lastObjective]);

    const handleSaveGoal = () => {
        updateMutation.mutate({ activeGoal: goalInput });
    };

    const handleSaveObjective = () => {
        updateMutation.mutate({ lastObjective: objectiveInput });
    };

    const toggleAutoDrive = () => {
        updateMutation.mutate({ isAutoDriveActive: !normalizedSessionState.isAutoDriveActive });
    };

    const sessions = normalizeSessionList(sessionsQuery.data);
    const catalog = normalizeSessionCatalog(catalogQuery.data);
    const installedHarnessCount = catalog.filter((entry) => entry.installed).length;
    const runningSessionCount = sessions.filter((session) => session.status === 'running').length;

    if (isLoading) {
        return <div className="p-8 flex items-center justify-center h-full"><Loader2 className="w-8 h-8 animate-spin text-zinc-500" /></div>;
    }

    return (
        <div className="p-8 max-w-5xl mx-auto space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
                        <ActivitySquare className="h-8 w-8 text-blue-500" />
                        Execution Session Control
                    </h1>
                    <p className="text-zinc-500 mt-2">
                        Manage supervised CLI sessions, Auto-Drive toggles, and the active operator objective.
                    </p>
                </div>
                <div className="flex gap-2">
                    <SessionCreateDialog catalog={catalog} onCreated={refreshSessions} />
                    <Button variant="outline" onClick={() => clearMutation.mutate()} className="border-red-500/20 text-red-500 hover:bg-red-500/10">
                        Reset Session
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Auto-Drive Control */}
                <Card className={`border-2 ${normalizedSessionState.isAutoDriveActive ? 'border-emerald-500/50 bg-emerald-950/10' : 'border-zinc-800 bg-zinc-900'} shadow-xl transition-all`}>
                    <CardHeader className="pb-2">
                        <CardTitle className="text-lg font-bold flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <Activity className={`h-5 w-5 ${normalizedSessionState.isAutoDriveActive ? 'text-emerald-500 animate-pulse' : 'text-zinc-500'}`} />
                                Auto-Drive Engine
                            </div>
                            <Badge variant={normalizedSessionState.isAutoDriveActive ? "default" : "secondary"} className={normalizedSessionState.isAutoDriveActive ? "bg-emerald-500 hover:bg-emerald-600" : ""}>
                                {normalizedSessionState.isAutoDriveActive ? "ACTIVE" : "PAUSED"}
                            </Badge>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4 pt-4">
                        <p className="text-sm text-zinc-400">
                            When Auto-Drive is active, background agents will continually pull objectives from the queue and execute them without waiting for manual intervention.
                        </p>
                        <Button
                            onClick={toggleAutoDrive}
                            disabled={updateMutation.isPending}
                            className={`w-full py-6 font-bold tracking-widest ${normalizedSessionState.isAutoDriveActive ? 'bg-red-900/50 hover:bg-red-900/80 text-red-400' : 'bg-emerald-600 hover:bg-emerald-500 text-white'}`}
                        >
                            {updateMutation.isPending ? <Loader2 className="w-5 h-5 animate-spin mx-auto" /> :
                                normalizedSessionState.isAutoDriveActive ? <><Square className="w-4 h-4 mr-2 fill-current" /> STOP AUTO-DRIVE</> : <><Play className="w-4 h-4 mr-2 fill-current" /> ENGAGE AUTO-DRIVE</>}
                        </Button>
                    </CardContent>
                </Card>

                {/* State Dump */}
                <Card className="bg-zinc-900 border-zinc-800 shadow-xl flex flex-col">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-lg font-bold flex items-center gap-2 text-zinc-300">
                            <HelpCircle className="h-5 w-5" />
                            Raw Session State
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1">
                        <div className="bg-black border border-zinc-800 rounded-md p-4 h-full">
                            <pre className="text-xs text-green-400 font-mono overflow-auto overflow-wrap-anywhere">
                                {JSON.stringify(sessionState, null, 2)}
                            </pre>
                        </div>
                    </CardContent>
                </Card>

                {/* Goal Management */}
                <Card className="bg-zinc-900 border-zinc-800 shadow-xl md:col-span-2">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-lg font-bold flex items-center gap-2 text-indigo-400">
                            <Target className="h-5 w-5" />
                            Current Operational Goal
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4 pt-4">
                        <div className="flex gap-2">
                            <input
                                value={goalInput}
                                onChange={e => setGoalInput(e.target.value)}
                                className="flex-1 bg-black border border-zinc-800 rounded-md p-3 text-sm text-white focus:ring-1 focus:ring-indigo-500 outline-none"
                                placeholder="Enter global objective for the system..."
                            />
                            <Button onClick={handleSaveGoal} disabled={updateMutation.isPending || goalInput === normalizedSessionState.activeGoal} className="bg-indigo-600 hover:bg-indigo-500">
                                Set Goal
                            </Button>
                        </div>
                        <div className="pt-2 border-t border-zinc-800">
                            <div className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 flex items-center gap-2">
                                <Crosshair className="w-3 h-3" /> Transient Objective
                            </div>
                            <div className="flex gap-2">
                                <input
                                    value={objectiveInput}
                                    onChange={e => setObjectiveInput(e.target.value)}
                                    className="flex-1 bg-black border border-zinc-800 rounded-md p-2 text-sm text-zinc-300 focus:ring-1 focus:ring-zinc-600 outline-none"
                                    placeholder="Enter current micro-task..."
                                />
                                <Button variant="secondary" onClick={handleSaveObjective} disabled={updateMutation.isPending || objectiveInput === normalizedSessionState.lastObjective}>
                                    Update
                                </Button>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-zinc-900 border-zinc-800 shadow-xl md:col-span-2">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-lg font-bold flex items-center justify-between gap-3">
                            <div className="flex items-center gap-2 text-cyan-400">
                                <ActivitySquare className="h-5 w-5" />
                                Supervised CLI Sessions
                            </div>
                            <div className="flex items-center gap-2 text-xs text-zinc-400">
                                <Badge variant="secondary">{runningSessionCount}/{sessions.length} running</Badge>
                                <Badge variant="secondary">{installedHarnessCount}/{catalog.length} harnesses detected</Badge>
                            </div>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4 pt-4">
                        {sessions.length === 0 ? (
                            <div className="rounded-lg border border-dashed border-zinc-800 bg-black/40 p-6 text-sm text-zinc-400">
                                No supervised sessions exist yet. Create one to launch Aider, Claude Code, Gemini CLI, Codex, or OpenCode under TormentNexus supervision.
                            </div>
                        ) : sessions.map((session) => {
                            const latestLog = session.logs.length > 0 ? session.logs[session.logs.length - 1] : null;
                            const isPending = pendingSessionActionId === session.id;
                            const canStart = session.status === 'created' || session.status === 'stopped' || session.status === 'error';
                            const canStop = session.status === 'starting' || session.status === 'running' || session.status === 'restarting';
                            const sessionDetails: SessionDetailsDialogSession = {
                                id: session.id ?? '',
                                name: session.name ?? 'Unnamed session',
                                cliType: session.cliType ?? 'cli',
                                workingDirectory: session.workingDirectory ?? '',
                                worktreePath: session.worktreePath,
                                executionProfile: session.executionProfile,
                                executionPolicy: session.executionPolicy,
                                autoRestart: session.autoRestart,
                                status: session.status ?? 'created',
                                restartCount: session.restartCount ?? 0,
                                maxRestartAttempts: session.maxRestartAttempts ?? 0,
                                scheduledRestartAt: session.scheduledRestartAt,
                                lastActivityAt: session.lastActivityAt ?? currentTimestamp,
                                lastError: session.lastError,
                                metadata: (session.metadata ?? {}) as SessionDetailsDialogSession['metadata'],
                            };

                            return (
                                <div key={session.id} className="rounded-xl border border-zinc-800 bg-black/50 p-4">
                                    <div className="flex flex-col gap-4 lg:flex-row lg:justify-between lg:items-start">
                                        <div className="min-w-0 flex-1 space-y-3">
                                            <div className="flex flex-wrap items-center gap-2">
                                                <h3 className="text-base font-semibold text-white">{session.name}</h3>
                                                <Badge className={getSessionTone(session.status)}>{session.status}</Badge>
                                                <Badge variant="outline" className="border-zinc-700 text-zinc-300">{session.cliType}</Badge>
                                                {session.executionPolicy?.shellLabel ? (
                                                    <Badge variant="outline" className="border-cyan-500/30 text-cyan-200">
                                                        {session.executionPolicy.shellLabel}
                                                    </Badge>
                                                ) : null}
                                                {session.autoRestart === false && (
                                                    <Badge variant="outline" className="border-amber-700/50 text-amber-500 bg-amber-950/20">Manual Restart</Badge>
                                                )}
                                                {session.status === 'error' && (
                                                    <Badge className="bg-red-950 text-red-400 border border-red-800">Crashed</Badge>
                                                )}
                                            </div>
                                            <p className="break-all font-mono text-xs text-zinc-500">{session.worktreePath ?? session.workingDirectory}</p>
                                            <p className="text-xs text-zinc-500">
                                                Last activity {formatRelativeTimestamp(session.lastActivityAt, currentTimestamp)} · Restarts {session.restartCount}/{session.maxRestartAttempts}
                                            </p>
                                            {session.status === 'restarting' && session.scheduledRestartAt ? (
                                                <p className="text-xs text-amber-300">
                                                    Restart queued {formatRestartCountdown(session.scheduledRestartAt, currentTimestamp)}
                                                </p>
                                            ) : null}
                                            {latestLog ? (
                                                <div className="rounded-lg border border-zinc-800 bg-zinc-950/80 p-3">
                                                    <div className="mb-2 flex items-center justify-between gap-3 text-[11px] uppercase tracking-[0.18em] text-zinc-500">
                                                        <span>Latest {latestLog.stream}</span>
                                                        <span>{formatRelativeTimestamp(latestLog.timestamp, currentTimestamp)}</span>
                                                    </div>
                                                    <p className="whitespace-pre-wrap break-words text-sm text-zinc-300 line-clamp-4">{latestLog.message}</p>
                                                </div>
                                            ) : null}
                                            {session.lastError ? (
                                                <div className="rounded-md border border-red-900/50 bg-red-950/30 p-3 mt-2">
                                                    <p className="text-sm font-semibold text-red-400 mb-1">Session Crashed</p>
                                                    <p className="text-xs text-red-300/80 break-words">{session.lastError}</p>
                                                </div>
                                            ) : null}
                                        </div>

                                        <div className="flex flex-wrap gap-2">
                                            <SessionDetailsDialog session={sessionDetails} currentTimestamp={currentTimestamp} />
                                            <Link href={`/dashboard/session/${session.id}`}>
                                                <Button
                                                    variant="outline"
                                                    className="border-blue-500/30 text-blue-200 hover:bg-blue-500/10"
                                                >
                                                    View Details
                                                </Button>
                                            </Link>
                                            <Button
                                                onClick={() => {
                                                    setPendingSessionActionId(session.id);
                                                    startSessionMutation.mutate({ id: session.id });
                                                }}
                                                disabled={!canStart || isPending}
                                                className="bg-emerald-700 hover:bg-emerald-600 text-white"
                                            >
                                                {isPending && canStart ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                                                Start
                                            </Button>
                                            <Button
                                                variant="outline"
                                                onClick={() => {
                                                    setPendingSessionActionId(session.id);
                                                    stopSessionMutation.mutate({ id: session.id });
                                                }}
                                                disabled={!canStop || isPending}
                                                className="border-zinc-700 text-zinc-200 hover:bg-zinc-800"
                                            >
                                                {isPending && canStop ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Square className="mr-2 h-4 w-4" />}
                                                Stop
                                            </Button>
                                            <Button
                                                variant="outline"
                                                onClick={() => {
                                                    setPendingSessionActionId(session.id);
                                                    restartSessionMutation.mutate({ id: session.id });
                                                }}
                                                disabled={isPending}
                                                className="border-cyan-500/30 text-cyan-200 hover:bg-cyan-500/10"
                                            >
                                                {isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RotateCcw className="mr-2 h-4 w-4" />}
                                                Restart
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
