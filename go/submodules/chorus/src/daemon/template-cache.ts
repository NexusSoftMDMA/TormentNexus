/**
 * Parsed-template cache. Every SSE attach used to re-yaml.parse +
 * zod.parse the template row, which is hot when 5+ tabs watch the same
 * run on a long chat. Keyed by (templateId, updated_at) so an upsert
 * through POST /templates naturally invalidates without an explicit
 * bust call.
 */

import yaml from 'yaml';
import { TemplateSchema } from '../lib/template-schema.js';

type ParsedTemplate = ReturnType<typeof TemplateSchema.parse>;

// Soft cap so the cache can't grow unbounded under a runaway template
// upsert workload. 50 is well above the realistic working set (10
// builtins + a handful of user clones); trim oldest-first via Map
// insertion order.
const TEMPLATE_CACHE_MAX = 50;
const cache = new Map<string, { stamp: number; parsed: ParsedTemplate }>();

export function getParsedTemplate(
  templateId: string,
  yamlText: string,
  stamp: number,
): ParsedTemplate {
  const hit = cache.get(templateId);
  if (hit && hit.stamp === stamp) return hit.parsed;
  const parsed = TemplateSchema.parse(yaml.parse(yamlText));
  cache.set(templateId, { stamp, parsed });
  // Map iteration order is insertion order; first key is oldest.
  while (cache.size > TEMPLATE_CACHE_MAX) {
    const oldest = cache.keys().next().value;
    if (oldest === undefined) break;
    cache.delete(oldest);
  }
  return parsed;
}
