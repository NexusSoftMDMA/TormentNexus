"use client";

import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../../components/ui/card';
import { Loader2 } from 'lucide-react';

export default function DecisionSystemPanel() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/mcp/decision/list-loaded')
      .then(res => res.json())
      .then(json => {
        setData(json);
        setLoading(false);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  }, []);

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">MCP Decision System</h1>
      <Card>
        <CardHeader>
          <CardTitle>Loaded Tools (Go-Native)</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center p-4">
              <Loader2 className="animate-spin w-6 h-6 text-muted-foreground" />
            </div>
          ) : data && data.tools ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {data.tools.map((t: any, i: number) => (
                <div key={i} className="border rounded p-4 shadow-sm bg-card text-card-foreground">
                  <p className="font-semibold text-lg">{t.AdvertisedName || t.name}</p>
                  <p className="text-sm text-muted-foreground mt-1 line-clamp-2">{t.Description || t.description}</p>
                  <div className="mt-4 flex justify-between text-xs text-muted-foreground font-mono">
                    <span>Rank: {t.score || 'Auto'}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-center p-4">No tools loaded or failed to connect to Go sidecar.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
