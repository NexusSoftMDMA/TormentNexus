"use client";
import React from 'react';
import { motion } from 'framer-motion';

const FEATURES = [
    {
        category: "🤖 AI Agents",
        items: [
            { name: "Director Chat", desc: "Communicate with the autonomous Director agent" },
            { name: "Council Debate", desc: "Watch AI personas debate decisions (Architect, Product, Critic)" },
            { name: "Squad Control", desc: "Spawn parallel agents in separate git worktrees" }
        ]
    },
    {
        category: "📊 Monitoring",
        items: [
            { name: "System Health", desc: "Real-time CPU, Memory, and load monitoring" },
            { name: "Activity Pulse", desc: "Live event timeline visualization" },
            { name: "Latency Monitor", desc: "Track response times across services" },
            { name: "Audit Logs", desc: "Complete record of all agent actions" }
        ]
    },
    {
        category: "🛠️ Development",
        items: [
            { name: "Code Sandbox", desc: "Execute Python/Node code in secure Docker containers" },
            { name: "Test Status", desc: "Auto-test results with pass/fail metrics" },
            { name: "Shell History", desc: "Browse and search command history" },
            { name: "Knowledge Graph", desc: "Interactive visualization of codebase structure" }
        ]
    },
    {
        category: "🔐 Security",
        items: [
            { name: "Autonomy Control", desc: "Adjust agent permission levels (Low/Medium/High)" },
            { name: "Security Shield", desc: "Policy-based action restrictions" },
            { name: "Self-Healing", desc: "Automatic error detection and repair" }
        ]
    },
    {
        category: "⚡ Productivity",
        items: [
            { name: "Global Search", desc: "Semantic codebase search (Ctrl+K)" },
            { name: "Click-to-Open", desc: "Click nodes in graph to open in VS Code" },
            { name: "Suggestions", desc: "AI-generated proactive recommendations" },
            { name: "Command Runner", desc: "Execute shell commands directly" }
        ]
    }
];

export const HelpWidget: React.FC = () => {
    return (
        <div className="bg-gradient-to-br from-zinc-900 to-black rounded-lg border border-zinc-800 p-4 h-full overflow-y-auto">
            <div className="text-center mb-4">
                <span className="text-3xl">📖</span>
                <h2 className="text-lg font-bold text-white mt-1">Feature Guide</h2>
                <p className="text-xs text-zinc-500">All capabilities at your fingertips</p>
            </div>

            <div className="space-y-4">
                {FEATURES.map((section, i) => (
                    <motion.div
                        key={i}
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: i * 0.1 }}
                        className="border border-zinc-800 rounded-lg overflow-hidden"
                    >
                        <div className="bg-zinc-800/50 px-3 py-2 font-bold text-sm text-zinc-300">
                            {section.category}
                        </div>
                        <div className="p-2 space-y-1">
                            {section.items.map((item, j) => (
                                <div key={j} className="flex items-start gap-2 p-2 hover:bg-zinc-800/30 rounded transition-colors">
                                    <span className="text-blue-400 font-medium text-xs whitespace-nowrap">{item.name}</span>
                                    <span className="text-zinc-500 text-xs">{item.desc}</span>
                                </div>
                            ))}
                        </div>
                    </motion.div>
                ))}
            </div>

            <div className="mt-4 text-center text-[10px] text-zinc-600 border-t border-zinc-800 pt-3">
                <p>Drag widgets to rearrange • Layout auto-saves</p>
                <p className="mt-1">Access Key: <code className="bg-zinc-800 px-1 rounded">admin</code> or <code className="bg-zinc-800 px-1 rounded">tormentnexus</code></p>
            </div>
        </div>
    );
};
