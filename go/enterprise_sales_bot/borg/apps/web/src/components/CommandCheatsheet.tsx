"use client";
import React from 'react';
import { motion } from 'framer-motion';

const COMMANDS = [
    { cmd: '/help', desc: 'List all commands', icon: '❓' },
    { cmd: '/git status', desc: 'Show repo status', icon: '📊' },
    { cmd: '/context add [file]', desc: 'Pin file to context', icon: '📌' },
    { cmd: '/director status', desc: 'Agent status info', icon: '🤖' },
    { cmd: '/council debate [topic]', desc: 'Start AI debate', icon: '🏛️' },
    { cmd: '/squad spawn [task]', desc: 'Create worker agent', icon: '👥' },
    { cmd: '/test run', desc: 'Execute test suite', icon: '🧪' },
    { cmd: '/heal', desc: 'Auto-fix errors', icon: '💚' },
];

export function CommandCheatsheet() {
    return (
        <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-zinc-900 to-black border border-zinc-800 p-4">
            {/* Background Glow */}
            <div className="absolute inset-0 opacity-10 bg-gradient-to-br from-green-500 to-emerald-500 blur-3xl" />

            <div className="relative z-10">
                <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-bold text-white flex items-center gap-2">
                        <span className="text-2xl">⌨️</span>
                        Slash Commands
                    </h3>
                    <span className="text-[10px] px-2 py-1 bg-green-500/20 text-green-400 rounded-full font-mono">
                        {COMMANDS.length} available
                    </span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-64 overflow-y-auto pr-1 custom-scrollbar">
                    {COMMANDS.map((c, i) => (
                        <motion.div
                            key={c.cmd}
                            initial={{ opacity: 0, x: -10 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ delay: i * 0.05 }}
                            className="group p-2 bg-zinc-800/50 hover:bg-zinc-800 rounded-lg transition-all cursor-pointer border border-transparent hover:border-green-500/30"
                        >
                            <div className="flex items-center gap-2">
                                <span className="text-lg opacity-70 group-hover:opacity-100 transition-opacity">{c.icon}</span>
                                <div className="flex-1 min-w-0">
                                    <code className="text-xs text-green-400 font-mono block truncate">{c.cmd}</code>
                                    <span className="text-[10px] text-zinc-500 block truncate">{c.desc}</span>
                                </div>
                            </div>
                        </motion.div>
                    ))}
                </div>

                <div className="mt-3 pt-3 border-t border-zinc-800 text-center">
                    <span className="text-[10px] text-zinc-600">Type commands in Director Chat</span>
                </div>
            </div>
        </div>
    );
}
