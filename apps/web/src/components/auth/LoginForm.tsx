
'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { Button, Input } from '@tormentnexus/ui';

export function LoginForm() {
    const router = useRouter();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setIsSubmitting(true);

        try {
            const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password }),
            });
            const data = await res.json();
            if (!res.ok || !data?.ok) {
                setError(data?.error ?? 'Login failed. Please try again.');
                return;
            }
            router.push('/');
            router.refresh();
        } catch {
            setError('Network error while logging in.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
                <Input
                    type="email"
                    placeholder="Email address"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="bg-white dark:bg-zinc-950 border-zinc-200 dark:border-zinc-800 focus-visible:ring-blue-500/50"
                />
                <div className="relative">
                    <Input
                        type="password"
                        placeholder="Password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="bg-white dark:bg-zinc-950 border-zinc-200 dark:border-zinc-800 focus-visible:ring-blue-500/50"
                    />
                    <div className="absolute right-0 top-11">
                        <Link href="/forgot-password" className="text-xs text-zinc-400 hover:text-blue-500 z-10 relative mr-2">
                            Forgot?
                        </Link>
                    </div>
                </div>
            </div>

            {error && (
                <div className="text-sm text-red-500">{error}</div>
            )}

            <Button
                type="submit"
                disabled={isSubmitting}
                className="w-full bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 disabled:opacity-60"
            >
                {isSubmitting ? 'Signing In...' : 'Sign In'}
            </Button>
        </form>
    );
}
