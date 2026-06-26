import React, { useState } from 'react';
import { useContextStore, type ContextItem } from '@src/stores';
import { Button, Icon, Typography, Textarea, Input } from '../ui';
import { useToastStore } from '@src/stores';
import { cn } from '@src/lib/utils';
import { Card } from '@src/components/ui/card';

interface ContextManagerProps {
  onInsert: (content: string) => void;
  onClose: () => void;
  initialContent?: string;
}

const ContextManager: React.FC<ContextManagerProps> = ({ onInsert, onClose, initialContent }) => {
  const { contexts, addContext, updateContext, deleteContext } = useContextStore();
  const { addToast } = useToastStore();
  const [editingId, setEditingId] = useState<string | null>(null);
  // Auto-start creation if initialContent is provided
  const [isCreating, setIsCreating] = useState(!!initialContent);
  const [editName, setEditName] = useState('');
  const [editContent, setEditContent] = useState(initialContent || '');

  const handleCreate = () => {
    setEditingId(null);
    setEditName('');
    setEditContent('');
    setIsCreating(true);
  };

  const handleEdit = (item: ContextItem) => {
    setEditingId(item.id);
    setEditName(item.name);
    setEditContent(item.content);
    setIsCreating(true);
  };

  const handleSave = () => {
    if (!editContent.trim()) {
      addToast({
        title: 'Error',
        message: 'Content cannot be empty',
        type: 'error',
        duration: 3000,
      });
      return;
    }

    const name = editName.trim() || (editContent.length > 30 ? editContent.substring(0, 30) + '...' : editContent);

    if (editingId) {
      updateContext(editingId, { name, content: editContent });
      addToast({
        title: 'Updated',
        message: 'Context updated',
        type: 'success',
        duration: 2000,
      });
    } else {
      addContext(editContent, name);
      addToast({
        title: 'Saved',
        message: 'New context saved',
        type: 'success',
        duration: 2000,
      });
    }
    setIsCreating(false);
    setEditingId(null);
  };

  const handleDelete = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm('Delete this context?')) {
      deleteContext(id);
    }
  };

  return (
    <div className="absolute inset-0 bg-white dark:bg-slate-900 z-50 flex flex-col animate-in slide-in-from-bottom-5 duration-200 shadow-xl border-t border-slate-200 dark:border-slate-700">
      <div className="flex items-center justify-between p-3 border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800">
        <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100 flex items-center gap-2">
          <Icon name="file-text" size="sm" />
          Context Library
        </Typography>
        <Button variant="ghost" size="icon" onClick={onClose}>
          <Icon name="x" size="sm" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {isCreating ? (
          <div className="space-y-3 animate-in fade-in zoom-in-95 duration-200">
            <div>
              <label className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-1 block">Name (Optional)</label>
              <Input
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                placeholder="e.g., Project Overview"
                className="w-full"
                autoFocus
              />
            </div>
            <div>
              <label className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-1 block">Content</label>
              <Textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                placeholder="Paste your context here..."
                className="w-full h-40 font-mono text-xs"
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="ghost" size="sm" onClick={() => setIsCreating(false)}>
                Cancel
              </Button>
              <Button size="sm" onClick={handleSave} className="bg-primary-600 hover:bg-primary-700 text-white">
                Save
              </Button>
            </div>
          </div>
        ) : (
          <>
            <Button onClick={handleCreate} size="sm" className="w-full mb-2 flex items-center justify-center gap-1 border-dashed border-2 border-slate-200 dark:border-slate-700 bg-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-800 h-12">
              <Icon name="plus" size="sm" />
              Add New Context
            </Button>

            {contexts.length === 0 ? (
              <div className="text-center py-8 text-slate-400 dark:text-slate-500">
                <Icon name="box" size="sm" className="text-slate-400 dark:text-slate-500 mr-2" />
                <Typography variant="body" className="text-sm">No saved contexts</Typography>
              </div>
            ) : (
              <div className="space-y-2">
                {contexts.map((ctx) => (
                  <Card key={ctx.id} className="group border hover:border-primary-300 dark:hover:border-primary-700 transition-colors">
                    <div className="p-3">
                      <div className="flex justify-between items-start mb-2">
                        <Typography variant="subtitle" className="font-medium text-slate-800 dark:text-slate-200 truncate pr-2">
                          {ctx.name}
                        </Typography>
                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => handleEdit(ctx)}
                            className="p-1 hover:bg-slate-100 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-blue-600"
                            title="Edit"
                          >
                            <Icon name="info" size="sm" />
                          </button>
                          <button
                            onClick={(e) => handleDelete(ctx.id, e)}
                            className="p-1 hover:bg-slate-100 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-red-600"
                            title="Delete"
                          >
                            <Icon name="x" size="sm" />
                          </button>
                        </div>
                      </div>
                      <Typography variant="body" className="text-xs text-slate-500 dark:text-slate-400 line-clamp-2 font-mono bg-slate-50 dark:bg-slate-800/50 p-1.5 rounded border border-slate-100 dark:border-slate-800">
                        {ctx.content}
                      </Typography>
                      <div className="mt-2 flex justify-end">
                        <Button
                          size="sm"
                          variant="outline"
                          className="bg-primary-50 hover:bg-primary-100 text-primary-700 dark:bg-primary-900/20 dark:hover:bg-primary-900/40 dark:text-primary-300 border-primary-100 dark:border-primary-800"
                          onClick={() => {
                            onInsert(ctx.content);
                            onClose();
                          }}
                        >
                          <Icon name="plus" size="xs" className="mr-1" />
                          Insert
                        </Button>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default ContextManager;
