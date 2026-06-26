import { EventEmitter } from 'events';

/**
 * Rate limit configuration for a provider
 */
export interface RateLimitConfig {
  /** Requests per minute */
  requestsPerMinute: number;
  /** Requests per hour */
  requestsPerHour: number;
  /** Tokens per minute (optional) */
  tokensPerMinute?: number;
  /** Tokens per day (optional) */
  tokensPerDay?: number;
  /** Cost per 1K tokens (for budget tracking) */
  costPer1kTokens?: number;
  /** Maximum concurrent requests */
  maxConcurrent?: number;
}

/**
 * Usage tracking for a provider
 */
interface ProviderUsage {
  /** Requests in current minute window */
  requestsThisMinute: number;
  /** Requests in current hour window */
  requestsThisHour: number;
  /** Tokens used in current minute window */
  tokensThisMinute: number;
  /** Tokens used today */
  tokensToday: number;
  /** Total cost accumulated */
  totalCost: number;
  /** Current concurrent requests */
  concurrentRequests: number;
  /** Minute window start */
  minuteWindowStart: number;
  /** Hour window start */
  hourWindowStart: number;
  /** Day window start */
  dayWindowStart: number;
  /** Request history for analytics */
  requestHistory: RequestRecord[];
  /** Whether provider is currently throttled */
  isThrottled: boolean;
  /** Throttle end time if throttled */
  throttleEndTime?: number;
}

/**
 * Record of a single request
 */
interface RequestRecord {
  timestamp: number;
  tokensUsed: number;
  latencyMs: number;
  success: boolean;
  cost: number;
}

/**
 * Quota check result
 */
export interface QuotaCheckResult {
  allowed: boolean;
  reason?: string;
  waitTimeMs?: number;
  currentUsage: {
    requestsPerMinute: number;
    requestsPerHour: number;
    tokensPerMinute: number;
    tokensPerDay: number;
    concurrentRequests: number;
    totalCost: number;
  };
  limits: RateLimitConfig;
}

/**
 * Provider statistics
 */
export interface ProviderStats {
  provider: string;
  usage: {
    requestsThisMinute: number;
    requestsThisHour: number;
    tokensThisMinute: number;
    tokensToday: number;
    concurrentRequests: number;
    totalCost: number;
  };
  limits: RateLimitConfig;
  health: {
    isThrottled: boolean;
    throttleEndTime?: number;
    successRate: number;
    avgLatencyMs: number;
  };
  utilizationPercent: {
    requestsPerMinute: number;
    requestsPerHour: number;
    tokensPerMinute: number;
    tokensPerDay: number;
  };
}

/**
 * Default rate limits for known providers
 */
const DEFAULT_PROVIDER_LIMITS: Record<string, RateLimitConfig> = {
  openai: {
    requestsPerMinute: 60,
    requestsPerHour: 3500,
    tokensPerMinute: 90000,
    tokensPerDay: 1000000,
    costPer1kTokens: 0.03,
    maxConcurrent: 10,
  },
  anthropic: {
    requestsPerMinute: 50,
    requestsPerHour: 2000,
    tokensPerMinute: 100000,
    tokensPerDay: 1000000,
    costPer1kTokens: 0.015,
    maxConcurrent: 10,
  },
  gemini: {
    requestsPerMinute: 60,
    requestsPerHour: 1500,
    tokensPerMinute: 60000,
    tokensPerDay: 500000,
    costPer1kTokens: 0.001,
    maxConcurrent: 10,
  },
  deepseek: {
    requestsPerMinute: 60,
    requestsPerHour: 3000,
    tokensPerMinute: 100000,
    tokensPerDay: 2000000,
    costPer1kTokens: 0.0014,
    maxConcurrent: 10,
  },
  grok: {
    requestsPerMinute: 30,
    requestsPerHour: 1000,
    tokensPerMinute: 50000,
    tokensPerDay: 500000,
    costPer1kTokens: 0.005,
    maxConcurrent: 5,
  },
  qwen: {
    requestsPerMinute: 60,
    requestsPerHour: 2000,
    tokensPerMinute: 80000,
    tokensPerDay: 1000000,
    costPer1kTokens: 0.002,
    maxConcurrent: 10,
  },
  kimi: {
    requestsPerMinute: 40,
    requestsPerHour: 1500,
    tokensPerMinute: 60000,
    tokensPerDay: 500000,
    costPer1kTokens: 0.002,
    maxConcurrent: 8,
  },
};

