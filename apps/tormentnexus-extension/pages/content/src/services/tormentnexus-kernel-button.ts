/**
 * TormentNexus Kernel Button Injection Service
 *
 * Detects supported AI chat websites (Claude.ai, ChatGPT) and injects
 * a "TormentNexus Kernel" button that attaches the local TormentNexus Kernel to the
 * active conversation session.
 *
 * The button provides:
 * - One-click attachment to the local TormentNexus Kernel (port 4300)
 * - Visual status indicator (connected/disconnected/warming)
 * - Context bridge: sends conversation context to TormentNexus for memory & healing
 */

import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('TormentNexusKernelButton');

// ─── Types ───────────────────────────────────────────────────────────────────

export type TormentNexusKernelStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export interface TormentNexusKernelConfig {
  kernelUrl: string;
  bridgeUrl: string;
  autoAttach: boolean;
  pollIntervalMs: number;
}

export interface TormentNexusKernelButtonState {
  status: TormentNexusKernelStatus;
  kernelUrl: string;
  sessionId: string | null;
  connectedAt: number | null;
  error: string | null;
}

// ─── Constants ───────────────────────────────────────────────────────────────

const TORMENTNEXUS_KERNEL_BUTTON_ID = 'tormentnexus-tormentnexus-kernel-btn';
const TORMENTNEXUS_KERNEL_STATUS_ID = 'tormentnexus-tormentnexus-kernel-status';
const TORMENTNEXUS_KERNEL_PANEL_ID = 'tormentnexus-tormentnexus-kernel-panel';

const DEFAULT_CONFIG: TormentNexusKernelConfig = {
  kernelUrl: 'http://127.0.0.1:4300',
  bridgeUrl: 'http://127.0.0.1:4100',
  autoAttach: false,
  pollIntervalMs: 5000,
};

// Supported websites with their injection selectors
const SUPPORTED_SITES = {
  'claude.ai': {
    name: 'Claude',
    buttonAnchor: '[data-testid="model-selector"], .claude-header, header',
    conversationSelector: '[data-testid="conversation-turn"], .conversation-turn',
    inputSelector: '[contenteditable="true"], textarea, [data-testid="chat-input"]',
  },
  'chatgpt.com': {
    name: 'ChatGPT',
    buttonAnchor: '#__next header, nav, [class*="header"], [class*="nav"]',
    conversationSelector: '[data-testid="conversation-turn"], [class*="conversation"], [class*="message"]',
    inputSelector: '#prompt-textarea, textarea[placeholder], [contenteditable="true"]',
  },
  'chat.openai.com': {
    name: 'ChatGPT',
    buttonAnchor: '#__next header, nav',
    conversationSelector: '[class*="conversation"], [class*="message"]',
    inputSelector: '#prompt-textarea, textarea, [contenteditable="true"]',
  },
} as const;

type SupportedSite = keyof typeof SUPPORTED_SITES;

// ─── Button Styles ───────────────────────────────────────────────────────────

