export interface BillingProviderQuotaSummary {
    provider: string;
    name: string;
    configured: boolean;
    authenticated?: boolean;
    authMethod?: string;
    tier?: string;
    limit?: number | null;
    used?: number;
    remaining?: number | null;
    resetDate?: string | null;
    rateLimitRpm?: number | null;
    availability?: string;
    lastError?: string | null;
    windows?: Array<{
        key: string;
        label: string;
        used: number;
        limit: number | null;
        remaining: number | null;
        resetDate: string | null;
        unit: string;
    }>;
    source?: string | null;
    connectionId?: string | null;
}

export interface BillingTaskRoutingRuleSummary {
    taskType: 'coding' | 'planning' | 'research' | 'general' | 'worker' | 'supervisor';
    strategy: 'cheapest' | 'best' | 'round-robin';
    fallbackPreview: Array<{
        provider: string;
        model?: string;
        reason?: string;
    }>;
}

export type BillingRoutingStrategy = BillingTaskRoutingRuleSummary['strategy'];

type ProviderPortalActionKind =
    | 'keys'
    | 'usage'
    | 'billing'
    | 'plan'
    | 'credits'
    | 'console'
    | 'cloud'
    | 'workspace'
    | 'token'
    | 'docs';

interface ProviderPortalAction {
    label: string;
    href: string;
    kind: ProviderPortalActionKind;
}

interface ProviderPortalDefinition {
    id: string;
    label: string;
    notes: string;
    actions: ProviderPortalAction[];
}

export interface ProviderPortalCard extends ProviderPortalDefinition {
    statusLabel: string;
    statusTone: 'success' | 'warning' | 'muted';
    authLabel: string;
    availabilityLabel: string;
    errorLabel: string | null;
}

interface ProviderQuickAccessLinkDefinition {
    providerId: string;
    preferredKinds: ProviderPortalActionKind[];
}

interface ProviderQuickAccessSectionDefinition {
    id: string;
    title: string;
    description: string;
    links: ProviderQuickAccessLinkDefinition[];
}

export interface ProviderQuickAccessLink {
    providerId: string;
    providerLabel: string;
    actionLabel: string;
    href: string;
    statusLabel: ProviderPortalCard['statusLabel'];
    statusTone: ProviderPortalCard['statusTone'];
}

export interface ProviderQuickAccessSection {
    id: string;
    title: string;
    description: string;
    links: ProviderQuickAccessLink[];
}

