import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';

import {
  buildDashboardAlerts,
  buildStartupChecklist,
  getGroupedStartupBlockingReasons,
  getStartupBlockingReasonGroupImpactedChecks,
  getStartupBlockingReasonGroupPriorityCounts,
  getStartupBlockingReasonGroupPrimaryReason,
  getStartupBlockingReasonGroupTopAction,
  getStartupBlockingReasonGroupSeverity,
  getPrioritizedStartupBlockingReasons,
  getStartupBlockingReasonAction,
  getStartupBlockingReasonActions,
  getStartupBlockingReasonImpactedChecks,
  getStartupBlockingReasonTitle,
  getStartupBlockingReasonPriorityCounts,
  getStartupBlockingReasonPriorityLabel,
  getStartupBlockingReasonPriorityTone,
  getStartupBlockingReasonPriority,
  getStartupBlockingReasonSubsystem,
  getStartupBlockingReasons,
  DashboardHomeView,
  buildOverviewMetrics,
  formatRelativeTimestamp,
  formatRestartCountdown,
  getQuotaUsagePercent,
  summarizeTrafficEvent,
  type DashboardProviderSummary,
  type DashboardSessionSummary,
  type DashboardStatusSummary,
  type DashboardStartupStatus,
} from './dashboard-home-view';

