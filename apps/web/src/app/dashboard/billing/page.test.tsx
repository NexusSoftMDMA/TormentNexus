import { describe, expect, it } from 'vitest';

import {
  formatRoutingStrategyLabel,
  formatTaskRoutingLabel,
  getProviderPortalCards,
  getProviderQuickAccessSections,
  getRoutingStrategyBadgeClasses,
  PROVIDER_PORTALS,
  type BillingProviderQuotaSummary,
} from './billing-portal-data';

describe('billing dashboard provider portals', () => {
  it('builds provider portal cards with live auth state when quotas exist', () => {
    const quotas: BillingProviderQuotaSummary[] = [
      {
        provider: 'openai',
        name: 'OpenAI',
        configured: true,
        authenticated: true,
        authMethod: 'api_key',
        availability: 'healthy',
      },
      {
        provider: 'anthropic',
        name: 'Anthropic',
        configured: true,
        authenticated: false,
        authMethod: 'api_key',
        availability: 'cooldown',
        lastError: 'quota exhausted',
      },
    ];

    const cards = getProviderPortalCards(quotas);
    const openai = cards.find((card) => card.id === 'openai');
    const anthropic = cards.find((card) => card.id === 'anthropic');

    expect(openai).toMatchObject({
      statusLabel: 'Connected',
      statusTone: 'success',
      authLabel: 'api key',
      availabilityLabel: 'healthy',
    });

    expect(anthropic).toMatchObject({
      statusLabel: 'Configured',
      statusTone: 'warning',
      authLabel: 'api key',
      availabilityLabel: 'cooldown',
      errorLabel: 'quota exhausted',
    });
  });

  it('keeps reference links available even when TormentNexus has no local auth state', () => {
    const cards = getProviderPortalCards(undefined);
    const providerIds = new Set(cards.map((card) => card.id));

    expect(cards).toHaveLength(PROVIDER_PORTALS.length);
    expect(providerIds.has('github-copilot')).toBe(true);
    expect(providerIds.has('antigravity')).toBe(true);
    expect(providerIds.has('kiro')).toBe(true);
    expect(providerIds.has('kimi-coding')).toBe(true);
    expect(cards.find((card) => card.id === 'azure-openai')).toMatchObject({
      statusLabel: 'Not connected',
      statusTone: 'muted',
      authLabel: 'No auth detected',
      availabilityLabel: 'reference only',
    });
  });

  it('falls back safely when quota payload is malformed', () => {
    const cards = getProviderPortalCards('invalid-quotas' as unknown as BillingProviderQuotaSummary[]);

    expect(cards).toHaveLength(PROVIDER_PORTALS.length);
    expect(cards.find((card) => card.id === 'openai')).toMatchObject({
      statusLabel: 'Not connected',
      authLabel: 'No auth detected',
      availabilityLabel: 'reference only',
    });
  });

  it('ignores malformed quota rows while preserving valid provider states', () => {
    const cards = getProviderPortalCards([
      {
        provider: 'openai',
        name: 'OpenAI',
        configured: true,
        authenticated: true,
        authMethod: 'api_key',
      },
      null as unknown as BillingProviderQuotaSummary,
      { provider: 42 } as unknown as BillingProviderQuotaSummary,
    ]);

    expect(cards.find((card) => card.id === 'openai')).toMatchObject({
      statusLabel: 'Connected',
      statusTone: 'success',
    });
    expect(cards.find((card) => card.id === 'anthropic')).toMatchObject({
      statusLabel: 'Not connected',
      statusTone: 'muted',
    });
  });

  it('builds curated quick-access sections for keys, plans, and cloud consoles', () => {
    const sections = getProviderQuickAccessSections(undefined);
    const apiKeys = sections.find((section) => section.id === 'api-keys');
    const plansBilling = sections.find((section) => section.id === 'plans-billing');
    const cloudOauth = sections.find((section) => section.id === 'cloud-oauth');

    expect(sections).toHaveLength(3);
    expect(apiKeys?.links.map((link) => [link.providerId, link.actionLabel])).toEqual([
      ['openai', 'API keys'],
      ['anthropic', 'API keys'],
      ['gemini', 'API keys'],
      ['openrouter', 'API keys'],
      ['deepseek', 'API keys'],
      ['github-copilot', 'PATs'],
    ]);
    expect(plansBilling?.links.map((link) => [link.providerId, link.actionLabel])).toEqual([
      ['openai', 'Billing'],
      ['anthropic', 'Plans'],
      ['openrouter', 'Credits'],
      ['azure-openai', 'Cost analysis'],
      ['github-copilot', 'Billing'],
      ['kiro', 'Kiro'],
    ]);
    expect(cloudOauth?.links.map((link) => [link.providerId, link.actionLabel])).toEqual([
      ['gemini', 'AI Studio'],
      ['azure-openai', 'Azure portal'],
      ['github-copilot', 'Copilot settings'],
      ['antigravity', 'Code Assist'],
      ['kiro', 'AWS Builder ID'],
      ['kimi-coding', 'Coding'],
    ]);
    expect(apiKeys?.links.every((link) => link.statusLabel === 'Not connected')).toBe(true);
  });

  it('formats task routing labels and strategy tones for dashboard display', () => {
    expect(formatTaskRoutingLabel('supervisor')).toBe('Supervisor tasks');
    expect(formatTaskRoutingLabel('general')).toBe('General');
    expect(formatRoutingStrategyLabel('round-robin')).toBe('Round robin');
    expect(getRoutingStrategyBadgeClasses('best')).toContain('fuchsia');
    expect(getRoutingStrategyBadgeClasses('round-robin')).toContain('blue');
  });
});