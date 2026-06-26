import {
	Server,
	LayoutDashboard,
	Globe,
	Key,
	Shield,
	Terminal,
	Settings,
	Search,
	Activity,
	Zap,
	Users,
	Brain,
	Scroll,
	Library,
	FileCode2,
	Workflow,
	Power,
	FlaskRound,
	Rocket,
	Wrench,
	Download,
	GitBranch,
	BookOpen,
} from "lucide-react";

export interface NavItem {
	title: string;
	href: string;
	icon: any;
	variant: "default" | "ghost";
}

export interface NavSection {
	title: string;
	items: NavItem[];
}

export const META_MCP_NAV: NavItem[] = [
	{
		title: "MCP Dashboard",
		href: "/dashboard/mcp",
		icon: Server,
		variant: "default",
	},
	{
		title: "Always-On Tools",
		href: "/dashboard/mcp/always-on",
		icon: Power,
		variant: "ghost",
	},
	{
		title: "Tool Catalog",
		href: "/dashboard/mcp/catalog",
		icon: Search,
		variant: "ghost",
	},
	{
		title: "Tools Inspector",
		href: "/dashboard/mcp/inspector",
		icon: Wrench,
		variant: "ghost",
	},
	{
		title: "MCP Registry",
		href: "/dashboard/mcp/registry",
		icon: Download,
		variant: "ghost",
	},
	{
		title: "MCP Settings",
		href: "/dashboard/mcp/settings",
		icon: Settings,
		variant: "ghost",
	},
];

export const MAIN_DASHBOARD_NAV: NavItem[] = [
	{
		title: "Dashboard Home",
		href: "/dashboard",
		icon: LayoutDashboard,
		variant: "ghost",
	},
	{
		title: "Swarm & Agents",
		href: "/dashboard/swarm",
		icon: Users,
		variant: "ghost",
	},
	{
		title: "Brain & Memory",
		href: "/dashboard/brain",
		icon: Brain,
		variant: "ghost",
	},
	{
		title: "Context & Sessions",
		href: "/dashboard/session",
		icon: Scroll,
		variant: "ghost",
	},
	{
		title: "Knowledge & Skills",
		href: "/dashboard/library",
		icon: Library,
		variant: "ghost",
	},
	{
		title: "Code Platform",
		href: "/dashboard/code",
		icon: FileCode2,
		variant: "ghost",
	},
];

export const OPERATIONS_NAV: NavItem[] = [
	{
		title: "Diagnostics & Research",
		href: "/dashboard/research",
		icon: FlaskRound,
		variant: "ghost",
	},
	{
		title: "Command Console",
		href: "/dashboard/command",
		icon: Terminal,
		variant: "ghost",
	},
	{
		title: "Git Chronicle",
		href: "/dashboard/chronicle",
		icon: GitBranch,
		variant: "ghost",
	},
	{
		title: "User Manual",
		href: "/dashboard/manual",
		icon: BookOpen,
		variant: "ghost",
	},
	{
		title: "Workflows",
		href: "/dashboard/workflows",
		icon: Workflow,
		variant: "ghost",
	},
	{
		title: "Security & Audits",
		href: "/dashboard/security",
		icon: Shield,
		variant: "ghost",
	},
	{
		title: "Integrations Hub",
		href: "/dashboard/integrations",
		icon: Globe,
		variant: "ghost",
	},
	{
		title: "Cloud Orchestrator",
		href: "/dashboard/cloud-orchestrator",
		icon: Rocket,
		variant: "ghost",
	},
	{
		title: "Billing & Plans",
		href: "/dashboard/billing",
		icon: Key,
		variant: "ghost",
	},
	{
		title: "Global Settings",
		href: "/dashboard/settings",
		icon: Settings,
		variant: "ghost",
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
		title: "Operations & Tools",
		items: OPERATIONS_NAV,
	},
];
