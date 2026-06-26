"use client";
import Link from "next/link";
import { Star } from "lucide-react";
import { useGithubStats } from "@/lib/useGithubStats";

const GITHUB_URL = "https://github.com/VasiHemanth/tokentelemetry";

export default function Footer() {
  const { stars, forks } = useGithubStats();
  return (
    <footer className="border-t border-[var(--tt-border)] py-[30px]">
      <div className="max-w-[1180px] mx-auto px-5 flex items-center justify-between gap-4 flex-wrap">
        <div className="font-mono text-[11.5px] text-[var(--tt-fg-dim)]">
          tokentelemetry · built by{" "}
          <a href="https://www.linkedin.com/in/vasi-hemanth/" target="_blank" rel="noopener noreferrer"
            className="text-[var(--tt-fg-muted)] hover:text-[var(--tt-fg)] transition-colors">
            Hemanth Vasi
          </a>
          <span className="mx-2 text-[var(--tt-fg-faint)]">·</span>
          <Link href="/privacy" className="text-[var(--tt-fg-muted)] hover:text-[var(--tt-fg)] transition-colors">Privacy</Link>
          <span className="mx-2 text-[var(--tt-fg-faint)]">·</span>MIT
        </div>
        <a href={GITHUB_URL} target="_blank" rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 h-[34px] px-3 rounded-[var(--tt-radius)] border border-[var(--tt-border-strong)] bg-[var(--tt-panel)] text-[12.5px] font-medium text-[var(--tt-fg-muted)] hover:text-[var(--tt-fg)] transition-colors">
          <Star size={14} className="text-[var(--tt-warn)]" fill="currentColor" />
          <span className="text-[var(--tt-fg)] font-semibold">{stars}</span> stars · {forks} forks
        </a>
      </div>
    </footer>
  );
}
