'use client';

import { useState } from 'react';
import {
  Play,
  Square,
  RefreshCw,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
  ExternalLink,
  Workflow,
  Users,
  Code,
  Search,
  TestTube,
  Layers,
  Activity,
  Server,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { ScrollArea } from './ui/scroll-area';
import { Input } from './ui/input';
import {
  useConductor,
  type ConductorTask,
  type TaskStatus,
  type TaskRole,
} from '../lib/hooks/use-conductor';

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString();
}

function getStatusColor(status: TaskStatus): string {
  switch (status) {
    case 'pending':
      return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30';
    case 'running':
      return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
    case 'completed':
      return 'bg-green-500/20 text-green-400 border-green-500/30';
    case 'failed':
      return 'bg-red-500/20 text-red-400 border-red-500/30';
    default:
      return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
  }
}

function getStatusIcon(status: TaskStatus) {
  switch (status) {
    case 'pending':
      return <Clock className="h-4 w-4 text-yellow-400" />;
    case 'running':
      return <Loader2 className="h-4 w-4 text-blue-400 animate-spin" />;
    case 'completed':
      return <CheckCircle2 className="h-4 w-4 text-green-400" />;
    case 'failed':
      return <XCircle className="h-4 w-4 text-red-400" />;
    default:
      return <AlertCircle className="h-4 w-4 text-gray-400" />;
  }
}

function getRoleIcon(role: TaskRole) {
  switch (role) {
    case 'architect':
      return <Layers className="h-4 w-4 text-purple-400" />;
    case 'developer':
      return <Code className="h-4 w-4 text-blue-400" />;
    case 'reviewer':
      return <Search className="h-4 w-4 text-orange-400" />;
    case 'tester':
      return <TestTube className="h-4 w-4 text-green-400" />;
    default:
      return <Users className="h-4 w-4 text-gray-400" />;
  }
}

function getRoleColor(role: TaskRole): string {
  switch (role) {
    case 'architect':
      return 'bg-purple-500/20 text-purple-400 border-purple-500/30';
    case 'developer':
      return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
    case 'reviewer':
      return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
    case 'tester':
      return 'bg-green-500/20 text-green-400 border-green-500/30';
    default:
      return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
  }
}

function getWorkerStatusColor(status: string): string {
  switch (status) {
    case 'idle':
      return 'text-gray-400';
    case 'busy':
      return 'text-yellow-400';
    case 'overloaded':
      return 'text-red-400';
    default:
      return 'text-white/60';
  }
}

interface TaskCardProps {
  task: ConductorTask;
}

function TaskCard({ task }: TaskCardProps) {
  return (
    <div className="p-4 rounded-lg bg-black/40 border border-white/5 space-y-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {getStatusIcon(task.status)}
          <span className="font-bold text-sm text-white truncate">{task.name}</span>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          <Badge className={`${getRoleColor(task.role)} border text-xs`}>
            <span className="flex items-center gap-1">
              {getRoleIcon(task.role)}
              {task.role}
            </span>
          </Badge>
          <Badge className={`${getStatusColor(task.status)} border text-xs`}>
            {task.status}
          </Badge>
        </div>
      </div>

      {task.status === 'running' && (
        <div className="space-y-1">
          <div className="flex justify-between text-xs">
            <span className="text-white/40">Progress</span>
            <span className="text-white/60 font-mono">{task.progress}%</span>
          </div>
          <div className="h-1.5 bg-black/40 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 rounded-full transition-all duration-300"
              style={{ width: `${task.progress}%` }}
            />
          </div>
        </div>
      )}

      <div className="flex items-center justify-between text-xs text-white/40">
        <div className="flex items-center gap-1">
          <Clock className="h-3 w-3" />
          <span>Created: {formatDate(task.createdAt)}</span>
        </div>
        <span className="font-mono text-white/30">{task.id.slice(0, 8)}</span>
      </div>

      {task.error && (
        <div className="p-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400">
          {task.error}
        </div>
      )}
    </div>
  );
}

