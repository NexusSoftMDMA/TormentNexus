import React, { useState } from 'react';
import { Card, CardContent } from '@src/components/ui/card';
import { Typography, Button, Icon } from '../ui';
import { cn } from '@src/lib/utils';
import { useDebuggerStore, type DebugPacket } from '@src/stores';

const PacketDetail: React.FC<{ packet: DebugPacket }> = ({ packet }) => {
  const [expanded, setExpanded] = useState(false);
  const isError = packet.type === 'error';
  const isRequest = packet.type === 'request';

  return (
    <div className={cn(
      "border-b border-slate-100 dark:border-slate-800 last:border-0",
      expanded ? "bg-slate-50 dark:bg-slate-800/50" : ""
    )}>
      <div
        className="flex items-center gap-2 p-3 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800/80 transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <div className={cn(
          "shrink-0 w-2 h-2 rounded-full",
          isError ? "bg-red-500" : isRequest ? "bg-blue-500" : "bg-green-500"
        )} />

        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between">
            <Typography variant="body" className="font-semibold truncate">
              {packet.method} {packet.toolName ? `(${packet.toolName})` : ''}
            </Typography>
            <span className="text-[10px] text-slate-400 font-mono shrink-0">
              {new Date(packet.timestamp).toLocaleTimeString()}
            </span>
          </div>
          <div className="flex items-center gap-2 mt-0.5">
            <span className={cn(
              "text-[10px] px-1.5 py-0.5 rounded font-medium",
              packet.direction === 'outbound' ? "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400" : "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400"
            )}>
              {packet.direction.toUpperCase()}
            </span>
            {packet.durationMs && (
              <span className="text-[10px] text-slate-500">
                {packet.durationMs}ms
              </span>
            )}
          </div>
        </div>

        <Icon
          name={expanded ? "chevron-up" : "chevron-down"}
          size="sm"
          className="text-slate-400 shrink-0"
        />
      </div>

      {expanded && (
        <div className="p-3 pt-0 border-t border-slate-100 dark:border-slate-800">
          <pre className="text-[11px] font-mono bg-slate-100 dark:bg-slate-900 p-2 rounded overflow-x-auto text-slate-700 dark:text-slate-300">
            {JSON.stringify(packet.payload, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
};

export const Debugger: React.FC = () => {
  const { packets, isRecording, toggleRecording, clearPackets } = useDebuggerStore();

  return (
    <div className="flex flex-col h-full space-y-4">
      <div className="flex items-center justify-between flex-shrink-0">
        <Typography variant="h4" className="text-slate-800 dark:text-slate-100 flex items-center gap-2">
          <Icon name="tools" className="text-primary-600 dark:text-primary-400" />
          Live Inspector
        </Typography>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={clearPackets}
            title="Clear logs"
          >
            <Icon name="trash-2" size="sm" />
          </Button>
          <Button
            variant={isRecording ? "default" : "outline"}
            size="sm"
            onClick={toggleRecording}
            className={isRecording ? "bg-red-500 hover:bg-red-600 border-red-500 text-white" : ""}
          >
            <Icon name={isRecording ? "x" : "play"} size="sm" className="mr-1" />
            {isRecording ? "Recording" : "Paused"}
          </Button>
        </div>
      </div>

      <Card className="flex-1 overflow-hidden flex flex-col border-slate-200 dark:border-slate-700 dark:bg-slate-800 shadow-sm">
        <div className="bg-slate-50 dark:bg-slate-900 px-3 py-2 border-b border-slate-200 dark:border-slate-700 flex justify-between items-center shrink-0">
          <Typography variant="caption" className="font-semibold text-slate-500">
            TRAFFIC ({packets.length})
          </Typography>
          {isRecording && <span className="animate-pulse w-2 h-2 rounded-full bg-red-500" />}
        </div>

        <CardContent className="flex-1 p-0 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
          {packets.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-slate-400 space-y-2 p-6 text-center">
              <Icon name="search" size="lg" className="opacity-50" />
              <Typography variant="body">
                No tool executions intercepted yet.
              </Typography>
              <Typography variant="caption" className="opacity-70">
                Trigger a tool instruction or wait for the AI to issue an MCP tool call to see packet traces here.
              </Typography>
            </div>
          ) : (
            <div className="flex flex-col">
              {packets.map((packet) => (
                <PacketDetail key={packet.id} packet={packet} />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
