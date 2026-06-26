"use client";

import React, { useEffect, useState } from 'react';
import { ScrollArea } from '@tormentnexus/ui';
import { Card, CardContent, CardHeader, CardTitle } from '@tormentnexus/ui';
import { Badge } from '@tormentnexus/ui';

interface AuditLog {
  timestamp: number;
  type: string;
  actor: string;
  resource: string;
  action: string;
  outcome: string;
  metadata?: any;
}

export function AuditLogViewer() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/audit-logs')
      .then(res => res.json())
      .then(data => {
        setLogs(data.logs || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch logs:', err);
        setLoading(false);
      });
  }, []);

  return (
    <Card className="w-full h-full bg-slate-950 text-slate-50 border-slate-800">
      <CardHeader>
        <CardTitle className="text-xl font-bold flex items-center justify-between">
          <span>Enterprise Audit Trail</span>
          <Badge variant="outline" className="text-slate-400">Phase 13</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[500px] w-full pr-4">
          <div className="space-y-2">
            {loading ? (
              <div className="text-center py-10 text-slate-500">Loading audit history...</div>
            ) : logs.length === 0 ? (
              <div className="text-center py-10 text-slate-500">No audit logs found.</div>
            ) : (
              logs.map((log, i) => (
                <div key={i} className="p-3 rounded bg-slate-900 border border-slate-800 hover:border-slate-700 transition-colors">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs font-mono text-slate-500">
                      {new Date(log.timestamp).toLocaleString()}
                    </span>
                    <Badge variant={log.outcome === 'success' ? 'default' : 'destructive'} className="text-[10px] uppercase">
                      {log.outcome}
                    </Badge>
                  </div>
                  <div className="flex gap-2 items-center">
                    <span className="text-blue-400 font-semibold">{log.actor}</span>
                    <span className="text-slate-400">performed</span>
                    <span className="text-emerald-400 font-mono">{log.action}</span>
                    <span className="text-slate-400">on</span>
                    <span className="text-purple-400">{log.resource}</span>
                  </div>
                  {log.metadata && (
                    <pre className="mt-2 text-[10px] bg-black p-2 rounded overflow-hidden text-slate-400">
                      {JSON.stringify(log.metadata, null, 2)}
                    </pre>
                  )}
                </div>
              ))
            )}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
