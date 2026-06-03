"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { usePathname } from "next/navigation";
import { ChevronDown, Command, Download, GripVertical, Pin, PinOff, RotateCcw, Search, Upload } from "lucide-react";
import { DndContext, PointerSensor, closestCenter, useSensor, useSensors, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, arrayMove, rectSortingStrategy, useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { cn } from "@/lib/utils";
import { SIDEBAR_SECTIONS } from "./mcp/nav-config";

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> { }

const SIDEBAR_COLLAPSE_STORAGE_KEY = 'tormentnexus_sidebar_collapsed_sections_v1';
const SIDEBAR_FAVORITES_STORAGE_KEY = 'tormentnexus_sidebar_favorites_v1';
const SIDEBAR_RECENT_STORAGE_KEY = 'tormentnexus_sidebar_recent_routes_v1';
const SIDEBAR_RECENT_SEARCHES_STORAGE_KEY = 'tormentnexus_sidebar_recent_searches_v1';
const MAX_RECENT_ROUTES = 8;
const MAX_RECENT_SEARCHES = 6;

type PaletteItem = {
    kind: 'route' | 'action';
    title: string;
    section: string;
    icon: any;
    href?: string;
    id?: string;
    description?: string;
};

export function Sidebar({ className }: SidebarProps) {
    const router = useRouter();
    const pathname = usePathname();
    const [query, setQuery] = useState('');
    const [collapsedSections, setCollapsedSections] = useState<Record<string, boolean>>({});
    const [favorites, setFavorites] = useState<string[]>([]);
    const [recentRoutes, setRecentRoutes] = useState<string[]>([]);
    const [recentSearches, setRecentSearches] = useState<string[]>([]);
    const [navNotice, setNavNotice] = useState<string | null>(null);
    const [isPaletteOpen, setIsPaletteOpen] = useState(false);
    const [paletteQuery, setPaletteQuery] = useState('');
    const [paletteIndex, setPaletteIndex] = useState(0);
    const fileInputRef = useRef<HTMLInputElement | null>(null);
    const searchInputRef = useRef<HTMLInputElement | null>(null);
    const paletteInputRef = useRef<HTMLInputElement | null>(null);

    useEffect(() => {
        try {
            const raw = localStorage.getItem(SIDEBAR_COLLAPSE_STORAGE_KEY);
            if (!raw) {
                return;
            }
            const parsed = JSON.parse(raw) as Record<string, boolean>;
            setCollapsedSections(parsed ?? {});
        } catch {
            // ignore invalid stored state
        }
    }, []);

    useEffect(() => {
        try {
            const raw = localStorage.getItem(SIDEBAR_FAVORITES_STORAGE_KEY);
            if (!raw) {
                return;
            }
            const parsed = JSON.parse(raw);
            if (Array.isArray(parsed)) {
                setFavorites(parsed.filter((value): value is string => typeof value === 'string'));
            }
        } catch {
            // ignore invalid stored state
        }
    }, []);

    useEffect(() => {
        try {
            const raw = localStorage.getItem(SIDEBAR_RECENT_STORAGE_KEY);
            if (!raw) {
                return;
            }
            const parsed = JSON.parse(raw);
            if (Array.isArray(parsed)) {
                setRecentRoutes(parsed.filter((value): value is string => typeof value === 'string'));
            }
        } catch {
            // ignore invalid stored state
        }
    }, []);

    useEffect(() => {
        try {
            const raw = localStorage.getItem(SIDEBAR_RECENT_SEARCHES_STORAGE_KEY);
            if (!raw) {
                return;
            }
            const parsed = JSON.parse(raw);
            if (Array.isArray(parsed)) {
                setRecentSearches(parsed.filter((value): value is string => typeof value === 'string'));
            }
        } catch {
            // ignore invalid stored state
        }
    }, []);

    const isActive = (href: string) => {
        if (pathname === href) {
            return true;
        }
        return href !== '/dashboard' && pathname.startsWith(`${href}/`);
    };

    const normalizedQuery = query.trim().toLowerCase();

    const allItemsByHref = useMemo(() => {
        const map = new Map<string, { title: string; href: string; icon: any }>();
        for (const section of SIDEBAR_SECTIONS) {
            for (const item of section.items) {
                map.set(item.href, item);
            }
        }
        return map;
    }, []);

    const filteredSections = useMemo(() => {
        return SIDEBAR_SECTIONS
            .map((section) => ({
                ...section,
                items: section.items.filter((item) => {
                    if (!normalizedQuery) {
                        return true;
                    }
                    return (
                        item.title.toLowerCase().includes(normalizedQuery) ||
                        item.href.toLowerCase().includes(normalizedQuery)
                    );
                }),
            }))
            .filter((section) => section.items.length > 0);
    }, [normalizedQuery]);

    const favoriteSet = useMemo(() => new Set(favorites), [favorites]);

    const paletteItems = useMemo<PaletteItem[]>(() => {
        const routeMeta = new Map<string, { title: string; href: string; icon: any; section: string }>();
        for (const section of SIDEBAR_SECTIONS) {
            for (const item of section.items) {
                routeMeta.set(item.href, {
                    ...item,
                    section: section.title,
                });
            }
        }

        const q = paletteQuery.trim().toLowerCase();
        const recencyRank = new Map<string, number>();
        recentRoutes.forEach((href, index) => {
            recencyRank.set(href, index);
        });

        const actions: PaletteItem[] = [
            {
                kind: 'action' as const,
                id: 'open-tormentnexus',
                title: 'Open TormentNexus Dashboard',
                section: 'Actions',
                description: 'Go to /dashboard/mcp',
                icon: Command,
            },
            {
                kind: 'action' as const,
                id: 'reset-nav',
                title: 'Reset Navigation Preferences',
                section: 'Actions',
                description: 'Clear search, favorites, recents, and collapse state',
                icon: RotateCcw,
            },
            {
                kind: 'action' as const,
                id: 'clear-favorites',
                title: 'Clear Favorites',
                section: 'Actions',
                description: 'Remove all pinned routes',
                icon: PinOff,
            },
            {
                kind: 'action' as const,
                id: 'clear-recents',
                title: 'Clear Recent Routes',
                section: 'Actions',
                description: 'Forget recently visited routes',
                icon: RotateCcw,
            },
        ].filter((item) => {
            if (!q) {
                return true;
            }
            return `${item.title} ${item.description ?? ''}`.toLowerCase().includes(q);
        });

        const rows = Array.from(routeMeta.values()).filter((item) => {
            if (!q) {
                return true;
            }
            return (
                item.title.toLowerCase().includes(q) ||
                item.href.toLowerCase().includes(q) ||
                item.section.toLowerCase().includes(q)
            );
        });

        rows.sort((a, b) => {
            const aFav = favoriteSet.has(a.href) ? 1 : 0;
            const bFav = favoriteSet.has(b.href) ? 1 : 0;
            if (aFav !== bFav) {
                return bFav - aFav;
            }

            const aRecent = recencyRank.has(a.href) ? recencyRank.get(a.href)! : Number.MAX_SAFE_INTEGER;
            const bRecent = recencyRank.has(b.href) ? recencyRank.get(b.href)! : Number.MAX_SAFE_INTEGER;
            if (aRecent !== bRecent) {
                return aRecent - bRecent;
            }

            return a.title.localeCompare(b.title);
        });

        const routeItems: PaletteItem[] = rows.map((item) => ({
            kind: 'route',
            ...item,
        }));

        return [...actions, ...routeItems].slice(0, 50);
    }, [favoriteSet, paletteQuery, recentRoutes]);

    const favoriteItems = useMemo(() => {
        return favorites
            .map((href) => allItemsByHref.get(href))
            .filter((item): item is { title: string; href: string; icon: any } => Boolean(item))
            .filter((item) => {
                if (!normalizedQuery) {
                    return true;
                }
                return item.title.toLowerCase().includes(normalizedQuery) || item.href.toLowerCase().includes(normalizedQuery);
            });
    }, [allItemsByHref, favorites, normalizedQuery]);

    const recentItems = useMemo(() => {
        return recentRoutes
            .map((href) => allItemsByHref.get(href))
            .filter((item): item is { title: string; href: string; icon: any } => Boolean(item))
            .filter((item) => {
                if (!normalizedQuery) {
                    return true;
                }
                return item.title.toLowerCase().includes(normalizedQuery) || item.href.toLowerCase().includes(normalizedQuery);
            });
    }, [allItemsByHref, normalizedQuery, recentRoutes]);

    const persistFavorites = (next: string[]) => {
        setFavorites(next);
        localStorage.setItem(SIDEBAR_FAVORITES_STORAGE_KEY, JSON.stringify(next));
    };

    const showNotice = (message: string) => {
        setNavNotice(message);
        window.setTimeout(() => {
            setNavNotice((current) => (current === message ? null : current));
        }, 2200);
    };

    const openPalette = () => {
        setIsPaletteOpen(true);
        setPaletteIndex(0);
    };

    const closePalette = () => {
        setIsPaletteOpen(false);
    };

    const selectPaletteItem = (href: string, openInNewTab = false) => {
        closePalette();
        if (openInNewTab) {
            window.open(href, '_blank', 'noopener,noreferrer');
            return;
        }
        router.push(href);
    };

    const toggleFavorite = (href: string) => {
        if (!allItemsByHref.has(href)) {
            return;
        }
        const exists = favoriteSet.has(href);
        if (exists) {
            persistFavorites(favorites.filter((entry) => entry !== href));
            return;
        }
        persistFavorites([...favorites, href]);
    };

    const persistRecentRoutes = (next: string[]) => {
        setRecentRoutes(next);
        localStorage.setItem(SIDEBAR_RECENT_STORAGE_KEY, JSON.stringify(next));
    };

    const persistRecentSearches = (next: string[]) => {
        setRecentSearches(next);
        localStorage.setItem(SIDEBAR_RECENT_SEARCHES_STORAGE_KEY, JSON.stringify(next));
    };

    const rememberPaletteSearch = (value: string) => {
        const trimmed = value.trim();
        if (!trimmed) {
            return;
        }
        const next = [trimmed, ...recentSearches.filter((entry) => entry !== trimmed)].slice(0, MAX_RECENT_SEARCHES);
        persistRecentSearches(next);
    };

    const runPaletteAction = (id: string) => {
        if (id === 'open-tormentnexus') {
            closePalette();
            router.push('/dashboard/mcp');
            return;
        }
        if (id === 'reset-nav') {
            resetNavigationState();
            closePalette();
            return;
        }
        if (id === 'clear-favorites') {
            resetFavoritesOnly();
            closePalette();
            return;
        }
        if (id === 'clear-recents') {
            clearRecentOnly();
            closePalette();
        }
    };

    const resetNavigationState = () => {
        setQuery('');
        setCollapsedSections({});
        setFavorites([]);
        setRecentRoutes([]);
        setRecentSearches([]);
        localStorage.removeItem(SIDEBAR_COLLAPSE_STORAGE_KEY);
        localStorage.removeItem(SIDEBAR_FAVORITES_STORAGE_KEY);
        localStorage.removeItem(SIDEBAR_RECENT_STORAGE_KEY);
        localStorage.removeItem(SIDEBAR_RECENT_SEARCHES_STORAGE_KEY);
        showNotice('Navigation preferences reset.');
    };

    const resetSearchOnly = () => {
        setQuery('');
        showNotice('Search cleared.');
    };

    const resetFavoritesOnly = () => {
        setFavorites([]);
        localStorage.removeItem(SIDEBAR_FAVORITES_STORAGE_KEY);
        showNotice('Favorites cleared.');
    };

    const clearRecentOnly = () => {
        setRecentRoutes([]);
        localStorage.removeItem(SIDEBAR_RECENT_STORAGE_KEY);
        showNotice('Recent routes cleared.');
    };

    const exportPreferences = () => {
        const payload = {
            collapsedSections,
            favorites,
            recentRoutes,
            recentSearches,
            exportedAt: new Date().toISOString(),
        };

        const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `tormentnexus-nav-preferences-${new Date().toISOString().slice(0, 10)}.json`;
        document.body.appendChild(link);
        link.click();
        link.remove();
        URL.revokeObjectURL(url);
        showNotice('Navigation preferences exported.');
    };

    const openImportDialog = () => {
        fileInputRef.current?.click();
    };

    const handleImportPreferences = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        event.target.value = '';
        if (!file) {
            return;
        }

        try {
            const text = await file.text();
            const parsed = JSON.parse(text) as {
                collapsedSections?: Record<string, boolean>;
                favorites?: string[];
                recentRoutes?: string[];
                recentSearches?: string[];
            };

            const nextCollapsed = parsed.collapsedSections ?? {};
            const nextFavoritesRaw = Array.isArray(parsed.favorites) ? parsed.favorites : [];
            const nextFavorites = nextFavoritesRaw
                .filter((href): href is string => typeof href === 'string')
                .filter((href) => allItemsByHref.has(href));
            const nextRecentRaw = Array.isArray(parsed.recentRoutes) ? parsed.recentRoutes : [];
            const nextRecent = nextRecentRaw
                .filter((href): href is string => typeof href === 'string')
                .filter((href) => allItemsByHref.has(href))
                .slice(0, MAX_RECENT_ROUTES);
            const nextSearchesRaw = Array.isArray(parsed.recentSearches) ? parsed.recentSearches : [];
            const nextSearches = nextSearchesRaw
                .filter((value): value is string => typeof value === 'string')
                .map((value) => value.trim())
                .filter(Boolean)
                .slice(0, MAX_RECENT_SEARCHES);

            setCollapsedSections(nextCollapsed);
            setFavorites(nextFavorites);
            setRecentRoutes(nextRecent);
            setRecentSearches(nextSearches);
            localStorage.setItem(SIDEBAR_COLLAPSE_STORAGE_KEY, JSON.stringify(nextCollapsed));
            localStorage.setItem(SIDEBAR_FAVORITES_STORAGE_KEY, JSON.stringify(nextFavorites));
            localStorage.setItem(SIDEBAR_RECENT_STORAGE_KEY, JSON.stringify(nextRecent));
            localStorage.setItem(SIDEBAR_RECENT_SEARCHES_STORAGE_KEY, JSON.stringify(nextSearches));
            showNotice('Navigation preferences imported.');
        } catch {
            showNotice('Failed to import preferences file.');
        }
    };

    useEffect(() => {
        if (!pathname || !allItemsByHref.has(pathname)) {
            return;
        }

        setRecentRoutes((current) => {
            const next = [pathname, ...current.filter((href) => href !== pathname)].slice(0, MAX_RECENT_ROUTES);
            localStorage.setItem(SIDEBAR_RECENT_STORAGE_KEY, JSON.stringify(next));
            return next;
        });
    }, [allItemsByHref, pathname]);

    useEffect(() => {
        const onKeyDown = (event: KeyboardEvent) => {
            const target = event.target as HTMLElement | null;
            const isTypingContext =
                target?.tagName === 'INPUT' ||
                target?.tagName === 'TEXTAREA' ||
                Boolean(target?.isContentEditable);

            if (event.key === '/' && !event.metaKey && !event.ctrlKey && !event.altKey && !isTypingContext) {
                event.preventDefault();
                searchInputRef.current?.focus();
                searchInputRef.current?.select();
            }

            if ((event.key.toLowerCase() === 'k') && (event.metaKey || event.ctrlKey)) {
                event.preventDefault();
                openPalette();
            }

            if (event.key === 'Escape' && document.activeElement === searchInputRef.current) {
                setQuery('');
                searchInputRef.current?.blur();
            }
        };

        window.addEventListener('keydown', onKeyDown);
        return () => window.removeEventListener('keydown', onKeyDown);
    }, []);

    useEffect(() => {
        if (!isPaletteOpen) {
            return;
        }
        window.setTimeout(() => {
            paletteInputRef.current?.focus();
            paletteInputRef.current?.select();
        }, 0);
    }, [isPaletteOpen]);

    useEffect(() => {
        setPaletteIndex(0);
    }, [paletteQuery]);

    useEffect(() => {
        if (!isPaletteOpen) {
            return;
        }

        const onPaletteKeyDown = (event: KeyboardEvent) => {
            if (event.key === 'Escape') {
                event.preventDefault();
                closePalette();
                return;
            }

            if (event.key === 'ArrowDown') {
                event.preventDefault();
                setPaletteIndex((prev) => Math.min(prev + 1, Math.max(0, paletteItems.length - 1)));
                return;
            }

            if (event.key === 'ArrowUp') {
                event.preventDefault();
                setPaletteIndex((prev) => Math.max(prev - 1, 0));
                return;
            }

            if (event.key === 'Enter') {
                const selected = paletteItems[paletteIndex];
                if (selected) {
                    event.preventDefault();
                    const trimmedQuery = paletteQuery.trim();
                    if (trimmedQuery) {
                        const nextSearches = [trimmedQuery, ...recentSearches.filter((entry) => entry !== trimmedQuery)]
                            .slice(0, MAX_RECENT_SEARCHES);
                        setRecentSearches(nextSearches);
                        localStorage.setItem(SIDEBAR_RECENT_SEARCHES_STORAGE_KEY, JSON.stringify(nextSearches));
                    }

                    if (selected.kind === 'action' && selected.id) {
                        runPaletteAction(selected.id);
                        return;
                    }

                    if (selected.href) {
                        selectPaletteItem(selected.href, event.metaKey || event.ctrlKey);
                    }
                }
            }
        };

        window.addEventListener('keydown', onPaletteKeyDown);
        return () => window.removeEventListener('keydown', onPaletteKeyDown);
    }, [isPaletteOpen, paletteIndex, paletteItems, paletteQuery, recentSearches]);

    const sensors = useSensors(
        useSensor(PointerSensor, {
            activationConstraint: { distance: 6 },
        })
    );

    const handleFavoriteDragEnd = (event: DragEndEvent) => {
        const { active, over } = event;
        if (!over || active.id === over.id) {
            return;
        }
        const oldIndex = favorites.indexOf(String(active.id));
        const newIndex = favorites.indexOf(String(over.id));
        if (oldIndex < 0 || newIndex < 0) {
            return;
        }
        persistFavorites(arrayMove(favorites, oldIndex, newIndex));
    };

    const toggleSection = (title: string) => {
        setCollapsedSections((current) => {
            const next = {
                ...current,
                [title]: !current[title],
            };
            localStorage.setItem(SIDEBAR_COLLAPSE_STORAGE_KEY, JSON.stringify(next));
            return next;
        });
    };

    return (
        <div className={cn("pb-12 w-72 border-r border-zinc-800 bg-zinc-950 block", className)}>
            <div className="space-y-4 py-4">
                <div className="px-3 py-2 space-y-5 max-h-[calc(100vh-2rem)] overflow-auto">
                    <h2 className="mb-1 px-4 text-lg font-semibold tracking-tight text-white">
                        TormentNexus Navigation
                    </h2>
                    <div className="px-3">
                        <input
                            ref={fileInputRef}
                            type="file"
                            accept="application/json"
                            onChange={handleImportPreferences}
                            className="hidden"
                        />
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
                            <input
                                ref={searchInputRef}
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                placeholder="Search pages... ( / )"
                                className="w-full h-9 rounded-md border border-zinc-800 bg-zinc-900 pl-9 pr-3 text-xs text-zinc-200 placeholder:text-zinc-500 outline-none focus:border-zinc-600"
                            />
                        </div>
                        <div className="mt-2 flex justify-end">
                            <button
                                type="button"
                                onClick={openPalette}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Quick switch (Ctrl/Cmd+K)"
                            >
                                <Command className="h-3 w-3" />
                                Quick Switch
                            </button>
                        </div>
                        <div className="mt-2 flex flex-wrap justify-end gap-1.5">
                            <button
                                type="button"
                                onClick={resetSearchOnly}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Clear current search text"
                            >
                                <Search className="h-3 w-3" />
                                Clear Search
                            </button>
                            <button
                                type="button"
                                onClick={resetFavoritesOnly}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Reset favorites only"
                            >
                                <PinOff className="h-3 w-3" />
                                Clear Favorites
                            </button>
                            <button
                                type="button"
                                onClick={clearRecentOnly}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Clear recently visited routes"
                            >
                                <RotateCcw className="h-3 w-3" />
                                Clear Recent
                            </button>
                            <button
                                type="button"
                                onClick={exportPreferences}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Export sidebar preferences"
                            >
                                <Download className="h-3 w-3" />
                                Export
                            </button>
                            <button
                                type="button"
                                onClick={openImportDialog}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Import sidebar preferences"
                            >
                                <Upload className="h-3 w-3" />
                                Import
                            </button>
                            <button
                                type="button"
                                onClick={resetNavigationState}
                                className="inline-flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900 px-2.5 py-1 text-[11px] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
                                title="Reset search, favorites, and section collapse state"
                            >
                                <RotateCcw className="h-3 w-3" />
                                Reset Nav
                            </button>
                        </div>
                        {navNotice ? (
                            <div className="mt-2 rounded-md border border-zinc-800 bg-zinc-900/80 px-2.5 py-1.5 text-[11px] text-zinc-300">
                                {navNotice}
                            </div>
                        ) : null}
                    </div>

                    {favoriteItems.length > 0 ? (
                        <div>
                            <h3 className="px-4 mb-1 text-[11px] uppercase tracking-wider text-zinc-500">
                                Favorites
                            </h3>
                            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleFavoriteDragEnd}>
                                <SortableContext items={favoriteItems.map((item) => item.href)} strategy={rectSortingStrategy}>
                                    <div className="space-y-1">
                                        {favoriteItems.map((item) => (
                                            <FavoriteNavRow
                                                key={item.href}
                                                item={item}
                                                active={isActive(item.href)}
                                                onToggleFavorite={toggleFavorite}
                                            />
                                        ))}
                                    </div>
                                </SortableContext>
                            </DndContext>
                        </div>
                    ) : null}

                    {recentItems.length > 0 ? (
                        <div>
                            <h3 className="px-4 mb-1 text-[11px] uppercase tracking-wider text-zinc-500">
                                Recent
                            </h3>
                            <div className="space-y-1">
                                {recentItems.map((item) => (
                                    <Link
                                        key={item.href}
                                        href={item.href}
                                        className={cn(
                                            "flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-zinc-800 hover:text-white transition-colors",
                                            isActive(item.href) ? "bg-zinc-800 text-white" : "text-zinc-300"
                                        )}
                                    >
                                        <item.icon className="mr-2 h-4 w-4" />
                                        {item.title}
                                    </Link>
                                ))}
                            </div>
                        </div>
                    ) : null}

                    {filteredSections.map((section) => {
                        const isCollapsed = normalizedQuery ? false : Boolean(collapsedSections[section.title]);

                        return (
                            <div key={section.title}>
                                <button
                                    type="button"
                                    onClick={() => toggleSection(section.title)}
                                    className="w-full px-4 mb-1 text-[11px] uppercase tracking-wider text-zinc-500 flex items-center justify-between hover:text-zinc-300 transition-colors"
                                >
                                    <span>{section.title}</span>
                                    <ChevronDown className={cn("h-3.5 w-3.5 transition-transform", isCollapsed ? "-rotate-90" : "rotate-0")} />
                                </button>
                                {!isCollapsed ? (
                                    <div className="space-y-1">
                                        {section.items.map((item) => (
                                            <div key={item.href} className="group flex items-center gap-1">
                                                <Link
                                                    href={item.href}
                                                    className={cn(
                                                        "flex-1 flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-zinc-800 hover:text-white transition-colors",
                                                        isActive(item.href) ? "bg-zinc-800 text-white" : "text-zinc-400",
                                                        "justify-start"
                                                    )}
                                                >
                                                    <item.icon className="mr-2 h-4 w-4" />
                                                    {item.title}
                                                </Link>
                                                <button
                                                    type="button"
                                                    onClick={() => toggleFavorite(item.href)}
                                                    className={cn(
                                                        "h-8 w-8 rounded-md border border-transparent hover:border-zinc-700 hover:bg-zinc-900 transition-colors flex items-center justify-center",
                                                        favoriteSet.has(item.href) ? "text-amber-300" : "text-zinc-600 group-hover:text-zinc-400"
                                                    )}
                                                    title={favoriteSet.has(item.href) ? 'Unpin from favorites' : 'Pin to favorites'}
                                                    aria-label={favoriteSet.has(item.href) ? `Unpin ${item.title}` : `Pin ${item.title}`}
                                                >
                                                    {favoriteSet.has(item.href) ? <PinOff className="h-3.5 w-3.5" /> : <Pin className="h-3.5 w-3.5" />}
                                                </button>
                                            </div>
                                        ))}
                                    </div>
                                ) : null}
                            </div>
                        )
                    })}
                    {filteredSections.length === 0 ? (
                        <div className="px-4 py-6 text-xs text-zinc-500 border border-zinc-800 rounded-md bg-zinc-900/60">
                            No pages match “{query}”.
                        </div>
                    ) : null}
                </div>
            </div>

            {isPaletteOpen ? (
                <div className="fixed inset-0 z-[120] bg-black/65 flex items-start justify-center p-8" onClick={closePalette}>
                    <div
                        className="w-full max-w-2xl rounded-xl border border-zinc-800 bg-zinc-950 shadow-2xl overflow-hidden"
                        onClick={(event) => event.stopPropagation()}
                    >
                        <div className="p-3 border-b border-zinc-800 flex items-center gap-2">
                            <Search className="h-4 w-4 text-zinc-500" />
                            <input
                                ref={paletteInputRef}
                                value={paletteQuery}
                                onChange={(e) => setPaletteQuery(e.target.value)}
                                placeholder="Jump to page, tool, or route..."
                                className="flex-1 bg-transparent text-sm text-zinc-100 outline-none placeholder:text-zinc-500"
                            />
                            <span className="text-[10px] text-zinc-500 border border-zinc-700 rounded px-1.5 py-0.5">Esc</span>
                        </div>

                        {recentSearches.length > 0 && !paletteQuery.trim() ? (
                            <div className="px-3 pt-2 pb-1 border-b border-zinc-900">
                                <div className="text-[10px] uppercase tracking-wider text-zinc-500 mb-1.5">Recent Searches</div>
                                <div className="flex flex-wrap gap-1.5">
                                    {recentSearches.map((entry) => (
                                        <button
                                            key={entry}
                                            type="button"
                                            onClick={() => setPaletteQuery(entry)}
                                            className="text-[11px] rounded border border-zinc-800 bg-zinc-900 px-2 py-1 text-zinc-300 hover:bg-zinc-800 hover:text-white"
                                        >
                                            {entry}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        ) : null}

                        <div className="max-h-[60vh] overflow-auto p-2 space-y-1">
                            {paletteItems.length === 0 ? (
                                <div className="px-3 py-8 text-center text-sm text-zinc-500">
                                    No routes match “{paletteQuery}”.
                                </div>
                            ) : (
                                paletteItems.map((item, index) => {
                                    const active = index === paletteIndex;
                                    const Icon = item.icon;
                                    const isFav = item.kind === 'route' && item.href ? favoriteSet.has(item.href) : false;
                                    const isRecent = item.kind === 'route' && item.href ? recentRoutes.includes(item.href) : false;

                                    return (
                                        <button
                                            key={item.kind === 'route' ? item.href : item.id}
                                            type="button"
                                            onMouseEnter={() => setPaletteIndex(index)}
                                            onClick={(event) => {
                                                const trimmedQuery = paletteQuery.trim();
                                                if (trimmedQuery) {
                                                    const nextSearches = [trimmedQuery, ...recentSearches.filter((entry) => entry !== trimmedQuery)]
                                                        .slice(0, MAX_RECENT_SEARCHES);
                                                    setRecentSearches(nextSearches);
                                                    localStorage.setItem(SIDEBAR_RECENT_SEARCHES_STORAGE_KEY, JSON.stringify(nextSearches));
                                                }

                                                if (item.kind === 'action' && item.id) {
                                                    runPaletteAction(item.id);
                                                    return;
                                                }

                                                if (item.href) {
                                                    selectPaletteItem(item.href, event.metaKey || event.ctrlKey);
                                                }
                                            }}
                                            className={cn(
                                                "w-full text-left rounded-md px-3 py-2 transition-colors border",
                                                active
                                                    ? "bg-zinc-800 border-zinc-700"
                                                    : "bg-transparent border-transparent hover:bg-zinc-900 hover:border-zinc-800"
                                            )}
                                        >
                                            <div className="flex items-center justify-between gap-2">
                                                <div className="min-w-0">
                                                    <div className="flex items-center gap-2 text-sm text-zinc-100">
                                                        <Icon className="h-4 w-4 text-zinc-400" />
                                                        <span className="truncate">{item.title}</span>
                                                        {item.kind === 'action' ? <span className="text-[10px] text-violet-300 border border-violet-600/40 px-1.5 py-0.5 rounded">Action</span> : null}
                                                        {isFav ? <span className="text-[10px] text-amber-300 border border-amber-600/40 px-1.5 py-0.5 rounded">Fav</span> : null}
                                                        {isRecent ? <span className="text-[10px] text-cyan-300 border border-cyan-600/40 px-1.5 py-0.5 rounded">Recent</span> : null}
                                                    </div>
                                                    <div className="text-[11px] text-zinc-500 truncate mt-0.5">
                                                        {item.section}
                                                        {item.kind === 'route' ? ` • ${item.href}` : item.description ? ` • ${item.description}` : ''}
                                                    </div>
                                                </div>
                                            </div>
                                        </button>
                                    );
                                })
                            )}
                        </div>
                    </div>
                </div>
            ) : null}
        </div>
    );
}

function FavoriteNavRow({
    item,
    active,
    onToggleFavorite,
}: {
    item: { title: string; href: string; icon: any };
    active: boolean;
    onToggleFavorite: (href: string) => void;
}) {
    const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: item.href });

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
    };

    return (
        <div
            ref={setNodeRef}
            style={style}
            className={cn(
                "group flex items-center gap-1 rounded-md",
                isDragging ? "opacity-70 ring-1 ring-blue-500" : "opacity-100"
            )}
        >
            <Link
                href={item.href}
                className={cn(
                    "flex-1 flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-zinc-800 hover:text-white transition-colors",
                    active ? "bg-zinc-800 text-white" : "text-zinc-300"
                )}
            >
                <item.icon className="mr-2 h-4 w-4" />
                {item.title}
            </Link>
            <button
                type="button"
                {...attributes}
                {...listeners}
                className="h-8 w-7 rounded-md text-zinc-500 hover:text-zinc-200 hover:bg-zinc-900 cursor-grab active:cursor-grabbing flex items-center justify-center"
                title="Drag to reorder favorites"
                aria-label={`Drag ${item.title}`}
            >
                <GripVertical className="h-3.5 w-3.5" />
            </button>
            <button
                type="button"
                onClick={() => onToggleFavorite(item.href)}
                className="h-8 w-8 rounded-md text-amber-300 hover:bg-zinc-900 flex items-center justify-center"
                title="Remove from favorites"
                aria-label={`Remove ${item.title} from favorites`}
            >
                <PinOff className="h-3.5 w-3.5" />
            </button>
        </div>
    );
}
