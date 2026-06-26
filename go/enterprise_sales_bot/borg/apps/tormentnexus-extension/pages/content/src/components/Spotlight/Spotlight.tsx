import React, { useState, useEffect, useRef } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { useProfileStore } from '@src/stores';
import { Icon } from '@src/components/sidebar/ui';
import { cn } from '@src/lib/utils';
import { eventBus } from '@src/events/event-bus';

interface CommandItem {
  id: string;
  title: string;
  icon: any;
  action: () => void;
  shortcut?: string;
  category: 'Navigation' | 'Tools' | 'Settings' | 'Profiles';
}

interface SpotlightProps {
  hostPointerEventsToggle: (enable: boolean) => void;
}

const Spotlight: React.FC<SpotlightProps> = ({ hostPointerEventsToggle }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const { profiles, setProfilesActive } = useProfileStore();

  // Listen for global Cmd+K
  useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      // Cmd+K (Mac) or Ctrl+K (Windows/Linux)
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setIsOpen(prev => !prev);
      }
    };
    window.addEventListener('keydown', handleGlobalKeyDown);
    return () => window.removeEventListener('keydown', handleGlobalKeyDown);
  }, []);

  useEffect(() => {
    hostPointerEventsToggle(isOpen);
    if (isOpen) {
      // slight delay to ensure render
      setTimeout(() => inputRef.current?.focus(), 50);
      setSelectedIndex(0);
    } else {
      setQuery('');
    }
  }, [isOpen, hostPointerEventsToggle]);

  const handleClose = () => setIsOpen(false);

  // Commands definition matching the original CommandPalette
  const commands: CommandItem[] = [
    {
      id: 'nav-tools',
      title: 'Go to Available Tools',
      icon: 'tool',
      category: 'Navigation',
      action: () => eventBus.emit('sidebar:navigate', 'availableTools'),
    },
    {
      id: 'nav-activity',
      title: 'Go to Activity Log',
      icon: 'activity',
      category: 'Navigation',
      action: () => eventBus.emit('sidebar:navigate', 'activity'),
    },
    {
      id: 'nav-dashboard',
      title: 'Go to Dashboard',
      icon: 'box',
      category: 'Navigation',
      action: () => eventBus.emit('sidebar:navigate', 'dashboard'),
    },
    {
      id: 'nav-settings',
      title: 'Go to Settings',
      icon: 'settings',
      category: 'Navigation',
      action: () => eventBus.emit('sidebar:navigate', 'settings'),
    },
    {
      id: 'nav-help',
      title: 'Go to Help',
      icon: 'help-circle',
      category: 'Navigation',
      action: () => eventBus.emit('sidebar:navigate', 'help'),
    },
    {
      id: 'toggle-push',
      title: 'Toggle Push Content Mode',
      icon: 'menu',
      category: 'Settings',
      action: () => {
        const evt = new CustomEvent('toggle-push-mode');
        window.dispatchEvent(evt);
      },
    },
    // Add profiles dynamically
    ...profiles.map(p => ({
      id: `profile-${p.id}`,
      title: `Switch to Profile: ${p.name}`,
      icon: 'server',
      category: 'Profiles' as const,
      action: () => setProfilesActive([p.id]),
    })),
  ];

  const filteredCommands = commands.filter(cmd => cmd.title.toLowerCase().includes(query.toLowerCase()));

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex(prev => (prev + 1) % filteredCommands.length);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex(prev => (prev - 1 + filteredCommands.length) % filteredCommands.length);
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (filteredCommands[selectedIndex]) {
          filteredCommands[selectedIndex].action();
          handleClose();
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        handleClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, filteredCommands, selectedIndex]);

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="fixed inset-0 z-[2147483647] flex items-start justify-center pt-[20vh] bg-black/50 backdrop-blur-sm"
          onClick={handleClose}>
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: -10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: -10 }}
            transition={{
              type: "spring",
              damping: 25,
              stiffness: 400
            }}
            className="w-full max-w-lg bg-white dark:bg-slate-900 rounded-xl shadow-2xl overflow-hidden border border-slate-200 dark:border-slate-700"
            onClick={e => e.stopPropagation()}>
            <div className="flex items-center px-4 py-3 border-b border-slate-100 dark:border-slate-800">
              <Icon name="search" size="sm" className="text-slate-400 mr-3" />
              <input
                ref={inputRef}
                className="flex-1 bg-transparent border-none outline-none text-sm text-slate-800 dark:text-slate-200 placeholder:text-slate-400"
                placeholder="Type a command or search..."
                value={query}
                onChange={e => {
                  setQuery(e.target.value);
                  setSelectedIndex(0);
                }}
              />
              <div className="text-[10px] bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-slate-500 font-mono">
                ESC
              </div>
            </div>

            <div className="max-h-[300px] overflow-y-auto py-2">
              {filteredCommands.length === 0 ? (
                <div className="px-4 py-8 text-center text-slate-500 text-sm">No commands found.</div>
              ) : (
                filteredCommands.map((cmd, index) => (
                  <div
                    key={cmd.id}
                    className={cn(
                      'px-4 py-2 flex items-center cursor-pointer text-sm transition-colors',
                      index === selectedIndex
                        ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 border-l-2 border-primary-500'
                        : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 border-l-2 border-transparent',
                    )}
                    onClick={() => {
                      cmd.action();
                      handleClose();
                    }}
                    onMouseEnter={() => setSelectedIndex(index)}>
                    <Icon
                      name={cmd.icon}
                      size="sm"
                      className={cn('mr-3 transition-colors', index === selectedIndex ? 'text-primary-500' : 'text-slate-400')}
                    />
                    <span className="flex-1">{cmd.title}</span>
                    {cmd.shortcut && <span className="text-xs text-slate-400 font-mono">{cmd.shortcut}</span>}
                  </div>
                ))
              )}
            </div>

            <div className="px-4 py-2 bg-slate-50 dark:bg-slate-800/50 border-t border-slate-100 dark:border-slate-800 flex justify-between items-center text-[10px] text-slate-500">
              <span>
                <strong className="font-medium text-slate-700 dark:text-slate-300">↑↓</strong> to navigate
              </span>
              <span>
                <strong className="font-medium text-slate-700 dark:text-slate-300">↵</strong> to select
              </span>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default Spotlight;
