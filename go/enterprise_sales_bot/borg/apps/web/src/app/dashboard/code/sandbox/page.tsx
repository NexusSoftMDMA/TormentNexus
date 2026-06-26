"use client";

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Code, Loader2, Play, Shield, Zap, Terminal, Cpu } from "lucide-react";

type ExecResult = {
  exitCode: number;
  stdout: string;
  stderr: string;
  duration: string;
  language: string;
  timedOut?: boolean;
  error?: string;
  sandboxed?: boolean;
  wasmSize?: number;
};

const CODE_TEMPLATES: Record<string, string> = {
  go: `package main

import "fmt"

func main() {
\tfmt.Println("Hello from WASM sandbox!")
\tfor i := 1; i <= 5; i++ {
\t\tfmt.Printf("Count: %d\\n", i)
\t}
}`,
  javascript: `// JavaScript execution
console.log("Hello from Node.js sandbox!");
for (let i = 1; i <= 5; i++) {
  console.log(\`Count: \${i}\`);
}`,
  python: `# Python execution
print("Hello from Python sandbox!")
for i in range(1, 6):
    print(f"Count: {i}")`,
  typescript: `// TypeScript execution
const greet = (name: string): string => \`Hello, \${name}!\`;
console.log(greet("WASM Sandbox"));
`,
  shell: `echo "Hello from shell sandbox!"
for i in 1 2 3 4 5; do
  echo "Count: $i"
done`,
};

const LANGUAGES = [
  { value: 'go', label: 'Go (WASM)', icon: Shield, wasmOnly: true },
  { value: 'javascript', label: 'JavaScript', icon: Zap },
  { value: 'typescript', label: 'TypeScript', icon: Code },
  { value: 'python', label: 'Python', icon: Terminal },
  { value: 'shell', label: 'Shell', icon: Cpu },
];

