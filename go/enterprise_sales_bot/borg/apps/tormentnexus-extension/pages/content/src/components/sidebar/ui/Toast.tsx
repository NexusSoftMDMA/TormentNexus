import type React from 'react';
import { useEffect, useState } from 'react';
import { useToastStore, type Toast as ToastType } from '@src/stores';
import { Icon, Typography } from './'; // Using local UI components
import { cn } from '@src/lib/utils';

const ToastItem: React.FC<{ toast: ToastType; onDismiss: (id: string) => void }> = ({ toast, onDismiss }) => {
  const [isExiting, setIsExiting] = useState(false);

  const handleDismiss = () => {
    setIsExiting(true);
    setTimeout(() => onDismiss(toast.id), 300); // Wait for exit animation
  };

  useEffect(() => {
    // Auto-dismiss handled by store logic, but we can add exit animation logic here if needed
  }, []);

  const getTypeStyles = (type: ToastType['type']) => {
    switch (type) {
      case 'success':
        return 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200';
      case 'error':
        return 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-800 text-red-800 dark:text-red-200';
      case 'warning':
        return 'bg-amber-50 dark:bg-amber-900/30 border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-200';
      default:
        return 'bg-blue-50 dark:bg-blue-900/30 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-200';
    }
  };

  const getIcon = (type: ToastType['type']) => {
    switch (type) {
      case 'success':
        return 'check';
      case 'error':
        return 'alert-triangle';
      case 'warning':
        return 'alert-triangle';
      default:
        return 'info';
    }
  };

  return (
    <div
      className={cn(
        'flex items-start gap-3 p-3 rounded-lg border shadow-lg transition-all duration-300 transform w-full max-w-xs pointer-events-auto',
        getTypeStyles(toast.type),
        isExiting ? 'translate-x-full opacity-0' : 'translate-x-0 opacity-100 animate-in slide-in-from-right-full',
      )}
      role="alert">
      <div className="flex-shrink-0 mt-0.5">
        <Icon name={getIcon(toast.type)} size="sm" />
      </div>
      <div className="flex-1 min-w-0">
        <Typography variant="subtitle" className="font-semibold text-sm leading-tight">
          {toast.title}
        </Typography>
        {toast.message && (
          <Typography variant="body" className="text-xs mt-1 opacity-90 leading-tight">
            {toast.message}
          </Typography>
        )}
      </div>
      <button
        onClick={handleDismiss}
        className="flex-shrink-0 opacity-60 hover:opacity-100 transition-opacity"
        aria-label="Dismiss notification">
        <Icon name="x" size="xs" />
      </button>
    </div>
  );
};

export const ToastContainer: React.FC = () => {
  const { toasts, removeToast } = useToastStore();

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none p-4 w-full max-w-sm items-end">
      {toasts.map(toast => (
        <ToastItem key={toast.id} toast={toast} onDismiss={removeToast} />
      ))}
    </div>
  );
};