const BUTTON_STYLES = `
  #${TORMENTNEXUS_KERNEL_BUTTON_ID} {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 8px;
    border: 1px solid rgba(56, 189, 248, 0.3);
    background: rgba(56, 189, 248, 0.08);
    color: #7dd3fc;
    font-size: 12px;
    font-weight: 600;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    cursor: pointer;
    transition: all 0.15s ease;
    white-space: nowrap;
    user-select: none;
    letter-spacing: 0.02em;
    z-index: 9999;
  }
  #${TORMENTNEXUS_KERNEL_BUTTON_ID}:hover {
    background: rgba(56, 189, 248, 0.15);
    border-color: rgba(56, 189, 248, 0.5);
    color: #38bdf8;
  }
  #${TORMENTNEXUS_KERNEL_BUTTON_ID}.tormentnexus-connected {
    border-color: rgba(52, 211, 153, 0.4);
    background: rgba(52, 211, 153, 0.08);
    color: #6ee7b7;
  }
  #${TORMENTNEXUS_KERNEL_BUTTON_ID}.tormentnexus-connected:hover {
    background: rgba(52, 211, 153, 0.15);
    border-color: rgba(52, 211, 153, 0.5);
  }
  #${TORMENTNEXUS_KERNEL_BUTTON_ID}.tormentnexus-error {
    border-color: rgba(248, 113, 113, 0.4);
    background: rgba(248, 113, 113, 0.08);
    color: #fca5a5;
  }
  #${TORMENTNEXUS_KERNEL_STATUS_ID} {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #64748b;
    flex-shrink: 0;
    transition: background 0.2s;
  }
  .tormentnexus-connected #${TORMENTNEXUS_KERNEL_STATUS_ID} {
    background: #34d399;
    box-shadow: 0 0 6px rgba(52, 211, 153, 0.5);
    animation: tormentnexus-pulse 2s ease-in-out infinite;
  }
  .tormentnexus-error #${TORMENTNEXUS_KERNEL_STATUS_ID} {
    background: #f87171;
  }
  .tormentnexus-connecting #${TORMENTNEXUS_KERNEL_STATUS_ID} {
    background: #fbbf24;
    animation: tormentnexus-blink 1s ease-in-out infinite;
  }
  @keyframes tormentnexus-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.6; }
  }
  @keyframes tormentnexus-blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }
  #${TORMENTNEXUS_KERNEL_PANEL_ID} {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 8px;
    width: 280px;
    padding: 12px;
    border-radius: 12px;
    border: 1px solid rgba(56, 189, 248, 0.2);
    background: rgba(15, 23, 42, 0.95);
    backdrop-filter: blur(12px);
    color: #e2e8f0;
    font-size: 12px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    z-index: 10000;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  }
`;

// ─── TormentNexus Kernel Button Service ─────────────────────────────────────────────

export class TormentNexusKernelButtonService {
  private config: TormentNexusKernelConfig;
  private state: TormentNexusKernelButtonState;
  private button: HTMLElement | null = null;
  private panel: HTMLElement | null = null;
  private pollTimer: ReturnType<typeof setInterval> | null = null;
  private siteConfig: (typeof SUPPORTED_SITES)[SupportedSite] | null = null;

  constructor(config: Partial<TormentNexusKernelConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.state = {
      status: 'disconnected',
      kernelUrl: this.config.kernelUrl,
      sessionId: null,
      connectedAt: null,
      error: null,
    };
  }

  /** Detect if the current page is a supported site */
  detectSite(): SupportedSite | null {
    const hostname = window.location.hostname.replace(/^www\./, '');
    if (hostname in SUPPORTED_SITES) {
      return hostname as SupportedSite;
    }
    return null;
  }

  /** Initialize the button injection */
  async initialize(): Promise<boolean> {
    const site = this.detectSite();
    if (!site) {
      logger.debug('TormentNexusKernel: unsupported site, skipping injection');
      return false;
    }

    this.siteConfig = SUPPORTED_SITES[site];
    logger.info(`TormentNexusKernel: detected ${SUPPORTED_SITES[site].name}, injecting button`);

    // Inject styles
    this.injectStyles();

    // Wait for the anchor element to appear
    const anchor = await this.waitForElement(this.siteConfig.buttonAnchor, 10000);
    if (!anchor) {
      logger.warn('TormentNexusKernel: could not find button anchor element');
      return false;
    }

    // Create and inject the button
    this.createButton(anchor);

    // Start health polling
    this.startPolling();

    // If auto-attach, connect immediately
    if (this.config.autoAttach) {
      this.attach();
    }

    return true;
  }

  /** Inject the button styles into the page */
  private injectStyles(): void {
    if (document.getElementById('tormentnexus-tormentnexus-kernel-styles')) return;

    const style = document.createElement('style');
    style.id = 'tormentnexus-tormentnexus-kernel-styles';
    style.textContent = BUTTON_STYLES;
    document.head.appendChild(style);
  }

