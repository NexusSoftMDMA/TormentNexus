import React, { useState } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { Icon } from '@src/components/sidebar/ui';
import { cn } from '@src/lib/utils';
import { isDarkTheme } from '@src/render_prescript/src/utils/themeDetector';

export interface ToolResultContent {
  type: string;
  text?: string;
  url?: string;
  code?: string;
  alt?: string;
  [key: string]: any;
}

export interface ToolResultWidgetProps {
  callId?: string;
  resultContent: string;
  timestamp?: number;
}

const formatJsonString = (str: string): { isJson: boolean; parsed: any; formatted: string } => {
  try {
    const parsed = JSON.parse(str);
    return { isJson: true, parsed, formatted: JSON.stringify(parsed, null, 2) };
  } catch {
    return { isJson: false, parsed: null, formatted: str };
  }
};

const JsonNode: React.FC<{ data: any; name?: string; level?: number; isLast?: boolean }> = ({ data, name, level = 0, isLast = true }) => {
  const [expanded, setExpanded] = useState(level < 2);
  const isObject = data !== null && typeof data === 'object';
  const isArray = Array.isArray(data);
  const isEmpty = isObject && Object.keys(data).length === 0;

  if (!isObject) {
    let valueClass = 'text-green-600 dark:text-green-400';
    if (typeof data === 'number') valueClass = 'text-blue-600 dark:text-blue-400';
    if (typeof data === 'boolean') valueClass = 'text-purple-600 dark:text-purple-400';
    if (data === null) valueClass = 'text-slate-400 italic';
    
    return (
      <div className="pl-4 font-mono text-xs hover:bg-black/5 dark:hover:bg-white/5 py-0.5 rounded flex group">
        {name && <span className="text-slate-600 dark:text-slate-300 mr-2">{name}:</span>}
        <span className={valueClass}>
          {typeof data === 'string' ? `"${data}"` : String(data)}
        </span>
        {!isLast && <span className="text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-300">,</span>}
      </div>
    );
  }

  const keys = Object.keys(data);
  const brackets = isArray ? ['[', ']'] : ['{', '}'];

  return (
    <div className="font-mono text-xs">
      <div 
        className="flex items-center hover:bg-black/5 dark:hover:bg-white/5 py-0.5 rounded cursor-pointer select-none"
        style={{ paddingLeft: `${Math.max(0, (level * 16) - 12)}px` }}
        onClick={(e) => { e.stopPropagation(); setExpanded(!expanded); }}
      >
        <span className="w-3 inline-block text-center mr-1 opacity-50 hover:opacity-100 flex-shrink-0">
          {!isEmpty && (expanded ? '▼' : '▶')}
        </span>
        {name && <span className="text-slate-600 dark:text-slate-300 mr-2">{name}:</span>}
        <span className="text-slate-500">{brackets[0]}</span>
        {!expanded && !isEmpty && <span className="text-slate-400 mx-1">{isArray ? `... ${keys.length} items` : '...'}</span>}
        {(!expanded || isEmpty) && <span className="text-slate-500">{brackets[1]}{!isLast && ','}</span>}
      </div>
      
      {expanded && !isEmpty && (
        <>
          <div className="pl-3 border-l border-black/10 dark:border-white/10 ml-[3px]">
            {keys.map((key, index) => (
              <JsonNode 
                key={key} 
                name={isArray ? undefined : key} 
                data={data[key as keyof typeof data]} 
                level={level + 1} 
                isLast={index === keys.length - 1} 
              />
            ))}
          </div>
          <div className="hover:bg-black/5 dark:hover:bg-white/5 py-0.5 rounded text-slate-500 flex" style={{ paddingLeft: `${level * 16}px` }}>
            {brackets[1]}{!isLast && ','}
          </div>
        </>
      )}
    </div>
  );
};

