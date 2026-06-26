'use client';

import { useState } from 'react';
import {
  MessageSquare,
  Play,
  Trash2,
  Plus,
  RefreshCw,
  Clock,
  Search,
  ChevronDown,
  ChevronRight,
  User,
  Bot,
  X,
  Hand,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import { Input } from './ui/input';
import { Textarea } from './ui/textarea';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import {
  useSessions,
  type Session,
  type Handoff,
  type Message,
} from '../lib/hooks/use-sessions';

function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch {
    return timestamp;
  }
}

function formatRelativeTime(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  } catch {
    return timestamp;
  }
}

function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';

  return (
    <div className={`flex gap-2 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div className={`flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center ${isUser ? 'bg-purple-600' : isSystem ? 'bg-yellow-600' : 'bg-zinc-700'
        }`}>
        {isUser ? (
          <User className="h-3 w-3 text-white" />
        ) : isSystem ? (
          <AlertCircle className="h-3 w-3 text-white" />
        ) : (
          <Bot className="h-3 w-3 text-white" />
        )}
      </div>
      <div className={`flex-1 max-w-[80%] ${isUser ? 'text-right' : ''}`}>
        <div className={`inline-block px-3 py-2 rounded-lg text-sm ${isUser
            ? 'bg-purple-600 text-white'
            : isSystem
              ? 'bg-yellow-600/20 text-yellow-200 border border-yellow-600/30'
              : 'bg-zinc-800 text-white/90'
          }`}>
          <p className="whitespace-pre-wrap break-words">{message.content}</p>
        </div>
        <div className="text-[10px] text-white/30 mt-1 font-mono">
          {formatTimestamp(message.timestamp)}
        </div>
      </div>
    </div>
  );
}

function SessionCard({
  session,
  isExpanded,
  onToggle,
  onResume,
  onDelete,
}: {
  session: Session;
  isExpanded: boolean;
  onToggle: () => void;
  onResume: () => void;
  onDelete: () => void;
}) {
  const statusColor = session.status === 'active'
    ? 'bg-green-500'
    : session.status === 'paused'
      ? 'bg-yellow-500'
      : 'bg-gray-500';

  return (
    <Card className="bg-zinc-900/50 border-white/10">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div
            className="flex items-center gap-2 cursor-pointer flex-1"
            onClick={onToggle}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => e.key === 'Enter' && onToggle()}
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 text-white/40" />
            ) : (
              <ChevronRight className="h-4 w-4 text-white/40" />
            )}
            <span className={`h-2 w-2 rounded-full ${statusColor}`} />
            <CardTitle className="text-white text-sm">{session.agentName}</CardTitle>
          </div>
          <div className="flex items-center gap-2">
            <Badge className="bg-white/10 text-white/60 border-0 font-mono text-[10px]">
              {session.messages.length} msgs
            </Badge>
          </div>
        </div>
        <CardDescription className="text-white/40 font-mono text-xs pl-8">
          {session.id.slice(0, 12)}... | {formatRelativeTime(session.timestamp)}
        </CardDescription>
      </CardHeader>

      {isExpanded && (
        <CardContent className="space-y-3">
          <ScrollArea className="h-[200px] pr-4">
            <div className="space-y-3">
              {session.messages.length === 0 ? (
                <div className="text-center text-white/40 py-4 text-xs">
                  No messages in this session
                </div>
              ) : (
                session.messages.map((msg) => (
                  <MessageBubble key={msg.id} message={msg} />
                ))
              )}
            </div>
          </ScrollArea>

          <div className="flex gap-2 pt-2 border-t border-white/5">
            <Button
              size="sm"
              onClick={onResume}
              className="flex-1 bg-purple-600 hover:bg-purple-500 text-white text-xs"
            >
              <Play className="h-3 w-3 mr-1" />
              Resume
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={onDelete}
              className="bg-transparent border-white/10 hover:bg-red-600/20 hover:text-red-400 text-xs"
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        </CardContent>
      )}
    </Card>
  );
}

function HandoffCard({
  handoff,
  isExpanded,
  onToggle,
  onClaim,
}: {
  handoff: Handoff;
  isExpanded: boolean;
  onToggle: () => void;
  onClaim: () => void;
}) {
  const isPending = handoff.status === 'pending';

  return (
    <Card className="bg-zinc-900/50 border-white/10">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div
            className="flex items-center gap-2 cursor-pointer flex-1"
            onClick={onToggle}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => e.key === 'Enter' && onToggle()}
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 text-white/40" />
            ) : (
              <ChevronRight className="h-4 w-4 text-white/40" />
            )}
            <Hand className={`h-4 w-4 ${isPending ? 'text-yellow-400' : 'text-green-400'}`} />
            <CardTitle className="text-white text-sm truncate max-w-[200px]">
              {handoff.description}
            </CardTitle>
          </div>
          <Badge className={`${isPending
              ? 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30'
              : 'bg-green-500/20 text-green-400 border-green-500/30'
            } border text-[10px]`}>
            {handoff.status}
          </Badge>
        </div>
        <CardDescription className="text-white/40 font-mono text-xs pl-8">
          {handoff.id.slice(0, 12)}... | {formatRelativeTime(handoff.timestamp)}
        </CardDescription>
      </CardHeader>

      {isExpanded && (
        <CardContent className="space-y-3">
          <div className="p-3 bg-black/40 rounded-lg border border-white/5">
            <div className="text-xs text-white/40 mb-1 uppercase tracking-wider">Context</div>
            <ScrollArea className="h-[100px]">
              <pre className="text-xs text-white/80 whitespace-pre-wrap font-mono">
                {handoff.context || 'No context provided'}
              </pre>
            </ScrollArea>
          </div>

          {handoff.claimedBy && (
            <div className="text-xs text-white/40">
              Claimed by <span className="text-green-400">{handoff.claimedBy}</span>
              {handoff.claimedAt && ` at ${formatTimestamp(handoff.claimedAt)}`}
            </div>
          )}

          {isPending && (
            <Button
              size="sm"
              onClick={onClaim}
              className="w-full bg-purple-600 hover:bg-purple-500 text-white text-xs"
            >
              <CheckCircle2 className="h-3 w-3 mr-1" />
              Claim Handoff
            </Button>
          )}
        </CardContent>
      )}
    </Card>
  );
}

function CreateHandoffDialog({
  isOpen,
  onClose,
  onSubmit,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (description: string, context: string) => void;
}) {
  const [description, setDescription] = useState('');
  const [context, setContext] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async () => {
    if (!description.trim()) return;
    setIsSubmitting(true);
    try {
      await onSubmit(description.trim(), context.trim());
      setDescription('');
      setContext('');
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <Card className="w-full max-w-md bg-zinc-900 border-white/10">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-white">Create Handoff</CardTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="h-6 w-6 p-0 text-white/40 hover:text-white"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
          <CardDescription className="text-white/40">
            Create a new handoff request for another agent or user
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label className="text-xs text-white/60 block mb-1">Description</label>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Brief description of the handoff..."
              className="bg-black/40 border-white/10 text-white"
            />
          </div>
          <div>
            <label className="text-xs text-white/60 block mb-1">Context</label>
            <Textarea
              value={context}
              onChange={(e) => setContext(e.target.value)}
              placeholder="Additional context, state, or information..."
              className="bg-black/40 border-white/10 text-white min-h-[120px] resize-none"
            />
          </div>
          <div className="flex gap-2 pt-2">
            <Button
              variant="outline"
              onClick={onClose}
              className="flex-1 bg-transparent border-white/10 text-white/60"
            >
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={!description.trim() || isSubmitting}
              className="flex-1 bg-purple-600 hover:bg-purple-500 text-white"
            >
              {isSubmitting ? 'Creating...' : 'Create Handoff'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export function SessionsDashboard() {
  const {
    sessions,
    sessionsLoading,
    sessionsError,
    deleteSessionById,
    resumeSessionById,
    refreshSessions,
    handoffs,
    handoffsLoading,
    handoffsError,
    createNewHandoff,
    claimHandoffById,
    refreshHandoffs,
    refreshAll,
  } = useSessions();

  const [searchQuery, setSearchQuery] = useState('');
  const [expandedSessions, setExpandedSessions] = useState<Set<string>>(new Set());
  const [expandedHandoffs, setExpandedHandoffs] = useState<Set<string>>(new Set());
  const [showCreateHandoff, setShowCreateHandoff] = useState(false);

  const filteredSessions = sessions.filter((session) => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      session.agentName.toLowerCase().includes(query) ||
      session.id.toLowerCase().includes(query)
    );
  });

  const pendingHandoffs = handoffs.filter((h) => h.status === 'pending');
  const claimedHandoffs = handoffs.filter((h) => h.status === 'claimed');

  const toggleSessionExpanded = (id: string) => {
    setExpandedSessions((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleHandoffExpanded = (id: string) => {
    setExpandedHandoffs((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleCreateHandoff = async (description: string, context: string) => {
    await createNewHandoff({ description, context });
  };

  const activeSessions = sessions.filter((s) => s.status === 'active').length;
  const totalMessages = sessions.reduce((sum, s) => sum + s.messages.length, 0);

  return (
    <div className="flex-1 overflow-y-auto bg-gray-900">
      <div className="p-6 space-y-6 max-w-7xl mx-auto">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight">Sessions & Handoffs</h1>
            <p className="text-white/40 text-sm">
              Manage agent sessions and coordinate handoffs
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Button
              variant="outline"
              size="sm"
              onClick={refreshAll}
              className="bg-transparent border-white/10 hover:bg-white/5"
            >
              <RefreshCw className="h-4 w-4 mr-1" />
              Refresh
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="bg-zinc-900/50 border-white/10">
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-white">{sessions.length}</div>
                  <div className="text-xs text-white/40 uppercase tracking-wider">Total Sessions</div>
                </div>
                <MessageSquare className="h-8 w-8 text-purple-400 opacity-50" />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-zinc-900/50 border-white/10">
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-green-400">{activeSessions}</div>
                  <div className="text-xs text-white/40 uppercase tracking-wider">Active Sessions</div>
                </div>
                <Play className="h-8 w-8 text-green-400 opacity-50" />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-zinc-900/50 border-white/10">
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-yellow-400">{pendingHandoffs.length}</div>
                  <div className="text-xs text-white/40 uppercase tracking-wider">Pending Handoffs</div>
                </div>
                <Hand className="h-8 w-8 text-yellow-400 opacity-50" />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-zinc-900/50 border-white/10">
            <CardContent className="pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-2xl font-bold text-blue-400">{totalMessages}</div>
                  <div className="text-xs text-white/40 uppercase tracking-wider">Total Messages</div>
                </div>
                <Clock className="h-8 w-8 text-blue-400 opacity-50" />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <Tabs defaultValue="sessions" className="w-full">
              <TabsList className="bg-zinc-900/50 border border-white/10">
                <TabsTrigger value="sessions" className="data-[state=active]:bg-purple-600">
                  <MessageSquare className="h-4 w-4 mr-2" />
                  Sessions ({sessions.length})
                </TabsTrigger>
                <TabsTrigger value="handoffs" className="data-[state=active]:bg-purple-600">
                  <Hand className="h-4 w-4 mr-2" />
                  Handoffs ({handoffs.length})
                </TabsTrigger>
              </TabsList>

              <TabsContent value="sessions" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">Agent Sessions</CardTitle>
                        <CardDescription className="text-white/40">
                          View and manage past agent sessions
                        </CardDescription>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={refreshSessions}
                        className="bg-transparent border-white/10 hover:bg-white/5"
                      >
                        <RefreshCw className="h-4 w-4" />
                      </Button>
                    </div>
                    <div className="pt-2">
                      <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-white/40" />
                        <Input
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                          placeholder="Search by agent name..."
                          className="bg-black/40 border-white/10 text-white pl-10"
                        />
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {sessionsLoading ? (
                      <div className="text-center text-white/40 py-8">Loading sessions...</div>
                    ) : sessionsError ? (
                      <div className="text-center text-red-400 py-8">
                        {sessionsError}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={refreshSessions}
                          className="mt-2 block mx-auto"
                        >
                          Retry
                        </Button>
                      </div>
                    ) : filteredSessions.length === 0 ? (
                      <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                        <MessageSquare className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>{searchQuery ? 'No matching sessions found' : 'No sessions yet'}</p>
                      </div>
                    ) : (
                      <ScrollArea className="h-[400px] pr-4">
                        <div className="space-y-3">
                          {filteredSessions.map((session) => (
                            <SessionCard
                              key={session.id}
                              session={session}
                              isExpanded={expandedSessions.has(session.id)}
                              onToggle={() => toggleSessionExpanded(session.id)}
                              onResume={() => resumeSessionById(session.id)}
                              onDelete={() => deleteSessionById(session.id)}
                            />
                          ))}
                        </div>
                      </ScrollArea>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="handoffs" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">Handoffs</CardTitle>
                        <CardDescription className="text-white/40">
                          View and manage handoff requests
                        </CardDescription>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={refreshHandoffs}
                          className="bg-transparent border-white/10 hover:bg-white/5"
                        >
                          <RefreshCw className="h-4 w-4" />
                        </Button>
                        <Button
                          size="sm"
                          onClick={() => setShowCreateHandoff(true)}
                          className="bg-purple-600 hover:bg-purple-500"
                        >
                          <Plus className="h-4 w-4 mr-1" />
                          New Handoff
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {handoffsLoading ? (
                      <div className="text-center text-white/40 py-8">Loading handoffs...</div>
                    ) : handoffsError ? (
                      <div className="text-center text-red-400 py-8">
                        {handoffsError}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={refreshHandoffs}
                          className="mt-2 block mx-auto"
                        >
                          Retry
                        </Button>
                      </div>
                    ) : handoffs.length === 0 ? (
                      <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                        <Hand className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>No handoffs yet</p>
                        <p className="text-xs mt-1">Click "New Handoff" to create one</p>
                      </div>
                    ) : (
                      <ScrollArea className="h-[400px] pr-4">
                        <div className="space-y-3">
                          {handoffs.map((handoff) => (
                            <HandoffCard
                              key={handoff.id}
                              handoff={handoff}
                              isExpanded={expandedHandoffs.has(handoff.id)}
                              onToggle={() => toggleHandoffExpanded(handoff.id)}
                              onClaim={() => claimHandoffById(handoff.id)}
                            />
                          ))}
                        </div>
                      </ScrollArea>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>

          <div className="space-y-6">
            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Clock className="h-5 w-5 text-blue-400" />
                  Recent Activity
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ScrollArea className="h-[300px]">
                  <div className="space-y-3">
                    {[...sessions, ...handoffs]
                      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
                      .slice(0, 10)
                      .map((item) => {
                        const isSession = 'agentName' in item;
                        return (
                          <div
                            key={item.id}
                            className="flex items-start gap-2 p-2 bg-black/40 rounded-lg"
                          >
                            {isSession ? (
                              <MessageSquare className="h-4 w-4 text-purple-400 mt-0.5" />
                            ) : (
                              <Hand className="h-4 w-4 text-yellow-400 mt-0.5" />
                            )}
                            <div className="flex-1 min-w-0">
                              <div className="text-xs text-white truncate">
                                {isSession
                                  ? (item as Session).agentName
                                  : (item as Handoff).description}
                              </div>
                              <div className="text-[10px] text-white/40 font-mono">
                                {formatRelativeTime(item.timestamp)}
                              </div>
                            </div>
                            {!isSession && (
                              <Badge className={`text-[9px] ${(item as Handoff).status === 'pending'
                                  ? 'bg-yellow-500/20 text-yellow-400'
                                  : 'bg-green-500/20 text-green-400'
                                } border-0`}>
                                {(item as Handoff).status}
                              </Badge>
                            )}
                          </div>
                        );
                      })}
                    {sessions.length === 0 && handoffs.length === 0 && (
                      <div className="text-center text-white/40 py-4 text-xs">
                        No recent activity
                      </div>
                    )}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <CardTitle className="text-white text-sm">Quick Stats</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Pending Handoffs</span>
                    <span className="text-sm font-bold text-yellow-400">{pendingHandoffs.length}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Claimed Handoffs</span>
                    <span className="text-sm font-bold text-green-400">{claimedHandoffs.length}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Active Sessions</span>
                    <span className="text-sm font-bold text-purple-400">{activeSessions}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Total Messages</span>
                    <span className="text-sm font-bold text-blue-400">{totalMessages}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <CreateHandoffDialog
        isOpen={showCreateHandoff}
        onClose={() => setShowCreateHandoff(false)}
        onSubmit={handleCreateHandoff}
      />
    </div>
  );
}