/**
 * Global configuration
 */
export interface QuotaManagerConfig {
  enabled: boolean;
  /** Global daily budget limit in USD */
  dailyBudgetLimit?: number;
  /** Alert threshold percentage (0-1) */
  alertThreshold: number;
  /** History retention in hours */
  historyRetentionHours: number;
  /** Enable automatic throttling on rate limit errors */
  autoThrottle: boolean;
  /** Throttle duration in seconds after rate limit error */
  throttleDurationSeconds: number;
}

/**
 * Rate limiting and quota management service for API providers
 */
class QuotaManagerService extends EventEmitter {
  private config: QuotaManagerConfig = {
    enabled: true,
    alertThreshold: 0.8,
    historyRetentionHours: 24,
    autoThrottle: true,
    throttleDurationSeconds: 60,
  };

  private providerLimits: Map<string, RateLimitConfig> = new Map();
  private providerUsage: Map<string, ProviderUsage> = new Map();
  private dailyCostTotal = 0;
  private dayStart = this.getStartOfDay();

  constructor() {
    super();
    // Initialize default limits
    for (const [provider, limits] of Object.entries(DEFAULT_PROVIDER_LIMITS)) {
      this.providerLimits.set(provider, limits);
    }

    // Start cleanup interval
    setInterval(() => this.cleanup(), 60000); // Every minute
  }

  /**
   * Check if a request is allowed for a provider
   */
  checkQuota(provider: string): QuotaCheckResult {
    if (!this.config.enabled) {
      return {
        allowed: true,
        currentUsage: this.getEmptyUsage(),
        limits: this.getLimits(provider),
      };
    }

    this.ensureUsageRecord(provider);
    this.updateWindows(provider);

    const usage = this.providerUsage.get(provider)!;
    const limits = this.getLimits(provider);

    // Check if throttled
    if (usage.isThrottled && usage.throttleEndTime) {
      const now = Date.now();
      if (now < usage.throttleEndTime) {
        return {
          allowed: false,
          reason: 'Provider is temporarily throttled due to rate limit errors',
          waitTimeMs: usage.throttleEndTime - now,
          currentUsage: this.formatUsage(usage),
          limits,
        };
      } else {
        usage.isThrottled = false;
        usage.throttleEndTime = undefined;
      }
    }

    // Check concurrent requests
    if (limits.maxConcurrent && usage.concurrentRequests >= limits.maxConcurrent) {
      return {
        allowed: false,
        reason: `Max concurrent requests (${limits.maxConcurrent}) reached`,
        waitTimeMs: 1000,
        currentUsage: this.formatUsage(usage),
        limits,
      };
    }

    // Check requests per minute
    if (usage.requestsThisMinute >= limits.requestsPerMinute) {
      const waitTime = 60000 - (Date.now() - usage.minuteWindowStart);
      return {
        allowed: false,
        reason: `Rate limit: ${limits.requestsPerMinute} requests per minute exceeded`,
        waitTimeMs: Math.max(0, waitTime),
        currentUsage: this.formatUsage(usage),
        limits,
      };
    }

    // Check requests per hour
    if (usage.requestsThisHour >= limits.requestsPerHour) {
      const waitTime = 3600000 - (Date.now() - usage.hourWindowStart);
      return {
        allowed: false,
        reason: `Rate limit: ${limits.requestsPerHour} requests per hour exceeded`,
        waitTimeMs: Math.max(0, waitTime),
        currentUsage: this.formatUsage(usage),
        limits,
      };
    }

    // Check tokens per minute (if configured)
    if (limits.tokensPerMinute && usage.tokensThisMinute >= limits.tokensPerMinute) {
      const waitTime = 60000 - (Date.now() - usage.minuteWindowStart);
      return {
        allowed: false,
        reason: `Token limit: ${limits.tokensPerMinute} tokens per minute exceeded`,
        waitTimeMs: Math.max(0, waitTime),
        currentUsage: this.formatUsage(usage),
        limits,
      };
    }

    // Check tokens per day (if configured)
    if (limits.tokensPerDay && usage.tokensToday >= limits.tokensPerDay) {
      const waitTime = this.getStartOfDay() + 86400000 - Date.now();
      return {
        allowed: false,
        reason: `Token limit: ${limits.tokensPerDay} tokens per day exceeded`,
        waitTimeMs: Math.max(0, waitTime),
        currentUsage: this.formatUsage(usage),
        limits,
      };
    }

    // Check daily budget
    if (this.config.dailyBudgetLimit && this.dailyCostTotal >= this.config.dailyBudgetLimit) {
      const waitTime = this.getStartOfDay() + 86400000 - Date.now();
      return {
        allowed: false,
        reason: `Daily budget limit ($${this.config.dailyBudgetLimit}) exceeded`,
        waitTimeMs: Math.max(0, waitTime),
        currentUsage: this.formatUsage(usage),
        limits,
      };
    }

    // Emit warning if approaching limits
    this.checkAlertThresholds(provider, usage, limits);

    return {
      allowed: true,
      currentUsage: this.formatUsage(usage),
      limits,
    };
  }

