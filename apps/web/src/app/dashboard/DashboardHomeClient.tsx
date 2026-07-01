"use client";

import { useEffect, useMemo, useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { trpc } from "../../utils/trpc";
import {
	DashboardHomeView,
	type DashboardFallbackSummary,
	type DashboardHealerSummary,
	type DashboardProviderSummary,
	type DashboardServerSummary,
	type DashboardSessionSummary,
	type DashboardStartupStatus,
	type DashboardStatusSummary,
	type DashboardTrafficSummary,
} from "./dashboard-home-view";
import { useGoSidecarDashboard } from "../../hooks/useGoSidecarData";
import { BrowserToolWidget } from "../../components/BrowserToolWidget";
import { VibeCheckWidget } from "../../components/VibeCheckWidget";

import { motion, AnimatePresence } from "framer-motion";

// Import sub-pages to consolidate
import ResearchPage from "./research/view";
import CommandDashboard from "./command/view";
import ChronicleDashboard from "./chronicle/view";
import ManualPage from "./manual/view";
import WorkflowsPage from "./workflows/view";
import SecurityPage from "./security/view";
import IntegrationsDashboard from "./integrations/view";
import CloudOrchestratorDashboardPage from "./cloud-orchestrator/view";
import ProviderAuthBillingMatrix from "./billing/view";
import SettingsDashboard from "./settings/view";

const SESSION_STATUS_PRIORITY: Record<
	DashboardSessionSummary["status"],
	number
> = {
	error: 6,
	restarting: 5,
	starting: 4,
	stopping: 3,
	running: 2,
	created: 1,
	stopped: 0,
};

const TABS = [
	{ id: "page-a", label: "Recovery & Sync" },
	{ id: "page-b", label: "Go MCP & Tools" },
	{ id: "page-c", label: "Memory & Skills" },
	{ id: "page-d", label: "Prompts & Deployments" },
] as const;

export function sortSessions(sessions: DashboardSessionSummary[]) {
	if (!Array.isArray(sessions)) return [];
	return [...sessions].sort((left, right) => {
		const priorityDelta =
			SESSION_STATUS_PRIORITY[right.status] -
			SESSION_STATUS_PRIORITY[left.status];
		if (priorityDelta !== 0) {
			return priorityDelta;
		}
		return right.lastActivityAt - left.lastActivityAt;
	});
}

function sortServers(servers: DashboardServerSummary[]) {
	if (!Array.isArray(servers)) return [];
	return [...servers].sort((left, right) =>
		left.name.localeCompare(right.name),
	);
}

export function DashboardHomeClient() {
	return (
		<Suspense
			fallback={<div className="p-8 text-zinc-500">Loading dashboard...</div>}
		>
			<DashboardHomeClientContent />
		</Suspense>
	);
}

function DashboardHomeClientContent() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const activeTab = searchParams.get("tab") || "page-a";

	const handleTabChange = (tabId: string) => {
		router.replace(`/dashboard?tab=${tabId}`);
	};

	const utils = trpc.useUtils();
	const toolsClient = trpc.tools as any;
	const [pendingSessionActionId, setPendingSessionActionId] = useState<
		string | null
	>(null);
	const [currentTimestamp, setCurrentTimestamp] = useState<number | null>(null);

	// Go sidecar fallback data — polled independently of tRPC
	const goData = useGoSidecarDashboard(5000);

	useEffect(() => {
		const refreshTimestamp = () => setCurrentTimestamp(Date.now());
		refreshTimestamp();
		const interval = window.setInterval(refreshTimestamp, 30_000);
		return () => window.clearInterval(interval);
	}, []);

	const mcpStatusQuery = trpc.mcp.getStatus.useQuery(undefined, {
		refetchInterval: 5000,
	});
	const startupStatusQuery = trpc.startupStatus.useQuery(undefined, {
		refetchInterval: 5000,
	});
	const serversQuery = trpc.mcp.listServers.useQuery(undefined, {
		refetchInterval: 5000,
	});
	const trafficQuery = trpc.mcp.traffic.useQuery(undefined, {
		refetchInterval: 3000,
	});
	const providerQuotasQuery = trpc.billing.getProviderQuotas.useQuery(
		undefined,
		{ refetchInterval: 10000 },
	);
	const fallbackChainQuery = trpc.billing.getFallbackChain.useQuery(undefined, {
		refetchInterval: 10000,
	});
	const sessionsQuery = trpc.session.list.useQuery(undefined, {
		refetchInterval: 3000,
	});
	const installArtifactsQuery = toolsClient?.detectInstallSurfaces?.useQuery
		? toolsClient.detectInstallSurfaces.useQuery(undefined, {
				refetchInterval: 10000,
			})
		: ({ data: null, refetch: async () => undefined } as {
				data: null;
				refetch: () => Promise<unknown>;
			});

	// Healer status — fetch history and vault record count from the Go HealerService
	const healerHistoryQuery = trpc.healer.getHistory.useQuery(undefined, {
		refetchInterval: 5000,
		retry: false,
	});
	const healerVaultCountQuery = trpc.healer.vaultRecordCount.useQuery(
		undefined,
		{
			refetchInterval: 10000,
			retry: false,
		},
	);

	// Determine if tRPC core is reachable
	const trpcReachable =
		mcpStatusQuery.data !== undefined &&
		mcpStatusQuery.data !== null &&
		!mcpStatusQuery.error;

	// Use Go sidecar data when tRPC is unreachable
	const useGoFallback = !trpcReachable && goData.connected;

	const isBootstrapping = !trpcReachable && !goData.connected;

	const refreshDashboard = async () => {
		await Promise.all([
			utils.mcp.getStatus.invalidate(),
			utils.startupStatus.invalidate(),
			utils.mcp.listServers.invalidate(),
			utils.mcp.traffic.invalidate(),
			utils.billing.getProviderQuotas.invalidate(),
			utils.billing.getFallbackChain.invalidate(),
			utils.session.list.invalidate(),
			installArtifactsQuery.refetch(),
			utils.healer.getHistory.invalidate(),
			utils.healer.vaultRecordCount.invalidate(),
		]);
	};

	const startSessionMutation = trpc.session.start.useMutation({
		onSettled: async () => {
			setPendingSessionActionId(null);
			await refreshDashboard();
		},
	});
	const stopSessionMutation = trpc.session.stop.useMutation({
		onSettled: async () => {
			setPendingSessionActionId(null);
			await refreshDashboard();
		},
	});
	const restartSessionMutation = trpc.session.restart.useMutation({
		onSettled: async () => {
			setPendingSessionActionId(null);
			await refreshDashboard();
		},
	});

	// ── MCP Status: tRPC → Go fallback → defaults ──
	const mcpStatus = useMemo<DashboardStatusSummary>(() => {
		if (trpcReachable && mcpStatusQuery.data) {
			return mcpStatusQuery.data as DashboardStatusSummary;
		}
		if (goData.mcpStatus) {
			return goData.mcpStatus as DashboardStatusSummary;
		}
		return {
			initialized: false,
			serverCount: 0,
			toolCount: 0,
			connectedCount: 0,
		};
	}, [trpcReachable, mcpStatusQuery.data, goData.mcpStatus]);

	// ── Startup Status: tRPC → Go fallback → defaults ──
	const startupStatus = useMemo<DashboardStartupStatus>(() => {
		if (trpcReachable && startupStatusQuery.data) {
			return startupStatusQuery.data as DashboardStartupStatus;
		}
		if (goData.startupStatus) {
			return goData.startupStatus as DashboardStartupStatus;
		}
		return {
			status: "starting",
			ready: false,
			uptime: 0,
			checks: {
				mcpAggregator: {
					ready: false,
					liveReady: false,
					serverCount: 0,
					connectedCount: 0,
					initialization: null,
					persistedServerCount: 0,
					persistedToolCount: 0,
					configuredServerCount: 0,
					advertisedServerCount: 0,
					advertisedToolCount: 0,
					advertisedAlwaysOnServerCount: 0,
					advertisedAlwaysOnToolCount: 0,
					inventoryReady: false,
					warmupInProgress: false,
				},
				configSync: { ready: false, status: null },
				memory: { ready: false, initialized: false, agentMemory: false },
				browser: { ready: false, active: false, pageCount: 0 },
				sessionSupervisor: { ready: false, sessionCount: 0, restore: null },
				extensionBridge: {
					ready: false,
					acceptingConnections: false,
					clientCount: 0,
					hasConnectedClients: false,
				},
				executionEnvironment: {
					ready: false,
					preferredShellId: null,
					preferredShellLabel: null,
					shellCount: 0,
					verifiedShellCount: 0,
					toolCount: 0,
					verifiedToolCount: 0,
					harnessCount: 0,
					verifiedHarnessCount: 0,
					supportsPowerShell: false,
					supportsPosixShell: false,
					notes: [],
				},
			},
		} as DashboardStartupStatus;
	}, [trpcReachable, startupStatusQuery.data, goData.startupStatus]);

	// ── Servers: tRPC → Go fallback → empty ──
	const servers = useMemo<DashboardServerSummary[]>(() => {
		if (trpcReachable && serversQuery.data) {
			return sortServers(serversQuery.data as DashboardServerSummary[]);
		}
		if (goData.servers.length > 0) {
			return sortServers(goData.servers as DashboardServerSummary[]);
		}
		return [];
	}, [trpcReachable, serversQuery.data, goData.servers]);

	// ── Traffic: tRPC only (Go sidecar doesn't track traffic yet) ──
	const traffic = useMemo<DashboardTrafficSummary[]>(
		() => {
			const data = trafficQuery.data;
			if (!Array.isArray(data)) return [];
			return [...data].sort((left, right) => right.timestamp - left.timestamp);
		},
		[trafficQuery.data],
	);

	// ── Providers: tRPC → Go fallback → empty ──
	const providers = useMemo<DashboardProviderSummary[]>(() => {
		if (trpcReachable && providerQuotasQuery.data) {
			return providerQuotasQuery.data as DashboardProviderSummary[];
		}
		if (goData.providers.length > 0) {
			return goData.providers as DashboardProviderSummary[];
		}
		return [];
	}, [trpcReachable, providerQuotasQuery.data, goData.providers]);

	// ── Fallback chain: tRPC → Go fallback → empty ──
	const fallbackChain = useMemo<DashboardFallbackSummary[]>(() => {
		if (trpcReachable && fallbackChainQuery.data?.chain) {
			return fallbackChainQuery.data.chain as DashboardFallbackSummary[];
		}
		if (goData.fallbackChain.length > 0) {
			return goData.fallbackChain as DashboardFallbackSummary[];
		}
		return [];
	}, [trpcReachable, fallbackChainQuery.data, goData.fallbackChain]);

	// ── Sessions: tRPC → Go fallback → empty ──
	const sessions = useMemo<DashboardSessionSummary[]>(() => {
		if (trpcReachable && sessionsQuery.data && Array.isArray(sessionsQuery.data)) {
			return sortSessions(sessionsQuery.data as DashboardSessionSummary[]);
		}
		if (goData.sessions && Array.isArray(goData.sessions) && goData.sessions.length > 0) {
			return sortSessions(goData.sessions as DashboardSessionSummary[]);
		}
		return [];
	}, [trpcReachable, sessionsQuery.data, goData.sessions]);

	// ── Healer Status: derived from healer history and vault queries ──
	const healerStatus = useMemo<DashboardHealerSummary>(() => {
		const history = (healerHistoryQuery.data ?? []) as any[];
		const vaultRecordCount = (healerVaultCountQuery.data as number) ?? 0;
		const activePathogens = history.filter((e: any) => !e.success).length;
		const resolvedCount = history.filter((e: any) => e.success).length;
		const total = activePathogens + resolvedCount;
		const successRate =
			total > 0 ? Math.round((resolvedCount / total) * 100) : 100;
		const lastSuccess = history.filter((e: any) => e.success);
		const lastHealTime =
			lastSuccess.length > 0
				? new Date(
						lastSuccess[lastSuccess.length - 1]?.timestamp,
					).toLocaleString()
				: null;
		const isLive =
			healerHistoryQuery.isSuccess || healerVaultCountQuery.isSuccess;
		return {
			activePathogens,
			resolvedCount,
			successRate,
			lastHealTime,
			vaultRecordCount,
			isLive,
		};
	}, [
		healerHistoryQuery.data,
		healerVaultCountQuery.data,
		healerHistoryQuery.isSuccess,
		healerVaultCountQuery.isSuccess,
	]);

	const renderActiveTab = () => {
		return (
			<DashboardHomeView
				activeTab={activeTab as "page-a" | "page-b" | "page-c" | "page-d"}
				generatedAtLabel={
					currentTimestamp
						? new Date(currentTimestamp).toLocaleTimeString()
						: "just now"
				}
				currentTimestamp={currentTimestamp}
				isBootstrapping={isBootstrapping}
				mcpStatus={mcpStatus}
				startupStatus={startupStatus}
				servers={servers}
				traffic={traffic}
				providers={providers}
				fallbackChain={fallbackChain}
				sessions={sessions}
				healerStatus={healerStatus}
				installSurfaceArtifacts={installArtifactsQuery.data ?? null}
				onStartSession={(sessionId) => {
					setPendingSessionActionId(sessionId);
					startSessionMutation.mutate({ id: sessionId });
				}}
				onStopSession={(sessionId) => {
					setPendingSessionActionId(sessionId);
					stopSessionMutation.mutate({ id: sessionId });
				}}
				onRestartSession={(sessionId) => {
					setPendingSessionActionId(sessionId);
					restartSessionMutation.mutate({ id: sessionId });
				}}
				pendingSessionActionId={pendingSessionActionId}
			>
				<BrowserToolWidget />
				<VibeCheckWidget />
			</DashboardHomeView>
		);
	};

	return (
		<div className="flex flex-col min-h-screen bg-black text-zinc-100">
			{/* Sleek Sub-navigation Tab Bar */}
			<div className="sticky top-0 z-20 flex overflow-x-auto border-b border-zinc-800 bg-zinc-950/95 backdrop-blur px-6 py-3 scrollbar-none gap-2">
				{TABS.map((tab) => {
					const isActive = activeTab === tab.id;
					return (
						<button
							key={tab.id}
							type="button"
							onClick={() => handleTabChange(tab.id)}
							className={`relative px-4 py-2 text-xs font-semibold rounded-md whitespace-nowrap transition-colors duration-200 ${
								isActive
									? "text-white shadow-sm"
									: "text-zinc-400 hover:text-zinc-200"
							}`}
						>
							{isActive && (
								<motion.div
									layoutId="activeTabHome"
									className="absolute inset-0 bg-gradient-to-r from-zinc-800 to-zinc-900 rounded-md -z-10 border border-zinc-700/50 shadow-inner"
									transition={{ type: "spring", stiffness: 380, damping: 30 }}
								/>
							)}
							{tab.label}
						</button>
					);
				})}
			</div>
			<div className="flex-1 p-4 md:p-6 min-h-0 overflow-auto">
				<AnimatePresence mode="wait">
					<motion.div
						key={activeTab}
						initial={{ opacity: 0, y: 10 }}
						animate={{ opacity: 1, y: 0 }}
						exit={{ opacity: 0, y: -10 }}
						transition={{ duration: 0.15 }}
						className="h-full"
					>
						{renderActiveTab()}
					</motion.div>
				</AnimatePresence>
			</div>
		</div>
	);
}
