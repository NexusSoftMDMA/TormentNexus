import React, { useState, useRef, useEffect } from 'react';
import { useUIStore } from '@src/stores';
import { Icon, Typography, Button } from '.';
import { cn } from '@src/lib/utils';
import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('NotificationCenter');

export const NotificationCenter: React.FC = () => {
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    const notifications = useUIStore(state => state.notifications);
    const markAsRead = useUIStore(state => state.markAsRead);
    const markAllAsRead = useUIStore(state => state.markAllAsRead);
    const clearNotifications = useUIStore(state => state.clearNotifications);
    const dismissNotification = useUIStore(state => state.dismissNotification);

    const unreadCount = notifications.filter(n => !n.read).length;

    // Handle clicking outside to close
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [isOpen]);

    const handleToggle = () => {
        setIsOpen(!isOpen);
        if (!isOpen && unreadCount > 0) {
            logger.debug('[NotificationCenter] Opened with unread messages');
        }
    };

    const getIconForType = (type: string) => {
        switch (type) {
            case 'success': return <Icon name="check" size="sm" className="text-green-500" />;
            case 'error': return <Icon name="x" size="sm" className="text-red-500" />;
            case 'warning': return <Icon name="alert-triangle" size="sm" className="text-amber-500" />;
            case 'info':
            default:
                return <Icon name="info" size="sm" className="text-blue-500" />;
        }
    };

    const formatTimeAgo = (timestamp: number) => {
        const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });
        const elapsed = Date.now() - timestamp;

        // Seconds
        if (elapsed < 60000) {
            return rtf.format(-Math.round(elapsed / 1000), 'second');
        }
        // Minutes
        if (elapsed < 3600000) {
            return rtf.format(-Math.round(elapsed / 60000), 'minute');
        }
        // Hours
        if (elapsed < 86400000) {
            return rtf.format(-Math.round(elapsed / 3600000), 'hour');
        }
        // Days
        return rtf.format(-Math.round(elapsed / 86400000), 'day');
    };

    return (
        <div className="relative inline-block text-left" ref={dropdownRef}>
            <Button
                variant="ghost"
                size="icon"
                onClick={handleToggle}
                aria-label={`Notifications (${unreadCount} unread)`}
                className="relative hover:bg-slate-100 dark:hover:bg-slate-700 rounded-full transition-all duration-200">
                <Icon name="bell" size="sm" className="text-slate-700 dark:text-slate-300" />
                {unreadCount > 0 && (
                    <span className="absolute top-1 right-1 flex h-2.5 w-2.5">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                        <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-red-500"></span>
                    </span>
                )}
            </Button>

            {isOpen && (
                <div className="origin-top-right absolute right-0 mt-2 w-80 rounded-md shadow-lg bg-white dark:bg-slate-800 ring-1 ring-black ring-opacity-5 z-50 divide-y divide-slate-100 dark:divide-slate-700">
                    <div className="px-4 py-3 flex items-center justify-between border-b border-slate-100 dark:border-slate-700">
                        <Typography variant="subtitle" className="font-semibold text-slate-800 dark:text-slate-200">
                            Notifications
                        </Typography>
                        {notifications.length > 0 && (
                            <div className="flex items-center gap-2">
                                {unreadCount > 0 && (
                                    <button
                                        onClick={() => markAllAsRead()}
                                        className="text-xs text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300">
                                        Mark all read
                                    </button>
                                )}
                                <button
                                    onClick={() => clearNotifications()}
                                    className="text-xs text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-300">
                                    Clear all
                                </button>
                            </div>
                        )}
                    </div>

                    <div className="max-h-96 overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
                        {notifications.length === 0 ? (
                            <div className="px-4 py-8 text-center text-slate-500 dark:text-slate-400">
                                <Icon name="bell" size="lg" className="mx-auto mb-2 opacity-20" />
                                <Typography variant="body" className="text-sm">You're all caught up!</Typography>
                            </div>
                        ) : (
                            <div className="divide-y divide-slate-50 dark:divide-slate-700/50">
                                {/* Sort notifications by date (newest first) but with unread prioritized */}
                                {[...notifications].sort((a, b) => {
                                    if (a.read === b.read) return b.timestamp - a.timestamp;
                                    return a.read ? 1 : -1;
                                }).map((notification) => (
                                    <div
                                        key={notification.id}
                                        className={cn(
                                            "px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors relative group",
                                            !notification.read ? "bg-blue-50/50 dark:bg-slate-800/80" : ""
                                        )}
                                        onClick={() => !notification.read && markAsRead(notification.id)}>

                                        {!notification.read && (
                                            <div className="absolute left-2 top-4 w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                                        )}

                                        <div className="flex gap-3 pl-2">
                                            <div className="mt-0.5 flex-shrink-0">
                                                {getIconForType(notification.type)}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex justify-between items-start mb-0.5">
                                                    <p className="text-sm font-medium text-slate-900 dark:text-slate-100 truncate pr-6">
                                                        {notification.title}
                                                    </p>
                                                    <span className="text-[10px] text-slate-400 whitespace-nowrap pt-0.5 whitespace-nowrap flex-shrink-0">
                                                        {formatTimeAgo(notification.timestamp)}
                                                    </span>
                                                </div>
                                                <p className="text-xs text-slate-600 dark:text-slate-300 line-clamp-2">
                                                    {notification.message}
                                                </p>

                                                {/* Render Remote Action Buttons if any */}
                                                {'actions' in notification && (notification as any).actions?.length > 0 && (
                                                    <div className="mt-2 flex flex-wrap gap-2">
                                                        {(notification as any).actions.map((action: any, idx: number) => (
                                                            <Button
                                                                key={idx}
                                                                size="sm"
                                                                variant={action.primary ? "default" : "outline"}
                                                                className={cn(
                                                                    "h-6 text-[10px] px-2",
                                                                    action.primary ? "bg-blue-600 hover:bg-blue-700 text-white" : ""
                                                                )}
                                                                onClick={(e: React.MouseEvent) => {
                                                                    e.stopPropagation();
                                                                    if (action.url) {
                                                                        window.open(action.url, '_blank');
                                                                    } else {
                                                                        logger.debug('Action clicked without URL:', action);
                                                                    }
                                                                    dismissNotification(notification.id, 'action_clicked');
                                                                }}>
                                                                {action.label}
                                                                {action.url && <Icon name="arrow-up-right" size="xs" className="ml-1" />}
                                                            </Button>
                                                        ))}
                                                    </div>
                                                )}
                                            </div>

                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    dismissNotification(notification.id);
                                                }}
                                                className="opacity-0 group-hover:opacity-100 absolute right-3 top-3 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
                                                aria-label="Dismiss">
                                                <Icon name="x" size="xs" />
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
};