  /**
   * Record that a request is starting
   */
  startRequest(provider: string): void {
    this.ensureUsageRecord(provider);
    const usage = this.providerUsage.get(provider)!;
    usage.concurrentRequests++;
  }

  /**
   * Record that a request has completed
   */
  recordRequest(
    provider: string,
    tokensUsed: number,
    latencyMs: number,
    success: boolean
  ): void {
    this.ensureUsageRecord(provider);
    this.updateWindows(provider);

    const usage = this.providerUsage.get(provider)!;
    const limits = this.getLimits(provider);

    // Update counters
    usage.requestsThisMinute++;
    usage.requestsThisHour++;
    usage.tokensThisMinute += tokensUsed;
    usage.tokensToday += tokensUsed;
    usage.concurrentRequests = Math.max(0, usage.concurrentRequests - 1);

    // Calculate cost
    const cost = limits.costPer1kTokens ? (tokensUsed / 1000) * limits.costPer1kTokens : 0;
    usage.totalCost += cost;
    this.dailyCostTotal += cost;

    // Record for history
    usage.requestHistory.push({
      timestamp: Date.now(),
      tokensUsed,
      latencyMs,
      success,
      cost,
    });

    // Trim history if needed
    const cutoff = Date.now() - this.config.historyRetentionHours * 3600000;
    usage.requestHistory = usage.requestHistory.filter((r) => r.timestamp > cutoff);

    // Emit metrics event
    this.emit('request', {
      provider,
      tokensUsed,
      latencyMs,
      success,
      cost,
      currentUsage: this.formatUsage(usage),
    });
  }

  /**
   * Record a rate limit error from the provider
   */
  recordRateLimitError(provider: string): void {
    if (!this.config.autoThrottle) return;

    this.ensureUsageRecord(provider);
    const usage = this.providerUsage.get(provider)!;

    usage.isThrottled = true;
    usage.throttleEndTime = Date.now() + this.config.throttleDurationSeconds * 1000;

    this.emit('throttled', {
      provider,
      throttleEndTime: usage.throttleEndTime,
      durationSeconds: this.config.throttleDurationSeconds,
    });
  }

  /**
   * Set rate limits for a provider
   */
  setProviderLimits(provider: string, limits: Partial<RateLimitConfig>): void {
    const existing = this.providerLimits.get(provider) || DEFAULT_PROVIDER_LIMITS.openai;
    this.providerLimits.set(provider, { ...existing, ...limits });
  }

  /**
   * Get rate limits for a provider
   */
  getLimits(provider: string): RateLimitConfig {
    return this.providerLimits.get(provider) || DEFAULT_PROVIDER_LIMITS.openai;
  }

