/**
 * swarmRouter.ts
 * 
 * Exposes the Multi-Agent Orchestration features (Swarm, Debate, Consensus) to the Next.js UI.
 */

import { z } from 'zod';
import { t, publicProcedure } from '../trpc.js';
import { SwarmOrchestrator, normalizeSwarmToolPolicy } from '../agents/swarm/SwarmOrchestrator.js';
import { DebateProtocol, DebateMode, DebateTopicType } from '../agents/swarm/DebateProtocol.js';
import { ConsensusEngine } from '../agents/swarm/ConsensusEngine.js';
import { MissionService } from '../services/MissionService.js';
import { SwarmMessageType } from '../mesh/MeshService.js';

// Global registry of active swarm orchestrators for approvals
const activeOrchestrators = new Map<string, SwarmOrchestrator>();

export const swarmRouter = t.router({

    startSwarm: publicProcedure
        .input(
            z.object({
                masterPrompt: z.string(),
                maxConcurrency: z.number().optional(),
                model: z.string().optional(),
                priority: z.number().min(1).max(5).optional(),
                tools: z.array(z.string()).optional(),
                toolPolicy: z.object({
                    allow: z.array(z.string()).optional(),
                    deny: z.array(z.string()).optional()
                }).optional()
            })
        )
        // @ts-expect-error -- type mismatch from inferred schema types
        .mutation(async ({ input }: {
            input: {
                masterPrompt: string;
                maxConcurrency?: number;
                model?: string;
                priority?: number;
                tools?: string[];
                toolPolicy?: { allow?: string[]; deny?: string[] };
            }
        }) => {
            const missionService = global.mcpServerInstance?.missionService;
            const healerService = global.mcpServerInstance?.healerService;
            const normalizedToolPolicy = normalizeSwarmToolPolicy(input.toolPolicy);

            const orchestrator = new SwarmOrchestrator(
                {
                    defaultModel: input.model || 'gpt-4o-mini',
                    maxConcurrency: input.maxConcurrency || 5,
                    gitWorktreeManager: global.mcpServerInstance?.gitWorktreeManager // Phase 95
                },
                missionService,
                healerService
            );

            // Decompose
            const tasks = await orchestrator.decomposeGoal(input.masterPrompt, input.tools, normalizedToolPolicy.effectivePolicy);
            const missionId = (orchestrator as any).currentMissionId;

            // Set mission priority if provided
            if (missionId && input.priority && missionService) {
                const mission = missionService.getMission(missionId);
                if (mission) {
                    mission.priority = input.priority;
                }
            }

            if (missionId) {
                activeOrchestrators.set(missionId, orchestrator);
                orchestrator.on('swarm:completed', () => activeOrchestrators.delete(missionId));
            }

            // Fire and forget
            orchestrator.executeSwarm();

            return {
                missionId,
                taskCount: tasks.length,
                effectiveToolPolicy: normalizedToolPolicy.effectivePolicy,
                policyWarnings: normalizedToolPolicy.warnings
            };
        }),

    resumeMission: publicProcedure
        .input(z.object({ missionId: z.string() }))
        .mutation(async ({ input }) => {
            const missionService = (global as any).mcpServerInstance?.missionService as MissionService;
            const healerService = (global as any).mcpServerInstance?.healerService;
            if (!missionService) throw new Error('MissionService unavailable');

            const orchestrator = new SwarmOrchestrator(
                { gitWorktreeManager: global.mcpServerInstance?.gitWorktreeManager },
                missionService,
                healerService
            );
            activeOrchestrators.set(input.missionId, orchestrator);

            orchestrator.on('swarm:completed', () => activeOrchestrators.delete(input.missionId));

            // Resume - this is async but we don't await the whole swarm
            orchestrator.resumeMission(input.missionId);

            return { success: true };
        }),

    approveTask: publicProcedure
        .input(z.object({
            missionId: z.string(),
            taskId: z.string(),
            approved: z.boolean()
        }))
        .mutation(async ({ input }) => {
            const orchestrator = activeOrchestrators.get(input.missionId);
            if (!orchestrator) throw new Error('No active orchestrator found for this mission');

            const success = orchestrator.approveTask(input.taskId, input.approved);
            return { success };
        }),

    decomposeTask: publicProcedure
        .input(z.object({
            missionId: z.string(),
            taskId: z.string()
        }))
        .mutation(async ({ input }) => {
            const orchestrator = activeOrchestrators.get(input.missionId);
            if (!orchestrator) throw new Error('No active orchestrator found for this mission');

            const subMission = await orchestrator.decomposeTask(input.taskId);
            return { success: !!subMission, subMissionId: subMission?.id };
        }),

    updateTaskPriority: publicProcedure
        .input(z.object({
            missionId: z.string(),
            taskId: z.string(),
            priority: z.number().min(1).max(5)
        }))
        .mutation(async ({ input }) => {
            const missionService = (global as any).mcpServerInstance?.missionService as MissionService;
            if (!missionService) throw new Error('MissionService unavailable');

            const mission = missionService.updateMissionTask(input.missionId, input.taskId, { priority: input.priority });
            return { success: !!mission };
        }),

    executeDebate: publicProcedure
        .input(
            z.object({
                topic: z.string(),
                proponentModel: z.string(),
                opponentModel: z.string(),
                judgeModel: z.string(),
                rounds: z.number().optional(),
                mode: z.enum(['standard', 'adversarial']).optional(),
                topicType: z.enum(['general', 'mission-plan']).optional()
            })
        )
        // @ts-expect-error -- type mismatch from inferred schema types
        .mutation(async ({ input }: { input: { topic: string; proponentModel: string; opponentModel: string; judgeModel: string; rounds?: number; mode?: DebateMode; topicType?: DebateTopicType } }) => {
            const debate = new DebateProtocol({
                topic: input.topic,
                proponentModel: input.proponentModel,
                opponentModel: input.opponentModel,
                judgeModel: input.judgeModel,
                maxRounds: input.rounds || 3,
                mode: input.mode,
                topicType: input.topicType
            });

            // Synchronous wait for simulation, standard TRPC streaming is better for prod
            const result = await debate.conductDebate();
            return result;
        }),

    seekConsensus: publicProcedure
        .input(
            z.object({
                prompt: z.string(),
                models: z.array(z.string()).min(2),
                requiredAgreement: z.number().optional()
            })
        )
        // @ts-expect-error -- type mismatch from inferred schema types
        .mutation(async ({ input }: { input: { prompt: string; models: string[]; requiredAgreement?: number } }) => {
            const engine = new ConsensusEngine();

            const result = await engine.seekConsensus({
                prompt: input.prompt,
                models: input.models,
                requiredAgreementPercentage: input.requiredAgreement
            });

            return result;
        }),

    getMissionHistory: publicProcedure
        .query(async () => {
            return global.mcpServerInstance?.missionService.getAllMissions() || [];
        }),

    getMissionRiskSummary: publicProcedure
        .query(async () => {
            const missions = global.mcpServerInstance?.missionService.getAllMissions() || [];
            const deniedToolCounts = new Map<string, number>();
            const deniedEventsByHour24 = Array.from({ length: 24 }, () => 0);
            const now = Date.now();
            const oneHourMs = 60 * 60 * 1000;
            const oneDayMs = 24 * oneHourMs;

            const summary = {
                totalMissions: missions.length,
                missionsWithDeniedEvents: 0,
                totalDeniedEvents: 0,
                severityScore: 0,
                topDeniedTools: [] as Array<{ tool: string; count: number }>,
                deniedEventsLast24h: 0,
                deniedEventsByHour24,
                statusBreakdown: {
                    active: 0,
                    completed: 0,
                    failed: 0,
                    paused: 0
                },
                topRiskMission: null as null | {
                    missionId: string;
                    goal: string;
                    deniedEventCount: number;
                }
            };

            for (const mission of missions as any[]) {
                const status = mission?.status as 'active' | 'completed' | 'failed' | 'paused' | undefined;
                if (status && summary.statusBreakdown[status] !== undefined) {
                    summary.statusBreakdown[status] += 1;
                }

                let missionDeniedEvents: Array<{ tool: string; reason: string; timestamp: number }> = [];
                const deniedEventCount = (mission.tasks || []).reduce((count: number, task: any) => {
                    if (Array.isArray(task?.deniedToolEvents)) {
                        missionDeniedEvents = missionDeniedEvents.concat(task.deniedToolEvents);
                    }
                    return count + (Array.isArray(task?.deniedToolEvents) ? task.deniedToolEvents.length : 0);
                }, 0);

                if (deniedEventCount > 0) {
                    summary.missionsWithDeniedEvents += 1;
                    summary.totalDeniedEvents += deniedEventCount;

                    for (const event of missionDeniedEvents) {
                        if (!event?.tool) continue;
                        deniedToolCounts.set(event.tool, (deniedToolCounts.get(event.tool) || 0) + 1);

                        const ts = typeof event.timestamp === 'number' ? event.timestamp : 0;
                        const ageMs = now - ts;
                        if (ageMs >= 0 && ageMs < oneDayMs) {
                            summary.deniedEventsLast24h += 1;
                            const hourIndex = 23 - Math.floor(ageMs / oneHourMs);
                            if (hourIndex >= 0 && hourIndex < 24) {
                                summary.deniedEventsByHour24[hourIndex] += 1;
                            }
                        }
                    }

                    if (!summary.topRiskMission || deniedEventCount > summary.topRiskMission.deniedEventCount) {
                        summary.topRiskMission = {
                            missionId: mission.id,
                            goal: mission.goal,
                            deniedEventCount
                        };
                    }
                }
            }

            const missionCount = Math.max(summary.totalMissions, 1);
            // Normalized severity heuristic in [0, 100]
            summary.severityScore = Math.min(100, Math.round((summary.totalDeniedEvents / missionCount) * 20));

            summary.topDeniedTools = Array.from(deniedToolCounts.entries())
                .map(([tool, count]) => ({ tool, count }))
                .sort((a, b) => b.count - a.count)
                .slice(0, 5);

            return summary;
        }),

    getMissionRiskRows: publicProcedure
        .input(
            z.object({
                statusFilter: z.enum(['all', 'active', 'completed', 'failed', 'paused']).optional(),
                sortBy: z.enum(['risk', 'recent']).optional(),
                minRisk: z.number().min(0).max(100).optional(),
                limit: z.number().min(1).max(500).optional()
            }).optional()
        )
        .query(async ({ input }) => {
            const missions = global.mcpServerInstance?.missionService.getAllMissions() || [];
            const now = Date.now();
            const oneDayMs = 24 * 60 * 60 * 1000;
            const twoDayMs = 2 * oneDayMs;

            const rows = (missions as any[])
                .map((mission) => {
                    const deniedEventCount = (mission.tasks || []).reduce((count: number, task: any) => {
                        return count + (Array.isArray(task?.deniedToolEvents) ? task.deniedToolEvents.length : 0);
                    }, 0);

                    const deniedEventsLast24h = (mission.tasks || []).reduce((count: number, task: any) => {
                        if (!Array.isArray(task?.deniedToolEvents)) return count;
                        return count + task.deniedToolEvents.filter((event: any) => {
                            const ts = typeof event?.timestamp === 'number' ? event.timestamp : 0;
                            const age = now - ts;
                            return age >= 0 && age < oneDayMs;
                        }).length;
                    }, 0);

                    const deniedEventsPrev24h = (mission.tasks || []).reduce((count: number, task: any) => {
                        if (!Array.isArray(task?.deniedToolEvents)) return count;
                        return count + task.deniedToolEvents.filter((event: any) => {
                            const ts = typeof event?.timestamp === 'number' ? event.timestamp : 0;
                            const age = now - ts;
                            return age >= oneDayMs && age < twoDayMs;
                        }).length;
                    }, 0);

                    const failedTasks = (mission.tasks || []).filter((task: any) => task?.status === 'failed').length;
                    const highPriorityTasks = (mission.tasks || []).filter((task: any) => (task?.priority || 0) >= 4).length;
                    const missionRiskScore = Math.min(
                        100,
                        deniedEventCount * 10 + deniedEventsLast24h * 6 + failedTasks * 8 + highPriorityTasks * 3
                    );

                    return {
                        mission,
                        deniedEventCount,
                        deniedEventsLast24h,
                        missionRiskScore,
                        updatedAt: new Date(mission.updatedAt || 0).getTime()
                    };
                })
                .filter((row) => {
                    if (!input?.statusFilter || input.statusFilter === 'all') return true;
                    return row.mission?.status === input.statusFilter;
                })
                .filter((row) => {
                    if (typeof input?.minRisk !== 'number') return true;
                    return row.missionRiskScore >= input.minRisk;
                })
                .sort((a, b) => {
                    if ((input?.sortBy || 'risk') === 'risk') {
                        if (b.missionRiskScore !== a.missionRiskScore) return b.missionRiskScore - a.missionRiskScore;
                        if (b.deniedEventCount !== a.deniedEventCount) return b.deniedEventCount - a.deniedEventCount;
                    }
                    return b.updatedAt - a.updatedAt;
                })
                .slice(0, input?.limit || 200)
                .map(({ updatedAt, ...rest }) => rest);

            return rows;
        }),

    getMissionRiskFacets: publicProcedure
        .input(
            z.object({
                statusFilter: z.enum(['all', 'active', 'completed', 'failed', 'paused']).optional(),
                minRisk: z.number().min(0).max(100).optional()
            }).optional()
        )
        .query(async ({ input }) => {
            const missions = global.mcpServerInstance?.missionService.getAllMissions() || [];
            const now = Date.now();
            const oneDayMs = 24 * 60 * 60 * 1000;
            const twoDayMs = 2 * oneDayMs;

            const scored = (missions as any[])
                .map((mission) => {
                    const deniedEventCount = (mission.tasks || []).reduce((count: number, task: any) => {
                        return count + (Array.isArray(task?.deniedToolEvents) ? task.deniedToolEvents.length : 0);
                    }, 0);

                    const deniedEventsLast24h = (mission.tasks || []).reduce((count: number, task: any) => {
                        if (!Array.isArray(task?.deniedToolEvents)) return count;
                        return count + task.deniedToolEvents.filter((event: any) => {
                            const ts = typeof event?.timestamp === 'number' ? event.timestamp : 0;
                            const age = now - ts;
                            return age >= 0 && age < oneDayMs;
                        }).length;
                    }, 0);

                    const deniedEventsPrev24h = (mission.tasks || []).reduce((count: number, task: any) => {
                        if (!Array.isArray(task?.deniedToolEvents)) return count;
                        return count + task.deniedToolEvents.filter((event: any) => {
                            const ts = typeof event?.timestamp === 'number' ? event.timestamp : 0;
                            const age = now - ts;
                            return age >= oneDayMs && age < twoDayMs;
                        }).length;
                    }, 0);

                    const failedTasks = (mission.tasks || []).filter((task: any) => task?.status === 'failed').length;
                    const highPriorityTasks = (mission.tasks || []).filter((task: any) => (task?.priority || 0) >= 4).length;
                    const missionRiskScore = Math.min(
                        100,
                        deniedEventCount * 10 + deniedEventsLast24h * 6 + failedTasks * 8 + highPriorityTasks * 3
                    );

                    return {
                        status: mission?.status as 'active' | 'completed' | 'failed' | 'paused' | undefined,
                        missionRiskScore,
                        deniedEventsLast24h,
                        deniedEventsPrev24h
                    };
                })
                .filter((row) => {
                    if (!input?.statusFilter || input.statusFilter === 'all') return true;
                    return row.status === input.statusFilter;
                })
                .filter((row) => {
                    if (typeof input?.minRisk !== 'number') return true;
                    return row.missionRiskScore >= input.minRisk;
                });

            const missionCount = scored.length;
            const riskScores = scored.map(row => row.missionRiskScore);
            const deniedLast24hTotal = scored.reduce((sum, row) => sum + row.deniedEventsLast24h, 0);
            const deniedPrev24hTotal = scored.reduce((sum, row) => sum + row.deniedEventsPrev24h, 0);
            const freshestUpdatedAtMs = (missions as any[])
                .map((mission) => new Date(mission?.updatedAt || 0).getTime())
                .filter((value) => Number.isFinite(value) && value > 0)
                .sort((a, b) => b - a)[0] || 0;
            const latestUpdateAgeSeconds = freshestUpdatedAtMs > 0
                ? Math.max(0, Math.round((now - freshestUpdatedAtMs) / 1000))
                : null;

            const freshnessBucket = latestUpdateAgeSeconds == null
                ? 'unknown'
                : latestUpdateAgeSeconds < 60
                    ? 'fresh'
                    : latestUpdateAgeSeconds < 300
                        ? 'recent'
                        : 'stale';
            const averageRisk = missionCount > 0
                ? Math.round(riskScores.reduce((sum, value) => sum + value, 0) / missionCount)
                : 0;

            const low = riskScores.filter(score => score < 40).length;
            const medium = riskScores.filter(score => score >= 40 && score < 70).length;
            const high = riskScores.filter(score => score >= 70).length;

            const lowPct = missionCount > 0 ? Math.round((low / missionCount) * 100) : 0;
            const mediumPct = missionCount > 0 ? Math.round((medium / missionCount) * 100) : 0;
            const highPct = missionCount > 0 ? Math.round((high / missionCount) * 100) : 0;

            const statusCounts = {
                active: scored.filter(row => row.status === 'active').length,
                completed: scored.filter(row => row.status === 'completed').length,
                failed: scored.filter(row => row.status === 'failed').length,
                paused: scored.filter(row => row.status === 'paused').length
            };
            const statusPercentages = {
                active: missionCount > 0 ? Math.round((statusCounts.active / missionCount) * 100) : 0,
                completed: missionCount > 0 ? Math.round((statusCounts.completed / missionCount) * 100) : 0,
                failed: missionCount > 0 ? Math.round((statusCounts.failed / missionCount) * 100) : 0,
                paused: missionCount > 0 ? Math.round((statusCounts.paused / missionCount) * 100) : 0
            };

            const dominantBand = [
                { band: 'low' as const, count: low },
                { band: 'medium' as const, count: medium },
                { band: 'high' as const, count: high }
            ].sort((a, b) => b.count - a.count)[0]?.band || 'low';

            const deniedDelta = deniedLast24hTotal - deniedPrev24hTotal;
            const deniedDeltaPct = deniedPrev24hTotal > 0
                ? Math.round((deniedDelta / deniedPrev24hTotal) * 100)
                : (deniedLast24hTotal > 0 ? 100 : 0);
            const deniedTrend = deniedDelta > 0 ? 'up' : deniedDelta < 0 ? 'down' : 'flat';

            const healthReasons: string[] = [];
            if (freshnessBucket === 'unknown') healthReasons.push('No mission update timestamp available');
            if (freshnessBucket === 'stale') healthReasons.push('Facet data appears stale (>5m since latest mission update)');
            if (statusPercentages.failed >= 40) healthReasons.push('High failed mission share (>=40%)');
            if (statusPercentages.failed >= 20 && statusPercentages.failed < 40) healthReasons.push('Elevated failed mission share (>=20%)');
            if (deniedTrend === 'up' && deniedDeltaPct >= 50 && deniedLast24hTotal >= 10) {
                healthReasons.push('Denied-tool activity accelerating strongly (+>=50%)');
            } else if (deniedTrend === 'up' && deniedDelta > 0) {
                healthReasons.push('Denied-tool activity increasing');
            }
            if (highPct >= 50) healthReasons.push('High-risk band dominates current filter scope');

            let healthSeverity: 'good' | 'warn' | 'critical' = 'good';
            if (
                freshnessBucket === 'unknown' ||
                freshnessBucket === 'stale' ||
                statusPercentages.failed >= 40 ||
                (deniedTrend === 'up' && deniedDeltaPct >= 50 && deniedLast24hTotal >= 10)
            ) {
                healthSeverity = 'critical';
            } else if (
                freshnessBucket === 'recent' ||
                statusPercentages.failed >= 20 ||
                deniedTrend === 'up' ||
                highPct >= 50
            ) {
                healthSeverity = 'warn';
            }

            const healthScore = Math.max(
                0,
                Math.min(
                    100,
                    100
                    - (healthSeverity === 'critical' ? 55 : healthSeverity === 'warn' ? 30 : 0)
                    - Math.round(statusPercentages.failed * 0.5)
                    - Math.round(Math.max(0, deniedDeltaPct) * 0.2)
                    - (freshnessBucket === 'stale' ? 20 : freshnessBucket === 'recent' ? 8 : freshnessBucket === 'unknown' ? 25 : 0)
                )
            );

            const sampleSizePenalty = missionCount < 5 ? 35 : missionCount < 15 ? 15 : 0;
            const freshnessPenalty = freshnessBucket === 'stale' ? 30 : freshnessBucket === 'recent' ? 12 : freshnessBucket === 'unknown' ? 40 : 0;
            const signalCongestionPenalty = Math.min(
                25,
                healthReasons.length * 6 + (deniedTrend === 'up' ? 6 : 0) + (statusPercentages.failed >= 20 ? 6 : 0)
            );

            const confidenceScore = Math.max(
                0,
                Math.min(
                    100,
                    100 - sampleSizePenalty - freshnessPenalty - signalCongestionPenalty
                )
            );
            const uncertaintyMargin = Math.min(
                35,
                5
                + Math.round(sampleSizePenalty * 0.35)
                + Math.round(freshnessPenalty * 0.25)
                + Math.round(signalCongestionPenalty * 0.2)
            );
            const scoreRange = {
                min: Math.max(0, confidenceScore - uncertaintyMargin),
                max: Math.min(100, confidenceScore + uncertaintyMargin)
            };
            const confidenceStability: 'stable' | 'watch' | 'volatile' = uncertaintyMargin >= 22
                ? 'volatile'
                : uncertaintyMargin >= 12 || deniedTrend === 'up'
                    ? 'watch'
                    : 'stable';
            const confidenceAdvice = confidenceStability === 'volatile'
                ? 'Pause high-risk expansion, refresh telemetry, and wait for signal convergence before major decisions.'
                : confidenceStability === 'watch'
                    ? 'Monitor another cycle and re-evaluate after fresh mission updates or reduced denial volatility.'
                    : 'Confidence is stable; proceed while continuing standard monitoring cadence.';
            const confidenceDrivers: string[] = [];
            if (missionCount < 5) {
                confidenceDrivers.push('Low sample size (<5 missions) limits confidence');
            } else if (missionCount < 15) {
                confidenceDrivers.push('Moderate sample size (<15 missions) provides partial confidence');
            } else {
                confidenceDrivers.push('Broad sample size (>=15 missions) supports confidence');
            }

            if (freshnessBucket === 'fresh') {
                confidenceDrivers.push('Telemetry is fresh (<60s since latest mission update)');
            } else if (freshnessBucket === 'recent') {
                confidenceDrivers.push('Telemetry is recent (<5m) but not fully fresh');
            } else if (freshnessBucket === 'stale') {
                confidenceDrivers.push('Telemetry staleness (>5m) lowers confidence');
            } else {
                confidenceDrivers.push('Missing telemetry timestamp lowers confidence');
            }

            if (healthReasons.length === 0) {
                confidenceDrivers.push('No active health warnings detected');
            } else if (healthReasons.length >= 3) {
                confidenceDrivers.push('Multiple concurrent risk signals reduce certainty');
            }
            if (signalCongestionPenalty >= 18) {
                confidenceDrivers.push('Signal congestion is elevated; confidence is intentionally discounted');
            }
            if (uncertaintyMargin >= 20) {
                confidenceDrivers.push('Wide uncertainty margin indicates lower precision in confidence estimate');
            }
            const confidenceLevel: 'high' | 'medium' | 'low' = confidenceScore >= 75 ? 'high' : confidenceScore >= 45 ? 'medium' : 'low';
            const confidenceAlerts: string[] = [];
            if (confidenceLevel === 'low') {
                confidenceAlerts.push('Low confidence level: avoid irreversible governance actions until confidence improves');
            }
            if (confidenceStability === 'volatile') {
                confidenceAlerts.push('Volatile confidence: confidence state may shift quickly between evaluation cycles');
            }
            if (freshnessBucket === 'stale' || freshnessBucket === 'unknown') {
                confidenceAlerts.push('Telemetry freshness is insufficient for high-assurance confidence interpretation');
            }
            if (missionCount < 5) {
                confidenceAlerts.push('Sample size is low; confidence interpretation may be noisy');
            }
            const confidenceAlertLevel: 'none' | 'warn' | 'critical' = confidenceLevel === 'low' || confidenceStability === 'volatile'
                ? 'critical'
                : confidenceAlerts.length > 0
                    ? 'warn'
                    : 'none';
            const confidenceAlertCount = confidenceAlerts.length;
            const hasCriticalAlert = confidenceAlertLevel === 'critical';

            let recommendedAction = 'Continue monitoring current mission mix';
            if (healthSeverity === 'critical') {
                if (freshnessBucket === 'stale' || freshnessBucket === 'unknown') {
                    recommendedAction = 'Refresh mission telemetry pipeline and verify mesh heartbeat/data flow';
                } else if (statusPercentages.failed >= 40) {
                    recommendedAction = 'Pause high-risk delegations and prioritize failed-mission recovery/resume workflows';
                } else if (deniedTrend === 'up') {
                    recommendedAction = 'Tighten tool policy or investigate denied-tool surge before continuing execution';
                } else {
                    recommendedAction = 'Escalate to operator review and reduce swarm concurrency until health improves';
                }
            } else if (healthSeverity === 'warn') {
                if (statusPercentages.failed >= 20) {
                    recommendedAction = 'Review failing mission cluster and apply targeted retries before launching new work';
                } else if (deniedTrend === 'up') {
                    recommendedAction = 'Audit recent policy denials and confirm expected tool allowances';
                } else {
                    recommendedAction = 'Increase observation cadence and keep risk threshold conservative';
                }
            }

            return {
                missionCount,
                averageRisk,
                maxRisk: missionCount > 0 ? Math.max(...riskScores) : 0,
                minObservedRisk: missionCount > 0 ? Math.min(...riskScores) : 0,
                dominantBand,
                health: {
                    severity: healthSeverity,
                    score: healthScore,
                    reasons: healthReasons.slice(0, 5),
                    recommendedAction,
                    confidence: {
                        score: confidenceScore,
                        level: confidenceLevel,
                        drivers: confidenceDrivers.slice(0, 4),
                        inputs: {
                            missionCount,
                            healthReasonCount: healthReasons.length,
                            freshnessBucket,
                            evaluatedAt: now
                        },
                        components: {
                            sampleSizePenalty,
                            freshnessPenalty,
                            signalCongestionPenalty,
                            totalPenalty: sampleSizePenalty + freshnessPenalty + signalCongestionPenalty
                        },
                        stability: confidenceStability,
                        advice: confidenceAdvice,
                        alertLevel: confidenceAlertLevel,
                        alertCount: confidenceAlertCount,
                        hasCriticalAlert,
                        alerts: confidenceAlerts.slice(0, 4),
                        uncertaintyMargin,
                        scoreRange
                    }
                },
                activity: {
                    deniedLast24h: deniedLast24hTotal,
                    deniedPrev24h: deniedPrev24hTotal,
                    deniedDelta,
                    deniedDeltaPct,
                    deniedTrend
                },
                freshness: {
                    generatedAt: now,
                    latestMissionUpdatedAt: freshestUpdatedAtMs || null,
                    latestUpdateAgeSeconds,
                    freshnessBucket
                },
                statusDistribution: {
                    counts: statusCounts,
                    percentages: statusPercentages
                },
                bands: {
                    low,
                    medium,
                    high
                },
                bandPercentages: {
                    low: lowPct,
                    medium: mediumPct,
                    high: highPct
                }
            };
        }),

    getMeshCapabilities: publicProcedure
        .query(async () => {
            return global.mcpServerInstance?.meshService?.getMeshCapabilities() || {};
        }),

    sendDirectMessage: publicProcedure
        .input(z.object({
            targetNodeId: z.string(),
            payload: z.any()
        }))
        .mutation(async ({ input }) => {
            const mesh = global.mcpServerInstance?.meshService;
            if (!mesh) throw new Error("MeshService not available");
            mesh.sendDirect(input.targetNodeId, SwarmMessageType.DIRECT_MESSAGE, input.payload);
            return { success: true };
        }),

    /**
     * Get the live transcript for the active SwarmController session.
     */
    getSwarmTranscript: publicProcedure.query(() => {
        return global.mcpServerInstance?.swarmController?.getTranscript() || [];
    }),
});
