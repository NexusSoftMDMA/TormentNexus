'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, MessageSquare, Terminal, Settings, LayoutDashboard, Users, Microscope, Network, Wrench, Zap } from 'lucide-react';
import { cn } from '../lib/utils';

export function MobileNav() {
    const pathname = usePathname();

    const items = [
        {
            href: '/dashboard/ecosystem',
            icon: Home,
            label: 'Home'
        },
        {
            href: '/sessions',
            icon: MessageSquare,
            label: 'Chat'
        },
        {
            href: '/autopilot',
            icon: Terminal,
            label: 'Agents'
        },
        {
            href: '/squads',
            icon: Users,
            label: 'Squads'
        },
        {
            href: '/research',
            icon: Microscope,
            label: 'Research'
        },
        {
            href: '/code/graph',
            icon: Network,
            label: 'Code'
        },
        {
            href: '/code/fix',
            icon: Wrench,
            label: 'Fix'
        },
        {
            href: '/code/symbols',
            icon: Zap,
            label: 'Symbols'
        },
        {
            href: '/conductor',
            icon: LayoutDashboard,
            label: 'Flows'
        },
        {
            href: '/settings',
            icon: Settings,
            label: 'Config'
        }
    ];

    return (
        <div className="md:hidden fixed bottom-0 left-0 right-0 h-16 bg-zinc-950 border-t border-white/10 flex items-center justify-around px-2 z-50 pb-safe">
            {items.map((item) => {
                const isActive = pathname === item.href;
                const Icon = item.icon;

                return (
                    <Link
                        key={item.href}
                        href={item.href}
                        className={cn(
                            "flex flex-col items-center justify-center w-full h-full space-y-1",
                            isActive ? "text-blue-400" : "text-zinc-500 hover:text-zinc-300"
                        )}
                    >
                        <Icon className="w-5 h-5" />
                        <span className="text-[10px] font-medium">{item.label}</span>
                    </Link>
                );
            })}
        </div>
    );
}
