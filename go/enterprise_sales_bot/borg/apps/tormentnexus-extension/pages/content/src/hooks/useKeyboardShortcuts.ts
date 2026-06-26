import { useEffect, useCallback } from 'react';

type ShortcutAction = 'toggleSidebar' | 'closeSidebar' | 'searchTools' | 'switchTab' | 'testConnection';

export const useKeyboardShortcuts = (actions: {
  toggleSidebar: () => void;
  closeSidebar: () => void;
  focusSearch: () => void;
  switchTab: (direction: 'next' | 'prev') => void;
  toggleCommandPalette?: () => void;
  testConnection?: () => void;
}) => {

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Ignore if typing in an input/textarea
      if (
        (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) &&
        event.key !== 'Escape'
      ) {
        return;
      }

      // Toggle Sidebar: Alt + Shift + S (or similar)
      if (event.altKey && event.shiftKey && (event.key === 's' || event.key === 'S')) {
        event.preventDefault();
        actions.toggleSidebar();
      }

      // Close Sidebar: Escape
      if (event.key === 'Escape') {
        // Only close if not in a modal or special context?
        // For now, let's just close sidebar if focused
        actions.closeSidebar();
      }

      // Focus Search: / (Forward Slash)
      if (event.key === '/' && !event.ctrlKey && !event.metaKey) {
        event.preventDefault();
        actions.focusSearch();
      }

      // Switch Tabs: Ctrl + Arrow Left/Right
      if (event.ctrlKey && event.key === 'ArrowRight') {
        event.preventDefault();
        actions.switchTab('next');
      }
      if (event.ctrlKey && event.key === 'ArrowLeft') {
        event.preventDefault();
        actions.switchTab('prev');
      }

      // Toggle Command Palette: Ctrl + K
      if (event.ctrlKey && (event.key === 'k' || event.key === 'K') && actions.toggleCommandPalette) {
        event.preventDefault();
        actions.toggleCommandPalette();
      }
    },
    [actions],
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
};
