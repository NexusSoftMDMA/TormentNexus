import React from 'react';
import { trpc } from '../utils/trpc';
import { Card } from './ui/card';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import { Activity, Brain, CheckCircle, Clock } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

type DirectorStatusViewModel = {
    active: boolean;
    status: string;
    goal: string;
    step: number;
    totalSteps: number;
    lastHistory: string[];
};

function normalizeDirectorStatus(value: unknown): DirectorStatusViewModel {
    if (!value || typeof value !== 'object') {
        return {
            active: false,
            status: 'OFFLINE',
            goal: 'No active task',
            step: 0,
            totalSteps: 0,
            lastHistory: [],
        };
    }

    const record = value as Record<string, unknown>;
    const active = typeof record.active === 'boolean' ? record.active : false;
    const status = typeof record.status === 'string' ? record.status : 'OFFLINE';
    const goal = typeof record.goal === 'string' ? record.goal : 'No active task';
    const step = typeof record.step === 'number' ? record.step : 0;
    const totalSteps = typeof record.totalSteps === 'number' ? record.totalSteps : 0;
    const lastHistory = Array.isArray(record.lastHistory)
        ? record.lastHistory.filter((entry): entry is string => typeof entry === 'string')
        : [];

    return {
        active,
        status,
        goal,
        step,
        totalSteps,
        lastHistory,
    };
}

export function DirectorStatusWidget() {
    const { data: status, isLoading } = trpc.director.status.useQuery(undefined, {
        refetchInterval: 1000, // Real-time updates
    });

    if (isLoading || !status) {
        return (
            <Card className="p-6 h-full flex items-center justify-center bg-zinc-900/50 border-zinc-800">
                <div className="flex flex-col items-center gap-2 text-zinc-500">
                    <Activity className="w-8 h-8 animate-pulse" />
                    <span>Connecting to Director...</span>
                </div>
            </Card>
        );
    }

    // Handle legacy or partial status response
    const normalized = normalizeDirectorStatus(status);
    const active = normalized.active;
    const currentStatus = normalized.status;
    const goal = normalized.goal;
    const step = normalized.step;
    const total = normalized.totalSteps;
    const history = normalized.lastHistory;

    return (
        <Card className="relative overflow-hidden bg-zinc-900 border-zinc-800 h-full flex flex-col">
            {/* Background Decor */}
            <div className="absolute top-0 right-0 p-32 bg-indigo-500/5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 pointer-events-none" />

            {/* Header */}
            <div className="p-6 border-b border-white/5 flex items-center justify-between z-10">
                <div className="flex items-center gap-3">
                    <div className={`p-2 rounded-lg ${active ? 'bg-indigo-500/20 text-indigo-400' : 'bg-zinc-800 text-zinc-500'}`}>
                        <Brain className="w-5 h-5" />
                    </div>
                    <div>
                        <h2 className="font-semibold text-white tracking-tight">Director Status</h2>
                        <div className="flex items-center gap-2 text-xs">
                            <span className={`w-2 h-2 rounded-full ${active ? 'bg-green-500 animate-pulse' : 'bg-zinc-600'}`} />
                            <span className="text-zinc-400 font-mono uppercase">{currentStatus}</span>
                        </div>
                    </div>
                </div>

                {active && (
                    <Badge variant="outline" className="font-mono border-indigo-500/30 text-indigo-400 bg-indigo-500/10">
                        STEP {step}/{total}
                    </Badge>
                )}
            </div>

            {/* Content */}
            <div className="flex-1 p-6 flex flex-col gap-6 z-10">
                {/* Active Goal */}
                <div className="space-y-2">
                    <label className="text-xs font-medium text-zinc-500 uppercase tracking-wider">Current Objective</label>
                    <div className="p-4 rounded-xl bg-black/40 border border-white/5 shadow-inner">
                        <p className="text-sm text-zinc-200 font-medium leading-relaxed">
                            {active ? goal : "System is idle. Waiting for instructions."}
                        </p>
                    </div>
                </div>

                {/* Progress Bar (if active) */}
                {active && total > 0 && (
                    <div className="space-y-2">
                        <div className="flex justify-between text-xs text-zinc-500">
                            <span>Progress</span>
                            <span>{Math.round((step / total) * 100)}%</span>
                        </div>
                        <div className="h-2 w-full bg-zinc-800 rounded-full overflow-hidden">
                            <motion.div
                                className="h-full bg-gradient-to-r from-indigo-500 to-purple-500"
                                initial={{ width: 0 }}
                                animate={{ width: `${(step / total) * 100}%` }}
                                transition={{ duration: 0.5 }}
                            />
                        </div>
                    </div>
                )}

                {/* Live Logs / Thoughts */}
                <div className="flex-1 min-h-0 flex flex-col">
                    <label className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-2 flex items-center gap-2">
                        <Clock className="w-3 h-3" /> Recent Activity
                    </label>
                    <ScrollArea className="flex-1 rounded-lg border border-white/5 bg-black/20 p-4">
                        <div className="space-y-3 font-mono text-xs">
                            <AnimatePresence>
                                {history.length > 0 ? (
                                    history.map((log: string, i: number) => (
                                        <motion.div
                                            key={i}
                                            initial={{ opacity: 0, x: -10 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            className="text-zinc-400 border-l-2 border-zinc-800 pl-3 py-1"
                                        >
                                            {log.replace(/^Observation:/, '🔍').replace(/^Action:/, '⚡').replace(/^Thinking:/, '🤔')}
                                        </motion.div>
                                    ))
                                ) : (
                                    <span className="text-zinc-600 italic">No recent logs...</span>
                                )}
                            </AnimatePresence>
                        </div>
                    </ScrollArea>
                </div>
            </div>
        </Card>
    );
}
