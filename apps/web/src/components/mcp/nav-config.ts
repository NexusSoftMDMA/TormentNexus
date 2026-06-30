import {
	Server,
	LayoutDashboard,
	Database,
	Globe,
	Key,
	Shield,
	Terminal,
	Settings,
	Search,
	Users,
	Brain,
	Scroll,
	Library,
	FileCode2,
	Workflow,
	Power,
	FlaskRound,
	Wrench,
	Download,
	GitBranch,
	BookOpen,
	Network,
	Radio,
	Eye,
	BarChart3,
	Cloud,
	Bug,
	Webhook,
	Bot,
	Cpu,
	DownloadCloud,
	Command,
	Snowflake,
	FileUp,
} from "lucide-react";

export interface NavItem {
	title: string;
	href: string;
	icon: any;
	variant: "default" | "ghost";
	tooltip?: string;
}

export interface NavSection {
	title: string;
	items: NavItem[];
}

// ── MCP & Tool Platform ──

export const META_MCP_NAV: NavItem[] = [
	{
		title: "MCP Dashboard",
		href: "/dashboard/mcp?tab=dashboard",
		icon: Server,
		variant: "default",
		tooltip:
			"MCP server overview: connected servers, tool counts, lifecycle status",
	},
	{
		title: "Always-On Tools",
		href: "/dashboard/mcp?tab=always-on",
		icon: Power,
		variant: "ghost",
		tooltip:
			"Toggle which built-in tools are always available to the MCP client",
	},
	{
		title: "Tool Catalog",
		href: "/dashboard/mcp?tab=catalog",
		icon: Search,
		variant: "ghost",
		tooltip:
			"Browse and search the full MCP tool catalog from all registered servers",
	},
	{
		title: "Tools Inspector",
		href: "/dashboard/mcp?tab=inspector",
		icon: Wrench,
		variant: "ghost",
		tooltip: "Inspect tool definitions, parameters, and schemas in detail",
	},
	{
		title: "MCP Registry",
		href: "/dashboard/mcp?tab=registry",
		icon: Download,
		variant: "ghost",
		tooltip:
			"Discover and install new MCP servers from public registries (Glama, Smithery)",
	},
	{
		title: "Tool Chains",
		href: "/dashboard/mcp?tab=tool-chains",
		icon: Webhook,
		variant: "ghost",
		tooltip: "Chain multiple tools together into reusable automated workflows",
	},
	{
		title: "MCP Settings",
		href: "/dashboard/mcp?tab=settings",
		icon: Settings,
		variant: "ghost",
		tooltip: "Configure MCP server connections, endpoints, and preferences",
	},
];

// ── Core System ──

export const MAIN_DASHBOARD_NAV: NavItem[] = [
	{
		title: "Dashboard Home",
		href: "/dashboard?tab=home",
		icon: LayoutDashboard,
		variant: "ghost",
		tooltip: "System overview: active sessions, recent activity, health status",
	},
	{
		title: "Swarm & Agents",
		href: "/dashboard/swarm?tab=swarm",
		icon: Users,
		variant: "ghost",
		tooltip:
			"Multi-agent orchestration: missions, debates, consensus, and agent management",
	},
	{
		title: "Council & Governance",
		href: "/dashboard/swarm?tab=council",
		icon: Shield,
		variant: "ghost",
		tooltip:
			"AI governance: council debates, approval workflow, autonomy levels, policies",
	},
	{
		title: "Brain & Memory",
		href: "/dashboard/swarm?tab=brain",
		icon: Brain,
		variant: "ghost",
		tooltip:
			"Memory system: L2 vault, spaced repetition, sleep cycle, FTS5 search",
	},
	{
		title: "Memory Explorer",
		href: "/dashboard/memory-search",
		icon: Database,
		variant: "ghost",
		tooltip: "Full-text search across 86K+ memories with L4 limbo management",
	},
	{
		title: "Context & Sessions",
		href: "/dashboard/swarm?tab=session",
		icon: Scroll,
		variant: "ghost",
		tooltip: "Imported sessions, context management, session export/import",
	},
	{
		title: "Knowledge & Skills",
		href: "/dashboard/swarm?tab=library",
		icon: Library,
		variant: "ghost",
		tooltip:
			"Skill registry, knowledge graph, RAG ingestion, and directory browser",
	},
	{
		title: "Code Platform",
		href: "/dashboard/swarm?tab=code",
		icon: FileCode2,
		variant: "ghost",
		tooltip:
			"AutoDev loops, code execution sandbox, LSP diagnostics, symbol search",
	},
];

// ── Infrastructure & Operations ──

