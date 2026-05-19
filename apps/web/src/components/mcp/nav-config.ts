import {
    Server,
    LayoutDashboard,
    Box,
    Globe,
    Key,
    Layers,
    Shield,
    FileCode,
    FileText,
    Settings,
    Search,
    BookOpen,
    Activity,
    Zap,
    Bot,
    Wrench,
    Download,
    Rocket,
    Brain,
    FlaskConical,
    Terminal,
    FileSearch,
    Settings2,
    Workflow,
    Library,
    BookOpenText,
    BarChart3,
    Hammer,
    Users,
    Eye,
    Heart,
    BookMarked,
    Building2,
    Lightbulb,
    Cog,
    FileCode2,
    ScrollText,
    Sparkles,
    Radio,
    Network,
    ShoppingBag,
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
        title: "MetaMCP Dashboard",
        href: "/dashboard/mcp",
        icon: Server,
        variant: "default",
    },
    {
        title: "MetaMCP Servers",
        href: "/dashboard/mcp/metamcp",
        icon: Network,
        variant: "ghost",
    },
    {
        title: "Namespaces",
        href: "/dashboard/mcp/namespaces",
        icon: Box,
        variant: "ghost",
    },
    {
        title: "Endpoints",
        href: "/dashboard/mcp/endpoints",
        icon: Globe,
        variant: "ghost",
    },
    {
        title: "API Keys",
        href: "/dashboard/mcp/api-keys",
        icon: Key,
        variant: "ghost",
    },
    {
        title: "Tool Sets",
        href: "/dashboard/mcp/tool-sets",
        icon: Layers,
        variant: "ghost",
    },
    {
        title: "Policies",
        href: "/dashboard/mcp/policies",
        icon: Shield,
        variant: "ghost",
    },
    {
        title: "Internal Scripts",
        href: "/dashboard/mcp/scripts",
        icon: FileCode,
        variant: "ghost",
    },
    {
        title: "System Audit",
        href: "/dashboard/mcp/audit",
        icon: FileText,
        variant: "ghost",
    },
    {
        title: "Logs",
        href: "/dashboard/mcp/logs",
        icon: Activity,
        variant: "ghost",
    },
    {
        title: "Observability",
        href: "/dashboard/mcp/observability",
        icon: Zap,
        variant: "ghost",
    },
    {
        title: "Inspector",
        href: "/dashboard/mcp/inspector",
        icon: Wrench,
        variant: "ghost",
    },
    {
        title: "Agent Playground",
        href: "/dashboard/mcp/agent",
        icon: Bot,
        variant: "ghost",
    },
    {
        title: "AI Tools",
        href: "/dashboard/mcp/ai-tools",
        icon: Sparkles,
        variant: "ghost",
    },
    {
        title: "System",
        href: "/dashboard/mcp/system",
        icon: Activity,
        variant: "ghost",
    },
    {
        title: "Search",
        href: "/dashboard/mcp/search",
        icon: Search,
        variant: "ghost",
    },
    {
        title: "Jules",
        href: "/dashboard/jules",
        icon: Rocket,
        variant: "ghost",
    },
    {
        title: "Registry",
        href: "/dashboard/mcp/registry",
        icon: Download,
        variant: "ghost",
    },
    {
        title: "Tool Catalog",
        href: "/dashboard/mcp/catalog",
        icon: Search,
        variant: "ghost",
    },
    {
        title: "Documentation",
        href: "/dashboard/mcp/docs",
        icon: BookOpen,
        variant: "ghost",
    },
    {
        title: "Configuration",
        href: "/dashboard/mcp/settings",
        icon: Settings,
        variant: "ghost",
    },
];

export const INTEGRATIONS_NAV: NavItem[] = [
    {
        title: "OpenCode Autopilot",
        href: "/dashboard/autopilot",
        icon: Sparkles,
        variant: "ghost",
    },
    {
        title: "Jules",
        href: "/dashboard/jules",
        icon: Rocket,
        variant: "ghost",
    },
];

