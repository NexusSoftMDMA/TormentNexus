
'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { motion } from 'framer-motion';

import { Button, Input } from '@tormentnexus/ui';

export default function ForgotPasswordPage() {
    const [email, setEmail] = useState('');
    const [sent, setSent] = useState(false);
    const [resetUrl, setResetUrl] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setIsSubmitting(true);

        try {
            const res = await fetch('/api/auth/forgot-password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email }),
            });
            const data = await res.json();
            if (!res.ok || !data?.ok) {
                setError(data?.error ?? 'Unable to process reset request.');
                return;
            }
            setResetUrl(typeof data?.resetUrl === 'string' ? data.resetUrl : null);
            setSent(true);
        } catch {
            setError('Network error while requesting reset link.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-black relative overflow-hidden">
            {/* Background Ambience */}
            <div className="absolute inset-0 z-0">
                <div className="absolute top-[-20%] left-[20%] w-[50%] h-[50%] bg-blue-500/20 rounded-full blur-[120px]" />
            </div>

            <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="relative z-10 w-full max-w-md p-8 bg-white/50 dark:bg-zinc-900/50 backdrop-blur-xl border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-2xl"
            >
                <div className="mb-8 text-center">
                    <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
                        Reset Password
                    </h1>
                    <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-sm">
                        Enter your email to receive a reset link
                    </p>
                </div>

                {!sent ? (
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <Input
                            type="email"
                            placeholder="Email address"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="bg-white dark:bg-zinc-950 border-zinc-200 dark:border-zinc-800 focus-visible:ring-blue-500/50"
                            required
                        />
                        {error && <p className="text-sm text-red-500">{error}</p>}
                        <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20">
                            {isSubmitting ? 'Sending...' : 'Send Reset Link'}
                        </Button>
                    </form>
                ) : (
                    <div className="text-center space-y-4">
                        <div className="w-12 h-12 bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 rounded-full flex items-center justify-center mx-auto text-xl">
                            ✓
                        </div>
                        <p className="text-zinc-600 dark:text-zinc-300">
                            Check your email for instructions to reset your password.
                        </p>
                        {resetUrl && (
                            <Link
                                href={resetUrl}
                                className="inline-flex items-center justify-center px-4 py-2 rounded-md text-sm bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20"
                            >
                                Continue to Reset Password
                            </Link>
                        )}
                    </div>
                )}

                <div className="mt-6 text-center text-xs">
                    <Link href="/login" className="text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
                        ← Back to Login
                    </Link>
                </div>
            </motion.div>
        </div>
    );
}