  /**
   * Get statistics for a provider
   */
  getProviderStats(provider: string): ProviderStats {
    this.ensureUsageRecord(provider);
    this.updateWindows(provider);

    const usage = this.providerUsage.get(provider)!;
    const limits = this.getLimits(provider);

    // Calculate success rate and avg latency from history
    const recentHistory = usage.requestHistory.filter(
      (r) => r.timestamp > Date.now() - 3600000
    );
    const successCount = recentHistory.filter((r) => r.success).length;
    const successRate = recentHistory.length > 0 ? successCount / recentHistory.length : 1;
    const avgLatencyMs =
      recentHistory.length > 0
        ? recentHistory.reduce((sum, r) => sum + r.latencyMs, 0) / recentHistory.length
        : 0;

    return {
      provider,
      usage: {
        requestsThisMinute: usage.requestsThisMinute,
        requestsThisHour: usage.requestsThisHour,
        tokensThisMinute: usage.tokensThisMinute,
        tokensToday: usage.tokensToday,
        concurrentRequests: usage.concurrentRequests,
        totalCost: usage.totalCost,
      },
      limits,
      health: {
        isThrottled: usage.isThrottled,
        throttleEndTime: usage.throttleEndTime,
        successRate,
        avgLatencyMs: Math.round(avgLatencyMs),
      },
      utilizationPercent: {
        requestsPerMinute: (usage.requestsThisMinute / limits.requestsPerMinute) * 100,
        requestsPerHour: (usage.requestsThisHour / limits.requestsPerHour) * 100,
        tokensPerMinute: limits.tokensPerMinute
          ? (usage.tokensThisMinute / limits.tokensPerMinute) * 100
          : 0,
        tokensPerDay: limits.tokensPerDay
          ? (usage.tokensToday / limits.tokensPerDay) * 100
          : 0,
      },
    };
  }

  /**
   * Get statistics for all providers
   */
  getAllStats(): {
    providers: ProviderStats[];
    global: {
      dailyCostTotal: number;
      dailyBudgetLimit?: number;
      budgetUtilizationPercent: number;
    };
  } {
    const providers: ProviderStats[] = [];
    for (const provider of this.providerUsage.keys()) {
      providers.push(this.getProviderStats(provider));
    }

    return {
      providers,
      global: {
        dailyCostTotal: this.dailyCostTotal,
        dailyBudgetLimit: this.config.dailyBudgetLimit,
        budgetUtilizationPercent: this.config.dailyBudgetLimit
          ? (this.dailyCostTotal / this.config.dailyBudgetLimit) * 100
          : 0,
      },
    };
  }

  /**
   * Get status summary
   */
  getStatus(): {
    enabled: boolean;
    config: QuotaManagerConfig;
    providersTracked: number;
    dailyCostTotal: number;
    throttledProviders: string[];
  } {
    const throttledProviders: string[] = [];
    for (const [provider, usage] of this.providerUsage.entries()) {
      if (usage.isThrottled) {
        throttledProviders.push(provider);
      }
    }

    return {
      enabled: this.config.enabled,
      config: this.config,
      providersTracked: this.providerUsage.size,
      dailyCostTotal: this.dailyCostTotal,
      throttledProviders,
    };
  }

  /**
   * Update configuration
   */
  setConfig(config: Partial<QuotaManagerConfig>): void {
    this.config = { ...this.config, ...config };
    this.emit('configChanged', this.config);
  }

  /**
   * Get configuration
   */
  getConfig(): QuotaManagerConfig {
    return { ...this.config };
  }

  /**
   * Enable/disable quota management
   */
  setEnabled(enabled: boolean): void {
    this.config.enabled = enabled;
    this.emit('enabledChanged', enabled);
  }

  /**
   * Check if enabled
   */
  isEnabled(): boolean {
    return this.config.enabled;
  }

  /**
   * Reset usage for a provider
   */
  resetProviderUsage(provider: string): void {
    this.providerUsage.delete(provider);
    this.emit('usageReset', { provider });
  }

  /**
   * Reset all usage data
   */
  resetAllUsage(): void {
    this.providerUsage.clear();
    this.dailyCostTotal = 0;
    this.dayStart = this.getStartOfDay();
    this.emit('allUsageReset', {});
  }

  /**
   * Unthrottle a provider manually
   */
  unthrottleProvider(provider: string): void {
    const usage = this.providerUsage.get(provider);
    if (usage) {
      usage.isThrottled = false;
      usage.throttleEndTime = undefined;
      this.emit('unthrottled', { provider });
    }
  }

