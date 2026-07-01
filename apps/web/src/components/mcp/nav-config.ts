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
	TrendingUp,
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
		href: "/dashboard?tab=page-b",
		icon: Server,
		variant: "default",
		tooltip:
			"MCP server overview: connected servers, tool counts, lifecycle status",
	},
	{
		title: "Always-On Tools",
		href: "/dashboard?tab=page-b",
		icon: Power,
		variant: "ghost",
		tooltip:
			"Toggle which built-in tools are always available to the MCP client",
	},
	{
		title: "Tool Catalog",
		href: "/dashboard?tab=page-b",
		icon: Search,
		variant: "ghost",
		tooltip:
			"Browse and search the full MCP tool catalog from all registered servers",
	},
	{
		title: "Tools Inspector",
		href: "/dashboard?tab=page-b",
		icon: Wrench,
		variant: "ghost",
		tooltip: "Inspect tool definitions, parameters, and schemas in detail",
	},
	{
		title: "MCP Registry",
		href: "/dashboard?tab=page-b",
		icon: Download,
		variant: "ghost",
		tooltip:
			"Discover and install new MCP servers from public registries (Glama, Smithery)",
	},
	{
		title: "Tool Chains",
		href: "/dashboard?tab=page-b",
		icon: Webhook,
		variant: "ghost",
		tooltip: "Chain multiple tools together into reusable automated workflows",
	},
	{
		title: "MCP Settings",
		href: "/dashboard?tab=page-b",
		icon: Settings,
		variant: "ghost",
		tooltip: "Configure MCP server connections, endpoints, and preferences",
	},
];

// ── Core System ──

export const MAIN_DASHBOARD_NAV: NavItem[] = [
	{
		title: "Dashboard Home",
		href: "/dashboard?tab=page-a",
		icon: LayoutDashboard,
		variant: "ghost",
		tooltip: "System overview: active sessions, recent activity, health status",
	},
	{
		title: "Swarm & Agents",
		href: "/dashboard?tab=page-b",
		icon: Users,
		variant: "ghost",
		tooltip:
			"Multi-agent orchestration: missions, debates, consensus, and agent management",
	},
	{
		title: "Council & Governance",
		href: "/dashboard?tab=page-b",
		icon: Shield,
		variant: "ghost",
		tooltip:
			"AI governance: council debates, approval workflow, autonomy levels, policies",
	},
	{
		title: "Brain & Memory",
		href: "/dashboard?tab=page-c",
		icon: Brain,
		variant: "ghost",
		tooltip:
			"Memory system: L2 vault, spaced repetition, sleep cycle, FTS5 search",
	},
	{
		title: "Memory Explorer",
		href: "/dashboard?tab=page-c",
		icon: Database,
		variant: "ghost",
		tooltip: "Full-text search across 86K+ memories with L4 limbo management",
	},
	{
		title: "Memory Analytics",
		href: "/dashboard?tab=page-c",
		icon: TrendingUp,
		variant: "ghost",
		tooltip:
			"Memory system overview: tier stats, heat distribution, kind breakdown, lifecycle pipeline",
	},
	{
		title: "Tool Karma",
		href: "/dashboard?tab=page-b",
		icon: Wrench,
		variant: "ghost",
		tooltip: "Native Go tool registry, handler health, and usage tracking",
	},
	{
		title: "Tool Console",
		href: "/dashboard?tab=page-b",
		icon: Terminal,
		variant: "ghost",
		tooltip: "Browse, inspect, and execute native Go tools interactively",
	},
	{
		title: "Context & Sessions",
		href: "/dashboard?tab=page-c",
		icon: Scroll,
		variant: "ghost",
		tooltip: "Imported sessions, context management, session export/import",
	},
	{
		title: "Knowledge & Skills",
		href: "/dashboard?tab=page-c",
		icon: Library,
		variant: "ghost",
		tooltip:
			"Skill registry, knowledge graph, RAG ingestion, and directory browser",
	},
	{
		title: "Code Platform",
		href: "/dashboard?tab=page-b",
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
		href: "/dashboard?tab=page-a",
		icon: Cpu,
		variant: "ghost",
		tooltip:
			"Live runtime overview: services, locks, startup readiness, imports",
	},
	{
		title: "Mesh Network",
		href: "/dashboard?tab=page-a",
		icon: Network,
		variant: "ghost",
		tooltip:
			"P2P memory sync mesh: peers, capabilities, broadcasts across machines",
	},
	{
		title: "Providers & Billing",
		href: "/dashboard?tab=page-a",
		icon: Key,
		variant: "ghost",
		tooltip:
			"LLM provider routing, fallback chains, quotas, cost history, model pricing",
	},
	{
		title: "Observability",
		href: "/dashboard?tab=page-a",
		icon: Eye,
		variant: "ghost",
		tooltip:
			"System pulse: event streams, provider status, real-time monitoring",
	},
	{
		title: "Logs & Metrics",
		href: "/dashboard?tab=page-a",
		icon: BarChart3,
		variant: "ghost",
		tooltip:
			"System logs, provider breakdown, routing history, system snapshots",
	},
	{
		title: "Browser Automation",
		href: "/dashboard?tab=page-a",
		icon: Globe,
		variant: "ghost",
		tooltip:
			"Browser controls: pages, history, console logs, scraping, screenshots",
	},
	{
		title: "Workflows",
		href: "/dashboard?tab=page-d",
		icon: Workflow,
		variant: "ghost",
		tooltip: "Workflow engine: definitions, executions, canvases, approvals",
	},
	{
		title: "Diagnostics & Research",
		href: "/dashboard?tab=page-c",
		icon: FlaskRound,
		variant: "ghost",
		tooltip:
			"Deep research, recursive web crawling, URL ingestion, research queue",
	},
	{
		title: "Command Console",
		href: "/dashboard?tab=page-b",
		icon: Terminal,
		variant: "ghost",
		tooltip: "CLI harness detection, command registry, shell history",
	},
	{
		title: "Healer & Auto-Repair",
		href: "/dashboard?tab=page-a",
		icon: Bug,
		variant: "ghost",
		tooltip: "Self-healing: diagnose errors, auto-repair, repair history",
	},
];

