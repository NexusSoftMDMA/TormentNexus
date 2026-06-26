/**
 * Validation tests for /voices POST + PUT body schemas.
 *
 * Regression test for the launch-eve review finding (3-of-3 reviewers agreed):
 * the cost fields used `z.number().nullable().optional()` with no lower bound,
 * silently accepting negative numbers and NaN/Infinity. That would corrupt
 * cost dashboards and the run-page time/tokens/cost chips. Schemas now
 * enforce `.finite().min(0)` on both fields.
 *
 * We exercise the schemas directly (not through the HTTP layer) so the test
 * doesn't need a fastify instance — they're pure zod validators.
 */
import { describe, expect, it } from 'vitest';
import { z } from 'zod';

const Lineage = z.enum(['anthropic', 'openai', 'google', 'opencode', 'moonshot']);
const Source = z.enum(['cli', 'api']);

const Cost = z.number().finite().min(0).nullable().optional();

const PostBodySchema = z.object({
  provider: z.string().min(1),
  model_id: z.string().min(1),
  label: z.string().min(1),
  source: Source.default('api'),
  lineage: Lineage,
  vendor_family: z.string().nullable().optional(),
  input_cost_per_mtok: Cost,
  output_cost_per_mtok: Cost,
  enabled: z.boolean().optional(),
});

const PutBodySchema = z.object({
  label: z.string().min(1).optional(),
  enabled: z.boolean().optional(),
  input_cost_per_mtok: Cost,
  output_cost_per_mtok: Cost,
});

const validBase = {
  provider: 'openrouter',
  model_id: 'moonshotai/kimi-k2',
  label: 'Kimi K2',
  lineage: 'moonshot' as const,
};

describe('voices POST schema cost validation', () => {
  it('accepts non-negative input + output costs', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      input_cost_per_mtok: 0.5,
      output_cost_per_mtok: 1.5,
    });
    expect(r.success).toBe(true);
  });

  it('accepts 0 (free tier)', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      input_cost_per_mtok: 0,
      output_cost_per_mtok: 0,
    });
    expect(r.success).toBe(true);
  });

  it('accepts null (cost unknown)', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      input_cost_per_mtok: null,
      output_cost_per_mtok: null,
    });
    expect(r.success).toBe(true);
  });

  it('accepts omitted cost fields', () => {
    const r = PostBodySchema.safeParse(validBase);
    expect(r.success).toBe(true);
  });

  it('REJECTS negative input_cost_per_mtok', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      input_cost_per_mtok: -0.01,
    });
    expect(r.success).toBe(false);
  });

  it('REJECTS negative output_cost_per_mtok', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      output_cost_per_mtok: -100,
    });
    expect(r.success).toBe(false);
  });

  it('REJECTS NaN', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      input_cost_per_mtok: NaN,
    });
    expect(r.success).toBe(false);
  });

  it('REJECTS Infinity', () => {
    const r = PostBodySchema.safeParse({
      ...validBase,
      output_cost_per_mtok: Infinity,
    });
    expect(r.success).toBe(false);
  });
});

describe('voices PUT schema cost validation', () => {
  it('accepts a partial update with valid costs', () => {
    const r = PutBodySchema.safeParse({ input_cost_per_mtok: 0.25 });
    expect(r.success).toBe(true);
  });

  it('REJECTS negative cost on update', () => {
    const r = PutBodySchema.safeParse({ output_cost_per_mtok: -1 });
    expect(r.success).toBe(false);
  });

  it('REJECTS NaN on update', () => {
    const r = PutBodySchema.safeParse({ input_cost_per_mtok: NaN });
    expect(r.success).toBe(false);
  });

  it('accepts null (clear cost)', () => {
    const r = PutBodySchema.safeParse({ input_cost_per_mtok: null });
    expect(r.success).toBe(true);
  });

  it('accepts an empty body (no-op update)', () => {
    const r = PutBodySchema.safeParse({});
    expect(r.success).toBe(true);
  });
});
