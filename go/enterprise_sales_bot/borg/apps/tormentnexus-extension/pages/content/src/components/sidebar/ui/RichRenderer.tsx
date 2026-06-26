import React from 'react';
import { cn } from '@src/lib/utils';
import { Typography, Icon } from './';

interface RichRendererProps {
  data: any;
  className?: string;
  defaultExpanded?: boolean;
}

const isImage = (data: any): boolean => {
  if (typeof data !== 'string') return false;
  return data.startsWith('data:image') || data.match(/\.(jpeg|jpg|gif|png|webp)$/) !== null;
};

const isMarkdown = (data: any): boolean => {
  if (typeof data !== 'string') return false;
  // Simple heuristic: look for markdown syntax
  return data.includes('```') || data.includes('# ') || data.includes('**') || data.includes('- ');
};

const JsonTree: React.FC<{ data: any; level?: number }> = ({ data, level = 0 }) => {
  const [expanded, setExpanded] = React.useState(level < 2);

  if (data === null) return <span className="text-gray-500">null</span>;
  if (data === undefined) return <span className="text-gray-500">undefined</span>;

  if (typeof data !== 'object') {
    if (isImage(data)) {
      return (
        <div className="mt-1">
          <img
            src={data}
            alt="Tool Result"
            className="max-w-full h-auto rounded border border-slate-200 dark:border-slate-700 max-h-60 object-contain"
          />
        </div>
      );
    }
    const colorClass =
      typeof data === 'string'
        ? 'text-green-600 dark:text-green-400'
        : typeof data === 'number'
          ? 'text-blue-600 dark:text-blue-400'
          : typeof data === 'boolean'
            ? 'text-purple-600 dark:text-purple-400'
            : 'text-gray-600';

    // Render markdown strings in a pre block for readability
    if (typeof data === 'string' && isMarkdown(data)) {
      return (
        <div className="whitespace-pre-wrap font-mono text-xs p-2 bg-slate-50 dark:bg-slate-900 rounded border border-slate-100 dark:border-slate-800 mt-1">
          {data}
        </div>
      );
    }

    return <span className={cn(colorClass, 'break-words')}>{String(data)}</span>;
  }

  const isArray = Array.isArray(data);
  const keys = Object.keys(data);

  if (keys.length === 0) {
    return <span className="text-slate-500">{isArray ? '[]' : '{}'}</span>;
  }

  return (
    <div className="ml-1">
      <div
        className="flex items-center cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800/50 rounded px-1 -ml-1 select-none"
        onClick={() => setExpanded(!expanded)}>
        <Icon
          name="chevron-right"
          size="xs"
          className={cn('text-slate-400 mr-1 transition-transform', expanded ? 'rotate-90' : '')}
        />
        <span className="text-slate-600 dark:text-slate-400 text-xs font-medium">
          {isArray ? `Array(${keys.length})` : 'Object'}
        </span>
      </div>

      {expanded && (
        <div className="pl-3 border-l border-slate-200 dark:border-slate-700 ml-1">
          {keys.map(key => (
            <div key={key} className="my-0.5">
              <span className="text-slate-500 dark:text-slate-400 mr-1 text-xs">{key}:</span>
              <JsonTree data={data[key]} level={level + 1} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export const RichRenderer: React.FC<RichRendererProps> = ({ data, className }) => {
  // If result is a complex object with specific MCP structure (content blocks), handle it
  if (data && typeof data === 'object' && Array.isArray(data.content)) {
    return (
      <div className={cn('space-y-2', className)}>
        {data.content.map((block: any, index: number) => {
          if (block.type === 'text') {
            return (
              <div key={index} className="prose dark:prose-invert max-w-none text-xs">
                {/* Basic text rendering, could be upgraded to full markdown */}
                <div className="whitespace-pre-wrap font-mono p-2 bg-slate-50 dark:bg-slate-900 rounded border border-slate-200 dark:border-slate-700">
                  {block.text}
                </div>
              </div>
            );
          }
          if (block.type === 'image') {
            return (
              <div key={index} className="mt-2">
                <img
                  src={`data:${block.mimeType};base64,${block.data}`}
                  alt="Tool Output"
                  className="max-w-full rounded-lg border border-slate-200 dark:border-slate-700"
                />
              </div>
            );
          }
          return <JsonTree key={index} data={block} />;
        })}
      </div>
    );
  }

  // Fallback to JSON Tree
  return (
    <div className={cn('font-mono text-xs overflow-x-auto', className)}>
      <JsonTree data={data} />
    </div>
  );
};
