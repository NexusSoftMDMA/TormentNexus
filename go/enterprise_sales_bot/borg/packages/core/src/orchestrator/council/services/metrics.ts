interface RequestMetric {
  count: number;
  totalLatencyMs: number;
  errors: number;
  lastRequestAt: number;
}

interface SupervisorMetric {
  calls: number;
  totalLatencyMs: number;
  errors: number;
  retries: number;
  lastCallAt: number;
}

interface DebateMetric {
  count: number;
  totalLatencyMs: number;
  totalRounds: number;
  consensusReached: number;
  lastDebateAt: number;
}

class MetricsService {
  private httpRequests: Map<string, RequestMetric> = new Map();
  private supervisorCalls: Map<string, SupervisorMetric> = new Map();
  private debateMetrics: DebateMetric = {
    count: 0,
    totalLatencyMs: 0,
    totalRounds: 0,
    consensusReached: 0,
    lastDebateAt: 0,
  };
  private startTime: number = Date.now();

  recordHttpRequest(method: string, path: string, statusCode: number, latencyMs: number): void {
    const key = `${method} ${this.normalizePath(path)}`;
    const existing = this.httpRequests.get(key) || {
      count: 0,
      totalLatencyMs: 0,
      errors: 0,
      lastRequestAt: 0,
    };

    existing.count++;
    existing.totalLatencyMs += latencyMs;
    existing.lastRequestAt = Date.now();
    if (statusCode >= 400) {
      existing.errors++;
    }

    this.httpRequests.set(key, existing);
  }

  recordSupervisorCall(
    supervisorName: string,
    latencyMs: number,
    success: boolean,
    retryCount: number = 0
  ): void {
    const existing = this.supervisorCalls.get(supervisorName) || {
      calls: 0,
      totalLatencyMs: 0,
      errors: 0,
      retries: 0,
      lastCallAt: 0,
    };

    existing.calls++;
    existing.totalLatencyMs += latencyMs;
    existing.retries += retryCount;
    existing.lastCallAt = Date.now();
    if (!success) {
      existing.errors++;
    }

    this.supervisorCalls.set(supervisorName, existing);
  }

  recordDebate(latencyMs: number, rounds: number, consensusReached: boolean): void {
    this.debateMetrics.count++;
    this.debateMetrics.totalLatencyMs += latencyMs;
    this.debateMetrics.totalRounds += rounds;
    if (consensusReached) {
      this.debateMetrics.consensusReached++;
    }
    this.debateMetrics.lastDebateAt = Date.now();
  }

  getSummary(): {
    uptime: number;
    http: {
      totalRequests: number;
      totalErrors: number;
      avgLatencyMs: number;
      byEndpoint: Record<string, { count: number; avgLatencyMs: number; errorRate: number }>;
    };
    supervisors: {
      totalCalls: number;
      totalErrors: number;
      avgLatencyMs: number;
      byProvider: Record<string, { calls: number; avgLatencyMs: number; errorRate: number; retries: number }>;
    };
    debates: {
      count: number;
      avgLatencyMs: number;
      avgRounds: number;
      consensusRate: number;
    };
  } {
    const httpTotal = { requests: 0, latency: 0, errors: 0 };
    const byEndpoint: Record<string, { count: number; avgLatencyMs: number; errorRate: number }> = {};

    for (const [key, metric] of this.httpRequests) {
      httpTotal.requests += metric.count;
      httpTotal.latency += metric.totalLatencyMs;
      httpTotal.errors += metric.errors;
      byEndpoint[key] = {
        count: metric.count,
        avgLatencyMs: Math.round(metric.totalLatencyMs / metric.count),
        errorRate: metric.count > 0 ? metric.errors / metric.count : 0,
      };
    }

    const supervisorTotal = { calls: 0, latency: 0, errors: 0 };
    const byProvider: Record<string, { calls: number; avgLatencyMs: number; errorRate: number; retries: number }> = {};

    for (const [name, metric] of this.supervisorCalls) {
      supervisorTotal.calls += metric.calls;
      supervisorTotal.latency += metric.totalLatencyMs;
      supervisorTotal.errors += metric.errors;
      byProvider[name] = {
        calls: metric.calls,
        avgLatencyMs: Math.round(metric.totalLatencyMs / metric.calls),
        errorRate: metric.calls > 0 ? metric.errors / metric.calls : 0,
        retries: metric.retries,
      };
    }

    return {
      uptime: Math.floor((Date.now() - this.startTime) / 1000),
      http: {
        totalRequests: httpTotal.requests,
        totalErrors: httpTotal.errors,
        avgLatencyMs: httpTotal.requests > 0 ? Math.round(httpTotal.latency / httpTotal.requests) : 0,
        byEndpoint,
      },
      supervisors: {
        totalCalls: supervisorTotal.calls,
        totalErrors: supervisorTotal.errors,
        avgLatencyMs: supervisorTotal.calls > 0 ? Math.round(supervisorTotal.latency / supervisorTotal.calls) : 0,
        byProvider,
      },
      debates: {
        count: this.debateMetrics.count,
        avgLatencyMs: this.debateMetrics.count > 0
          ? Math.round(this.debateMetrics.totalLatencyMs / this.debateMetrics.count)
          : 0,
        avgRounds: this.debateMetrics.count > 0
          ? Math.round((this.debateMetrics.totalRounds / this.debateMetrics.count) * 10) / 10
          : 0,
        consensusRate: this.debateMetrics.count > 0
          ? this.debateMetrics.consensusReached / this.debateMetrics.count
          : 0,
      },
    };
  }

