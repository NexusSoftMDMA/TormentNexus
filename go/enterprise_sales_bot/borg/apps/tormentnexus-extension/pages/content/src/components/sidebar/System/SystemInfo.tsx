import React from 'react';
import { Typography, Icon } from '../ui';
import { Card, CardContent, CardHeader, CardTitle } from '@src/components/ui/card';

const SystemInfo: React.FC = () => {
  const buildDate = new Date().toLocaleString();
  const version = "1.1.0"; // Should match VERSION file

  return (
    <div className="flex flex-col h-full space-y-4 p-4 overflow-y-auto">
      <div className="flex flex-col space-y-2">
        <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100">
          System Status
        </Typography>
        <Typography variant="caption" className="text-slate-500 dark:text-slate-400">
          Build information and project structure.
        </Typography>
      </div>

      <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
        <CardHeader className="pb-2 border-b border-slate-100 dark:border-slate-800">
          <CardTitle className="text-sm font-semibold text-slate-700 dark:text-slate-300">Version Info</CardTitle>
        </CardHeader>
        <CardContent className="p-4 space-y-2">
          <div className="flex justify-between items-center">
            <span className="text-sm text-slate-600 dark:text-slate-400">Version</span>
            <span className="text-sm font-mono font-medium text-slate-900 dark:text-slate-100">{version}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-slate-600 dark:text-slate-400">Build Date</span>
            <span className="text-sm font-mono text-slate-900 dark:text-slate-100">{buildDate}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-slate-600 dark:text-slate-400">Environment</span>
            <span className="text-sm font-mono text-slate-900 dark:text-slate-100">{(import.meta as any).env.MODE}</span>
          </div>
        </CardContent>
      </Card>

      <Card className="border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
        <CardHeader className="pb-2 border-b border-slate-100 dark:border-slate-800">
          <CardTitle className="text-sm font-semibold text-slate-700 dark:text-slate-300">Project Structure</CardTitle>
        </CardHeader>
        <CardContent className="p-4">
          <div className="text-xs font-mono bg-slate-50 dark:bg-slate-950 p-2 rounded border border-slate-100 dark:border-slate-800 overflow-x-auto whitespace-pre">
{`tormentnexus-extension/
├── chrome-extension/       (v1.1.0)
│   ├── background/         (Service Worker)
│   └── mcpclient/          (Protocol Layer)
├── pages/
│   └── content/            (Sidebar UI - v1.1.0)
│       └── src/
│           ├── components/ (React Components)
│           ├── stores/     (Zustand Stores)
│           └── lib/        (Logic & Services)
├── packages/               (Shared Libraries)
│   ├── shared/             (Utils, Logger)
│   └── storage/            (Chrome Storage)
└── docs/                   (Documentation)`}
          </div>
        </CardContent>
      </Card>

      <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-100 dark:border-blue-800">
        <div className="flex items-start gap-2">
          <Icon name="info" size="sm" className="text-blue-500 mt-0.5" />
          <Typography variant="body" className="text-xs text-blue-700 dark:text-blue-300">
            This dashboard reflects the current build configuration. To update submodules, please use the project build scripts.
          </Typography>
        </div>
      </div>
    </div>
  );
};

export default SystemInfo;
