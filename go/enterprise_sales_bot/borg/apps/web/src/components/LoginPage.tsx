"use client";
import React, { useState } from 'react';
import { motion } from 'framer-motion';

interface LoginPageProps {
    onLogin: () => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onLogin }) => {
    const [code, setCode] = useState('');
    const [error, setError] = useState(false);
    const [loading, setLoading] = useState(false);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(false);

        // Simple mock auth for visual effect
        setTimeout(() => {
            if (code === 'admin' || code === 'tormentnexus' || code === '') { // Allow empty for ease of dev
                onLogin();
            } else {
                setError(true);
                setLoading(false);
            }
        }, 1500);
    };

    return (
        <div className="fixed inset-0 bg-black flex items-center justify-center overflow-hidden z-50">
            {/* Animated Background */}
            <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20" />
            <div className="absolute inset-0 bg-gradient-to-br from-indigo-900/30 via-black to-cyan-900/20" />

            {/* Floating Orbs */}
            <motion.div
                animate={{ x: [0, 100, 0], y: [0, -50, 0] }}
                transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
                className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-600/20 rounded-full blur-[100px]"
            />
            <motion.div
                animate={{ x: [0, -100, 0], y: [0, 50, 0] }}
                transition={{ duration: 25, repeat: Infinity, ease: "linear" }}
                className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-cyan-600/10 rounded-full blur-[100px]"
            />

            <motion.div
                initial={{ opacity: 0, scale: 0.9, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                transition={{ duration: 0.8, ease: "easeOut" }}
                className="relative z-10 w-full max-w-md p-8"
            >
                {/* Glass Card */}
                <div className="absolute inset-0 bg-zinc-900/80 backdrop-blur-xl rounded-2xl border border-zinc-800 shadow-2xl" />

                <div className="relative z-20 flex flex-col items-center">
                    {/* Logo / Icon */}
                    <div className="w-16 h-16 mb-6 rounded-xl bg-gradient-to-tr from-blue-500 to-cyan-400 flex items-center justify-center shadow-lg shadow-blue-500/20">
                        <span className="text-3xl">💠</span>
                    </div>

                    <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-b from-white to-zinc-400 mb-2">
                        TormentNexus OS
                    </h1>
                    <p className="text-zinc-500 text-sm mb-8 tracking-wide">SYSTEM ACCESS REQUIRED</p>

                    <form onSubmit={handleSubmit} className="w-full space-y-4">
                        <div className="relative group">
                            <input
                                type="password"
                                value={code}
                                onChange={(e) => setCode(e.target.value)}
                                placeholder="Access Key"
                                className="w-full bg-black/50 border border-zinc-700 rounded-lg px-4 py-3 text-center text-white placeholder-zinc-600 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500/50 transition-all font-mono"
                                autoFocus
                            />
                            <div className="absolute inset-0 rounded-lg bg-gradient-to-r from-blue-500/20 to-cyan-500/20 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                        </div>

                        {error && (
                            <motion.p
                                initial={{ opacity: 0, y: -10 }}
                                animate={{ opacity: 1, y: 0 }}
                                className="text-red-500 text-xs text-center"
                            >
                                ACCESS DENIED. KEY INVALID.
                            </motion.p>
                        )}

                        <button
                            type="submit"
                            disabled={loading}
                            className={`w-full py-3 rounded-lg font-bold text-sm tracking-widest uppercase transition-all
                                ${loading
                                    ? 'bg-zinc-800 text-zinc-500 cursor-not-allowed'
                                    : 'bg-white text-black hover:bg-zinc-200 hover:shadow-lg hover:shadow-white/10'
                                }
                            `}
                        >
                            {loading ? (
                                <span className="flex items-center justify-center gap-2">
                                    <span className="w-2 h-2 bg-zinc-500 rounded-full animate-bounce" />
                                    <span className="w-2 h-2 bg-zinc-500 rounded-full animate-bounce delay-75" />
                                    <span className="w-2 h-2 bg-zinc-500 rounded-full animate-bounce delay-150" />
                                </span>
                            ) : (
                                "Initialize"
                            )}
                        </button>
                    </form>

                    <div className="mt-8 flex items-center gap-4 text-[10px] text-zinc-600 uppercase tracking-widest">
                        <span>Ver 0.1.0</span>
                        <span className="w-1 h-1 bg-zinc-700 rounded-full" />
                        <span>Secure Connection</span>
                    </div>
                </div>
            </motion.div>
        </div>
    );
};
