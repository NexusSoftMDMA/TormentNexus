// Hermes session source pill — replaces the project field for Hermes sessions
// in the sessions list, and renders in the detail-page cross-cutting header.
//
// Taxonomy verified against ~/.hermes/hermes-agent/ source code:
// 38 distinct source values across 6 categories. See HERMES_INTERNALS.md §1.1.

import {
  Terminal, Globe, Server, Send, MessageSquare, Hash, MessageCircle,
  Shield, Triangle, Clock, Webhook, HelpCircle, Mail, Smartphone,
  Home, Sparkles, Bell, Briefcase, Bot, Coins, Lock, Users,
  MessagesSquare, Anchor, Network,
  type LucideIcon,
} from "lucide-react";

export type HermesSource =
  // Interactive
  | "cli" | "local" | "tui" | "webui"
  // Service / API
  | "api_server" | "webhook" | "msgraph_webhook" | "gateway"
  // Chat — Western
  | "telegram" | "discord" | "slack" | "whatsapp" | "signal" | "matrix"
  | "mattermost" | "google_chat" | "irc" | "line" | "simplex" | "teams"
  | "bluebubbles"
  // Chat — Chinese
  | "feishu" | "dingtalk" | "weixin" | "wecom" | "wecom_callback"
  | "qqbot" | "yuanbao"
  // Notification-only
  | "email" | "sms" | "homeassistant"
  // Autonomous
  | "cron"
  // Fallback
  | "unknown";

interface SourceMeta {
  label: string;
  icon: LucideIcon;
  /** Tailwind classes for text + background + border. */
  cls: string;
}