// ── Data & Integrations ──

export const DATA_NAV: NavItem[] = [
	{
		title: "Session Imports",
		href: "/dashboard?tab=page-a",
		icon: DownloadCloud,
		variant: "ghost",
		tooltip:
			"Import external sessions from Claude, Gemini, Aider, and other tools",
	},
	{
		title: "CLI Harnesses",
		href: "/dashboard?tab=page-a",
		icon: Command,
		variant: "ghost",
		tooltip: "Detected CLI harnesses: versions, capabilities, install surfaces",
	},
	{
		title: "Browser Extension",
		href: "/dashboard?tab=page-a",
		icon: Bot,
		variant: "ghost",
		tooltip: "Browser extension bridge: memories, DOM parsing, stats",
	},
	{
		title: "Cloud Development",
		href: "/dashboard?tab=page-a",
		icon: Cloud,
		variant: "ghost",
		tooltip: "Cloud dev sessions: providers, messages, plans, logs",
	},
	{
		title: "DeerFlow",
		href: "/dashboard?tab=page-a",
		icon: Radio,
		variant: "ghost",
		tooltip: "DeerFlow bridge: models, skills, memory status",
	},
	{
		title: "Integrations Hub",
		href: "/dashboard?tab=page-a",
		icon: Globe,
		variant: "ghost",
		tooltip:
			"External integrations: Open WebUI, Ollama, and third-party bridges",
	},
	{
		title: "Git Chronicle",
		href: "/dashboard?tab=page-a",
		icon: GitBranch,
		variant: "ghost",
		tooltip: "Git history, commit log, repository change tracking",
	},
	{
		title: "Cold Archive",
		href: "/dashboard?tab=page-c",
		icon: Snowflake,
		variant: "ghost",
		tooltip: "L3 cold storage for low-heat memories: browse, search, promote",
	},
	{
		title: "Session Import",
		href: "/dashboard?tab=page-c",
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
		href: "/dashboard?tab=page-a",
		icon: Key,
		variant: "ghost",
		tooltip: "Manage API keys, OAuth clients, authentication providers",
	},
	{
		title: "Security & Audits",
		href: "/dashboard?tab=page-a",
		icon: Shield,
		variant: "ghost",
		tooltip: "Audit logs, security policies, access control, compliance",
	},
	{
		title: "User Manual",
		href: "/dashboard?tab=page-a",
		icon: BookOpen,
		variant: "ghost",
		tooltip: "Built-in documentation for all TormentNexus features",
	},
	{
		title: "Global Settings",
		href: "/dashboard?tab=page-a",
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
