"use client";

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../../components/ui/card';
import { Loader2, Code2 } from 'lucide-react';

export default function CodeExecutorPanel() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);

  const testExecutor = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/native/codeexec/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ language: 'javascript', code: 'console.log("Hello from Go Sandbox")' })
      });
      const data = await res.json();
      setResult(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center space-x-3">
        <Code2 className="w-8 h-8 text-primary" />
        <h1 className="text-2xl font-bold tracking-tight">Native Code Sandbox</h1>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Test Execution</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Dispatches a test JavaScript script to the Go `CodeModeEngine`.
          </p>
          <button
            onClick={testExecutor}
            disabled={loading}
            className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
          >
            {loading ? 'Executing...' : 'Run Test'}
          </button>

          {result && (
            <div className="mt-4 p-4 bg-muted border rounded">
              <pre className="text-xs font-mono">{JSON.stringify(result, null, 2)}</pre>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