const ContentItem: React.FC<{ item: any }> = ({ item }) => {
  if (item.type === 'text') {
    return <div className="mb-2 whitespace-pre-wrap word-break shrink-0 font-sans text-[13px]">{item.text}</div>;
  }
  if (item.type === 'image' && item.url) {
    return (
      <div className="my-2 p-1 bg-black/5 dark:bg-white/5 rounded-md inline-block">
        <img src={item.url} alt={item.alt || 'Image result'} className="max-w-[300px] h-auto rounded" />
      </div>
    );
  }
  if (item.type === 'code' && item.code) {
    return (
      <div className="my-2 bg-slate-100 dark:bg-slate-800 rounded-md p-3 overflow-x-auto text-xs font-mono border border-slate-200 dark:border-slate-700">
        <pre className="!m-0">{item.code}</pre>
      </div>
    );
  }
  return (
    <pre className="my-1 whitespace-pre-wrap text-xs text-slate-500 font-mono">
      {JSON.stringify(item, null, 2)}
    </pre>
  );
};

export const ToolResultWidget: React.FC<ToolResultWidgetProps> = ({ callId, resultContent, timestamp }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [copied, setCopied] = useState(false);
  const isDark = isDarkTheme();

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(resultContent);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const { isJson, parsed, formatted } = formatJsonString(resultContent);

  const hasArrayContent = isJson && parsed && parsed.content && Array.isArray(parsed.content);

  return (
    <div
      className={cn(
        'w-full max-w-full my-3 rounded-lg border overflow-hidden transition-all shadow-sm',
        isDark ? 'bg-[#1b1b1b] border-white/10 text-slate-200' : 'bg-slate-50 border-black/10 text-slate-800'
      )}>
      {/* Header */}
      <div
        className={cn(
          'flex items-center justify-between px-3 py-2 cursor-pointer transition-colors',
          isDark ? 'hover:bg-white/5' : 'hover:bg-black/5'
        )}
        onClick={() => setIsExpanded(!isExpanded)}>
        <div className="flex items-center gap-2 overflow-hidden">
          <div className={cn(
            'p-1 rounded-md shrink-0 flex items-center justify-center', 
            isDark ? 'bg-primary-900/50 text-primary-400' : 'bg-primary-100 text-primary-600'
          )}>
            <Icon name="check" size="xs" />
          </div>
          <span className="font-semibold text-[13px] whitespace-nowrap">Tool Result</span>
          {callId && (
            <span className="text-[10px] font-mono opacity-50 truncate max-w-[100px]" title={callId}>
              {callId.substring(0, 12)}...
            </span>
          )}
        </div>
        
        <div className="flex items-center gap-2 shrink-0">
          {timestamp && (
            <span className="text-[10px] opacity-40">
              {new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
            </span>
          )}
          <button 
            type="button"
            className="p-1 rounded opacity-50 hover:opacity-100 hover:bg-black/5 dark:hover:bg-white/5 disabled:opacity-30 disabled:pointer-events-none"
            onClick={handleCopy}
            disabled={!resultContent}
            title="Copy raw result"
          >
            <Icon name={copied ? "check" : "copy"} size="xs" className={copied ? "text-green-500" : ""} />
          </button>
          <div className="p-1 rounded opacity-50 hover:opacity-100 hover:bg-black/5 dark:hover:bg-white/5 ml-1">
            <Icon name={isExpanded ? 'chevron-up' : 'chevron-down'} size="xs" />
          </div>
        </div>
      </div>

      {/* Expandable Content Area */}
      <AnimatePresence initial={false}>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2, ease: "easeInOut" }}
            className="overflow-hidden border-t border-black/5 dark:border-white/5 bg-white dark:bg-[#121212]"
          >
            <div className="p-3 overflow-x-auto max-h-[400px] overflow-y-auto custom-scrollbar">
              {hasArrayContent ? (
                <div className="flex flex-col gap-1">
                  {parsed.content.map((item: any, idx: number) => (
                    <ContentItem key={idx} item={item} />
                  ))}
                </div>
              ) : isJson ? (
                <div className="py-1 min-w-[300px]">
                  <JsonNode data={parsed} />
                </div>
              ) : (
                <pre className="text-xs font-mono !m-0 whitespace-pre-wrap break-words opacity-80">
                  {formatted}
                </pre>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