describe('dashboard home helpers', () => {
  it('builds overview metrics from live dashboard summaries', () => {
    const mcpStatus: DashboardStatusSummary = {
      initialized: true,
      serverCount: 3,
      toolCount: 18,
      connectedCount: 2,
    };

    const sessions: DashboardSessionSummary[] = [
      {
        id: 'session-1',
        name: 'Aider workspace',
        cliType: 'aider',
        workingDirectory: 'c:/repo',
        status: 'running',
        restartCount: 1,
        maxRestartAttempts: 5,
        lastActivityAt: 1_700_000_000_000,
        logs: [],
      },
      {
        id: 'session-2',
        name: 'Claude Code workspace',
        cliType: 'claude-code',
        workingDirectory: 'c:/repo-2',
        status: 'stopped',
        restartCount: 0,
        maxRestartAttempts: 5,
        lastActivityAt: 1_700_000_000_000,
        logs: [],
      },
    ];

    const providers: DashboardProviderSummary[] = [
      {
        provider: 'anthropic',
        name: 'Anthropic',
        configured: true,
        authenticated: true,
        authMethod: 'api_key',
        tier: 'pro',
        limit: 1000,
        used: 250,
        remaining: 750,
        availability: 'healthy',
      },
      {
        provider: 'google',
        name: 'Google Gemini',
        configured: true,
        authenticated: false,
        authMethod: 'api_key',
        tier: 'free',
        limit: 100,
        used: 100,
        remaining: 0,
        availability: 'degraded',
        lastError: 'quota exhausted',
      },
    ];

    expect(buildOverviewMetrics(mcpStatus, sessions, providers)).toEqual([
      {
        label: 'MCP servers',
        value: '2/3',
        detail: '18 tools indexed across the router',
      },
      {
        label: 'Supervised sessions',
        value: '1/2',
        detail: 'running right now',
      },
      {
        label: 'Configured providers',
        value: '2',
        detail: '1 need attention',
      },
    ]);
  });

  it('uses neutral overview metrics while the first live snapshot is still loading', () => {
    expect(buildOverviewMetrics(
      {
        initialized: false,
        serverCount: 0,
        toolCount: 0,
        connectedCount: 0,
      },
      [],
      [],
      true,
    )).toEqual([
      {
        label: 'MCP servers',
        value: '—',
        detail: 'Connecting to live router telemetry',
      },
      {
        label: 'Supervised sessions',
        value: '—',
        detail: 'Waiting for the first session supervisor snapshot',
      },
      {
        label: 'Configured providers',
        value: '—',
        detail: 'Waiting for the first provider routing snapshot',
      },
    ]);
  });

  it('uses first-run guidance when no providers are configured yet', () => {
    expect(buildOverviewMetrics(
      {
        initialized: true,
        serverCount: 0,
        toolCount: 0,
        connectedCount: 0,
      },
      [],
      [],
    )).toContainEqual({
      label: 'Configured providers',
      value: '0',
      detail: 'configure your first provider',
    });
  });

  it('formats traffic summaries and quota usage safely', () => {
    expect(summarizeTrafficEvent({
      server: 'github',
      method: 'tools/call',
      toolName: 'create_issue',
      paramsSummary: 'title=Bug',
      latencyMs: 42,
      success: true,
      timestamp: 1_700_000_000_000,
    })).toContain('tools/call · create_issue — title=Bug');

    expect(getQuotaUsagePercent({
      provider: 'anthropic',
      name: 'Anthropic',
      configured: true,
      authenticated: true,
      tier: 'pro',
      limit: 400,
      used: 100,
      remaining: 300,
    })).toBe(25);

    expect(getQuotaUsagePercent({
      provider: 'local',
      name: 'Local',
      configured: true,
      authenticated: true,
      tier: 'local',
      limit: null,
      used: 0,
      remaining: null,
    })).toBeNull();

    expect(formatRelativeTimestamp(1_700_000_000_000, null)).toBe('just now');
    expect(formatRelativeTimestamp(1_700_000_000_000, 1_700_000_060_000)).toBe('1m ago');
    expect(formatRestartCountdown(1_700_000_075_000, 1_700_000_060_000)).toBe('in 15s');
  });

  it('treats normalized provider availability states as degraded even without a last error message', () => {
    const mcpStatus: DashboardStatusSummary = {
      initialized: true,
      serverCount: 1,
      toolCount: 6,
      connectedCount: 1,
    };

    const providers: DashboardProviderSummary[] = [
      {
        provider: 'openai',
        name: 'OpenAI',
        configured: true,
        authenticated: true,
        authMethod: 'api_key',
        tier: 'pro',
        limit: 1000,
        used: 100,
        remaining: 900,
        availability: 'rate_limited',
      },
    ];

    expect(buildOverviewMetrics(mcpStatus, [], providers)).toContainEqual({
      label: 'Configured providers',
      value: '1',
      detail: '1 need attention',
    });

    const alerts = buildDashboardAlerts(
      mcpStatus,
      {
        status: 'running',
        ready: true,
        uptime: 42,
        checks: {
          mcpAggregator: {
            ready: true,
            liveReady: true,
            residentReady: true,
            serverCount: 1,
            connectedCount: 1,
            residentConnectedCount: 0,
            initialization: {
              inProgress: false,
              initialized: true,
              connectedClientCount: 1,
              configuredServerCount: 1,
            },
            persistedServerCount: 1,
            persistedToolCount: 6,
            advertisedServerCount: 1,
            advertisedToolCount: 6,
            advertisedAlwaysOnServerCount: 0,
            advertisedAlwaysOnToolCount: 0,
            inventoryReady: true,
          },
          configSync: {
            ready: true,
            status: {
              inProgress: false,
              lastServerCount: 1,
              lastToolCount: 6,
            },
          },
          memory: {
            ready: true,
            initialized: true,
            agentMemory: true,
          },
          browser: {
            ready: true,
            active: false,
            pageCount: 0,
          },
          sessionSupervisor: {
            ready: true,
            sessionCount: 0,
            restore: {
              restoredSessionCount: 0,
              autoResumeCount: 0,
            },
          },
          extensionBridge: {
            ready: true,
            clientCount: 0,
          },
          executionEnvironment: {
            ready: true,
            preferredShellId: 'pwsh',
            preferredShellLabel: 'PowerShell 7',
            shellCount: 2,
            verifiedShellCount: 2,
            toolCount: 4,
            verifiedToolCount: 4,
            harnessCount: 1,
            verifiedHarnessCount: 1,
            supportsPowerShell: true,
            supportsPosixShell: false,
          },
        },
      },
      [],
      providers,
      [],
    );

    expect(alerts).toContainEqual(expect.objectContaining({
      id: 'provider-degraded',
      severity: 'warning',
      title: 'Provider routing has degraded capacity',
    }));
  });

  it('builds startup checklist details from core boot state', () => {
    const startupStatus: DashboardStartupStatus = {
      status: 'running',
      ready: false,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: true,
          liveReady: true,
          residentReady: true,
          serverCount: 2,
          connectedCount: 1,
          residentConnectedCount: 1,
          initialization: {
            inProgress: false,
            initialized: true,
            connectedClientCount: 1,
            configuredServerCount: 2,
          },
          persistedServerCount: 2,
          persistedToolCount: 14,
          configuredServerCount: 2,
          advertisedServerCount: 2,
          advertisedToolCount: 14,
          advertisedAlwaysOnServerCount: 1,
          advertisedAlwaysOnToolCount: 3,
          inventoryReady: true,
        },
        configSync: {
          ready: true,
          status: {
            inProgress: false,
            lastCompletedAt: 1_700_000_000_000,
            lastServerCount: 2,
            lastToolCount: 14,
          },
        },
        memory: {
          ready: true,
          initialized: true,
          agentMemory: true,
        },
        browser: {
          ready: true,
          active: false,
          pageCount: 0,
        },
        sessionSupervisor: {
          ready: false,
          sessionCount: 1,
          restore: {
            restoredSessionCount: 1,
            autoResumeCount: 0,
          },
        },
        extensionBridge: {
          ready: true,
          acceptingConnections: true,
          clientCount: 1,
          hasConnectedClients: true,
        },
        executionEnvironment: {
          ready: true,
          preferredShellId: 'pwsh',
          preferredShellLabel: 'PowerShell 7',
          shellCount: 2,
          verifiedShellCount: 2,
          toolCount: 4,
          verifiedToolCount: 4,
          harnessCount: 1,
          verifiedHarnessCount: 1,
          supportsPowerShell: true,
          supportsPosixShell: false,
        },
      },
    };

    expect(buildStartupChecklist(startupStatus)).toEqual([
      {
        label: 'Cached inventory',
        ready: true,
        detail: '2 cached servers · 14 advertised tools from cached snapshot · 3 always-on advertised immediately',
      },
      {
        label: 'Resident MCP runtime',
        ready: true,
        detail: '1/1 resident server connection ready · on-demand tools can still cold-start as needed',
      },
      {
        label: 'Memory / context',
        ready: true,
        detail: 'Memory manager initialized and agent context services are available',
      },
      {
        label: 'Session restore',
        ready: false,
        detail: '1 restored · 0 auto-resumed',
      },
      {
        label: 'Client bridge',
        ready: true,
        detail: '1 connected bridge client · browser/editor bridge listener ready for new clients',
      },
      {
        label: 'Execution environment',
        ready: true,
        detail: 'PowerShell 7 preferred · 4/4 verified tools',
      },
    ]);
  });

  it('shows ready startup bridge status even before any clients attach', () => {
    const startupStatus: DashboardStartupStatus = {
      status: 'running',
      ready: true,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: true,
          liveReady: true,
          residentReady: true,
          serverCount: 0,
          connectedCount: 0,
          residentConnectedCount: 0,
          initialization: {
            inProgress: false,
            initialized: true,
            connectedClientCount: 0,
            configuredServerCount: 0,
          },
          persistedServerCount: 0,
          persistedToolCount: 0,
          configuredServerCount: 0,
          advertisedServerCount: 0,
          advertisedToolCount: 0,
          advertisedAlwaysOnServerCount: 0,
          advertisedAlwaysOnToolCount: 0,
          inventoryReady: true,
        },
        configSync: {
          ready: true,
          status: {
            inProgress: false,
            lastCompletedAt: 1_700_000_000_000,
            lastServerCount: 0,
            lastToolCount: 0,
          },
        },
        memory: {
          ready: true,
          initialized: true,
          agentMemory: true,
        },
        browser: {
          ready: true,
          active: false,
          pageCount: 0,
        },
        sessionSupervisor: {
          ready: true,
          sessionCount: 0,
          restore: {
            restoredSessionCount: 0,
            autoResumeCount: 0,
          },
        },
        extensionBridge: {
          ready: true,
          acceptingConnections: true,
          clientCount: 0,
          hasConnectedClients: false,
        },
        executionEnvironment: {
          ready: true,
          preferredShellId: 'pwsh',
          preferredShellLabel: 'PowerShell 7',
          shellCount: 1,
          verifiedShellCount: 1,
          toolCount: 3,
          verifiedToolCount: 3,
          harnessCount: 0,
          verifiedHarnessCount: 0,
          supportsPowerShell: true,
          supportsPosixShell: false,
        },
      },
    };

    expect(buildStartupChecklist(startupStatus)).toEqual([
      {
        label: 'Cached inventory',
        ready: true,
        detail: 'No configured servers yet · empty cached inventory is ready',
      },
      {
        label: 'Resident MCP runtime',
        ready: true,
        detail: 'No downstream servers configured · on-demand MCP launches are ready when needed',
      },
      {
        label: 'Memory / context',
        ready: true,
        detail: 'Memory manager initialized and agent context services are available',
      },
      {
        label: 'Session restore',
        ready: true,
        detail: '0 restored · 0 auto-resumed',
      },
      {
        label: 'Client bridge',
        ready: true,
        detail: '0 connected bridge clients · browser/editor bridge listener ready for new clients',
      },
      {
        label: 'Execution environment',
        ready: true,
        detail: 'PowerShell 7 preferred · 3/3 verified tools',
      },
    ]);
  });

  it('adds extension install artifact readiness when install-surface telemetry is provided', () => {
    const startupStatus: DashboardStartupStatus = {
      status: 'running',
      ready: true,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: true,
          liveReady: true,
          residentReady: true,
          serverCount: 0,
          connectedCount: 0,
          residentConnectedCount: 0,
          initialization: {
            inProgress: false,
            initialized: true,
            connectedClientCount: 0,
            configuredServerCount: 0,
          },
          persistedServerCount: 0,
          persistedToolCount: 0,
          configuredServerCount: 0,
          advertisedServerCount: 0,
          advertisedToolCount: 0,
          advertisedAlwaysOnServerCount: 0,
          advertisedAlwaysOnToolCount: 0,
          inventoryReady: true,
        },
        configSync: {
          ready: true,
          status: {
            inProgress: false,
            lastCompletedAt: 1_700_000_000_000,
            lastServerCount: 0,
            lastToolCount: 0,
          },
        },
        memory: {
          ready: true,
          initialized: true,
          agentMemory: true,
        },
        browser: {
          ready: true,
          active: false,
          pageCount: 0,
        },
        sessionSupervisor: {
          ready: true,
          sessionCount: 0,
          restore: {
            restoredSessionCount: 0,
            autoResumeCount: 0,
          },
        },
        extensionBridge: {
          ready: true,
          acceptingConnections: true,
          clientCount: 0,
          hasConnectedClients: false,
        },
        executionEnvironment: {
          ready: true,
          preferredShellId: 'pwsh',
          preferredShellLabel: 'PowerShell 7',
          shellCount: 1,
          verifiedShellCount: 1,
          toolCount: 3,
          verifiedToolCount: 3,
          harnessCount: 0,
          verifiedHarnessCount: 0,
          supportsPowerShell: true,
          supportsPosixShell: false,
        },
      },
    };

    expect(buildStartupChecklist(startupStatus, false, [
      { id: 'browser-extension-chromium', status: 'ready' },
      { id: 'browser-extension-firefox', status: 'ready' },
    ])).toContainEqual({
      label: 'Extension install artifacts',
      ready: true,
      detail: 'Chromium/Edge and Firefox extension bundles are ready to load.',
    });
  });

  it('gracefully renders startup checklist details from partial older payloads', () => {
    const startupStatus = {
      status: 'running',
      ready: false,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: true,
          liveReady: false,
          residentReady: false,
          serverCount: 1,
          connectedCount: 0,
          residentConnectedCount: 0,
          initialization: null,
          persistedServerCount: 1,
          persistedToolCount: 3,
          advertisedAlwaysOnServerCount: 0,
          inventoryReady: true,
        },
        sessionSupervisor: {
          ready: false,
          sessionCount: 0,
          restore: null,
        },
        extensionBridge: {
          ready: false,
          clientCount: 0,
        },
      },
    } as DashboardStartupStatus;

    expect(buildStartupChecklist(startupStatus)).toEqual([
      {
        label: 'Cached inventory',
        ready: true,
        detail: '1 cached servers · 3 advertised tools from cached snapshot',
      },
      {
        label: 'Resident MCP runtime',
        ready: false,
        detail: '1 on-demand server can launch when needed · no resident MCP runtime is required',
      },
      {
        label: 'Memory / context',
        ready: false,
        detail: 'Waiting for memory initialization',
      },
      {
        label: 'Session restore',
        ready: false,
        detail: 'Waiting for supervisor restore',
      },
      {
        label: 'Client bridge',
        ready: false,
        detail: 'Browser/editor bridge listener is offline',
      },
      {
        label: 'Execution environment',
        ready: false,
        detail: '0/0 verified shells · 0/0 verified tools',
      },
    ]);

    expect(() => renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: true, serverCount: 1, toolCount: 3, connectedCount: 0 }}
        startupStatus={startupStatus}
        servers={[]}
        traffic={[]}
        providers={[]}
        fallbackChain={[]}
        sessions={[]}
      />,
    )).not.toThrow();
  });

  it('renders a neutral connecting state before the first live startup snapshot arrives', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        isBootstrapping
        mcpStatus={{ initialized: false, serverCount: 0, toolCount: 0, connectedCount: 0 }}
        startupStatus={{
          status: 'starting',
          ready: false,
          uptime: 0,
          checks: {
            mcpAggregator: {
              ready: false,
              serverCount: 0,
              initialization: null,
              persistedServerCount: 0,
              persistedToolCount: 0,
              inventoryReady: false,
            },
            configSync: { ready: false, status: null },
            memory: { ready: false, initialized: false, agentMemory: false },
            browser: { ready: false, active: false, pageCount: 0 },
            sessionSupervisor: { ready: false, sessionCount: 0, restore: null },
            extensionBridge: { ready: false, clientCount: 0 },
            executionEnvironment: {
              ready: false,
              shellCount: 0,
              verifiedShellCount: 0,
              toolCount: 0,
              verifiedToolCount: 0,
              harnessCount: 0,
              verifiedHarnessCount: 0,
              supportsPowerShell: false,
              supportsPosixShell: false,
            },
          },
        }}
        servers={[]}
        traffic={[]}
        providers={[]}
        fallbackChain={[]}
        sessions={[]}
      />,
    );

    expect(html).toContain('Connecting to live core telemetry.');
    expect(html).toContain('Connecting to live startup telemetry from core.');
    expect(html).toContain('Connecting');
    expect(html).toContain('>—<');
    expect(html).not.toContain('All clear');
    expect(html).not.toContain('0 active');
  });

  it('surfaces cross-panel operator alerts in priority order', () => {
    const mcpStatus: DashboardStatusSummary = {
      initialized: true,
      serverCount: 2,
      toolCount: 14,
      connectedCount: 0,
    };

    const startupStatus: DashboardStartupStatus = {
      status: 'starting',
      ready: false,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: false,
          liveReady: false,
          serverCount: 2,
          connectedCount: 0,
          initialization: {
            inProgress: true,
            initialized: false,
            connectedClientCount: 0,
            configuredServerCount: 2,
          },
          persistedServerCount: 2,
          persistedToolCount: 14,
          advertisedServerCount: 2,
          advertisedToolCount: 14,
          advertisedAlwaysOnServerCount: 0,
          advertisedAlwaysOnToolCount: 0,
          inventoryReady: false,
        },
        configSync: {
          ready: false,
          status: {
            inProgress: true,
            lastServerCount: 2,
            lastToolCount: 14,
          },
        },
        memory: {
          ready: true,
          initialized: true,
          agentMemory: true,
        },
        browser: {
          ready: true,
          active: false,
          pageCount: 0,
        },
        sessionSupervisor: {
          ready: true,
          sessionCount: 1,
          restore: {
            restoredSessionCount: 1,
            autoResumeCount: 0,
          },
        },
        extensionBridge: {
          ready: true,
          clientCount: 1,
        },
        executionEnvironment: {
          ready: false,
          preferredShellId: null,
          preferredShellLabel: null,
          shellCount: 1,
          verifiedShellCount: 0,
          toolCount: 0,
          verifiedToolCount: 0,
          harnessCount: 0,
          verifiedHarnessCount: 0,
          supportsPowerShell: false,
          supportsPosixShell: false,
        },
      },
    };

    const alerts = buildDashboardAlerts(
      mcpStatus,
      startupStatus,
      [
        {
          name: 'github',
          status: 'error',
          toolCount: 8,
          config: { command: 'node', args: ['github.js'], env: ['GITHUB_TOKEN'] },
        },
      ],
      [
        {
          provider: 'anthropic',
          name: 'Anthropic',
          configured: true,
          authenticated: false,
          authMethod: 'api_key',
          tier: 'pro',
          limit: 1000,
          used: 950,
          remaining: 50,
          availability: 'degraded',
          lastError: 'quota exhausted',
        },
      ],
      [
        {
          id: 'session-1',
          name: 'Aider workspace',
          cliType: 'aider',
          workingDirectory: 'c:/repo',
          status: 'error',
          restartCount: 3,
          maxRestartAttempts: 5,
          lastActivityAt: 1_700_000_000_000,
          lastError: 'session crashed',
          logs: [],
        },
      ],
    );

    expect(alerts.map((alert) => alert.title)).toEqual([
      'Supervised sessions have failed',
      'Some MCP servers need attention',
      'Startup sequence is still warming up',
      'Provider routing has degraded capacity',
    ]);
  });

  it('uses a dedicated startup alert when compat fallback is active', () => {
    const alerts = buildDashboardAlerts(
      {
        initialized: true,
        serverCount: 64,
        toolCount: 0,
        connectedCount: 0,
      },
      {
        status: 'degraded',
        ready: false,
        uptime: 42,
        summary: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        checks: {
          mcpAggregator: {
            ready: true,
            liveReady: true,
            residentReady: false,
            serverCount: 64,
            connectedCount: 0,
            residentConnectedCount: 0,
            initialization: null,
            persistedServerCount: 64,
            persistedToolCount: 0,
            configuredServerCount: 64,
            advertisedServerCount: 64,
            advertisedToolCount: 0,
            advertisedAlwaysOnServerCount: 0,
            advertisedAlwaysOnToolCount: 0,
            inventoryReady: false,
          },
          configSync: {
            ready: true,
            status: {
              inProgress: false,
              lastServerCount: 64,
              lastToolCount: 0,
            },
          },
          memory: {
            ready: false,
            initialized: false,
            agentMemory: false,
          },
          browser: {
            ready: false,
            active: false,
            pageCount: 0,
          },
          sessionSupervisor: {
            ready: false,
            sessionCount: 0,
            restore: null,
          },
          extensionBridge: {
            ready: false,
            clientCount: 0,
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
          },
        },
      },
      [],
      [],
      [],
    );

    expect(alerts).toContainEqual(expect.objectContaining({
      id: 'startup-compat-fallback',
      title: 'Startup is using local compat fallback',
      detail: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
    }));
    expect(alerts.find((alert) => alert.id === 'startup-pending')).toBeUndefined();
  });

  it('includes resident MCP warmup posture in startup checklist details', () => {
    const startupStatus = {
      status: 'running',
      ready: false,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: true,
          liveReady: true,
          residentReady: false,
          serverCount: 3,
          connectedCount: 1,
          residentConnectedCount: 0,
          warmingServerCount: 2,
          failedWarmupServerCount: 1,
          initialization: null,
          persistedServerCount: 3,
          persistedToolCount: 9,
          advertisedAlwaysOnServerCount: 1,
          inventoryReady: true,
        },
        configSync: {
          ready: true,
          status: null,
        },
        memory: {
          ready: true,
          initialized: true,
          agentMemory: true,
        },
        browser: {
          ready: true,
          active: false,
          pageCount: 0,
        },
        sessionSupervisor: {
          ready: true,
          sessionCount: 0,
          restore: {
            restoredSessionCount: 0,
            autoResumeCount: 0,
          },
        },
        extensionBridge: {
          ready: true,
          acceptingConnections: true,
          clientCount: 0,
          hasConnectedClients: false,
        },
        executionEnvironment: {
          ready: true,
          preferredShellId: 'pwsh',
          preferredShellLabel: 'PowerShell 7',
          shellCount: 1,
          verifiedShellCount: 1,
          toolCount: 2,
          verifiedToolCount: 2,
          harnessCount: 0,
          verifiedHarnessCount: 0,
          supportsPowerShell: true,
          supportsPosixShell: false,
          notes: [],
        },
      },
    } as DashboardStartupStatus;

    expect(buildStartupChecklist(startupStatus)).toContainEqual({
      label: 'Resident MCP runtime',
      ready: false,
      detail: 'Cached inventory is already advertised · resident always-on servers are still warming · on-demand tools remain launchable · 2 warming · 1 failed',
    });
  });

  it('surfaces tormentnexus seeding state in memory/context startup details', () => {
    const startupStatus = {
      status: 'running',
      ready: false,
      uptime: 42,
      checks: {
        mcpAggregator: {
          ready: true,
          liveReady: true,
          residentReady: true,
          serverCount: 0,
          connectedCount: 0,
          residentConnectedCount: 0,
          initialization: null,
          persistedServerCount: 0,
          persistedToolCount: 0,
          advertisedAlwaysOnServerCount: 0,
          inventoryReady: true,
        },
        configSync: {
          ready: true,
          status: null,
        },
        memory: {
          ready: false,
          initialized: true,
          agentMemory: true,
          claudeMem: {
            ready: false,
            enabled: true,
            storeExists: true,
            defaultSectionCount: 7,
            presentDefaultSectionCount: 3,
            missingSections: ['project_overview'],
          },
        },
        browser: {
          ready: true,
          active: false,
          pageCount: 0,
        },
        sessionSupervisor: {
          ready: true,
          sessionCount: 0,
          restore: {
            restoredSessionCount: 0,
            autoResumeCount: 0,
          },
        },
        extensionBridge: {
          ready: true,
          acceptingConnections: true,
          clientCount: 0,
          hasConnectedClients: false,
        },
        executionEnvironment: {
          ready: true,
          preferredShellId: 'pwsh',
          preferredShellLabel: 'PowerShell 7',
          shellCount: 1,
          verifiedShellCount: 1,
          toolCount: 2,
          verifiedToolCount: 2,
          harnessCount: 0,
          verifiedHarnessCount: 0,
          supportsPowerShell: true,
          supportsPosixShell: false,
          notes: [],
        },
      },
    } as DashboardStartupStatus;

    expect(buildStartupChecklist(startupStatus)).toContainEqual({
      label: 'Memory / context',
      ready: false,
      detail: 'Memory manager is initialized, but tormentnexus is still seeding default sections (3/7 present)',
    });
  });

  it('returns sanitized startup blocking reasons from startup payloads', () => {
    expect(getStartupBlockingReasons({
      status: 'running',
      ready: false,
      uptime: 12,
      blockingReasons: [
        { code: 'memory_not_ready', detail: 'Memory manager initialization is still pending.' },
        { code: 'extension_bridge_not_ready', detail: 'Extension bridge listener is offline.' },
      ],
      checks: {
        mcpAggregator: {
          ready: false,
          serverCount: 0,
          initialization: null,
          persistedServerCount: 0,
          persistedToolCount: 0,
          inventoryReady: false,
        },
        configSync: { ready: false, status: null },
        memory: { ready: false, initialized: false, agentMemory: false },
        browser: { ready: false, active: false, pageCount: 0 },
        sessionSupervisor: { ready: false, sessionCount: 0, restore: null },
        extensionBridge: { ready: false, clientCount: 0 },
        executionEnvironment: {
          ready: false,
          shellCount: 0,
          verifiedShellCount: 0,
          toolCount: 0,
          verifiedToolCount: 0,
          harnessCount: 0,
          verifiedHarnessCount: 0,
          supportsPowerShell: false,
          supportsPosixShell: false,
        },
      },
    })).toEqual([
      { code: 'memory_not_ready', detail: 'Memory manager initialization is still pending.' },
      { code: 'extension_bridge_not_ready', detail: 'Extension bridge listener is offline.' },
    ]);
  });

  it('maps startup blocking reason codes to actionable dashboard links', () => {
    expect(getStartupBlockingReasonAction('memory_not_ready')).toEqual({
      href: '/dashboard/memory',
      label: 'Open memory dashboard',
    });

    expect(getStartupBlockingReasonAction('mcp_config_sync_pending')).toEqual({
      href: '/dashboard/mcp/system',
      label: 'Open MCP system',
    });

    expect(getStartupBlockingReasonAction('unknown_reason')).toEqual({
      href: '/dashboard',
      label: 'Open startup overview',
    });
  });

  it('deduplicates startup blocking actions while preserving first-seen order', () => {
    expect(getStartupBlockingReasonActions([
      { code: 'memory_not_ready', detail: 'Memory pending' },
      { code: 'claude_mem_not_ready', detail: 'TormentNexus pending' },
      { code: 'extension_bridge_not_ready', detail: 'Bridge offline' },
      { code: 'browser_service_not_ready', detail: 'Browser service offline' },
    ])).toEqual([
      { href: '/dashboard/memory', label: 'Open memory dashboard' },
      { href: '/dashboard/integrations', label: 'Open Integration Hub' },
    ]);
  });

  it('scores startup blocker priorities for operator triage', () => {
    expect(getStartupBlockingReasonPriority('mcp_aggregator_not_initialized')).toBe(100);
    expect(getStartupBlockingReasonPriority('extension_bridge_not_ready')).toBe(80);
    expect(getStartupBlockingReasonPriority('memory_not_ready')).toBe(60);
    expect(getStartupBlockingReasonPriority('browser_service_not_ready')).toBe(40);
    expect(getStartupBlockingReasonPriority('unknown_reason')).toBe(20);
  });

  it('orders startup blockers by descending priority while preserving stable order within a tier', () => {
    expect(getPrioritizedStartupBlockingReasons([
      { code: 'memory_not_ready', detail: 'Memory pending' },
      { code: 'extension_bridge_not_ready', detail: 'Bridge offline' },
      { code: 'mcp_aggregator_not_initialized', detail: 'Aggregator offline' },
      { code: 'claude_mem_not_ready', detail: 'Claude mem pending' },
    ])).toEqual([
      { code: 'mcp_aggregator_not_initialized', detail: 'Aggregator offline', priority: 100 },
      { code: 'extension_bridge_not_ready', detail: 'Bridge offline', priority: 80 },
      { code: 'memory_not_ready', detail: 'Memory pending', priority: 60 },
      { code: 'claude_mem_not_ready', detail: 'Claude mem pending', priority: 60 },
    ]);
  });

  it('maps startup blocker priority values to readable labels', () => {
    expect(getStartupBlockingReasonPriorityLabel(100)).toBe('High');
    expect(getStartupBlockingReasonPriorityLabel(80)).toBe('High');
    expect(getStartupBlockingReasonPriorityLabel(60)).toBe('Medium');
    expect(getStartupBlockingReasonPriorityLabel(40)).toBe('Low');
  });

  it('maps startup blocker priority labels to tone classes', () => {
    expect(getStartupBlockingReasonPriorityTone('High')).toContain('border-rose-500/40');
    expect(getStartupBlockingReasonPriorityTone('Medium')).toContain('border-amber-500/40');
    expect(getStartupBlockingReasonPriorityTone('Low')).toContain('border-emerald-500/40');
  });

  it('counts startup blockers by priority band', () => {
    expect(getStartupBlockingReasonPriorityCounts([
      { code: 'mcp_aggregator_not_initialized', detail: 'A', priority: 100 },
      { code: 'extension_bridge_not_ready', detail: 'B', priority: 80 },
      { code: 'memory_not_ready', detail: 'C', priority: 60 },
      { code: 'browser_service_not_ready', detail: 'D', priority: 40 },
    ])).toEqual({
      high: 2,
      medium: 1,
      low: 1,
    });
  });

  it('maps startup blocker reason codes to subsystem groups', () => {
    expect(getStartupBlockingReasonSubsystem('mcp_inventory_not_ready')).toEqual({
      key: 'mcp',
      label: 'MCP router',
    });

    expect(getStartupBlockingReasonSubsystem('memory_not_ready')).toEqual({
      key: 'memory',
      label: 'Memory / context',
    });

    expect(getStartupBlockingReasonSubsystem('unknown_reason')).toEqual({
      key: 'startup',
      label: 'Startup platform',
    });
  });

  it('maps startup blocker codes to readable titles', () => {
    expect(getStartupBlockingReasonTitle('mcp_aggregator_not_initialized')).toBe('MCP router is not initialized');
    expect(getStartupBlockingReasonTitle('memory_not_ready')).toBe('Memory manager is still initializing');
    expect(getStartupBlockingReasonTitle('unknown_reason')).toBe('Startup blocker requires operator attention');
  });

  it('groups prioritized startup blockers by subsystem while preserving reason order', () => {
    expect(getGroupedStartupBlockingReasons([
      { code: 'mcp_aggregator_not_initialized', detail: 'Aggregator offline', priority: 100 },
      { code: 'extension_bridge_not_ready', detail: 'Bridge offline', priority: 80 },
      { code: 'mcp_inventory_not_ready', detail: 'Inventory stale', priority: 80 },
      { code: 'memory_not_ready', detail: 'Memory pending', priority: 60 },
    ])).toEqual([
      {
        key: 'mcp',
        label: 'MCP router',
        reasons: [
          { code: 'mcp_aggregator_not_initialized', detail: 'Aggregator offline', priority: 100 },
          { code: 'mcp_inventory_not_ready', detail: 'Inventory stale', priority: 80 },
        ],
      },
      {
        key: 'memory',
        label: 'Memory / context',
        reasons: [
          { code: 'memory_not_ready', detail: 'Memory pending', priority: 60 },
        ],
      },
      {
        key: 'integrations',
        label: 'Integrations',
        reasons: [
          { code: 'extension_bridge_not_ready', detail: 'Bridge offline', priority: 80 },
        ],
      },
    ]);
  });

  it('derives subsystem group severity from the highest reason priority in that group', () => {
    expect(getStartupBlockingReasonGroupSeverity([
      { code: 'browser_service_not_ready', detail: 'browser', priority: 40 },
      { code: 'memory_not_ready', detail: 'memory', priority: 60 },
    ])).toBe('Medium');

    expect(getStartupBlockingReasonGroupSeverity([
      { code: 'extension_bridge_not_ready', detail: 'bridge', priority: 80 },
      { code: 'mcp_aggregator_not_initialized', detail: 'aggregator', priority: 100 },
    ])).toBe('High');
  });

  it('derives a top action per blocker group from the highest-priority reason', () => {
    expect(getStartupBlockingReasonGroupTopAction([
      { code: 'memory_not_ready', detail: 'memory', priority: 60 },
      { code: 'extension_bridge_not_ready', detail: 'bridge', priority: 80 },
    ])).toEqual({
      href: '/dashboard/integrations',
      label: 'Open Integration Hub',
    });

    expect(getStartupBlockingReasonGroupTopAction([])).toBeNull();
  });

  it('derives a primary blocker per group from the highest-priority reason', () => {
    expect(getStartupBlockingReasonGroupPrimaryReason([
      { code: 'memory_not_ready', detail: 'memory', priority: 60 },
      { code: 'extension_bridge_not_ready', detail: 'bridge', priority: 80 },
    ])).toEqual({
      code: 'extension_bridge_not_ready',
      detail: 'bridge',
      priority: 80,
    });

    expect(getStartupBlockingReasonGroupPrimaryReason([])).toBeNull();
  });

  it('derives per-group priority mix counts', () => {
    expect(getStartupBlockingReasonGroupPriorityCounts([
      { code: 'mcp_aggregator_not_initialized', detail: 'agg', priority: 100 },
      { code: 'extension_bridge_not_ready', detail: 'bridge', priority: 80 },
      { code: 'memory_not_ready', detail: 'memory', priority: 60 },
      { code: 'browser_service_not_ready', detail: 'browser', priority: 40 },
    ])).toEqual({
      high: 2,
      medium: 1,
      low: 1,
    });
  });

  it('maps startup blocker codes to impacted startup checks and dedupes per group', () => {
    expect(getStartupBlockingReasonImpactedChecks('mcp_inventory_not_ready')).toEqual([
      { key: 'cached-inventory', label: 'Cached inventory' },
      { key: 'resident-runtime', label: 'Resident MCP runtime' },
    ]);

    expect(getStartupBlockingReasonGroupImpactedChecks([
      { code: 'mcp_inventory_not_ready', detail: 'inventory', priority: 80 },
      { code: 'mcp_resident_runtime_not_ready', detail: 'runtime', priority: 100 },
      { code: 'mcp_config_sync_pending', detail: 'config', priority: 80 },
    ])).toEqual([
      { key: 'cached-inventory', label: 'Cached inventory' },
      { key: 'resident-runtime', label: 'Resident MCP runtime' },
    ]);
  });
});

