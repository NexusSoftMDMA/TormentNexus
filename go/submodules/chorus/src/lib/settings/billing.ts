/**
 * Billing-mode setting. Captures whether the user pays for the CLIs
 * chorus drives via per-token API keys, monthly subscriptions, or a mix.
 *
 * Why this matters: chorus's cost estimate on the /new page used to assume
 * everyone paid spot API rates (Opus ~$15/M input, GPT-5.5 ~$3/M, etc.),
 * which produced fictional dollar amounts for users on Claude Pro
 * ($20/mo flat) or ChatGPT Plus. Subscription users should see "Subscription
 * quota" not "$0.47" — the latter implies they're charged when they aren't.
 *
 * v0.7 keeps this as a single global enum. v0.8 will split into per-lineage
 * overrides (Claude on subscription, Gemini on API) once we have the
 * onboarding UI for it.
 *
 * Default is 'subscription' — the path-of-least-cost-shock for new users.
 * Most people land here from a Claude Pro / ChatGPT Plus / Gemini Advanced
 * subscription they already pay for; showing "$0.47 estimate" on every
 * /new chat would imply per-call billing they don't owe and scare them
 * off. Subscription users see "Subscription quota — ~12,000 input tokens"
 * instead. API-key users flip the toggle once and never look back.
 */

import { settings } from '../db';

export type BillingMode = 'api' | 'subscription' | 'mixed';

const BILLING_MODE_KEY = 'billing_mode';
export const DEFAULT_BILLING_MODE: BillingMode = 'subscription';

export const BILLING_MODE_LABELS: Record<BillingMode, { label: string; description: string }> = {
  api: {
    label: 'Pay-per-token API keys',
    description:
      'You bring API keys (Anthropic, OpenAI, Google) and pay per request. Cost estimates are accurate.',
  },
  subscription: {
    label: 'CLI subscriptions',
    description:
      'You run Claude Code, Codex, Gemini CLI etc. through their monthly subscriptions (Claude Pro, ChatGPT Plus, Gemini Advanced). Chorus calls them through the CLI; no per-call cost. Cost shown as "Subscription quota."',
  },
  mixed: {
    label: 'Mixed — some subscription, some API',
    description:
      'A subset of CLIs are on subscriptions and a subset on API keys. Estimates show worst-case API cost; real bill may be lower.',
  },
};

function isBillingMode(value: unknown): value is BillingMode {
  return value === 'api' || value === 'subscription' || value === 'mixed';
}

export async function getBillingMode(): Promise<BillingMode> {
  const raw = await settings.get(BILLING_MODE_KEY);
  return isBillingMode(raw) ? raw : DEFAULT_BILLING_MODE;
}

export async function setBillingMode(value: BillingMode): Promise<BillingMode> {
  if (!isBillingMode(value)) {
    throw new Error(`invalid billing mode: ${value}`);
  }
  await settings.set(BILLING_MODE_KEY, value);
  return getBillingMode();
}
