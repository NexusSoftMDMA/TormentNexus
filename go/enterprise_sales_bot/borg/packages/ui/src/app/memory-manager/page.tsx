"use client";

import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../../components/ui/card';
import { Loader2, Database } from 'lucide-react';

export default function MemoryManagerPanel() {
  const [data, setData] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/native/memory/get')
      .then(res => res.json())
      .then(json => {
        setData(json || []);
        setLoading(false);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  }, []);

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center space-x-3">
        <Database className="w-8 h-8 text-primary" />
        <h1 className="text-2xl font-bold tracking-tight">Go Memory Manager</h1>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Stored Tiers</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center p-4">
              <Loader2 className="animate-spin w-6 h-6 text-muted-foreground" />
            </div>
          ) : data.length > 0 ? (
            <ul className="list-disc list-inside space-y-2">
              {data.map((mem, i) => (
                <li key={i} className="text-sm font-mono bg-muted p-2 rounded">{mem}</li>
              ))}
            </ul>
          ) : (
            <p className="text-muted-foreground text-center p-4">Memory manager is initialized but empty.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