export const PROVIDER_PORTALS: ProviderPortalDefinition[] = [
    {
        id: 'openai',
        label: 'OpenAI',
        notes: 'Platform billing, usage, and API key controls for ChatGPT / Codex API workloads.',
        actions: [
            { label: 'API keys', href: 'https://platform.openai.com/api-keys', kind: 'keys' },
            { label: 'Usage', href: 'https://platform.openai.com/usage', kind: 'usage' },
            { label: 'Billing', href: 'https://platform.openai.com/settings/organization/billing/overview', kind: 'billing' },
            { label: 'Docs', href: 'https://platform.openai.com/docs/overview', kind: 'docs' },
        ],
    },
    {
        id: 'anthropic',
        label: 'Anthropic',
        notes: 'Claude API console, workspace usage, and plan management links.',
        actions: [
            { label: 'API keys', href: 'https://console.anthropic.com/settings/keys', kind: 'keys' },
            { label: 'Usage', href: 'https://console.anthropic.com/settings/usage', kind: 'usage' },
            { label: 'Plans', href: 'https://console.anthropic.com/settings/plans', kind: 'plan' },
            { label: 'Docs', href: 'https://docs.anthropic.com/', kind: 'docs' },
        ],
    },
    {
        id: 'gemini',
        label: 'Google Gemini / AI Studio',
        notes: 'Gemini API keys, Google AI Studio, and Google Cloud / Vertex entry points.',
        actions: [
            { label: 'AI Studio', href: 'https://aistudio.google.com/', kind: 'workspace' },
            { label: 'API keys', href: 'https://aistudio.google.com/app/apikey', kind: 'keys' },
            { label: 'Vertex AI', href: 'https://console.cloud.google.com/vertex-ai', kind: 'cloud' },
            { label: 'Docs', href: 'https://ai.google.dev/', kind: 'docs' },
        ],
    },
    {
        id: 'openrouter',
        label: 'OpenRouter',
        notes: 'Centralized multi-model routing, credits, and usage controls.',
        actions: [
            { label: 'API keys', href: 'https://openrouter.ai/keys', kind: 'keys' },
            { label: 'Usage', href: 'https://openrouter.ai/activity', kind: 'usage' },
            { label: 'Credits', href: 'https://openrouter.ai/settings/credits', kind: 'credits' },
            { label: 'Docs', href: 'https://openrouter.ai/docs/quickstart', kind: 'docs' },
        ],
    },
    {
        id: 'xai',
        label: 'xAI / Grok',
        notes: 'xAI developer console, Grok model usage, and API onboarding.',
        actions: [
            { label: 'Console', href: 'https://console.x.ai/', kind: 'console' },
            { label: 'Docs', href: 'https://docs.x.ai/', kind: 'docs' },
            { label: 'API keys', href: 'https://console.x.ai/', kind: 'keys' },
        ],
    },
    {
        id: 'deepseek',
        label: 'DeepSeek',
        notes: 'DeepSeek platform entry point for API credentials and account usage.',
        actions: [
            { label: 'Platform', href: 'https://platform.deepseek.com/', kind: 'console' },
            { label: 'API keys', href: 'https://platform.deepseek.com/api_keys', kind: 'keys' },
            { label: 'Docs', href: 'https://api-docs.deepseek.com/', kind: 'docs' },
        ],
    },
    {
        id: 'mistral',
        label: 'Mistral',
        notes: 'Mistral console and API documentation for hosted models.',
        actions: [
            { label: 'Console', href: 'https://console.mistral.ai/', kind: 'console' },
            { label: 'API keys', href: 'https://console.mistral.ai/', kind: 'keys' },
            { label: 'Docs', href: 'https://docs.mistral.ai/', kind: 'docs' },
        ],
    },
    {
        id: 'groq',
        label: 'Groq',
        notes: 'GroqCloud keys, usage, and low-latency model docs.',
        actions: [
            { label: 'Console', href: 'https://console.groq.com/', kind: 'console' },
            { label: 'API keys', href: 'https://console.groq.com/keys', kind: 'keys' },
            { label: 'Docs', href: 'https://console.groq.com/docs/overview', kind: 'docs' },
        ],
    },
    {
        id: 'azure-openai',
        label: 'Azure OpenAI',
        notes: 'Azure subscription, billing, and OpenAI deployment management.',
        actions: [
            { label: 'Azure portal', href: 'https://portal.azure.com/', kind: 'cloud' },
            { label: 'Cost analysis', href: 'https://portal.azure.com/#view/Microsoft_Azure_CostManagement/Menu/~/overview', kind: 'billing' },
            { label: 'Docs', href: 'https://learn.microsoft.com/azure/ai-services/openai/', kind: 'docs' },
        ],
    },
    {
        id: 'github-copilot',
        label: 'GitHub Copilot',
        notes: 'Copilot subscription and GitHub personal access / billing surfaces.',
        actions: [
            { label: 'Copilot settings', href: 'https://github.com/settings/copilot', kind: 'workspace' },
            { label: 'Billing', href: 'https://github.com/settings/billing', kind: 'billing' },
            { label: 'PATs', href: 'https://github.com/settings/tokens', kind: 'token' },
            { label: 'Docs', href: 'https://docs.github.com/copilot', kind: 'docs' },
        ],
    },
    {
        id: 'antigravity',
        label: 'Antigravity',
        notes: 'Google Cloud Code Assist / Antigravity subscription and quota surfaces.',
        actions: [
            { label: 'Code Assist', href: 'https://console.cloud.google.com/', kind: 'workspace' },
            { label: 'Google Cloud', href: 'https://console.cloud.google.com/', kind: 'cloud' },
            { label: 'Docs', href: 'https://cloud.google.com/code-assist/docs', kind: 'docs' },
        ],
    },
    {
        id: 'kiro',
        label: 'Kiro',
        notes: 'Kiro / AWS CodeWhisperer subscription and quota management links.',
        actions: [
            { label: 'Kiro', href: 'https://kiro.dev/', kind: 'plan' },
            { label: 'AWS Builder ID', href: 'https://view.awsapps.com/start', kind: 'cloud' },
            { label: 'Docs', href: 'https://docs.aws.amazon.com/codewhisperer/', kind: 'docs' },
        ],
    },
    {
        id: 'kimi-coding',
        label: 'Kimi Coding',
        notes: 'Kimi Coding membership and usage overview.',
        actions: [
            { label: 'Kimi', href: 'https://kimi.com/', kind: 'plan' },
            { label: 'Coding', href: 'https://kimi.com/', kind: 'workspace' },
            { label: 'Docs', href: 'https://platform.moonshot.ai/docs', kind: 'docs' },
        ],
    },
];

