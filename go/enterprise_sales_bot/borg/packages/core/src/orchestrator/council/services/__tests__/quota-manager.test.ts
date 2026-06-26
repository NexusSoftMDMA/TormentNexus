import { describe, it, expect, beforeEach } from 'vitest';
import { quotaManager, type RateLimitConfig } from '../quota-manager.js';

describe('QuotaManagerService', () => {
  beforeEach(() => {
    quotaManager.resetAllUsage();
    quotaManager.setEnabled(true);
    quotaManager.setConfig({
      enabled: true,
      alertThreshold: 0.8,
      historyRetentionHours: 24,
      autoThrottle: true,
      throttleDurationSeconds: 60,
    });
  });

  describe('checkQuota', () => {
    it('should allow requests when under limits', () => {
      const result = quotaManager.checkQuota('openai');
      expect(result.allowed).toBe(true);
      expect(result.currentUsage.requestsPerMinute).toBe(0);
    });

    it('should return allowed true when disabled', () => {
      quotaManager.setEnabled(false);
      const result = quotaManager.checkQuota('openai');
      expect(result.allowed).toBe(true);
    });

    it('should block when requests per minute exceeded', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 2,
        requestsPerHour: 100,
      });

      quotaManager.recordRequest('testprovider', 100, 50, true);
      quotaManager.recordRequest('testprovider', 100, 50, true);

      const result = quotaManager.checkQuota('testprovider');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('requests per minute');
    });

    it('should block when requests per hour exceeded', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 1000,
        requestsPerHour: 2,
      });

      quotaManager.recordRequest('testprovider', 100, 50, true);
      quotaManager.recordRequest('testprovider', 100, 50, true);

      const result = quotaManager.checkQuota('testprovider');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('requests per hour');
    });

    it('should block when concurrent requests exceeded', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 100,
        requestsPerHour: 1000,
        maxConcurrent: 2,
      });

      quotaManager.startRequest('testprovider');
      quotaManager.startRequest('testprovider');

      const result = quotaManager.checkQuota('testprovider');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('concurrent requests');
    });

    it('should block when tokens per minute exceeded', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 100,
        requestsPerHour: 1000,
        tokensPerMinute: 1000,
      });

      quotaManager.recordRequest('testprovider', 1000, 50, true);

      const result = quotaManager.checkQuota('testprovider');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('tokens per minute');
    });

    it('should block when tokens per day exceeded', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 100,
        requestsPerHour: 1000,
        tokensPerDay: 500,
      });

      quotaManager.recordRequest('testprovider', 500, 50, true);

      const result = quotaManager.checkQuota('testprovider');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('tokens per day');
    });

    it('should block when daily budget exceeded', () => {
      quotaManager.setConfig({ dailyBudgetLimit: 0.001 });
      quotaManager.setProviderLimits('budgettest', {
        requestsPerMinute: 100,
        requestsPerHour: 1000,
        tokensPerMinute: 100000,
        tokensPerDay: 100000,
        costPer1kTokens: 0.01,
      });

      quotaManager.recordRequest('budgettest', 1000, 50, true);

      const result = quotaManager.checkQuota('budgettest');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('budget limit');
    });

    it('should block throttled providers', () => {
      quotaManager.recordRateLimitError('testprovider');

      const result = quotaManager.checkQuota('testprovider');
      expect(result.allowed).toBe(false);
      expect(result.reason).toContain('throttled');
    });
  });

  describe('recordRequest', () => {
    it('should track request counts', () => {
      quotaManager.recordRequest('openai', 100, 50, true);
      quotaManager.recordRequest('openai', 200, 100, true);

      const stats = quotaManager.getProviderStats('openai');
      expect(stats.usage.requestsThisMinute).toBe(2);
      expect(stats.usage.requestsThisHour).toBe(2);
      expect(stats.usage.tokensThisMinute).toBe(300);
    });

    it('should track costs', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 100,
        requestsPerHour: 1000,
        costPer1kTokens: 0.03,
      });

      quotaManager.recordRequest('testprovider', 1000, 50, true);
      quotaManager.recordRequest('testprovider', 2000, 50, true);

      const stats = quotaManager.getProviderStats('testprovider');
      expect(stats.usage.totalCost).toBe(0.09);
    });

    it('should decrement concurrent requests', () => {
      quotaManager.startRequest('openai');
      quotaManager.startRequest('openai');
      
      let stats = quotaManager.getProviderStats('openai');
      expect(stats.usage.concurrentRequests).toBe(2);

      quotaManager.recordRequest('openai', 100, 50, true);
      
      stats = quotaManager.getProviderStats('openai');
      expect(stats.usage.concurrentRequests).toBe(1);
    });
  });

  describe('recordRateLimitError', () => {
    it('should throttle provider', () => {
      quotaManager.recordRateLimitError('openai');

      const stats = quotaManager.getProviderStats('openai');
      expect(stats.health.isThrottled).toBe(true);
      expect(stats.health.throttleEndTime).toBeDefined();
    });

    it('should not throttle when autoThrottle disabled', () => {
      quotaManager.setConfig({ autoThrottle: false });
      quotaManager.recordRateLimitError('openai');

      const stats = quotaManager.getProviderStats('openai');
      expect(stats.health.isThrottled).toBe(false);
    });
  });

  describe('unthrottleProvider', () => {
    it('should unthrottle a throttled provider', () => {
      quotaManager.recordRateLimitError('openai');
      expect(quotaManager.getProviderStats('openai').health.isThrottled).toBe(true);

      quotaManager.unthrottleProvider('openai');
      expect(quotaManager.getProviderStats('openai').health.isThrottled).toBe(false);
    });
  });

  describe('setProviderLimits', () => {
    it('should set custom limits', () => {
      const customLimits: Partial<RateLimitConfig> = {
        requestsPerMinute: 30,
        requestsPerHour: 500,
        tokensPerMinute: 50000,
      };

      quotaManager.setProviderLimits('custom', customLimits);
      const limits = quotaManager.getLimits('custom');

      expect(limits.requestsPerMinute).toBe(30);
      expect(limits.requestsPerHour).toBe(500);
      expect(limits.tokensPerMinute).toBe(50000);
    });

    it('should use defaults for unknown providers', () => {
      const limits = quotaManager.getLimits('unknownprovider');
      expect(limits.requestsPerMinute).toBe(60);
    });
  });

  describe('getProviderStats', () => {
    it('should return stats with utilization percentages', () => {
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 10,
        requestsPerHour: 100,
        tokensPerMinute: 1000,
        tokensPerDay: 10000,
      });

      quotaManager.recordRequest('testprovider', 100, 50, true);
      quotaManager.recordRequest('testprovider', 100, 50, true);

      const stats = quotaManager.getProviderStats('testprovider');
      expect(stats.utilizationPercent.requestsPerMinute).toBe(20);
      expect(stats.utilizationPercent.requestsPerHour).toBe(2);
      expect(stats.utilizationPercent.tokensPerMinute).toBe(20);
      expect(stats.utilizationPercent.tokensPerDay).toBe(2);
    });

    it('should calculate success rate', () => {
      quotaManager.recordRequest('testprovider', 100, 50, true);
      quotaManager.recordRequest('testprovider', 100, 50, true);
      quotaManager.recordRequest('testprovider', 100, 50, false);

      const stats = quotaManager.getProviderStats('testprovider');
      expect(stats.health.successRate).toBeCloseTo(0.667, 2);
    });
  });

  describe('getAllStats', () => {
    it('should return stats for all tracked providers', () => {
      quotaManager.recordRequest('openai', 100, 50, true);
      quotaManager.recordRequest('anthropic', 200, 100, true);

      const allStats = quotaManager.getAllStats();
      expect(allStats.providers.length).toBe(2);
      expect(allStats.providers.map((p: { provider: string }) => p.provider)).toContain('openai');
      expect(allStats.providers.map((p: { provider: string }) => p.provider)).toContain('anthropic');
    });

    it('should include global budget info', () => {
      quotaManager.setConfig({ dailyBudgetLimit: 10 });
      quotaManager.setProviderLimits('testprovider', {
        requestsPerMinute: 100,
        requestsPerHour: 1000,
        costPer1kTokens: 0.01,
      });
      quotaManager.recordRequest('testprovider', 1000, 50, true);

      const allStats = quotaManager.getAllStats();
      expect(allStats.global.dailyBudgetLimit).toBe(10);
      expect(allStats.global.dailyCostTotal).toBe(0.01);
      expect(allStats.global.budgetUtilizationPercent).toBe(0.1);
    });
  });

  describe('getStatus', () => {
    it('should return status summary', () => {
      quotaManager.recordRequest('openai', 100, 50, true);
      quotaManager.recordRateLimitError('anthropic');

      const status = quotaManager.getStatus();
      expect(status.enabled).toBe(true);
      expect(status.providersTracked).toBe(2);
      expect(status.throttledProviders).toContain('anthropic');
    });
  });

  describe('config management', () => {
    it('should update config', () => {
      quotaManager.setConfig({
        dailyBudgetLimit: 100,
        alertThreshold: 0.9,
      });

      const config = quotaManager.getConfig();
      expect(config.dailyBudgetLimit).toBe(100);
      expect(config.alertThreshold).toBe(0.9);
    });

    it('should toggle enabled state', () => {
      quotaManager.setEnabled(false);
      expect(quotaManager.isEnabled()).toBe(false);

      quotaManager.setEnabled(true);
      expect(quotaManager.isEnabled()).toBe(true);
    });
  });

  describe('reset', () => {
    it('should reset provider usage', () => {
      quotaManager.recordRequest('openai', 100, 50, true);
      expect(quotaManager.getProviderStats('openai').usage.requestsThisMinute).toBe(1);

      quotaManager.resetProviderUsage('openai');
      expect(quotaManager.getProviderStats('openai').usage.requestsThisMinute).toBe(0);
    });

    it('should reset all usage', () => {
      quotaManager.recordRequest('openai', 100, 50, true);
      quotaManager.recordRequest('anthropic', 100, 50, true);

      quotaManager.resetAllUsage();

      const allStats = quotaManager.getAllStats();
      expect(allStats.providers.length).toBe(0);
    });
  });

  describe('events', () => {
    it('should emit request event', () => {
      return new Promise<void>((resolve) => {
        quotaManager.once('request', (data: { provider: string; tokensUsed: number }) => {
          expect(data.provider).toBe('openai');
          expect(data.tokensUsed).toBe(100);
          resolve();
        });

        quotaManager.recordRequest('openai', 100, 50, true);
      });
    });

    it('should emit throttled event', () => {
      return new Promise<void>((resolve) => {
        quotaManager.once('throttled', (data: { provider: string; durationSeconds: number }) => {
          expect(data.provider).toBe('openai');
          expect(data.durationSeconds).toBe(60);
          resolve();
        });

        quotaManager.recordRateLimitError('openai');
      });
    });

    it('should emit unthrottled event', () => {
      return new Promise<void>((resolve) => {
        quotaManager.recordRateLimitError('openai');

        quotaManager.once('unthrottled', (data: { provider: string }) => {
          expect(data.provider).toBe('openai');
          resolve();
        });

        quotaManager.unthrottleProvider('openai');
      });
    });
  });

  describe('default provider limits', () => {
    it('should have defaults for openai', () => {
      const limits = quotaManager.getLimits('openai');
      expect(limits.requestsPerMinute).toBe(60);
      expect(limits.requestsPerHour).toBe(3500);
    });

    it('should have defaults for anthropic', () => {
      const limits = quotaManager.getLimits('anthropic');
      expect(limits.requestsPerMinute).toBe(50);
      expect(limits.requestsPerHour).toBe(2000);
    });

    it('should have defaults for gemini', () => {
      const limits = quotaManager.getLimits('gemini');
      expect(limits.requestsPerMinute).toBe(60);
      expect(limits.requestsPerHour).toBe(1500);
    });

    it('should have defaults for deepseek', () => {
      const limits = quotaManager.getLimits('deepseek');
      expect(limits.requestsPerMinute).toBe(60);
    });
  });
});