export default function CodeSandboxPage() {
  const [language, setLanguage] = useState('go');
  const [code, setCode] = useState(CODE_TEMPLATES.go);
  const [result, setResult] = useState<ExecResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [wasmStatus, setWasmStatus] = useState<Record<string, unknown> | null>(null);

  const selectedLang = LANGUAGES.find(l => l.value === language);
  const isWASM = selectedLang?.wasmOnly ?? false;

  const handleLanguageChange = (lang: string) => {
    setLanguage(lang);
    setCode(CODE_TEMPLATES[lang] || '');
    setResult(null);
    setError(null);
  };

  const handleExecute = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const endpoints = [
        '/api/go/code/wasm/exec',
        '/api/go/code/exec',
      ];

      // For non-WASM languages, use the process sandbox
      const endpoint = isWASM ? endpoints[0] : endpoints[1];

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          language: language,
          code: code,
          timeoutSec: 15,
          maxMemoryMB: 64,
        }),
        signal: AbortSignal.timeout(30000),
      });

      const data = await response.json();

      if (data.success) {
        setResult(data.data as ExecResult);
      } else {
        setError(data.error || 'Execution failed');
        if (data.hint) {
          setError(prev => prev ? `${prev}\n\n💡 ${data.hint}` : prev);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reach sandbox');
    } finally {
      setLoading(false);
    }
  };

  const handleCheckWASMStatus = async () => {
    try {
      const response = await fetch('/api/go/code/wasm/status', {
        signal: AbortSignal.timeout(5000),
      });
      const data = await response.json();
      setWasmStatus(data.data as Record<string, unknown>);
    } catch {
      setWasmStatus({ available: false, error: 'Failed to check status' });
    }
  };

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white">Code Sandbox</h1>
        <p className="text-zinc-500 mt-1">
          Execute code in isolated sandboxes — WASM for Go, process-based for other languages
        </p>
      </div>

      {/* Language Selector */}
      <div className="flex flex-wrap gap-2">
        {LANGUAGES.map((lang) => {
          const Icon = lang.icon;
          const isActive = language === lang.value;
          return (
            <button
              key={lang.value}
              type="button"
              onClick={() => handleLanguageChange(lang.value)}
              className={`rounded-lg border px-4 py-2 text-sm flex items-center gap-2 transition-colors ${
                isActive
                  ? 'border-blue-500/50 bg-blue-500/15 text-blue-200'
                  : 'border-zinc-700 bg-zinc-950/70 text-zinc-300 hover:bg-zinc-800'
              }`}
            >
              <Icon className="h-4 w-4" />
              {lang.label}
              {lang.wasmOnly ? (
                <span className="text-[10px] bg-emerald-500/10 border border-emerald-500/20 px-1.5 py-0.5 rounded text-emerald-300 uppercase tracking-wider">
                  sandboxed
                </span>
              ) : null}
            </button>
          );
        })}
      </div>

      <div className="grid gap-6 xl:grid-cols-[1fr_380px]">
        {/* Code Editor */}
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader className="pb-3 border-b border-zinc-800">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm text-white flex items-center gap-2">
                <Code className="h-4 w-4 text-blue-400" />
                {selectedLang?.label || language}
              </CardTitle>
              <div className="flex items-center gap-2">
                {isWASM && (
                  <span className="flex items-center gap-1 text-xs text-emerald-300">
                    <Shield className="h-3.5 w-3.5" /> WASM Sandbox
                  </span>
                )}
                <Button
                  onClick={handleExecute}
                  disabled={loading || !code.trim()}
                  className="bg-blue-600 hover:bg-blue-500 text-white"
                >
                  {loading ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Play className="mr-2 h-4 w-4" />
                  )}
                  Run
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <textarea
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className="w-full bg-zinc-950 text-zinc-100 font-mono text-sm p-4 min-h-[300px] resize-y focus:outline-none"
              spellCheck={false}
              placeholder="Write your code here..."
            />
          </CardContent>
        </Card>

        {/* Results Panel */}
        <div className="space-y-4">
          {error && (
            <Card className="bg-zinc-900 border-red-500/30">
              <CardContent className="p-4 text-red-300 text-sm whitespace-pre-wrap break-words">
                <div className="font-semibold mb-1">Error</div>
                {error}
              </CardContent>
            </Card>
          )}

          {result && (
            <Card className="bg-zinc-900 border-zinc-800">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm text-white flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    Result
                    {result.sandboxed && (
                      <span className="text-[10px] bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded text-emerald-300 uppercase tracking-wider">
                        WASM sandboxed
                      </span>
                    )}
                  </span>
                  <span className={`text-xs ${result.exitCode === 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                    Exit {result.exitCode} · {result.duration}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {result.stdout && (
                  <div>
                    <div className="text-xs uppercase tracking-wider text-zinc-500 mb-1">Stdout</div>
                    <pre className="bg-zinc-950 border border-zinc-800 rounded p-3 text-sm text-emerald-300 font-mono whitespace-pre-wrap overflow-auto max-h-48">
                      {result.stdout}
                    </pre>
                  </div>
                )}
                {result.stderr && (
                  <div>
                    <div className="text-xs uppercase tracking-wider text-zinc-500 mb-1">Stderr</div>
                    <pre className="bg-zinc-950 border border-zinc-800 rounded p-3 text-sm text-amber-300 font-mono whitespace-pre-wrap overflow-auto max-h-32">
                      {result.stderr}
                    </pre>
                  </div>
                )}
                {result.wasmSize ? (
                  <div className="text-xs text-zinc-500">
                    WASM binary: {(result.wasmSize / 1024).toFixed(1)} KB
                  </div>
                ) : null}
              </CardContent>
            </Card>
          )}

          {/* WASM Status */}
          <Card className="bg-zinc-900 border-zinc-800">
            <CardHeader className="pb-3">
              <CardTitle className="text-sm text-white flex items-center gap-2">
                <Shield className="h-4 w-4 text-emerald-400" />
                WASM Sandbox Status
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Button
                variant="outline"
                className="border-zinc-700 text-zinc-300 hover:bg-zinc-800 w-full mb-3"
                onClick={handleCheckWASMStatus}
              >
                Check Availability
              </Button>
              {wasmStatus && (
                <pre className="text-xs text-zinc-400 bg-zinc-950 border border-zinc-800 rounded p-3 whitespace-pre-wrap overflow-auto max-h-48">
                  {JSON.stringify(wasmStatus, null, 2)}
                </pre>
              )}
            </CardContent>
          </Card>

          {/* Security Notice */}
          <Card className="bg-zinc-900 border-zinc-800">
            <CardContent className="p-4 space-y-2 text-xs text-zinc-500">
              <div className="flex items-center gap-2 text-emerald-300 font-semibold">
                <Shield className="h-4 w-4" />
                Security Model
              </div>
              <p><span className="text-white font-medium">WASM Sandbox (Go):</span> No filesystem, no network, memory-limited, timeout-enforced. Code runs in WebAssembly with wasmtime/wasmer runtime.</p>
              <p><span className="text-white font-medium">Process Sandbox (JS/TS/Python/Shell):</span> Process-level isolation with configurable timeouts. Runs in the host OS with the respective runtime.</p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