  /** Create and inject the TormentNexus Kernel button */
  private createButton(anchor: Element): void {
    if (document.getElementById(TORMENTNEXUS_KERNEL_BUTTON_ID)) return;

    const btn = document.createElement('button');
    btn.id = TORMENTNEXUS_KERNEL_BUTTON_ID;
    btn.type = 'button';
    btn.title = 'Attach to TormentNexus Kernel';

    btn.innerHTML = `
      <span id="${TORMENTNEXUS_KERNEL_STATUS_ID}"></span>
      <span>TormentNexus Kernel</span>
    `;

    btn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      this.handleClick();
    });

    // Insert the button near the anchor
    // Try to append to the anchor's parent or after the anchor
    const container = anchor.parentElement || anchor;
    container.style.position = 'relative';
    container.appendChild(btn);
    this.button = btn;

    logger.info('TormentNexusKernel: button injected');
  }

  /** Handle button click — toggle attach/detach or show panel */
  private handleClick(): void {
    if (this.state.status === 'connected') {
      this.showPanel();
    } else {
      this.attach();
    }
  }

  /** Attach to the local TormentNexus Kernel */
  async attach(): Promise<void> {
    this.setState({ status: 'connecting', error: null });

    try {
      const response = await fetch(`${this.config.kernelUrl}/api/native/status`, {
        method: 'GET',
        signal: AbortSignal.timeout(5000),
      });

      if (!response.ok) {
        throw new Error(`Kernel returned ${response.status}`);
      }

      const data = await response.json();

      this.setState({
        status: 'connected',
        sessionId: data.sessionId || this.generateSessionId(),
        connectedAt: Date.now(),
        error: null,
      });

      logger.info('TormentNexusKernel: attached to kernel', data);

      // Register this chat session with the kernel
      await this.registerSession();
    } catch (err: any) {
      this.setState({
        status: 'error',
        error: err.message || 'Connection failed',
      });
      logger.warn('TormentNexusKernel: attach failed', err.message);
    }
  }

  /** Detach from the kernel */
  detach(): void {
    this.setState({
      status: 'disconnected',
      sessionId: null,
      connectedAt: null,
      error: null,
    });
    logger.info('TormentNexusKernel: detached');
  }

  /** Register this session with the TormentNexus Kernel */
  private async registerSession(): Promise<void> {
    if (!this.state.sessionId) return;

    try {
      const site = this.detectSite();
      const payload = {
        sessionId: this.state.sessionId,
        source: site ? SUPPORTED_SITES[site].name : 'unknown',
        url: window.location.href,
        attachedAt: Date.now(),
      };

      await fetch(`${this.config.kernelUrl}/api/native/session/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: AbortSignal.timeout(3000),
      });
    } catch (err: any) {
      logger.debug('TormentNexusKernel: session registration failed (non-critical)', err.message);
    }
  }

  /** Show the info panel */
  private showPanel(): void {
    if (this.panel) {
      this.panel.remove();
      this.panel = null;
      return;
    }

    const panel = document.createElement('div');
    panel.id = TORMENTNEXUS_KERNEL_PANEL_ID;

    const uptime = this.state.connectedAt
      ? Math.round((Date.now() - this.state.connectedAt) / 1000)
      : 0;
    const uptimeStr = uptime >= 60 ? `${Math.floor(uptime / 60)}m ${uptime % 60}s` : `${uptime}s`;

    panel.innerHTML = `
      <div style="margin-bottom: 8px; font-weight: 700; color: #38bdf8;">
        TormentNexus Kernel Connected
      </div>
      <div style="color: #94a3b8; line-height: 1.6;">
        <div>Session: <span style="color: #e2e8f0;">${this.state.sessionId?.slice(0, 12)}...</span></div>
        <div>Uptime: <span style="color: #e2e8f0;">${uptimeStr}</span></div>
        <div>Kernel: <span style="color: #e2e8f0;">${this.state.kernelUrl}</span></div>
      </div>
      <div style="margin-top: 10px; display: flex; gap: 6px;">
        <button id="tormentnexus-detach-btn" style="
          flex: 1; padding: 6px; border-radius: 6px;
          border: 1px solid rgba(248,113,113,0.3);
          background: rgba(248,113,113,0.08);
          color: #fca5a5; font-size: 11px; cursor: pointer;
        ">Detach</button>
        <button id="tormentnexus-reconnect-btn" style="
          flex: 1; padding: 6px; border-radius: 6px;
          border: 1px solid rgba(56,189,248,0.3);
          background: rgba(56,189,248,0.08);
          color: #7dd3fc; font-size: 11px; cursor: pointer;
        ">Reconnect</button>
      </div>
    `;

    this.button?.appendChild(panel);
    this.panel = panel;

    // Wire up panel buttons
    document.getElementById('tormentnexus-detach-btn')?.addEventListener('click', () => {
      this.detach();
      this.showPanel(); // Toggle panel off
    });

    document.getElementById('tormentnexus-reconnect-btn')?.addEventListener('click', () => {
      this.detach();
      this.attach();
      this.showPanel(); // Toggle panel off
    });

    // Close panel when clicking outside
    const closePanel = (e: MouseEvent) => {
      if (!panel.contains(e.target as Node) && !this.button?.contains(e.target as Node)) {
        this.showPanel(); // Toggle off
        document.removeEventListener('click', closePanel);
      }
    };
    setTimeout(() => document.addEventListener('click', closePanel), 0);
  }

  /** Start polling for kernel health */
  private startPolling(): void {
    this.stopPolling();
    this.pollTimer = setInterval(() => this.healthCheck(), this.config.pollIntervalMs);
  }

  /** Stop polling */
  private stopPolling(): void {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = null;
    }
  }

  /** Check kernel health */
  private async healthCheck(): Promise<void> {
    if (this.state.status === 'disconnected') return;

    try {
      const response = await fetch(`${this.config.kernelUrl}/api/native/status`, {
        method: 'GET',
        signal: AbortSignal.timeout(3000),
      });

      if (!response.ok) {
        throw new Error(`Kernel returned ${response.status}`);
      }

      if (this.state.status === 'error') {
        // Kernel recovered
        this.setState({ status: 'connected', error: null });
      }
    } catch (err: any) {
      if (this.state.status === 'connected') {
        this.setState({ status: 'error', error: 'Kernel unreachable' });
      }
    }
  }

  /** Update state and re-render button */
  private setState(partial: Partial<TormentNexusKernelButtonState>): void {
    this.state = { ...this.state, ...partial };

    if (this.button) {
      this.button.className = '';
      switch (this.state.status) {
        case 'connected':
          this.button.classList.add('tormentnexus-connected');
          this.button.title = 'TormentNexus Kernel: Connected';
          break;
        case 'connecting':
          this.button.classList.add('tormentnexus-connecting');
          this.button.title = 'TormentNexus Kernel: Connecting...';
          break;
        case 'error':
          this.button.classList.add('tormentnexus-error');
          this.button.title = `TormentNexus Kernel: Error - ${this.state.error}`;
          break;
        default:
          this.button.title = 'Attach to TormentNexus Kernel';
      }
    }
  }

  /** Generate a session ID */
  private generateSessionId(): string {
    const site = this.detectSite();
    const prefix = site ? SUPPORTED_SITES[site].name.toLowerCase() : 'unknown';
    return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
  }

  /** Wait for an element matching a selector to appear in the DOM */
  private waitForElement(selector: string, timeoutMs: number): Promise<Element | null> {
    // Check if already present
    const existing = document.querySelector(selector);
    if (existing) return Promise.resolve(existing);

    return new Promise((resolve) => {
      const observer = new MutationObserver(() => {
        const el = document.querySelector(selector);
        if (el) {
          observer.disconnect();
          resolve(el);
        }
      });

      observer.observe(document.body, { childList: true, subtree: true });

      setTimeout(() => {
        observer.disconnect();
        resolve(document.querySelector(selector));
      }, timeoutMs);
    });
  }

  /** Clean up the button and stop polling */
  destroy(): void {
    this.stopPolling();
    this.button?.remove();
    this.panel?.remove();
    this.button = null;
    this.panel = null;

    const styles = document.getElementById('tormentnexus-tormentnexus-kernel-styles');
    styles?.remove();
  }
}

// ─── Singleton ───────────────────────────────────────────────────────────────

let instance: TormentNexusKernelButtonService | null = null;

/** Get or create the TormentNexus Kernel button service */
export function getTormentNexusKernelService(config?: Partial<TormentNexusKernelConfig>): TormentNexusKernelButtonService {
  if (!instance) {
    instance = new TormentNexusKernelButtonService(config);
  }
  return instance;
}

/** Auto-initialize the TormentNexus Kernel button if on a supported site */
export async function initTormentNexusKernelButton(config?: Partial<TormentNexusKernelConfig>): Promise<boolean> {
  const service = getTormentNexusKernelService(config);
  return service.initialize();
}
