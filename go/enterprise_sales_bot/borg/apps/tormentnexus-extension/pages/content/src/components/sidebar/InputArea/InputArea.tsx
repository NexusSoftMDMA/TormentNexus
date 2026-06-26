import type React from 'react';
import { useState, useEffect } from 'react';
import { cn } from '@src/lib/utils';
import { Icon, Button } from '../ui';
import { useToastStore } from '@src/stores';
import ContextManager from '../ContextManager/ContextManager';

interface InputAreaProps {
  onSubmit: (text: string) => void;
  onToggleMinimize: () => void;
}

import { eventBus } from '@src/events/event-bus';

const InputArea: React.FC<InputAreaProps> = ({ onSubmit, onToggleMinimize }) => {
  const [inputText, setInputText] = useState('');
  const [selectedText, setSelectedText] = useState('');
  const [isListening, setIsListening] = useState(false);
  const [showContextManager, setShowContextManager] = useState(false);
  const [contextManagerInitialContent, setContextManagerInitialContent] = useState<string>('');
  const { addToast } = useToastStore.getState();

  // Listen for selection changes on the page
  useEffect(() => {
    const handleSelectionChange = () => {
      const selection = window.getSelection();
      if (selection && selection.toString().trim().length > 0) {
        setSelectedText(selection.toString().trim());
      } else {
        setSelectedText('');
      }
    };

    document.addEventListener('selectionchange', handleSelectionChange);

    // Listen for context save events from sidebar/background
    const unsubscribeContextSave = eventBus.on('context:save', (data) => {
      if (data.openManager === false) {
        return;
      }

      setContextManagerInitialContent(data.content);
      setShowContextManager(true);
    });

    return () => {
      document.removeEventListener('selectionchange', handleSelectionChange);
      unsubscribeContextSave();
    };
  }, []);

  const handleSubmit = () => {
    if (!inputText.trim()) return;
    onSubmit(inputText);
    setInputText('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleImportSelection = () => {
    if (selectedText) {
      setInputText(prev => {
        const prefix = prev ? prev + '\n\n' : '';
        return prefix + `Context: """\n${selectedText}\n"""`;
      });
      addToast({
        title: 'Context Added',
        message: 'Selected text added to input',
        type: 'success',
        duration: 2000,
      });
    }
  };

  const handleInsertContext = (content: string) => {
    setInputText(prev => {
      const prefix = prev ? prev + '\n\n' : '';
      return prefix + content;
    });
    setShowContextManager(false);
  };

  const toggleListening = () => {
    if (isListening) {
      return;
    }

    if (!('webkitSpeechRecognition' in window)) {
      addToast({
        title: 'Not Supported',
        message: 'Voice input is not supported in this browser.',
        type: 'error',
        duration: 3000,
      });
      return;
    }

    const recognition = new (window as any).webkitSpeechRecognition();
    recognition.continuous = false;
    recognition.interimResults = false;
    recognition.lang = 'en-US';

    recognition.onstart = () => {
      setIsListening(true);
      addToast({
        title: 'Listening...',
        message: 'Speak now.',
        type: 'info',
        duration: 2000,
      });
    };

    recognition.onend = () => setIsListening(false);

    recognition.onerror = (event: any) => {
      setIsListening(false);
      addToast({
        title: 'Error',
        message: `Voice input error: ${event.error}`,
        type: 'error',
        duration: 3000,
      });
    };

    recognition.onresult = (event: any) => {
      const transcript = event.results[0][0].transcript;
      if (transcript) {
        setInputText(prev => prev + (prev ? ' ' : '') + transcript);
      }
    };

    try {
      recognition.start();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="p-3 relative">
      {/* Context Manager Overlay */}
      {showContextManager && (
        <div className="absolute bottom-full left-0 right-0 h-[400px] mb-2 z-50 shadow-2xl rounded-t-lg overflow-hidden border border-slate-200 dark:border-slate-700">
          <ContextManager
            onInsert={handleInsertContext}
            onClose={() => {
              setShowContextManager(false);
              setContextManagerInitialContent('');
            }}
            initialContent={contextManagerInitialContent}
          />
        </div>
      )}

      {/* Context Action Bar */}
      {selectedText && (
        <div className="mb-2 flex items-center justify-between bg-primary-50 dark:bg-primary-900/30 p-2 rounded border border-primary-100 dark:border-primary-800 animate-in slide-in-from-bottom-2 fade-in duration-200">
          <div className="flex items-center gap-2 overflow-hidden">
            <Icon name="file-text" size="xs" className="text-primary-500 flex-shrink-0" />
            <span className="text-xs text-primary-700 dark:text-primary-300 truncate max-w-[200px]">
              "{selectedText.substring(0, 30)}..."
            </span>
          </div>
          <Button
            size="sm"
            variant="ghost"
            className="h-6 px-2 text-xs hover:bg-primary-100 dark:hover:bg-primary-800 text-primary-600 dark:text-primary-400"
            onClick={handleImportSelection}>
            Import
          </Button>
        </div>
      )}

      <div className="relative">
        <textarea
          value={inputText}
          onChange={e => setInputText(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask AI or use a tool..."
          className="w-full min-h-[80px] max-h-[200px] p-3 pr-10 text-sm border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600"
        />
        <div className="absolute bottom-2 right-2 flex gap-1">
          {/* Context Manager Button */}
          <Button
            size="sm"
            variant="ghost"
            className={cn(
              "h-8 w-8 p-0 rounded-full hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors text-slate-400 hover:text-slate-600 dark:text-slate-500 dark:hover:text-slate-300",
              showContextManager ? "bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300" : ""
            )}
            onClick={() => setShowContextManager(!showContextManager)}
            title="Manage Saved Context"
          >
            <Icon name="book" size="sm" />
          </Button>

          {/* Voice Input Button */}
          <Button
            size="sm"
            variant="ghost"
            className={cn(
              "h-8 w-8 p-0 rounded-full hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors",
              isListening ? "text-red-500 animate-pulse bg-red-50 dark:bg-red-900/20" : "text-slate-400 hover:text-slate-600 dark:text-slate-500 dark:hover:text-slate-300"
            )}
            onClick={toggleListening}
            title="Voice Input"
          >
            <Icon name={isListening ? "mic-off" : "mic"} size="sm" />
          </Button>

          <Button
            size="sm"
            className="h-8 w-8 p-0 rounded-full bg-blue-600 hover:bg-blue-700 text-white shadow-sm"
            onClick={handleSubmit}
            disabled={!inputText.trim()}>
            <Icon name="arrow-up-right" size="sm" />
          </Button>
        </div>
      </div>
      <div className="mt-2 flex justify-between items-center">
        <button
          onClick={onToggleMinimize}
          className="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 flex items-center gap-1">
          <Icon name="chevron-down" size="xs" />
          Hide Input
        </button>
        <span className="text-[10px] text-slate-400">Shift+Enter for new line</span>
      </div>
    </div>
  );
};

export default InputArea;