  getPrometheusFormat(): string {
    const lines: string[] = [];
    const prefix = 'autopilot';

    lines.push(`# HELP ${prefix}_uptime_seconds Server uptime in seconds`);
    lines.push(`# TYPE ${prefix}_uptime_seconds gauge`);
    lines.push(`${prefix}_uptime_seconds ${Math.floor((Date.now() - this.startTime) / 1000)}`);
    lines.push('');

    lines.push(`# HELP ${prefix}_http_requests_total Total HTTP requests`);
    lines.push(`# TYPE ${prefix}_http_requests_total counter`);
    for (const [key, metric] of this.httpRequests) {
      const [method, path] = key.split(' ');
      lines.push(`${prefix}_http_requests_total{method="${method}",path="${path}"} ${metric.count}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_http_errors_total Total HTTP errors (4xx, 5xx)`);
    lines.push(`# TYPE ${prefix}_http_errors_total counter`);
    for (const [key, metric] of this.httpRequests) {
      const [method, path] = key.split(' ');
      lines.push(`${prefix}_http_errors_total{method="${method}",path="${path}"} ${metric.errors}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_http_latency_ms_total Total HTTP latency in milliseconds`);
    lines.push(`# TYPE ${prefix}_http_latency_ms_total counter`);
    for (const [key, metric] of this.httpRequests) {
      const [method, path] = key.split(' ');
      lines.push(`${prefix}_http_latency_ms_total{method="${method}",path="${path}"} ${metric.totalLatencyMs}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_supervisor_calls_total Total supervisor API calls`);
    lines.push(`# TYPE ${prefix}_supervisor_calls_total counter`);
    for (const [name, metric] of this.supervisorCalls) {
      lines.push(`${prefix}_supervisor_calls_total{supervisor="${name}"} ${metric.calls}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_supervisor_errors_total Total supervisor errors`);
    lines.push(`# TYPE ${prefix}_supervisor_errors_total counter`);
    for (const [name, metric] of this.supervisorCalls) {
      lines.push(`${prefix}_supervisor_errors_total{supervisor="${name}"} ${metric.errors}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_supervisor_retries_total Total supervisor retries`);
    lines.push(`# TYPE ${prefix}_supervisor_retries_total counter`);
    for (const [name, metric] of this.supervisorCalls) {
      lines.push(`${prefix}_supervisor_retries_total{supervisor="${name}"} ${metric.retries}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_supervisor_latency_ms_total Total supervisor latency in milliseconds`);
    lines.push(`# TYPE ${prefix}_supervisor_latency_ms_total counter`);
    for (const [name, metric] of this.supervisorCalls) {
      lines.push(`${prefix}_supervisor_latency_ms_total{supervisor="${name}"} ${metric.totalLatencyMs}`);
    }
    lines.push('');

    lines.push(`# HELP ${prefix}_debates_total Total debates conducted`);
    lines.push(`# TYPE ${prefix}_debates_total counter`);
    lines.push(`${prefix}_debates_total ${this.debateMetrics.count}`);
    lines.push('');

    lines.push(`# HELP ${prefix}_debates_consensus_total Debates that reached consensus`);
    lines.push(`# TYPE ${prefix}_debates_consensus_total counter`);
    lines.push(`${prefix}_debates_consensus_total ${this.debateMetrics.consensusReached}`);
    lines.push('');

    lines.push(`# HELP ${prefix}_debates_rounds_total Total debate rounds`);
    lines.push(`# TYPE ${prefix}_debates_rounds_total counter`);
    lines.push(`${prefix}_debates_rounds_total ${this.debateMetrics.totalRounds}`);
    lines.push('');

    lines.push(`# HELP ${prefix}_debates_latency_ms_total Total debate latency in milliseconds`);
    lines.push(`# TYPE ${prefix}_debates_latency_ms_total counter`);
    lines.push(`${prefix}_debates_latency_ms_total ${this.debateMetrics.totalLatencyMs}`);

    return lines.join('\n');
  }

  reset(): void {
    this.httpRequests.clear();
    this.supervisorCalls.clear();
    this.debateMetrics = {
      count: 0,
      totalLatencyMs: 0,
      totalRounds: 0,
      consensusReached: 0,
      lastDebateAt: 0,
    };
    this.startTime = Date.now();
  }

  private normalizePath(path: string): string {
    const UUID_PATTERN = /\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi;
    const NUMERIC_ID_PATTERN = /\/[0-9]+/g;
    const LONG_ID_PATTERN = /\/[a-z0-9]{20,}/gi;
    
    return path
      .replace(UUID_PATTERN, '/:id')
      .replace(NUMERIC_ID_PATTERN, '/:id')
      .replace(LONG_ID_PATTERN, '/:id');
  }
}

export const metrics = new MetricsService();
