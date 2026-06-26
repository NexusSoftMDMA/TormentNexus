'use client';

import { useState } from 'react';
import {
  Terminal,
  Play,
  Square,
  RotateCcw,
  Trash2,
  Plus,
  RefreshCw,
  Settings,
  Clock,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Timer,
  Search,
  Download,
  ChevronDown,
  FolderOpen,
  Cpu,
  Zap,
  Shield,
  History,
  PlayCircle,
  PauseCircle,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Switch } from './ui/switch';
import { ScrollArea } from './ui/scroll-area';
import { Input } from './ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import {
  useAutopilot,
  type CLITool,
  type CLISession,
  type VetoRequest,
  type DebateDecision,
  type RiskLevel,
} from '../lib/hooks/use-autopilot';

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${minutes}m`;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function getRiskColor(risk: RiskLevel): string {
  switch (risk) {
    case 'low':
      return 'bg-green-500/20 text-green-400 border-green-500/30';
    case 'medium':
      return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30';
    case 'high':
      return 'bg-red-500/20 text-red-400 border-red-500/30';
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'active':
    case 'running':
    case 'available':
      return 'text-green-400';
    case 'stopped':
    case 'paused':
      return 'text-gray-400';
    case 'error':
    case 'unavailable':
      return 'text-red-400';
    default:
      return 'text-white/60';
  }
}

function CLIToolCard({ tool }: { tool: CLITool }) {
  const isAvailable = tool.status === 'available';
  
  return (
    <div className="p-4 rounded-lg bg-black/40 border border-white/5 space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Terminal className="h-4 w-4 text-purple-400" />
          <span className="font-bold text-sm text-white">{tool.name}</span>
        </div>
        <Badge className={`${isAvailable ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'} border-0`}>
          {tool.status}
        </Badge>
      </div>
      <div className="text-xs font-mono text-white/60 space-y-1">
        <div className="flex items-center gap-2">
          <span className="text-white/40">Version:</span>
          <span className="text-purple-300">{tool.version}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-white/40">Path:</span>
          <span className="truncate">{tool.path}</span>
        </div>
      </div>
    </div>
  );
}