const PROVIDER_QUICK_ACCESS_SECTIONS: ProviderQuickAccessSectionDefinition[] = [
    {
        id: 'api-keys',
        title: 'API keys & tokens',
        description: 'Fast path to the credentials most likely to block first-run setup in TormentNexus.',
        links: [
            { providerId: 'openai', preferredKinds: ['keys'] },
            { providerId: 'anthropic', preferredKinds: ['keys'] },
            { providerId: 'gemini', preferredKinds: ['keys'] },
            { providerId: 'openrouter', preferredKinds: ['keys'] },
            { providerId: 'deepseek', preferredKinds: ['keys'] },
            { providerId: 'github-copilot', preferredKinds: ['token'] },
        ],
    },
    {
        id: 'plans-billing',
        title: 'Plans, billing & credits',
        description: 'Keep spend, credits, and membership surfaces one click away when quotas or subscriptions need attention.',
        links: [
            { providerId: 'openai', preferredKinds: ['billing'] },
            { providerId: 'anthropic', preferredKinds: ['plan'] },
            { providerId: 'openrouter', preferredKinds: ['credits', 'billing'] },
            { providerId: 'azure-openai', preferredKinds: ['billing'] },
            { providerId: 'github-copilot', preferredKinds: ['billing'] },
            { providerId: 'kiro', preferredKinds: ['plan'] },
        ],
    },
    {
        id: 'cloud-oauth',
        title: 'Cloud & OAuth consoles',
        description: 'Use these entry points for cloud-managed providers, OAuth-backed surfaces, and hosted console workflows.',
        links: [
            { providerId: 'gemini', preferredKinds: ['workspace', 'cloud'] },
            { providerId: 'azure-openai', preferredKinds: ['cloud'] },
            { providerId: 'github-copilot', preferredKinds: ['workspace'] },
            { providerId: 'antigravity', preferredKinds: ['workspace', 'cloud'] },
            { providerId: 'kiro', preferredKinds: ['cloud'] },
            { providerId: 'kimi-coding', preferredKinds: ['workspace', 'plan'] },
        ],
    },
];

export const ROUTING_STRATEGY_OPTIONS: Array<{ value: BillingRoutingStrategy; label: string }> = [
    { value: 'best', label: 'Best quality' },
    { value: 'cheapest', label: 'Lowest cost' },
    { value: 'round-robin', label: 'Round robin' },
];