const SOURCES: Record<HermesSource, SourceMeta> = {
  // ── Interactive ────────────────────────────────────────────────────────
  cli:             { label: "CLI",      icon: Terminal,       cls: "text-amber-300 bg-amber-500/10 border-amber-500/30" },
  local:           { label: "LOCAL",    icon: Terminal,       cls: "text-amber-300 bg-amber-500/10 border-amber-500/30" },
  tui:             { label: "TUI",      icon: Terminal,       cls: "text-amber-300 bg-amber-500/10 border-amber-500/30" },
  webui:           { label: "WEBUI",    icon: Globe,          cls: "text-violet-300 bg-violet-500/10 border-violet-500/30" },

  // ── Service / API ──────────────────────────────────────────────────────
  api_server:      { label: "API",      icon: Server,         cls: "text-violet-300 bg-violet-500/10 border-violet-500/30" },
  webhook:         { label: "WEBHOOK",  icon: Webhook,        cls: "text-orange-300 bg-orange-500/10 border-orange-500/30" },
  msgraph_webhook: { label: "MS GRAPH", icon: Webhook,        cls: "text-orange-300 bg-orange-500/10 border-orange-500/30" },
  gateway:         { label: "GATEWAY",  icon: Network,        cls: "text-slate-300 bg-slate-500/10 border-slate-500/30" },

  // ── Chat — Western ─────────────────────────────────────────────────────
  telegram:        { label: "TELEGRAM",  icon: Send,           cls: "text-sky-300 bg-sky-500/10 border-sky-500/30" },
  discord:         { label: "DISCORD",   icon: MessageSquare,  cls: "text-indigo-300 bg-indigo-500/10 border-indigo-500/30" },
  slack:           { label: "SLACK",     icon: Hash,           cls: "text-pink-300 bg-pink-500/10 border-pink-500/30" },
  whatsapp:        { label: "WHATSAPP",  icon: MessageCircle,  cls: "text-emerald-300 bg-emerald-500/10 border-emerald-500/30" },
  signal:          { label: "SIGNAL",    icon: Shield,         cls: "text-blue-300 bg-blue-500/10 border-blue-500/30" },
  matrix:          { label: "MATRIX",    icon: Triangle,       cls: "text-green-300 bg-green-500/10 border-green-500/30" },
  mattermost:      { label: "MATTERMOST",icon: Anchor,         cls: "text-blue-300 bg-blue-500/10 border-blue-500/30" },
  google_chat:     { label: "GCHAT",     icon: MessageSquare,  cls: "text-red-300 bg-red-500/10 border-red-500/30" },
  irc:             { label: "IRC",       icon: Hash,           cls: "text-zinc-300 bg-zinc-500/10 border-zinc-500/30" },
  line:            { label: "LINE",      icon: MessageSquare,  cls: "text-green-300 bg-green-500/10 border-green-500/30" },
  simplex:         { label: "SIMPLEX",   icon: Lock,           cls: "text-cyan-300 bg-cyan-500/10 border-cyan-500/30" },
  teams:           { label: "TEAMS",     icon: Users,          cls: "text-violet-300 bg-violet-500/10 border-violet-500/30" },
  bluebubbles:     { label: "BLUEBUBBLES",icon: MessageCircle, cls: "text-blue-300 bg-blue-500/10 border-blue-500/30" },

  // ── Chat — Chinese ─────────────────────────────────────────────────────
  feishu:          { label: "FEISHU",    icon: Sparkles,       cls: "text-cyan-300 bg-cyan-500/10 border-cyan-500/30" },
  dingtalk:        { label: "DINGTALK",  icon: Bell,           cls: "text-sky-300 bg-sky-500/10 border-sky-500/30" },
  weixin:          { label: "WEIXIN",    icon: MessagesSquare, cls: "text-green-300 bg-green-500/10 border-green-500/30" },
  wecom:           { label: "WECOM",     icon: Briefcase,      cls: "text-emerald-300 bg-emerald-500/10 border-emerald-500/30" },
  wecom_callback:  { label: "WECOM CB",  icon: Briefcase,      cls: "text-emerald-300 bg-emerald-500/10 border-emerald-500/30" },
  qqbot:           { label: "QQ",        icon: Bot,            cls: "text-red-300 bg-red-500/10 border-red-500/30" },
  yuanbao:         { label: "YUANBAO",   icon: Coins,          cls: "text-amber-300 bg-amber-500/10 border-amber-500/30" },

  // ── Notification only ──────────────────────────────────────────────────
  email:           { label: "EMAIL",     icon: Mail,           cls: "text-blue-300 bg-blue-500/10 border-blue-500/30" },
  sms:             { label: "SMS",       icon: Smartphone,     cls: "text-teal-300 bg-teal-500/10 border-teal-500/30" },
  homeassistant:   { label: "HOME ASSIST",icon: Home,          cls: "text-yellow-300 bg-yellow-500/10 border-yellow-500/30" },

  // ── Autonomous ─────────────────────────────────────────────────────────
  cron:            { label: "CRON",      icon: Clock,          cls: "text-orange-300 bg-orange-500/10 border-orange-500/30" },

  // ── Fallback ───────────────────────────────────────────────────────────
  unknown:         { label: "UNKNOWN",   icon: HelpCircle,     cls: "text-[var(--tt-fg-muted)] bg-[var(--tt-panel)] border-[var(--tt-border)]" },
};

const FALLBACK: SourceMeta = SOURCES.unknown;

export default function SourceBadge({
  source,
  size = "sm",
}: {
  source: string | null | undefined;
  size?: "xs" | "sm" | "md";
}) {
  const meta = (source && SOURCES[source as HermesSource]) || FALLBACK;
  const Icon = meta.icon;
  const sizing =
    size === "xs" ? "text-[9px] px-1.5 py-[1px] gap-1" :
    size === "md" ? "text-[11px] px-2.5 py-1 gap-1.5" :
    "text-[10px] px-2 py-[2px] gap-1";
  const iconSize = size === "xs" ? 9 : size === "md" ? 13 : 11;
  return (
    <span
      className={`inline-flex items-center font-mono uppercase tracking-wider rounded border ${sizing} ${meta.cls}`}
      title={`Hermes session source: ${meta.label.toLowerCase()}`}
    >
      <Icon size={iconSize} />
      {meta.label}
    </span>
  );
}