  // ============ Private Methods ============

  private ensureUsageRecord(provider: string): void {
    if (!this.providerUsage.has(provider)) {
      const now = Date.now();
      this.providerUsage.set(provider, {
        requestsThisMinute: 0,
        requestsThisHour: 0,
        tokensThisMinute: 0,
        tokensToday: 0,
        totalCost: 0,
        concurrentRequests: 0,
        minuteWindowStart: now,
        hourWindowStart: now,
        dayWindowStart: this.getStartOfDay(),
        requestHistory: [],
        isThrottled: false,
      });
    }
  }

  private updateWindows(provider: string): void {
    const usage = this.providerUsage.get(provider);
    if (!usage) return;

    const now = Date.now();

    // Reset minute window
    if (now - usage.minuteWindowStart >= 60000) {
      usage.requestsThisMinute = 0;
      usage.tokensThisMinute = 0;
      usage.minuteWindowStart = now;
    }

    // Reset hour window
    if (now - usage.hourWindowStart >= 3600000) {
      usage.requestsThisHour = 0;
      usage.hourWindowStart = now;
    }

    // Reset day window
    const today = this.getStartOfDay();
    if (today > usage.dayWindowStart) {
      usage.tokensToday = 0;
      usage.totalCost = 0;
      usage.dayWindowStart = today;
    }

    // Reset global daily cost
    if (today > this.dayStart) {
      this.dailyCostTotal = 0;
      this.dayStart = today;
    }
  }

  private getStartOfDay(): number {
    const now = new Date();
    return new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  }

  private formatUsage(usage: ProviderUsage): QuotaCheckResult['currentUsage'] {
    return {
      requestsPerMinute: usage.requestsThisMinute,
      requestsPerHour: usage.requestsThisHour,
      tokensPerMinute: usage.tokensThisMinute,
      tokensPerDay: usage.tokensToday,
      concurrentRequests: usage.concurrentRequests,
      totalCost: usage.totalCost,
    };
  }

  private getEmptyUsage(): QuotaCheckResult['currentUsage'] {
    return {
      requestsPerMinute: 0,
      requestsPerHour: 0,
      tokensPerMinute: 0,
      tokensPerDay: 0,
      concurrentRequests: 0,
      totalCost: 0,
    };
  }

  private checkAlertThresholds(
    provider: string,
    usage: ProviderUsage,
    limits: RateLimitConfig
  ): void {
    const threshold = this.config.alertThreshold;

    // Check requests per minute
    if (usage.requestsThisMinute / limits.requestsPerMinute >= threshold) {
      this.emit('alert', {
        provider,
        type: 'requestsPerMinute',
        utilization: usage.requestsThisMinute / limits.requestsPerMinute,
        message: `${provider} approaching requests per minute limit (${usage.requestsThisMinute}/${limits.requestsPerMinute})`,
      });
    }

    // Check requests per hour
    if (usage.requestsThisHour / limits.requestsPerHour >= threshold) {
      this.emit('alert', {
        provider,
        type: 'requestsPerHour',
        utilization: usage.requestsThisHour / limits.requestsPerHour,
        message: `${provider} approaching requests per hour limit (${usage.requestsThisHour}/${limits.requestsPerHour})`,
      });
    }

    // Check daily budget
    if (
      this.config.dailyBudgetLimit &&
      this.dailyCostTotal / this.config.dailyBudgetLimit >= threshold
    ) {
      this.emit('alert', {
        provider: 'global',
        type: 'dailyBudget',
        utilization: this.dailyCostTotal / this.config.dailyBudgetLimit,
        message: `Approaching daily budget limit ($${this.dailyCostTotal.toFixed(2)}/$${this.config.dailyBudgetLimit})`,
      });
    }
  }

  private cleanup(): void {
    const now = Date.now();
    const cutoff = now - this.config.historyRetentionHours * 3600000;

    for (const usage of this.providerUsage.values()) {
      usage.requestHistory = usage.requestHistory.filter((r) => r.timestamp > cutoff);
    }
  }
}

export const quotaManager = new QuotaManagerService();
