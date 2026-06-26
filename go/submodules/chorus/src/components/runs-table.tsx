"use client";

import { Search, X } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { chatDisplayTitle } from "@/lib/chat-title";
import type { Chat } from "@/lib/types";

interface RunsTableProps {
  chats: Chat[];
}

/**
 * Client-side filter shell on top of the server-fetched chats list.
 *
 * Filters are local state today. URL-param persistence (shareable filtered
 * views) is a follow-up; for now Reload resets to "all".
 */
export function RunsTable({ chats }: RunsTableProps) {
  const [query, setQuery] = useState("");
  const [templateFilter, setTemplateFilter] = useState<string>("");
  const [statusFilter, setStatusFilter] = useState<string>("");

  const templateOptions = useMemo(() => {
    const set = new Set<string>();
    for (const c of chats) set.add(c.templateId);
    return Array.from(set).sort();
  }, [chats]);

  const statusOptions = useMemo(() => {
    const set = new Set<string>();
    for (const c of chats) set.add(c.status);
    return Array.from(set).sort();
  }, [chats]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return chats.filter((c) => {
      if (templateFilter && c.templateId !== templateFilter) return false;
      if (statusFilter && c.status !== statusFilter) return false;
      if (!q) return true;
      const title = chatDisplayTitle(c.work).toLowerCase();
      return (
        title.includes(q) ||
        c.work.toLowerCase().includes(q) ||
        c.id.toLowerCase().includes(q) ||
        (c.slug ?? "").toLowerCase().includes(q)
      );
    });
  }, [chats, query, templateFilter, statusFilter]);

  const activeFilters = Boolean(query || templateFilter || statusFilter);
  const clearAll = () => {
    setQuery("");
    setTemplateFilter("");
    setStatusFilter("");
  };

  return (
    <>
      <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search by title, work, chat id, or slug…"
            className="w-full rounded-md border border-border bg-card py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary/40 focus:outline-none"
          />
        </div>
        <select
          value={templateFilter}
          onChange={(e) => setTemplateFilter(e.target.value)}
          className="rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary/40 focus:outline-none"
          aria-label="Filter by template"
        >
          <option value="">All templates</option>
          {templateOptions.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary/40 focus:outline-none"
          aria-label="Filter by status"
        >
          <option value="">All statuses</option>
          {statusOptions.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
        {activeFilters && (
          <button
            type="button"
            onClick={clearAll}
            className="flex items-center gap-1 rounded-md border border-border bg-card px-3 py-2 text-xs text-muted-foreground transition hover:border-destructive/40 hover:text-destructive"
          >
            <X className="h-3 w-3" />
            Clear
          </button>
        )}
      </div>

      <p className="mb-3 text-xs text-muted-foreground">
        {filtered.length === chats.length
          ? `${chats.length} chat${chats.length === 1 ? "" : "s"}`
          : `${filtered.length} of ${chats.length} chat${chats.length === 1 ? "" : "s"}`}
      </p>

      <div className="space-y-2">
        {filtered.length === 0 ? (
          <div className="rounded-lg border border-border bg-card/30 p-8 text-center">
            <p className="text-sm text-muted-foreground">
              {activeFilters ? "No chats match these filters." : "No chats yet. Start a new one!"}
            </p>
          </div>
        ) : (
          filtered.map((chat) => (
            <Link
              key={chat.id}
              href={`/runs/${chat.slug || chat.id}`}
              className="group flex items-start justify-between gap-4 rounded-lg border border-border bg-card p-4 transition hover:border-muted-foreground/30 hover:bg-card/80"
            >
              <div className="min-w-0 flex-1">
                <div className="mb-1 flex items-center gap-2">
                  <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    {chat.status}
                  </span>
                  <Badge
                    variant="outline"
                    className="border-border font-mono text-[10px]"
                  >
                    {chat.templateId}
                  </Badge>
                </div>
                <h3 className="line-clamp-1 text-sm font-medium text-foreground">
                  {chatDisplayTitle(chat.work)}
                </h3>
                <p className="mt-1 text-xs text-muted-foreground">
                  {new Date(chat.createdAt).toLocaleString()}
                </p>
              </div>
            </Link>
          ))
        )}
      </div>
    </>
  );
}
