export type NormalizedSubmoduleStatus = 'clean' | 'dirty' | 'missing' | 'error';

export interface NormalizedSubmoduleInfo {
  path: string;
  url: string;
  status: NormalizedSubmoduleStatus;
  lastCommit?: string;
  lastCommitDate?: string;
  lastCommitMessage?: string;
  version?: string;
  pkgName?: string;
}

export interface NormalizedLinkCategory {
  name: string;
  links: string[];
}

export interface SubmoduleSummaryCounts {
  clean: number;
  dirty: number;
  missing: number;
  error: number;
  resources: number;
}

const asRecord = (value: unknown): Record<string, unknown> => (
  value && typeof value === 'object' ? (value as Record<string, unknown>) : {}
);

const asTrimmedString = (value: unknown, fallback: string): string => {
  if (typeof value !== 'string') return fallback;
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : fallback;
};

const asOptionalTrimmedString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') return undefined;
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
};

const normalizeStatus = (value: unknown): NormalizedSubmoduleStatus => {
  switch (value) {
    case 'clean':
    case 'dirty':
    case 'missing':
    case 'error':
      return value;
    default:
      return 'error';
  }
};

export const normalizeSubmodules = (payload: unknown): NormalizedSubmoduleInfo[] => {
  if (!Array.isArray(payload)) return [];

  return payload.map((row, index) => {
    const sub = asRecord(row);
    return {
      path: asTrimmedString(sub.path, `unknown/submodule-${index + 1}`),
      url: asTrimmedString(sub.url, 'unknown-url'),
      status: normalizeStatus(sub.status),
      lastCommit: asOptionalTrimmedString(sub.lastCommit),
      lastCommitDate: asOptionalTrimmedString(sub.lastCommitDate),
      lastCommitMessage: asOptionalTrimmedString(sub.lastCommitMessage),
      version: asOptionalTrimmedString(sub.version),
      pkgName: asOptionalTrimmedString(sub.pkgName),
    };
  });
};

export const normalizeUserLinks = (payload: unknown): NormalizedLinkCategory[] => {
  if (!Array.isArray(payload)) return [];

  return payload.map((row, index) => {
    const category = asRecord(row);
    const links = Array.isArray(category.links)
      ? category.links
          .filter((link): link is string => typeof link === 'string')
          .map((link) => link.trim())
          .filter((link) => link.length > 0)
      : [];

    return {
      name: asTrimmedString(category.name, `Category ${index + 1}`),
      links,
    };
  });
};

export const summarizeSubmoduleCounts = (
  submodules: NormalizedSubmoduleInfo[],
  userLinks: NormalizedLinkCategory[],
): SubmoduleSummaryCounts => {
  return {
    clean: submodules.filter((s) => s.status === 'clean').length,
    dirty: submodules.filter((s) => s.status === 'dirty').length,
    missing: submodules.filter((s) => s.status === 'missing').length,
    error: submodules.filter((s) => s.status === 'error').length,
    resources: userLinks.reduce((acc, cat) => acc + cat.links.length, 0),
  };
};