export const MAIN_DASHBOARD_NAV: NavItem[] = [
    { title: "Dashboard Home", href: "/dashboard", icon: LayoutDashboard, variant: "ghost" },
    { title: "Director", href: "/dashboard/director", icon: Bot, variant: "ghost" },
    { title: "Council", href: "/dashboard/council", icon: Users, variant: "ghost" },
    { title: "Supervisor", href: "/dashboard/supervisor", icon: Eye, variant: "ghost" },
    { title: "Swarm", href: "/dashboard/swarm", icon: Users, variant: "ghost" },
    { title: "Brain", href: "/dashboard/brain", icon: Brain, variant: "ghost" },
    { title: "Research", href: "/dashboard/research", icon: FlaskConical, variant: "ghost" },
    { title: "Memory", href: "/dashboard/memory", icon: Brain, variant: "ghost" },
    { title: "Knowledge", href: "/dashboard/knowledge", icon: Network, variant: "ghost" },
    { title: "Code", href: "/dashboard/code", icon: FileCode2, variant: "ghost" },
    { title: "Command", href: "/dashboard/command", icon: Terminal, variant: "ghost" },
    { title: "Inspector", href: "/dashboard/inspector", icon: FileSearch, variant: "ghost" },
    { title: "Settings", href: "/dashboard/settings", icon: Settings2, variant: "ghost" },
    { title: "Workflows", href: "/dashboard/workflows", icon: Workflow, variant: "ghost" },
    { title: "Library", href: "/dashboard/library", icon: Library, variant: "ghost" },
    { title: "Manual", href: "/dashboard/manual", icon: BookOpenText, variant: "ghost" },
    { title: "Plans", href: "/dashboard/plans", icon: Lightbulb, variant: "ghost" },
    { title: "Metrics", href: "/dashboard/metrics", icon: BarChart3, variant: "ghost" },
    { title: "Marketplace", href: "/dashboard/marketplace", icon: ShoppingBag, variant: "ghost" },
    { title: "Skills", href: "/dashboard/skills", icon: Hammer, variant: "ghost" },
    { title: "Squads", href: "/dashboard/squads", icon: Users, variant: "ghost" },
    { title: "Healer", href: "/dashboard/healer", icon: Heart, variant: "ghost" },
    { title: "Security", href: "/dashboard/security", icon: Shield, variant: "ghost" },
    { title: "Events", href: "/dashboard/events", icon: Activity, variant: "ghost" },
    { title: "Pulse", href: "/dashboard/pulse", icon: Radio, variant: "ghost" },
    { title: "Reader", href: "/dashboard/reader", icon: BookMarked, variant: "ghost" },
    { title: "Architecture", href: "/dashboard/architecture", icon: Building2, variant: "ghost" },
    { title: "Evolution", href: "/dashboard/evolution", icon: Sparkles, variant: "ghost" },
    { title: "Config", href: "/dashboard/config", icon: Cog, variant: "ghost" },
    { title: "Chronicle", href: "/dashboard/chronicle", icon: ScrollText, variant: "ghost" },
    { title: "Submodules", href: "/dashboard/submodules", icon: FileCode, variant: "ghost" },
    { title: "Workshop", href: "/dashboard/workshop", icon: Wrench, variant: "ghost" },
    { title: "Super Assistant", href: "/dashboard/super-assistant", icon: Bot, variant: "ghost" },
    { title: "Billing", href: "/dashboard/billing", icon: Key, variant: "ghost" },
];

export const SIDEBAR_SECTIONS: NavSection[] = [
    {
        title: "MetaMCP Tools",
        items: META_MCP_NAV,
    },
    {
        title: "Integrations",
        items: INTEGRATIONS_NAV,
    },
    {
        title: "Main Dashboard + Subpages",
        items: MAIN_DASHBOARD_NAV,
    },
];