interface RoleSelectorProps {
  selectedRole: TaskRole;
  onRoleSelect: (role: TaskRole) => void;
  onStart: () => void;
  disabled?: boolean;
}

function RoleSelector({ selectedRole, onRoleSelect, onStart, disabled }: RoleSelectorProps) {
  const roles: TaskRole[] = ['architect', 'developer', 'reviewer', 'tester'];

  return (
    <div className="flex items-center gap-3">
      <div className="flex gap-1 bg-black/40 p-1 rounded-lg border border-white/10">
        {roles.map((role) => (
          <button
            key={role}
            onClick={() => onRoleSelect(role)}
            disabled={disabled}
            className={`
              flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium transition-colors
              ${selectedRole === role
                ? 'bg-purple-600 text-white'
                : 'text-white/60 hover:text-white hover:bg-white/5'
              }
              disabled:opacity-50 disabled:cursor-not-allowed
            `}
          >
            {getRoleIcon(role)}
            <span className="capitalize">{role}</span>
          </button>
        ))}
      </div>
      <Button
        onClick={onStart}
        disabled={disabled}
        className="bg-purple-600 hover:bg-purple-500"
      >
        <Play className="h-4 w-4 mr-1" />
        Start Task
      </Button>
    </div>
  );
}

export function ConductorDashboard() {
  const {
    tasks,
    tasksLoading,
    tasksError,
    startTask,
    refreshTasks,
    conductorStatus,
    conductorStatusLoading,
    conductorStatusError,
    refreshConductorStatus,
    vibeKanbanStatus,
    vibeKanbanLoading,
    vibeKanbanError,
    startVibeKanbanInstance,
    stopVibeKanbanInstance,
    refreshVibeKanbanStatus,
    refreshAll,
  } = useConductor();

  const [selectedRole, setSelectedRole] = useState<TaskRole>('developer');
  const [frontendPort, setFrontendPort] = useState('3000');
  const [backendPort, setBackendPort] = useState('3001');
  const [isStartingTask, setIsStartingTask] = useState(false);

  const handleStartTask = async () => {
    setIsStartingTask(true);
    try {
      await startTask(selectedRole);
    } finally {
      setIsStartingTask(false);
    }
  };

  const handleStartVibeKanban = async () => {
    await startVibeKanbanInstance({
      frontendPort: parseInt(frontendPort, 10),
      backendPort: parseInt(backendPort, 10),
    });
  };

  const pendingTasks = tasks.filter(t => t.status === 'pending');
  const runningTasks = tasks.filter(t => t.status === 'running');
  const completedTasks = tasks.filter(t => t.status === 'completed');
  const failedTasks = tasks.filter(t => t.status === 'failed');

  return (
    <div className="flex-1 overflow-y-auto bg-gray-900">
      <div className="p-6 space-y-6 max-w-7xl mx-auto">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-3">
              <Workflow className="h-7 w-7 text-purple-400" />
              Conductor Dashboard
            </h1>
            <p className="text-white/40 text-sm">Task orchestration and VibeKanban integration</p>
          </div>
          <Button
            variant="outline"
            onClick={refreshAll}
            className="bg-transparent border-white/10 hover:bg-white/5"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh All
          </Button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-white flex items-center gap-2">
                      <Activity className="h-5 w-5 text-blue-400" />
                      Task Management
                    </CardTitle>
                    <CardDescription className="text-white/40">
                      Create and monitor orchestration tasks
                    </CardDescription>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={refreshTasks}
                    disabled={tasksLoading}
                    className="bg-transparent border-white/10 hover:bg-white/5"
                  >
                    <RefreshCw className={`h-4 w-4 ${tasksLoading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="p-4 bg-black/40 rounded-lg border border-white/10">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-medium text-white">Start New Task</span>
                  </div>
                  <RoleSelector
                    selectedRole={selectedRole}
                    onRoleSelect={setSelectedRole}
                    onStart={handleStartTask}
                    disabled={isStartingTask}
                  />
                </div>

                <div className="grid grid-cols-4 gap-3">
                  <div className="p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg text-center">
                    <div className="text-xl font-bold text-yellow-400">{pendingTasks.length}</div>
                    <div className="text-xs text-white/40">Pending</div>
                  </div>
                  <div className="p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg text-center">
                    <div className="text-xl font-bold text-blue-400">{runningTasks.length}</div>
                    <div className="text-xs text-white/40">Running</div>
                  </div>
                  <div className="p-3 bg-green-500/10 border border-green-500/20 rounded-lg text-center">
                    <div className="text-xl font-bold text-green-400">{completedTasks.length}</div>
                    <div className="text-xs text-white/40">Completed</div>
                  </div>
                  <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-center">
                    <div className="text-xl font-bold text-red-400">{failedTasks.length}</div>
                    <div className="text-xs text-white/40">Failed</div>
                  </div>
                </div>

                {tasksError && (
                  <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-sm text-red-400">
                    {tasksError}
                  </div>
                )}

                {tasksLoading && tasks.length === 0 ? (
                  <div className="text-center text-white/40 py-8">
                    <Loader2 className="h-6 w-6 mx-auto mb-2 animate-spin" />
                    Loading tasks...
                  </div>
                ) : tasks.length === 0 ? (
                  <div className="text-center text-white/40 py-8 border border-dashed border-white/10 rounded-lg">
                    <Workflow className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No tasks found</p>
                    <p className="text-xs mt-1">Start a new task to begin orchestration</p>
                  </div>
                ) : (
                  <ScrollArea className="h-[400px]">
                    <div className="space-y-3">
                      {tasks.map((task) => (
                        <TaskCard key={task.id} task={task} />
                      ))}
                    </div>
                  </ScrollArea>
                )}
              </CardContent>
            </Card>
          </div>

          <div className="space-y-6">
            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-white flex items-center gap-2">
                    <Server className="h-5 w-5 text-green-400" />
                    Conductor Status
                  </CardTitle>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={refreshConductorStatus}
                    disabled={conductorStatusLoading}
                    className="bg-transparent border-white/10 hover:bg-white/5 h-7 w-7 p-0"
                  >
                    <RefreshCw className={`h-3 w-3 ${conductorStatusLoading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {conductorStatusError && (
                  <div className="p-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400 mb-3">
                    {conductorStatusError}
                  </div>
                )}

                {conductorStatusLoading && !conductorStatus ? (
                  <div className="text-center text-white/40 py-4">
                    <Loader2 className="h-5 w-5 mx-auto mb-2 animate-spin" />
                    Loading...
                  </div>
                ) : conductorStatus ? (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                      <span className="text-xs text-white/40">Active Tasks</span>
                      <span className="text-sm font-bold text-blue-400">{conductorStatus.activeTasks}</span>
                    </div>
                    <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                      <span className="text-xs text-white/40">Queue Depth</span>
                      <span className="text-sm font-bold text-yellow-400">{conductorStatus.queueDepth}</span>
                    </div>
                    <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                      <span className="text-xs text-white/40">Worker Status</span>
                      <Badge className={`${getWorkerStatusColor(conductorStatus.workerStatus)} bg-transparent border border-white/10 text-xs`}>
                        {conductorStatus.workerStatus}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                      <span className="text-xs text-white/40">Total Completed</span>
                      <span className="text-sm font-bold text-green-400">{conductorStatus.totalCompleted}</span>
                    </div>
                    <div className="flex items-center justify-between p-2 bg-black/40 rounded">
                      <span className="text-xs text-white/40">Total Failed</span>
                      <span className="text-sm font-bold text-red-400">{conductorStatus.totalFailed}</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-center text-white/40 py-4">
                    <AlertCircle className="h-5 w-5 mx-auto mb-2 opacity-40" />
                    Unable to load status
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="bg-zinc-900/50 border-white/10">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-white flex items-center gap-2">
                    <Layers className="h-5 w-5 text-purple-400" />
                    VibeKanban
                  </CardTitle>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={refreshVibeKanbanStatus}
                    disabled={vibeKanbanLoading}
                    className="bg-transparent border-white/10 hover:bg-white/5 h-7 w-7 p-0"
                  >
                    <RefreshCw className={`h-3 w-3 ${vibeKanbanLoading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>
                <CardDescription className="text-white/40">
                  Visual kanban board integration
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {vibeKanbanError && (
                  <div className="p-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400">
                    {vibeKanbanError}
                  </div>
                )}

                <div className="flex items-center justify-between p-3 bg-black/40 rounded-lg border border-white/10">
                  <span className="text-sm text-white">Status</span>
                  <div className="flex items-center gap-2">
                    <span className={`h-2 w-2 rounded-full ${vibeKanbanStatus?.running ? 'bg-green-500 animate-pulse' : 'bg-gray-500'}`} />
                    <span className={`text-sm font-medium ${vibeKanbanStatus?.running ? 'text-green-400' : 'text-gray-400'}`}>
                      {vibeKanbanStatus?.running ? 'Running' : 'Stopped'}
                    </span>
                  </div>
                </div>

                {!vibeKanbanStatus?.running && (
                  <div className="space-y-3">
                    <div>
                      <label className="text-xs text-white/40 block mb-1">Frontend Port</label>
                      <Input
                        type="number"
                        value={frontendPort}
                        onChange={(e) => setFrontendPort(e.target.value)}
                        placeholder="3000"
                        className="bg-black/40 border-white/10 text-white"
                      />
                    </div>
                    <div>
                      <label className="text-xs text-white/40 block mb-1">Backend Port</label>
                      <Input
                        type="number"
                        value={backendPort}
                        onChange={(e) => setBackendPort(e.target.value)}
                        placeholder="3001"
                        className="bg-black/40 border-white/10 text-white"
                      />
                    </div>
                    <Button
                      onClick={handleStartVibeKanban}
                      disabled={vibeKanbanLoading}
                      className="w-full bg-purple-600 hover:bg-purple-500"
                    >
                      {vibeKanbanLoading ? (
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      ) : (
                        <Play className="h-4 w-4 mr-2" />
                      )}
                      Start VibeKanban
                    </Button>
                  </div>
                )}

                {vibeKanbanStatus?.running && (
                  <div className="space-y-3">
                    {vibeKanbanStatus.frontendUrl && (
                      <a
                        href={vibeKanbanStatus.frontendUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center justify-between p-2 bg-black/40 rounded hover:bg-white/5 transition-colors group"
                      >
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-white/40">Frontend</span>
                          <span className="text-sm text-white font-mono">{vibeKanbanStatus.frontendUrl}</span>
                        </div>
                        <ExternalLink className="h-4 w-4 text-white/40 group-hover:text-purple-400" />
                      </a>
                    )}
                    {vibeKanbanStatus.backendUrl && (
                      <a
                        href={vibeKanbanStatus.backendUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center justify-between p-2 bg-black/40 rounded hover:bg-white/5 transition-colors group"
                      >
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-white/40">Backend</span>
                          <span className="text-sm text-white font-mono">{vibeKanbanStatus.backendUrl}</span>
                        </div>
                        <ExternalLink className="h-4 w-4 text-white/40 group-hover:text-purple-400" />
                      </a>
                    )}

                    <div className="flex gap-3">
                      <div className="flex-1 p-2 bg-black/40 rounded text-center">
                        <div className="text-xs text-white/40">Frontend Port</div>
                        <div className="text-sm font-mono text-white">{vibeKanbanStatus.frontendPort}</div>
                      </div>
                      <div className="flex-1 p-2 bg-black/40 rounded text-center">
                        <div className="text-xs text-white/40">Backend Port</div>
                        <div className="text-sm font-mono text-white">{vibeKanbanStatus.backendPort}</div>
                      </div>
                    </div>

                    <Button
                      onClick={stopVibeKanbanInstance}
                      disabled={vibeKanbanLoading}
                      variant="outline"
                      className="w-full bg-transparent border-white/10 hover:bg-red-600/20 hover:text-red-400 hover:border-red-500/30"
                    >
                      {vibeKanbanLoading ? (
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      ) : (
                        <Square className="h-4 w-4 mr-2" />
                      )}
                      Stop VibeKanban
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
