"use client";

import { createContext, useContext } from "react";

export interface PlanSnippet {
  session_id: string;
  agent: string;
  timestamp: string;
  content: string;
}

export interface SessionRow {
  id: string;
  agent: string;
  project: string;
  timestamp: string;
  display?: string;
  text?: string;
  mcp_tools: string[];
  subagents: string[];
  has_plan: boolean;
  copilot_source?: string;
  antigravity_source?: string;
  tokens?: { input: number; output: number; cached: number; total: number };
  cost?: number;
  /* Delegation & ecosystem telemetry (see DESIGN.md) */
  delegation?: {
    supported: boolean;
    tokens_recorded?: boolean;
    spawn_count?: number;
    delegated_total?: number;
    linked_children?: number;
    by_type?: Record<string, { count: number; total?: number; cost?: number; child_session_ids?: string[] }>;
  };
  delegated_cost?: number;
  parent_session_id?: string | null;
  child_session_ids?: string[];
  subagent_info?: { role?: string; nickname?: string; depth?: number };
  skills_used?: { name: string; count: number }[];
  mcp_usage?: Record<string, Record<string, number>>;
}

export interface WorktreeSummary {
  name: string;
  path: string;
  session_count: number;
  tokens: { input: number; output: number; cached: number; total: number; cost: number };
  agents: string[];
  status: string;
}

export interface ProjectData {
  name: string;
  path: string;
  session_count: number;
  agents: string[];
  mcp_tools: string[];
  subagent_count: number;
  configured_subagent_count?: number;
  plan_count: number;
  plans: PlanSnippet[];
  tokens?: { input: number; output: number; cached: number; total: number };
  // Git-worktree grouping (added by the backend)
  canonical_repo?: string;
  is_worktree?: boolean;
  worktree_name?: string;
  parent_path?: string;
  parent_name?: string;
  is_repo_root?: boolean;
  synthesized?: boolean;
  worktrees?: WorktreeSummary[];
  aggregate?: {
    session_count: number;
    subagent_count: number;
    plan_count: number;
    configured_subagent_count: number;
    tokens: { input: number; output: number; cached: number; total: number; cost: number };
    agents: string[];
    mcp_tools: string[];
    worktree_count: number;
  };
}

interface ProjectCtx {
  decodedPath: string;
  projectName: string;
  project: ProjectData | undefined;
  sessions: SessionRow[];
  loading: boolean;
}

const Ctx = createContext<ProjectCtx | null>(null);

export function ProjectProvider({ value, children }: { value: ProjectCtx; children: React.ReactNode }) {
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useProject(): ProjectCtx {
  const v = useContext(Ctx);
  if (!v) throw new Error("useProject must be used inside <ProjectProvider>");
  return v;
}