describe('DashboardHomeView', () => {
  it('renders all four v1 panels with live-style content', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: true, serverCount: 2, toolCount: 14, connectedCount: 1 }}
        startupStatus={{
          status: 'running',
          ready: true,
          uptime: 120,
          checks: {
            mcpAggregator: {
              ready: true,
              liveReady: true,
              residentReady: true,
              serverCount: 2,
              connectedCount: 1,
              residentConnectedCount: 1,
              initialization: {
                inProgress: false,
                initialized: true,
                connectedClientCount: 1,
                configuredServerCount: 2,
              },
              persistedServerCount: 2,
              persistedToolCount: 14,
              advertisedServerCount: 2,
              advertisedToolCount: 14,
              advertisedAlwaysOnServerCount: 0,
              advertisedAlwaysOnToolCount: 2,
              inventoryReady: true,
            },
            configSync: {
              ready: true,
              status: {
                inProgress: false,
                lastCompletedAt: 1_700_000_000_000,
                lastServerCount: 2,
                lastToolCount: 14,
              },
            },
            memory: {
              ready: true,
              initialized: true,
              agentMemory: true,
            },
            browser: {
              ready: true,
              active: false,
              pageCount: 0,
            },
            sessionSupervisor: {
              ready: true,
              sessionCount: 1,
              restore: {
                restoredSessionCount: 1,
                autoResumeCount: 0,
              },
            },
            extensionBridge: {
              ready: true,
              clientCount: 1,
            },
            executionEnvironment: {
              ready: true,
              preferredShellId: 'pwsh',
              preferredShellLabel: 'PowerShell 7',
              shellCount: 2,
              verifiedShellCount: 2,
              toolCount: 5,
              verifiedToolCount: 5,
              harnessCount: 1,
              verifiedHarnessCount: 1,
              supportsPowerShell: true,
              supportsPosixShell: true,
            },
          },
        }}
        servers={[
          {
            name: 'github',
            status: 'connected',
            toolCount: 8,
            config: { command: 'node', args: ['github.js'], env: ['GITHUB_TOKEN'] },
          },
        ]}
        traffic={[
          {
            server: 'github',
            method: 'tools/call',
            toolName: 'create_issue',
            paramsSummary: 'title=Bug',
            latencyMs: 15,
            success: true,
            timestamp: 1_700_000_000_000,
          },
        ]}
        providers={[
          {
            provider: 'anthropic',
            name: 'Anthropic',
            configured: true,
            authenticated: true,
            authMethod: 'api_key',
            tier: 'pro',
            limit: 1000,
            used: 400,
            remaining: 600,
            availability: 'healthy',
          },
        ]}
        fallbackChain={[
          {
            priority: 1,
            provider: 'anthropic',
            model: 'claude-3-7-sonnet',
            reason: 'configured',
          },
        ]}
        sessions={[
          {
            id: 'session-1',
            name: 'Aider workspace',
            cliType: 'aider',
            workingDirectory: 'c:/repo',
            autoRestart: false,
            status: 'restarting',
            restartCount: 1,
            maxRestartAttempts: 5,
            scheduledRestartAt: 1_700_000_075_000,
            lastActivityAt: 1_700_000_000_000,
            logs: [
              { timestamp: 1_700_000_000_000, stream: 'stdout', message: 'Ready for instructions' },
            ],
          },
        ]}
      />,
    );

    expect(html).toContain('Overview');
    expect(html).toContain('MCP Router');
    expect(html).toContain('Sessions');
    expect(html).toContain('Providers');
    expect(html).toContain('Integration Hub');
    expect(html).toContain('Server health and traffic');
    expect(html).toContain('Startup readiness');
    expect(html).toContain('Install &amp; connect TormentNexus');
    expect(html).toContain('Browser extensions');
    expect(html).toContain('Editor surfaces');
    expect(html).toContain('Client config sync');
    expect(html).toContain('Cached inventory');
    expect(html).toContain('Resident MCP runtime');
    expect(html).toContain('Memory / context');
    expect(html).toContain('Supervised CLI runtime');
    expect(html).toContain('Quota and fallback posture');
    expect(html).toContain('Detailed MCP view');
    expect(html).toContain('Detailed provider view');
    expect(html).toContain('Open inspector');
    expect(html).toContain('create_issue');
    expect(html).toContain('Ready for instructions');
    expect(html).toContain('Manual restart only');
    expect(html).toContain('Restart queued in 15s');
    expect(html).toContain('Operator alerts');
    expect(html).toContain('All clear');
    expect(html).toContain('All major systems look healthy');
  });

  it('renders actionable first-run provider guidance when no providers are configured', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: false, serverCount: 0, toolCount: 0, connectedCount: 0 }}
        startupStatus={{
          status: 'starting',
          ready: false,
          uptime: 1,
          checks: {
            mcpAggregator: {
              ready: false,
              liveReady: false,
              residentReady: false,
              serverCount: 0,
              connectedCount: 0,
              residentConnectedCount: 0,
              initialization: null,
              persistedServerCount: 0,
              persistedToolCount: 0,
              advertisedServerCount: 0,
              advertisedToolCount: 0,
              advertisedAlwaysOnServerCount: 0,
              advertisedAlwaysOnToolCount: 0,
              inventoryReady: false,
            },
            configSync: {
              ready: false,
              status: null,
            },
            memory: {
              ready: false,
              initialized: false,
              agentMemory: false,
            },
            browser: {
              ready: false,
              active: false,
              pageCount: 0,
            },
            sessionSupervisor: {
              ready: false,
              sessionCount: 0,
              restore: null,
            },
            extensionBridge: {
              ready: false,
              clientCount: 0,
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
            },
          },
        }}
        servers={[]}
        traffic={[]}
        providers={[]}
        fallbackChain={[]}
        sessions={[]}
      />,
    );

    expect(html).toContain('configure your first provider');
    expect(html).toContain('Configure an API key or OAuth-backed provider in Billing to unlock fallback routing.');
    expect(html).toContain('No fallback chain is exposed yet. Configure providers to populate the routing order.');
  });

  it('renders active alerts when the router, providers, or sessions degrade', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: false, serverCount: 2, toolCount: 14, connectedCount: 0 }}
        startupStatus={{
          status: 'starting',
          ready: false,
          uptime: 120,
          checks: {
            mcpAggregator: {
              ready: false,
              liveReady: false,
              residentReady: false,
              serverCount: 2,
              connectedCount: 0,
              residentConnectedCount: 0,
              initialization: {
                inProgress: true,
                initialized: false,
                connectedClientCount: 0,
                configuredServerCount: 2,
              },
              persistedServerCount: 2,
              persistedToolCount: 14,
              advertisedServerCount: 2,
              advertisedToolCount: 14,
              advertisedAlwaysOnServerCount: 0,
              advertisedAlwaysOnToolCount: 0,
              inventoryReady: false,
            },
            configSync: {
              ready: false,
              status: {
                inProgress: true,
                lastServerCount: 2,
                lastToolCount: 14,
              },
            },
            memory: {
              ready: true,
              initialized: true,
              agentMemory: true,
            },
            browser: {
              ready: true,
              active: false,
              pageCount: 0,
            },
            sessionSupervisor: {
              ready: false,
              sessionCount: 1,
              restore: {
                restoredSessionCount: 1,
                autoResumeCount: 0,
              },
            },
            extensionBridge: {
              ready: true,
              clientCount: 1,
            },
            executionEnvironment: {
              ready: false,
              preferredShellId: null,
              preferredShellLabel: null,
              shellCount: 1,
              verifiedShellCount: 0,
              toolCount: 0,
              verifiedToolCount: 0,
              harnessCount: 0,
              verifiedHarnessCount: 0,
              supportsPowerShell: false,
              supportsPosixShell: false,
            },
          },
        }}
        servers={[
          {
            name: 'github',
            status: 'error',
            toolCount: 8,
            config: { command: 'node', args: ['github.js'], env: ['GITHUB_TOKEN'] },
          },
        ]}
        traffic={[]}
        providers={[
          {
            provider: 'anthropic',
            name: 'Anthropic',
            configured: true,
            authenticated: false,
            authMethod: 'api_key',
            tier: 'pro',
            limit: 1000,
            used: 400,
            remaining: 600,
            availability: 'degraded',
            lastError: 'quota exhausted',
          },
        ]}
        fallbackChain={[]}
        sessions={[
          {
            id: 'session-1',
            name: 'Aider workspace',
            cliType: 'aider',
            workingDirectory: 'c:/repo',
            status: 'error',
            restartCount: 1,
            maxRestartAttempts: 5,
            lastActivityAt: 1_700_000_000_000,
            lastError: 'session crashed',
            logs: [],
          },
        ]}
      />,
    );

    expect(html).toContain('Operator alerts');
    expect(html).toContain('4 active');
    expect(html).toContain('MCP router is not initialized');
    expect(html).toContain('Provider routing has degraded capacity');
    expect(html).toContain('Supervised sessions have failed');
    expect(html).toContain('Inspect MCP router');
  });

  it('renders compat fallback startup copy when live startup telemetry is unavailable', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: true, serverCount: 64, toolCount: 0, connectedCount: 0 }}
        startupStatus={{
          status: 'degraded',
          ready: false,
          uptime: 120,
          summary: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
          checks: {
            mcpAggregator: {
              ready: true,
              liveReady: true,
              residentReady: false,
              serverCount: 64,
              connectedCount: 0,
              residentConnectedCount: 0,
              initialization: null,
              persistedServerCount: 64,
              persistedToolCount: 0,
              configuredServerCount: 64,
              advertisedServerCount: 64,
              advertisedToolCount: 0,
              advertisedAlwaysOnServerCount: 0,
              advertisedAlwaysOnToolCount: 0,
              inventoryReady: false,
            },
            configSync: {
              ready: true,
              status: {
                inProgress: false,
                lastServerCount: 64,
                lastToolCount: 0,
              },
            },
            memory: {
              ready: false,
              initialized: false,
              agentMemory: false,
            },
            browser: {
              ready: false,
              active: false,
              pageCount: 0,
            },
            sessionSupervisor: {
              ready: false,
              sessionCount: 0,
              restore: null,
            },
            extensionBridge: {
              ready: false,
              clientCount: 0,
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
            },
          },
        }}
        servers={[]}
        traffic={[]}
        providers={[]}
        fallbackChain={[]}
        sessions={[]}
      />,
    );

    expect(html).toContain('Compat fallback');
    expect(html).toContain('Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.');
    expect(html).toContain('Startup is using local compat fallback');
  });

  it('renders startup blocking reasons beneath readiness checks when startup is pending', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: true, serverCount: 1, toolCount: 4, connectedCount: 0 }}
        startupStatus={{
          status: 'running',
          ready: false,
          uptime: 120,
          summary: 'Startup pending: Memory manager initialization is still pending.',
          blockingReasons: [
            { code: 'memory_not_ready', detail: 'Memory manager initialization is still pending.' },
            { code: 'extension_bridge_not_ready', detail: 'Extension bridge listener is offline.' },
          ],
          checks: {
            mcpAggregator: {
              ready: true,
              liveReady: true,
              residentReady: true,
              serverCount: 1,
              connectedCount: 0,
              residentConnectedCount: 0,
              initialization: null,
              persistedServerCount: 1,
              persistedToolCount: 4,
              advertisedServerCount: 1,
              advertisedToolCount: 4,
              advertisedAlwaysOnServerCount: 0,
              advertisedAlwaysOnToolCount: 0,
              inventoryReady: true,
            },
            configSync: {
              ready: true,
              status: {
                inProgress: false,
                lastServerCount: 1,
                lastToolCount: 4,
              },
            },
            memory: {
              ready: false,
              initialized: false,
              agentMemory: false,
            },
            browser: {
              ready: true,
              active: false,
              pageCount: 0,
            },
            sessionSupervisor: {
              ready: true,
              sessionCount: 0,
              restore: {
                restoredSessionCount: 0,
                autoResumeCount: 0,
              },
            },
            extensionBridge: {
              ready: false,
              clientCount: 0,
            },
            executionEnvironment: {
              ready: true,
              preferredShellId: 'pwsh',
              preferredShellLabel: 'PowerShell 7',
              shellCount: 1,
              verifiedShellCount: 1,
              toolCount: 2,
              verifiedToolCount: 2,
              harnessCount: 0,
              verifiedHarnessCount: 0,
              supportsPowerShell: true,
              supportsPosixShell: false,
            },
          },
        }}
        servers={[]}
        traffic={[]}
        providers={[]}
        fallbackChain={[]}
        sessions={[]}
      />,
    );

    expect(html).toContain('Blocking reasons');
    expect(html).toContain('Memory manager is still initializing');
    expect(html).toContain('Extension bridge listener is offline');
    expect(html).toContain('memory_not_ready');
    expect(html).toContain('extension_bridge_not_ready');
    expect(html).toContain('2 pending');
    expect(html).toContain('Open memory dashboard');
    expect(html).toContain('Open Integration Hub');
    expect(html).toContain('Suggested actions:');
    expect(html).toContain('Priority mix: 1 high · 1 medium · 0 low');
    expect(html).toContain('Memory / context');
    expect(html).toContain('Integrations');
    expect(html).toContain('Top action: Open memory dashboard');
    expect(html).toContain('Top action: Open Integration Hub');
    expect(html).toContain('Impacts: Memory / context');
    expect(html).toContain('Impacts: Client bridge');
    expect(html).toContain('Primary blocker: Memory manager is still initializing');
    expect(html).toContain('Primary blocker: Extension bridge listener is offline');
    expect(html).toContain('Group mix: 0 high · 1 medium · 0 low');
    expect(html).toContain('Group mix: 1 high · 0 medium · 0 low');
    expect(html).toContain('High group');
    expect(html).toContain('Medium group');
    expect(html).toContain('1 item');
    expect(html).toContain('High priority');
    expect(html).toContain('Medium priority');
    expect(html).toContain('border-rose-500/40');
    expect(html).toContain('border-amber-500/40');
    expect(html).toContain('href="/dashboard/memory"');
    expect(html).toContain('href="/dashboard/integrations"');
  });

  it('renders provider availability badges as degraded for normalized throttling states', () => {
    const html = renderToStaticMarkup(
      <DashboardHomeView
        generatedAtLabel="12:00:00 PM"
        currentTimestamp={1_700_000_060_000}
        mcpStatus={{ initialized: true, serverCount: 1, toolCount: 4, connectedCount: 1 }}
        startupStatus={{
          status: 'running',
          ready: true,
          uptime: 120,
          checks: {
            mcpAggregator: {
              ready: true,
              liveReady: true,
              residentReady: true,
              serverCount: 1,
              connectedCount: 1,
              residentConnectedCount: 0,
              initialization: {
                inProgress: false,
                initialized: true,
                connectedClientCount: 1,
                configuredServerCount: 1,
              },
              persistedServerCount: 1,
              persistedToolCount: 4,
              advertisedServerCount: 1,
              advertisedToolCount: 4,
              advertisedAlwaysOnServerCount: 0,
              advertisedAlwaysOnToolCount: 0,
              inventoryReady: true,
            },
            configSync: {
              ready: true,
              status: {
                inProgress: false,
                lastServerCount: 1,
                lastToolCount: 4,
              },
            },
            memory: {
              ready: true,
              initialized: true,
              agentMemory: true,
            },
            browser: {
              ready: true,
              active: false,
              pageCount: 0,
            },
            sessionSupervisor: {
              ready: true,
              sessionCount: 0,
              restore: {
                restoredSessionCount: 0,
                autoResumeCount: 0,
              },
            },
            extensionBridge: {
              ready: true,
              clientCount: 1,
            },
            executionEnvironment: {
              ready: true,
              preferredShellId: 'pwsh',
              preferredShellLabel: 'PowerShell 7',
              shellCount: 1,
              verifiedShellCount: 1,
              toolCount: 3,
              verifiedToolCount: 3,
              harnessCount: 0,
              verifiedHarnessCount: 0,
              supportsPowerShell: true,
              supportsPosixShell: false,
            },
          },
        }}
        servers={[]}
        traffic={[]}
        providers={[
          {
            provider: 'openai',
            name: 'OpenAI',
            configured: true,
            authenticated: true,
            authMethod: 'api_key',
            tier: 'pro',
            limit: 1000,
            used: 350,
            remaining: 650,
            availability: 'quota_exhausted',
          },
        ]}
        fallbackChain={[]}
        sessions={[]}
      />,
    );

    expect(html).toContain('Quota exhausted');
    expect(html).toContain('Provider routing has degraded capacity');
  });
});