function SessionCard({
  session,
  onStart,
  onStop,
  onRestart,
  onDelete,
}: {
  session: CLISession;
  onStart: () => void;
  onStop: () => void;
  onRestart: () => void;
  onDelete: () => void;
}) {
  const isActive = session.status === 'active';
  
  return (
    <Card className="bg-zinc-900/50 border-white/10">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-white text-sm flex items-center gap-2">
            <span className={`h-2 w-2 rounded-full ${isActive ? 'bg-green-500 animate-pulse' : 'bg-gray-500'}`} />
            {session.cliType}
          </CardTitle>
          <Badge className={`${getStatusColor(session.status)} bg-transparent border border-white/10`}>
            {session.status}
          </Badge>
        </div>
        <CardDescription className="text-white/40 font-mono text-xs truncate">
          {session.id}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="text-xs font-mono text-white/60 space-y-1">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-3 w-3" />
            <span className="truncate">{session.workingDirectory}</span>
          </div>
          {isActive && (
            <div className="flex items-center gap-2">
              <Clock className="h-3 w-3" />
              <span>Uptime: {formatUptime(session.uptime)}</span>
            </div>
          )}
        </div>
        <div className="flex gap-2">
          {isActive ? (
            <Button
              variant="outline"
              size="sm"
              onClick={onStop}
              className="flex-1 text-xs bg-transparent border-white/10 hover:bg-red-600/20 hover:text-red-400"
            >
              <Square className="h-3 w-3 mr-1" />
              Stop
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={onStart}
              className="flex-1 text-xs bg-transparent border-white/10 hover:bg-green-600/20 hover:text-green-400"
            >
              <Play className="h-3 w-3 mr-1" />
              Start
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={onRestart}
            className="text-xs bg-transparent border-white/10 hover:bg-yellow-600/20 hover:text-yellow-400"
          >
            <RotateCcw className="h-3 w-3" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={onDelete}
            className="text-xs bg-transparent border-white/10 hover:bg-red-600/20 hover:text-red-400"
          >
            <Trash2 className="h-3 w-3" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function VetoRequestCard({
  request,
  onApprove,
  onReject,
  onExtend,
}: {
  request: VetoRequest;
  onApprove: () => void;
  onReject: () => void;
  onExtend: (seconds: number) => void;
}) {
  const [showExtend, setShowExtend] = useState(false);
  const timePercent = Math.max(0, Math.min(100, (request.timeRemaining / 300) * 100));
  
  return (
    <Card className="bg-zinc-900/50 border-white/10">
      <CardContent className="pt-4 space-y-3">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 space-y-1">
            <div className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-yellow-400" />
              <span className="font-bold text-sm text-white">{request.action}</span>
            </div>
            <p className="text-xs text-white/60">{request.description}</p>
            <div className="flex items-center gap-2 text-xs text-white/40">
              <span>Agent: {request.requestingAgent}</span>
              {request.sessionId && <span>Session: {request.sessionId.slice(0, 8)}</span>}
            </div>
          </div>
          <Badge className={`${getRiskColor(request.riskLevel)} border`}>
            {request.riskLevel}
          </Badge>
        </div>
        
        <div className="space-y-1">
          <div className="flex justify-between text-xs">
            <span className="text-white/40">Time remaining</span>
            <span className={`font-mono ${request.timeRemaining < 30 ? 'text-red-400' : 'text-white/60'}`}>
              {formatUptime(request.timeRemaining)}
            </span>
          </div>
          <div className="h-1.5 bg-black/40 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full transition-all ${
                request.timeRemaining < 30 ? 'bg-red-500' : request.timeRemaining < 60 ? 'bg-yellow-500' : 'bg-purple-500'
              }`}
              style={{ width: `${timePercent}%` }}
            />
          </div>
        </div>
        
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={onApprove}
            className="flex-1 bg-green-600 hover:bg-green-500 text-white"
          >
            <CheckCircle2 className="h-3 w-3 mr-1" />
            Approve
          </Button>
          <Button
            size="sm"
            onClick={onReject}
            className="flex-1 bg-red-600 hover:bg-red-500 text-white"
          >
            <XCircle className="h-3 w-3 mr-1" />
            Reject
          </Button>
          <div className="relative">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowExtend(!showExtend)}
              className="bg-transparent border-white/10 hover:bg-white/5"
            >
              <Timer className="h-3 w-3" />
              <ChevronDown className="h-3 w-3 ml-1" />
            </Button>
            {showExtend && (
              <div className="absolute right-0 top-full mt-1 bg-zinc-800 border border-white/10 rounded-md shadow-lg z-10 p-1">
                {[30, 60, 300].map((seconds) => (
                  <button
                    key={seconds}
                    onClick={() => { onExtend(seconds); setShowExtend(false); }}
                    className="block w-full text-left px-3 py-1.5 text-xs text-white/80 hover:bg-white/10 rounded"
                  >
                    +{seconds < 60 ? `${seconds}s` : `${seconds / 60}m`}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function DebateHistoryRow({ decision }: { decision: DebateDecision }) {
  const decisionColor = {
    approved: 'text-green-400',
    rejected: 'text-red-400',
    timeout: 'text-yellow-400',
  }[decision.decision];
  
  return (
    <tr className="border-b border-white/5 hover:bg-white/5">
      <td className="px-3 py-2 text-xs text-white/60 font-mono">
        {new Date(decision.timestamp).toLocaleString()}
      </td>
      <td className="px-3 py-2 text-xs text-white">{decision.action}</td>
      <td className="px-3 py-2">
        <span className={`text-xs font-bold uppercase ${decisionColor}`}>
          {decision.decision}
        </span>
      </td>
      <td className="px-3 py-2 text-xs text-white/60">
        {decision.participants.join(', ')}
      </td>
      <td className="px-3 py-2 text-xs text-white/60 font-mono">
        {formatDuration(decision.duration)}
      </td>
      <td className="px-3 py-2">
        <div className="flex items-center gap-1">
          <div className="h-1.5 w-12 bg-black/40 rounded-full overflow-hidden">
            <div
              className="h-full bg-purple-500 rounded-full"
              style={{ width: `${decision.confidence}%` }}
            />
          </div>
          <span className="text-xs text-white/40">{decision.confidence}%</span>
        </div>
      </td>
    </tr>
  );
}

export function AutopilotDashboard() {
  const {
    cliTools,
    cliToolsLoading,
    refreshTools,
    sessions,
    sessionsLoading,
    createNewSession,
    startSessionById,
    stopSessionById,
    restartSessionById,
    deleteSessionById,
    startAll,
    stopAll,
    smartPilotStatus,
    smartPilotLoading,
    toggleSmartPilot,
    updateConfig,
    pause,
    resume,
    resetApprovalCount,
    vetoQueue,
    vetoLoading,
    approve,
    reject,
    extendTimeout,
    debateHistory,
    debateAnalytics,
    debateLoading,
    searchHistory,
    exportHistory,
  } = useAutopilot();

  const [searchQuery, setSearchQuery] = useState('');
  const [newSessionCli, setNewSessionCli] = useState('');
  const [newSessionDir, setNewSessionDir] = useState('');
  const [showNewSession, setShowNewSession] = useState(false);

  const handleSearch = () => {
    if (searchQuery.trim()) {
      searchHistory(searchQuery);
    }
  };

  const handleCreateSession = async () => {
    if (newSessionCli && newSessionDir) {
      await createNewSession(newSessionCli, newSessionDir);
      setNewSessionCli('');
      setNewSessionDir('');
      setShowNewSession(false);
    }
  };

  const activeSessions = sessions.filter(s => s.status === 'active').length;
  const stoppedSessions = sessions.filter(s => s.status === 'stopped').length;

  return (
    <div className="flex-1 overflow-y-auto bg-gray-900">
      <div className="p-6 space-y-6 max-w-7xl mx-auto">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight">Autopilot Dashboard</h1>
            <p className="text-white/40 text-sm">Manage CLI sessions, smart pilot controls, and decision workflows</p>
          </div>
          <div className="flex items-center gap-3">
            <Badge variant="outline" className="border-white/20 text-white/60">
              {activeSessions} active / {stoppedSessions} stopped
            </Badge>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <Tabs defaultValue="sessions" className="w-full">
              <TabsList className="bg-zinc-900/50 border border-white/10">
                <TabsTrigger value="sessions" className="data-[state=active]:bg-purple-600">
                  <Terminal className="h-4 w-4 mr-2" />
                  Sessions
                </TabsTrigger>
                <TabsTrigger value="tools" className="data-[state=active]:bg-purple-600">
                  <Cpu className="h-4 w-4 mr-2" />
                  CLI Tools
                </TabsTrigger>
              </TabsList>

              <TabsContent value="sessions" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">CLI Sessions</CardTitle>
                        <CardDescription className="text-white/40">
                          Manage active and stopped CLI sessions
                        </CardDescription>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={startAll}
                          className="bg-transparent border-white/10 hover:bg-green-600/20 hover:text-green-400"
                        >
                          <PlayCircle className="h-4 w-4 mr-1" />
                          Start All
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={stopAll}
                          className="bg-transparent border-white/10 hover:bg-red-600/20 hover:text-red-400"
                        >
                          <Square className="h-4 w-4 mr-1" />
                          Stop All
                        </Button>
                        <Button
                          size="sm"
                          onClick={() => setShowNewSession(!showNewSession)}
                          className="bg-purple-600 hover:bg-purple-500"
                        >
                          <Plus className="h-4 w-4 mr-1" />
                          New Session
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {showNewSession && (
                      <div className="mb-4 p-4 bg-black/40 rounded-lg border border-white/10 space-y-3">
                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <label className="text-xs text-white/40 block mb-1">CLI Type</label>
                            <Input
                              value={newSessionCli}
                              onChange={(e) => setNewSessionCli(e.target.value)}
                              placeholder="e.g., opencode, claude"
                              className="bg-black/40 border-white/10 text-white"
                            />
                          </div>
                          <div>
                            <label className="text-xs text-white/40 block mb-1">Working Directory</label>
                            <Input
                              value={newSessionDir}
                              onChange={(e) => setNewSessionDir(e.target.value)}
                              placeholder="/path/to/project"
                              className="bg-black/40 border-white/10 text-white"
                            />
                          </div>
                        </div>
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setShowNewSession(false)}
                            className="bg-transparent border-white/10"
                          >
                            Cancel
                          </Button>
                          <Button
                            size="sm"
                            onClick={handleCreateSession}
                            className="bg-purple-600 hover:bg-purple-500"
                          >
                            Create
                          </Button>
                        </div>
                      </div>
                    )}
                    
                    {sessionsLoading ? (
                      <div className="text-center text-white/40 py-8">Loading sessions...</div>
                    ) : sessions.length === 0 ? (
                      <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                        <Terminal className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>No sessions configured</p>
                        <p className="text-xs mt-1">Click "New Session" to create one</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {sessions.map((session) => (
                          <SessionCard
                            key={session.id}
                            session={session}
                            onStart={() => startSessionById(session.id)}
                            onStop={() => stopSessionById(session.id)}
                            onRestart={() => restartSessionById(session.id)}
                            onDelete={() => deleteSessionById(session.id)}
                          />
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="tools" className="mt-4">
                <Card className="bg-zinc-900/50 border-white/10">
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-white">CLI Registry</CardTitle>
                        <CardDescription className="text-white/40">
                          Detected CLI tools available for sessions
                        </CardDescription>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={refreshTools}
                          disabled={cliToolsLoading}
                          className="bg-transparent border-white/10 hover:bg-white/5"
                        >
                          <RefreshCw className={`h-4 w-4 mr-1 ${cliToolsLoading ? 'animate-spin' : ''}`} />
                          Refresh Detection
                        </Button>
                        <Button
                          size="sm"
                          className="bg-purple-600 hover:bg-purple-500"
                        >
                          <Plus className="h-4 w-4 mr-1" />
                          Register Custom
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {cliToolsLoading ? (
                      <div className="text-center text-white/40 py-8">Scanning for CLI tools...</div>
                    ) : cliTools.length === 0 ? (
                      <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                        <Cpu className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>No CLI tools detected</p>
                        <p className="text-xs mt-1">Click "Refresh Detection" to scan</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {cliTools.map((tool) => (
                          <CLIToolCard key={tool.id} tool={tool} />
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-white flex items-center gap-2">
                      <Shield className="h-5 w-5 text-yellow-400" />
                      Veto Queue
                    </CardTitle>
                    <CardDescription className="text-white/40">
                      Pending actions requiring human approval
                    </CardDescription>
                  </div>
                  <Badge variant="outline" className={`${vetoQueue.length > 0 ? 'border-yellow-500/50 text-yellow-400' : 'border-white/20 text-white/60'}`}>
                    {vetoQueue.length} pending
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                {vetoLoading ? (
                  <div className="text-center text-white/40 py-8">Loading veto queue...</div>
                ) : vetoQueue.length === 0 ? (
                  <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                    <CheckCircle2 className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No pending approvals</p>
                    <p className="text-xs mt-1">All actions have been processed</p>
                  </div>
                ) : (
                  <ScrollArea className="h-[300px]">
                    <div className="space-y-4">
                      {vetoQueue.map((request) => (
                        <VetoRequestCard
                          key={request.id}
                          request={request}
                          onApprove={() => approve(request.id)}
                          onReject={() => reject(request.id)}
                          onExtend={(seconds) => extendTimeout(request.id, seconds)}
                        />
                      ))}
                    </div>
                  </ScrollArea>
                )}
              </CardContent>
            </Card>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-white flex items-center gap-2">
                      <History className="h-5 w-5 text-blue-400" />
                      Debate History
                    </CardTitle>
                    <CardDescription className="text-white/40">
                      Past decisions and their outcomes
                    </CardDescription>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => exportHistory('json')}
                      className="bg-transparent border-white/10 hover:bg-white/5"
                    >
                      <Download className="h-4 w-4 mr-1" />
                      JSON
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => exportHistory('csv')}
                      className="bg-transparent border-white/10 hover:bg-white/5"
                    >
                      <Download className="h-4 w-4 mr-1" />
                      CSV
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                {debateAnalytics && (
                  <div className="grid grid-cols-4 gap-4 mb-4">
                    <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                      <div className="text-2xl font-bold text-white">{debateAnalytics.totalDecisions}</div>
                      <div className="text-xs text-white/40">Total Decisions</div>
                    </div>
                    <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                      <div className="text-2xl font-bold text-green-400">{debateAnalytics.approvalRate.toFixed(1)}%</div>
                      <div className="text-xs text-white/40">Approval Rate</div>
                    </div>
                    <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                      <div className="text-2xl font-bold text-purple-400">{formatDuration(debateAnalytics.avgDecisionTime)}</div>
                      <div className="text-xs text-white/40">Avg Decision Time</div>
                    </div>
                    <div className="p-3 bg-black/40 rounded-lg border border-white/5 text-center">
                      <div className="text-2xl font-bold text-blue-400">{debateAnalytics.todayDecisions}</div>
                      <div className="text-xs text-white/40">Today</div>
                    </div>
                  </div>
                )}
                
                <div className="flex gap-2 mb-4">
                  <Input
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                    placeholder="Search history..."
                    className="bg-black/40 border-white/10 text-white"
                  />
                  <Button
                    variant="outline"
                    onClick={handleSearch}
                    className="bg-transparent border-white/10 hover:bg-white/5"
                  >
                    <Search className="h-4 w-4" />
                  </Button>
                </div>

                {debateLoading ? (
                  <div className="text-center text-white/40 py-8">Loading history...</div>
                ) : debateHistory.length === 0 ? (
                  <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                    <History className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No debate history</p>
                  </div>
                ) : (
                  <ScrollArea className="h-[300px]">
                    <table className="w-full">
                      <thead className="sticky top-0 bg-zinc-900">
                        <tr className="border-b border-white/10">
                          <th className="px-3 py-2 text-left text-xs font-medium text-white/40 uppercase tracking-wider">Timestamp</th>
                          <th className="px-3 py-2 text-left text-xs font-medium text-white/40 uppercase tracking-wider">Action</th>
                          <th className="px-3 py-2 text-left text-xs font-medium text-white/40 uppercase tracking-wider">Decision</th>
                          <th className="px-3 py-2 text-left text-xs font-medium text-white/40 uppercase tracking-wider">Participants</th>
                          <th className="px-3 py-2 text-left text-xs font-medium text-white/40 uppercase tracking-wider">Duration</th>
                          <th className="px-3 py-2 text-left text-xs font-medium text-white/40 uppercase tracking-wider">Confidence</th>
                        </tr>
                      </thead>
                      <tbody>
                        {debateHistory.map((decision) => (
                          <DebateHistoryRow key={decision.id} decision={decision} />
                        ))}
                      </tbody>
                    </table>
                  </ScrollArea>
                )}
              </CardContent>
            </Card>
          </div>

          <div className="space-y-6">
            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Zap className="h-5 w-5 text-purple-400" />
                  Smart Pilot Controls
                </CardTitle>
                <CardDescription className="text-white/40">
                  Configure autonomous decision-making
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                {smartPilotLoading ? (
                  <div className="text-center text-white/40 py-4">Loading...</div>
                ) : smartPilotStatus ? (
                  <>
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-sm font-medium text-white">Enable Smart Pilot</div>
                        <div className="text-xs text-white/40">Allow autonomous decisions</div>
                      </div>
                      <Switch
                        checked={smartPilotStatus.enabled}
                        onCheckedChange={toggleSmartPilot}
                      />
                    </div>

                    <div className="p-3 bg-black/40 rounded-lg border border-white/5">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-xs text-white/40">Status</span>
                        <Badge className={`${getStatusColor(smartPilotStatus.state)} bg-transparent border border-white/10`}>
                          {smartPilotStatus.state}
                        </Badge>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-white/40">Remaining Approvals</span>
                        <span className="text-lg font-bold text-purple-400">{smartPilotStatus.remainingApprovals}</span>
                      </div>
                    </div>

                    <div className="space-y-4">
                      <div>
                        <div className="flex justify-between text-xs mb-1">
                          <span className="text-white/40">Auto-approval Limit</span>
                          <span className="text-white font-mono">{smartPilotStatus.config.autoApprovalLimit}</span>
                        </div>
                        <input
                          type="range"
                          min="0"
                          max="100"
                          value={smartPilotStatus.config.autoApprovalLimit}
                          onChange={(e) => updateConfig({ autoApprovalLimit: parseInt(e.target.value) })}
                          className="w-full h-1.5 bg-black/40 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-purple-500"
                        />
                      </div>

                      <div>
                        <div className="flex justify-between text-xs mb-1">
                          <span className="text-white/40">Max Concurrent Sessions</span>
                          <span className="text-white font-mono">{smartPilotStatus.config.maxConcurrentSessions}</span>
                        </div>
                        <input
                          type="range"
                          min="1"
                          max="10"
                          value={smartPilotStatus.config.maxConcurrentSessions}
                          onChange={(e) => updateConfig({ maxConcurrentSessions: parseInt(e.target.value) })}
                          className="w-full h-1.5 bg-black/40 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-purple-500"
                        />
                      </div>

                      <div>
                        <div className="flex justify-between text-xs mb-1">
                          <span className="text-white/40">Decision Timeout (seconds)</span>
                          <span className="text-white font-mono">{smartPilotStatus.config.decisionTimeout}s</span>
                        </div>
                        <input
                          type="range"
                          min="10"
                          max="300"
                          step="10"
                          value={smartPilotStatus.config.decisionTimeout}
                          onChange={(e) => updateConfig({ decisionTimeout: parseInt(e.target.value) })}
                          className="w-full h-1.5 bg-black/40 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-purple-500"
                        />
                      </div>
                    </div>

                    <div className="flex gap-2">
                      {smartPilotStatus.state === 'running' ? (
                        <Button
                          variant="outline"
                          onClick={pause}
                          className="flex-1 bg-transparent border-white/10 hover:bg-yellow-600/20 hover:text-yellow-400"
                        >
                          <PauseCircle className="h-4 w-4 mr-1" />
                          Pause
                        </Button>
                      ) : (
                        <Button
                          variant="outline"
                          onClick={resume}
                          className="flex-1 bg-transparent border-white/10 hover:bg-green-600/20 hover:text-green-400"
                        >
                          <PlayCircle className="h-4 w-4 mr-1" />
                          Resume
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        onClick={resetApprovalCount}
                        className="flex-1 bg-transparent border-white/10 hover:bg-purple-600/20 hover:text-purple-400"
                      >
                        <RotateCcw className="h-4 w-4 mr-1" />
                        Reset
                      </Button>
                    </div>
                  </>
                ) : (
                  <div className="text-center text-red-400 py-4">Failed to load smart pilot status</div>
                )}
              </CardContent>
            </Card>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Settings className="h-5 w-5 text-gray-400" />
                  Quick Stats
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Active Sessions</span>
                    <span className="text-sm font-bold text-green-400">{activeSessions}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Pending Vetos</span>
                    <span className="text-sm font-bold text-yellow-400">{vetoQueue.length}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">CLI Tools</span>
                    <span className="text-sm font-bold text-purple-400">{cliTools.length}</span>
                  </div>
                  <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                    <span className="text-xs text-white/40">Total Decisions</span>
                    <span className="text-sm font-bold text-blue-400">{debateAnalytics?.totalDecisions ?? 0}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