export function getProviderPortalCards(quotas: BillingProviderQuotaSummary[] | undefined): ProviderPortalCard[] {
    const normalizedQuotas = (Array.isArray(quotas) ? quotas : [])
        .filter((quota): quota is BillingProviderQuotaSummary => Boolean(quota) && typeof quota === 'object' && typeof quota.provider === 'string');
    const quotaMap = new Map(normalizedQuotas.map((quota) => [quota.provider, quota]));

    return PROVIDER_PORTALS.map((portal) => {
        const quota = quotaMap.get(portal.id);
        const authenticated = !!quota?.authenticated;
        const configured = !!quota?.configured;
        const authMethod = quota?.authMethod && quota.authMethod !== 'none'
            ? quota.authMethod.replace(/_/g, ' ')
            : 'manual setup';
        const availability = quota?.availability?.replace(/_/g, ' ') ?? 'reference only';

        return {
            ...portal,
            statusLabel: authenticated ? 'Connected' : configured ? 'Configured' : 'Not connected',
            statusTone: authenticated ? 'success' : configured ? 'warning' : 'muted',
            authLabel: authenticated || configured ? authMethod : 'No auth detected',
            availabilityLabel: availability,
            errorLabel: quota?.lastError ?? null,
        };
    });
}

export function getProviderQuickAccessSections(quotas: BillingProviderQuotaSummary[] | undefined): ProviderQuickAccessSection[] {
    const portalCards = getProviderPortalCards(quotas);
    const portalCardMap = new Map(portalCards.map((card) => [card.id, card]));
    const portalDefinitionMap = new Map(PROVIDER_PORTALS.map((portal) => [portal.id, portal]));

    return PROVIDER_QUICK_ACCESS_SECTIONS.map((section) => {
        const links = section.links.flatMap((linkDefinition) => {
            const portal = portalDefinitionMap.get(linkDefinition.providerId);
            const portalCard = portalCardMap.get(linkDefinition.providerId);

            if (!portal || !portalCard) {
                return [];
            }

            const action = linkDefinition.preferredKinds
                .map((kind) => portal.actions.find((candidate) => candidate.kind === kind))
                .find((candidate): candidate is ProviderPortalAction => !!candidate);

            if (!action) {
                return [];
            }

            return [{
                providerId: portal.id,
                providerLabel: portal.label,
                actionLabel: action.label,
                href: action.href,
                statusLabel: portalCard.statusLabel,
                statusTone: portalCard.statusTone,
            }];
        });

        return {
            id: section.id,
            title: section.title,
            description: section.description,
            links,
        };
    }).filter((section) => section.links.length > 0);
}

export function getPortalBadgeClasses(tone: ProviderPortalCard['statusTone']): string {
    switch (tone) {
        case 'success':
            return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
        case 'warning':
            return 'bg-amber-500/10 text-amber-300 border-amber-500/20';
        default:
            return 'bg-zinc-800 text-zinc-400 border-zinc-700';
    }
}

export function formatTaskRoutingLabel(taskType: BillingTaskRoutingRuleSummary['taskType']): string {
    switch (taskType) {
        case 'coding':
            return 'Coding';
        case 'planning':
            return 'Planning';
        case 'research':
            return 'Research';
        case 'worker':
            return 'Worker tasks';
        case 'supervisor':
            return 'Supervisor tasks';
        default:
            return 'General';
    }
}

export function formatRoutingStrategyLabel(strategy: BillingRoutingStrategy): string {
    return ROUTING_STRATEGY_OPTIONS.find((option) => option.value === strategy)?.label ?? strategy;
}

export function getRoutingStrategyBadgeClasses(strategy: BillingTaskRoutingRuleSummary['strategy']): string {
    switch (strategy) {
        case 'best':
            return 'bg-fuchsia-500/10 text-fuchsia-300 border-fuchsia-500/20';
        case 'cheapest':
            return 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20';
        default:
            return 'bg-blue-500/10 text-blue-300 border-blue-500/20';
    }
}