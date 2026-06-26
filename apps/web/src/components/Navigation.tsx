"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Sheet, SheetContent, SheetTrigger } from "@tormentnexus/ui";
import { Button, StreamStatus } from "@tormentnexus/ui";
import { Menu } from "lucide-react";
import { useState } from "react";

const NAV_ITEMS = [
    { href: '/dashboard?tab=home', label: 'Mission Control', color: 'hover:text-blue-500', activeColor: 'text-blue-500' },
    { href: '/dashboard/mcp', label: 'MCP Registry', color: 'hover:text-teal-500', activeColor: 'text-teal-500' },
    { href: '/dashboard/swarm', label: 'Agent Swarm', color: 'hover:text-purple-500', activeColor: 'text-purple-500' },
    { href: '/docs', label: 'Documentation', color: 'hover:text-blue-100', activeColor: 'text-blue-500' },
];

interface NavigationProps {
    versionLabel?: string;
}

export function Navigation({ versionLabel = 'dev' }: NavigationProps) {
    const pathname = usePathname();
    const [open, setOpen] = useState(false);

    const isActive = (path: string) => pathname === path;

    return (
        <nav className="w-full bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 px-6 py-4 flex items-center justify-between sticky top-0 z-50">
            <div className="flex items-center gap-6">
                <div className="text-xl font-bold bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
                    TORMENTNEXUS
                </div>

                {/* Desktop Navigation */}
                <div className="hidden md:flex gap-4">
                    {NAV_ITEMS.map((item) => (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={`text-sm font-medium transition-colors ${item.color} ${isActive(item.href) ? item.activeColor : 'text-zinc-500 dark:text-zinc-400'}`}
                        >
                            {item.label}
                        </Link>
                    ))}
                </div>
            </div>

            {/* Mobile Navigation */}
            <div className="md:hidden">
                <Sheet open={open} onOpenChange={setOpen}>
                    <SheetTrigger asChild>
                        <Button variant="ghost" size="icon">
                            <Menu className="h-6 w-6" />
                        </Button>
                    </SheetTrigger>
                    <SheetContent side="left" className="w-[300px] sm:w-[400px]">
                        <div className="flex flex-col gap-4 mt-8">
                            {NAV_ITEMS.map((item) => (
                                <Link
                                    key={item.href}
                                    href={item.href}
                                    onClick={() => setOpen(false)}
                                    className={`text-lg font-medium transition-colors ${item.color} ${isActive(item.href) ? item.activeColor : 'text-zinc-500 dark:text-zinc-400'}`}
                                >
                                    {item.label}
                                </Link>
                            ))}
                            <div className="mt-auto pt-8 border-t border-zinc-200 dark:border-zinc-800">
                                <span className="text-xs text-zinc-400">v{versionLabel}</span>
                            </div>
                        </div>
                    </SheetContent>
                </Sheet>
            </div>

            <div className="hidden md:flex items-center gap-4">
                <StreamStatus />
                <div className="text-xs text-zinc-400 border-l border-zinc-800 pl-4">
                    v{versionLabel}
                </div>
            </div>
        </nav>
    );
}
