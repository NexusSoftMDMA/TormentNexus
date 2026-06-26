import React, { useState, useRef } from 'react';
import { useMacroStore, type Macro } from '@src/stores';
import { Icon, Typography, Button } from '../ui';
import { Card } from '@src/components/ui/card';
import MacroBuilder from './MacroBuilder';
import { MacroRunner } from '@src/lib/macro.runner';
import { useToastStore } from '@src/stores';
import { cn } from '@src/lib/utils';

interface MacroListProps {
  onExecute: (toolName: string, args: any) => Promise<any>;
}

const MacroList: React.FC<MacroListProps> = ({ onExecute }) => {
  const { macros, addMacro, deleteMacro } = useMacroStore();
  const { addToast } = useToastStore();
  const [editingMacro, setEditingMacro] = useState<Macro | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [runningMacroId, setRunningMacroId] = useState<string | null>(null);
  const [importUrl, setImportUrl] = useState('');
  const [showImportUrlInput, setShowImportUrlInput] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDelete = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm('Are you sure you want to delete this macro?')) {
      deleteMacro(id);
    }
  };

  const handleEdit = (macro: Macro) => {
    setEditingMacro(macro);
    setIsCreating(false);
  };

  const handleCreate = () => {
    setEditingMacro(null);
    setIsCreating(true);
  };

  const handleCloseBuilder = () => {
    setEditingMacro(null);
    setIsCreating(false);
  };

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  const handleImportUrlClick = () => {
    setShowImportUrlInput(!showImportUrlInput);
  };

  const handleImportFromUrl = async () => {
    if (!importUrl.trim()) return;

    try {
      addToast({
        title: 'Importing...',
        message: 'Fetching macro from URL...',
        type: 'info',
        duration: 2000,
      });

      const response = await fetch(importUrl);
      if (!response.ok) throw new Error('Failed to fetch macro');

      const macroData = await response.json();

      // Basic validation for V2 macros
      if (!macroData.name || !Array.isArray(macroData.nodes)) {
        throw new Error('Invalid macro file format (must be V2 graph format with nodes/edges)');
      }

      addMacro({
        name: macroData.name + ' (Imported)',
        description: macroData.description || '',
        nodes: macroData.nodes,
        edges: macroData.edges || [],
      });

      addToast({
        title: 'Import Successful',
        message: `Imported "${macroData.name}"`,
        type: 'success',
        duration: 3000,
      });
      setImportUrl('');
      setShowImportUrlInput(false);
    } catch (error) {
      addToast({
        title: 'Import Failed',
        message: error instanceof Error ? error.message : 'Unknown error',
        type: 'error',
        duration: 5000,
      });
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      try {
        const content = event.target?.result as string;
        const macroData = JSON.parse(content);

        // Basic validation for V2 macros
        if (!macroData.name || !Array.isArray(macroData.nodes)) {
          throw new Error('Invalid macro file format (must be V2 graph format with nodes/edges)');
        }

        addMacro({
          name: macroData.name + ' (Imported)',
          description: macroData.description || '',
          nodes: macroData.nodes,
          edges: macroData.edges || [],
        });

        addToast({
          title: 'Import Successful',
          message: `Imported "${macroData.name}"`,
          type: 'success',
          duration: 3000,
        });
      } catch (error) {
        addToast({
          title: 'Import Failed',
          message: error instanceof Error ? error.message : 'Unknown error',
          type: 'error',
          duration: 5000,
        });
      } finally {
        // Reset input
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      }
    };
    reader.readAsText(file);
  };

  const handleRun = async (macro: Macro, e: React.MouseEvent) => {
    e.stopPropagation();
    if (runningMacroId) return;

    setRunningMacroId(macro.id);
    addToast({
      title: 'Starting Macro',
      message: `Running "${macro.name}"...`,
      type: 'info',
      duration: 2000,
    });

    const runner = new MacroRunner(onExecute, (msg, type) => {
      // Optional: more granular feedback could go here
      // For now we rely on the runner throwing or completing
      console.log(`[Macro: ${macro.name}] [${type}] ${msg}`);
    });

    try {
      await runner.run(macro);
      addToast({
        title: 'Macro Completed',
        message: `"${macro.name}" finished successfully.`,
        type: 'success',
        duration: 3000,
      });
    } catch (error) {
      addToast({
        title: 'Macro Failed',
        message: error instanceof Error ? error.message : String(error),
        type: 'error',
        duration: 5000,
      });
    } finally {
      setRunningMacroId(null);
    }
  };

  if (isCreating || editingMacro) {
    return (
      <MacroBuilder
        existingMacro={editingMacro}
        onClose={handleCloseBuilder}
      />
    );
  }

  return (
    <div className="flex flex-col h-full space-y-4 p-4">
      <div className="flex justify-between items-center mb-2">
        <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100">
          Macros
        </Typography>
        <div className="flex items-center gap-2">
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileChange}
            accept=".json"
            className="hidden"
          />
          <div className="flex gap-1">
            <Button onClick={handleImportUrlClick} size="sm" variant="ghost" className="h-8 w-8 p-0" title="Import from URL">
              <Icon name="server" size="xs" />
            </Button>
            <Button onClick={() => window.open('https://mcp.localhost', '_blank')} size="sm" variant="outline" className="flex items-center gap-1 hidden sm:flex" title="Browse Community Macros">
              <Icon name="server" size="sm" />
              Community
            </Button>
            <Button onClick={handleImportClick} size="sm" variant="outline" className="flex items-center gap-1 hidden sm:flex" title="Import from JSON">
              <Icon name="cloud" size="sm" className="text-slate-400 dark:text-slate-500" />
              Import
            </Button>
          </div>
          <Button onClick={handleCreate} size="sm" className="flex items-center gap-1">
            <Icon name="plus" size="xs" />
            New
          </Button>
        </div>
      </div>

      {showImportUrlInput && (
        <div className="flex gap-2 mb-2 animate-in slide-in-from-top-2">
          <input
            type="text"
            value={importUrl}
            onChange={(e) => setImportUrl(e.target.value)}
            placeholder="https://example.com/macro.json"
            className="flex-1 px-3 py-1.5 text-xs border rounded bg-white dark:bg-slate-800 border-slate-300 dark:border-slate-600 focus:ring-1 focus:ring-primary-500"
            onKeyDown={(e) => e.key === 'Enter' && handleImportFromUrl()}
          />
          <Button size="sm" onClick={handleImportFromUrl} disabled={!importUrl}>
            Go
          </Button>
        </div>
      )}

      <div className="flex-1 overflow-y-auto min-h-0 space-y-3 pr-1 scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
        {macros.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-40 text-slate-400 dark:text-slate-500 border-2 border-dashed border-slate-200 dark:border-slate-700 rounded-lg">
            <Icon name="braces" size="lg" className="mb-2 opacity-50" />
            <Typography variant="body" className="text-sm">
              No macros created yet
            </Typography>
            <div className="flex gap-2 mt-3">
              <Button variant="outline" size="sm" onClick={handleImportClick}>
                Import Macro
              </Button>
              <Button variant="ghost" size="sm" onClick={handleCreate} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                Create Macro
              </Button>
            </div>
          </div>
        ) : (
          macros.map((macro) => (
            <Card
              key={macro.id}
              className={cn(
                "border overflow-hidden transition-all duration-200 hover:shadow-md cursor-pointer group",
                runningMacroId === macro.id ? "ring-2 ring-primary-500 border-transparent" : ""
              )}
              onClick={() => handleEdit(macro)}
            >
              <div className="p-3 bg-white dark:bg-slate-900 flex items-start justify-between">
                <div className="flex items-start gap-3 overflow-hidden w-full">
                  <div
                    className={cn(
                      "p-2 rounded-full flex-shrink-0 transition-colors z-10",
                      runningMacroId === macro.id
                        ? "bg-primary-100 text-primary-700 animate-pulse"
                        : "bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400 hover:bg-primary-100 dark:hover:bg-primary-900/40"
                    )}
                    onClick={(e) => handleRun(macro, e)}
                    title="Run Macro"
                  >
                    <Icon name={runningMacroId === macro.id ? "refresh" : "play"} size="sm" className={runningMacroId === macro.id ? "animate-spin" : ""} />
                  </div>
                  <div className="min-w-0 flex-1">
                    <Typography variant="subtitle" className="font-medium text-slate-800 dark:text-slate-200 truncate">
                      {macro.name}
                    </Typography>
                    <Typography variant="body" className="text-xs text-slate-500 dark:text-slate-400 line-clamp-1 mt-0.5">
                      {macro.description || 'No description'}
                    </Typography>
                    <div className="mt-2 flex items-center gap-2">
                      <span className="text-[10px] bg-slate-100 dark:bg-slate-800 text-slate-500 px-1.5 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                        {macro.nodes?.length || 0} nodes
                      </span>
                      <span className="text-[10px] text-slate-400">
                        Updated {new Date(macro.updatedAt).toLocaleDateString()}
                      </span>
                    </div>
                  </div>

                  <div className="flex flex-col gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
                      onClick={(e) => handleDelete(macro.id, e)}
                      title="Delete Macro"
                    >
                      <Icon name="x" size="xs" />
                    </Button>
                  </div>
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  );
};

export default MacroList;
