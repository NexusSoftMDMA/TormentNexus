import { GLOBAL_SECRET_PATH_PATTERNS } from '../config/schema.js';
import { normalizePathSeparators, toVirtualRelative } from './mount-registry.js';

export function matchesPathPatterns(relativePath: string, patterns: readonly string[]): boolean {
  const normalized = toVirtualRelative(relativePath);
  return patterns.some((pattern) => {
    const normalizedPattern = toVirtualRelative(pattern);
    if (normalizedPattern.endsWith('/**')) {
      const prefix = normalizedPattern.slice(0, -3);
      return normalized === prefix || normalized.startsWith(`${prefix}/`);
    }
    return globToRegExp(pattern).test(normalized);
  });
}

export function matchesGlobalSecretPattern(relativePath: string): boolean {
  return matchesPathPatterns(relativePath, GLOBAL_SECRET_PATH_PATTERNS);
}

export function matchesConfiguredOrGlobalPattern(relativePath: string, patterns: readonly string[]): boolean {
  return matchesGlobalSecretPattern(relativePath) || matchesPathPatterns(relativePath, patterns);
}

export function isGloballyDeniedPath(relativePath: string, realPath: string): boolean {
  return matchesGlobalSecretPattern(relativePath) || realPathHasSensitiveSegment(realPath);
}

const GLOBALLY_DENIED_SEGMENTS = new Set(['.mvmt', '.ssh', '.gnupg', '.aws', '.kube', '.docker']);
const GLOBALLY_DENIED_SEGMENT_PATHS = [
  ['.config', 'gh'],
  ['.config', 'gcloud'],
  ['.config', 'azure'],
];

function realPathHasSensitiveSegment(realPath: string): boolean {
  const segments = normalizePathSeparators(realPath).split('/').filter(Boolean);
  if (segments.some((segment) => GLOBALLY_DENIED_SEGMENTS.has(segment))) return true;
  return GLOBALLY_DENIED_SEGMENT_PATHS.some((denied) => containsSegmentSequence(segments, denied));
}

function containsSegmentSequence(segments: string[], needle: string[]): boolean {
  for (let start = 0; start <= segments.length - needle.length; start += 1) {
    if (needle.every((segment, offset) => segments[start + offset] === segment)) return true;
  }
  return false;
}

function globToRegExp(pattern: string): RegExp {
  const normalized = toVirtualRelative(pattern);
  const escaped = escapeRegExp(normalized)
    .replace(/\\\*\\\*/g, '.*')
    .replace(/\\\*/g, '[^/]*');
  return new RegExp(`^${escaped}$`);
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
