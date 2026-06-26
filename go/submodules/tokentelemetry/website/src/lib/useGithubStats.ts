"use client";
import { useEffect, useState } from "react";

const REPO = "VasiHemanth/tokentelemetry";
const CACHE_KEY = "tt-gh-stats";

// Fallback shown before the live fetch resolves (and if GitHub is unreachable /
// rate-limited). Keep roughly current so the first paint isn't wildly off.
export const FALLBACK_STATS = { stars: 108, forks: 17 };

export type GithubStats = { stars: number; forks: number };

/**
 * Live star/fork counts from the public GitHub API. Client-side (this is the
 * marketing site, not the local app, so an outbound call is fine) and cached in
 * sessionStorage so we don't refetch on every navigation or burn the
 * unauthenticated rate limit (60/hr/IP). Always returns usable numbers.
 */
export function useGithubStats(): GithubStats {
  const [stats, setStats] = useState<GithubStats>(FALLBACK_STATS);

  useEffect(() => {
    try {
      const cached = sessionStorage.getItem(CACHE_KEY);
      if (cached) setStats(JSON.parse(cached));
    } catch {
      /* ignore */
    }
    fetch(`https://api.github.com/repos/${REPO}`, { headers: { Accept: "application/vnd.github+json" } })
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => {
        if (d && typeof d.stargazers_count === "number") {
          const next = { stars: d.stargazers_count, forks: d.forks_count ?? FALLBACK_STATS.forks };
          setStats(next);
          try { sessionStorage.setItem(CACHE_KEY, JSON.stringify(next)); } catch { /* ignore */ }
        }
      })
      .catch(() => {});
  }, []);

  return stats;
}