export const OPERATIONS_NAV: NavItem[] = [
	{
		title: "Runtime Status",
		href: "/dashboard/runtime",
		icon: Cpu,
		variant: "ghost",
		tooltip:
			"Live runtime overview: services, locks, startup readiness, imports",
	},
	{
		title: "Mesh Network",
		href: "/dashboard/mesh",
		icon: Network,
		variant: "ghost",
		tooltip:
			"P2P memory sync mesh: peers, capabilities, broadcasts across machines",
	},
	{
		title: "Providers & Billing",
		href: "/dashboard?tab=billing",
		icon: Key,
		variant: "ghost",
		tooltip:
			"LLM provider routing, fallback chains, quotas, cost history, model pricing",
	},
	{
		title: "Observability",
		href: "/dashboard/observability",
		icon: Eye,
		variant: "ghost",
		tooltip:
			"System pulse: event streams, provider status, real-time monitoring",
	},
	{
		title: "Logs & Metrics",
		href: "/dashboard/logs-metrics",
		icon: BarChart3,
		variant: "ghost",
		tooltip:
			"System logs, provider breakdown, routing history, system snapshots",
	},
	{
		title: "Browser Automation",
		href: "/dashboard/browser",
		icon: Globe,
		variant: "ghost",
		tooltip:
			"Browser controls: pages, history, console logs, scraping, screenshots",
	},
	{
		title: "Workflows",
		href: "/dashboard?tab=workflows",
		icon: Workflow,
		variant: "ghost",
		tooltip: "Workflow engine: definitions, executions, canvases, approvals",
	},
	{
		title: "Diagnostics & Research",
		href: "/dashboard?tab=research",
		icon: FlaskRound,
		variant: "ghost",
		tooltip:
			"Deep research, recursive web crawling, URL ingestion, research queue",
	},
	{
		title: "Command Console",
		href: "/dashboard?tab=command",
		icon: Terminal,
		variant: "ghost",
		tooltip: "CLI harness detection, command registry, shell history",
	},
	{
		title: "Healer & Auto-Repair",
		href: "/dashboard/healer",
		icon: Bug,
		variant: "ghost",
		tooltip: "Self-healing: diagnose errors, auto-repair, repair history",
	},
];

// ── Data & Integrations ──

export const DATA_NAV: NavItem[] = [
	{
		title: "Session Imports",
		href: "/dashboard/imports",
		icon: DownloadCloud,
		variant: "ghost",
		tooltip:
			"Import external sessions from Claude, Gemini, Aider, and other tools",
	},
	{
		title: "CLI Harnesses",
		href: "/dashboard/cli-harnesses",
		icon: Command,
		variant: "ghost",
		tooltip: "Detected CLI harnesses: versions, capabilities, install surfaces",
	},
	{
		title: "Browser Extension",
		href: "/dashboard/browser-extension",
		icon: Bot,
		variant: "ghost",
		tooltip: "Browser extension bridge: memories, DOM parsing, stats",
	},
	{
		title: "Cloud Development",
		href: "/dashboard/cloud-dev",
		icon: Cloud,
		variant: "ghost",
		tooltip: "Cloud dev sessions: providers, messages, plans, logs",
	},
	{
		title: "DeerFlow",
		href: "/dashboard/deerflow",
		icon: Radio,
		variant: "ghost",
		tooltip: "DeerFlow bridge: models, skills, memory status",
	},
	{
		title: "Integrations Hub",
		href: "/dashboard?tab=integrations",
		icon: Globe,
		variant: "ghost",
		tooltip:
			"External integrations: Open WebUI, Ollama, and third-party bridges",
	},
	{
		title: "Git Chronicle",
		href: "/dashboard?tab=chronicle",
		icon: GitBranch,
		variant: "ghost",
		tooltip: "Git history, commit log, repository change tracking",
	},
	{
		title: "Cold Archive",
		href: "/dashboard/cold-archive",
		icon: Snowflake,
		variant: "ghost",
		tooltip: "L3 cold storage for low-heat memories: browse, search, promote",
	},
	{
		title: "Session Import",
		href: "/dashboard/sessions/import",
		icon: FileUp,
		variant: "ghost",
		tooltip:
			"Scan and import sessions from external tools into the memory vault",
	},
];

// ── Security, Settings & Admin ──

export const ADMIN_NAV: NavItem[] = [
	{
		title: "API Keys & Auth",
		href: "/dashboard/api-keys",
		icon: Key,
		variant: "ghost",
		tooltip: "Manage API keys, OAuth clients, authentication providers",
	},
	{
		title: "Security & Audits",
		href: "/dashboard?tab=security",
		icon: Shield,
		variant: "ghost",
		tooltip: "Audit logs, security policies, access control, compliance",
	},
	{
		title: "User Manual",
		href: "/dashboard?tab=manual",
		icon: BookOpen,
		variant: "ghost",
		tooltip: "Built-in documentation for all TormentNexus features",
	},
	{
		title: "Global Settings",
		href: "/dashboard?tab=settings",
		icon: Settings,
		variant: "ghost",
		tooltip: "System settings: environment, providers, config files",
	},
];

export const SIDEBAR_SECTIONS: NavSection[] = [
	{
		title: "MCP Control",
		items: META_MCP_NAV,
	},
	{
		title: "Agent Core",
		items: MAIN_DASHBOARD_NAV,
	},
	{
		title: "Infrastructure",
		items: OPERATIONS_NAV,
	},
	{
		title: "Data & Integrations",
		items: DATA_NAV,
	},
	{
		title: "Admin",
		items: ADMIN_NAV,
	},
];
