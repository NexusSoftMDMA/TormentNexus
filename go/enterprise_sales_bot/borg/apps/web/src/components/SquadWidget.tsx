"use client";
import React, { useState } from 'react';
import { trpc } from '../utils/trpc';
import { motion, AnimatePresence } from 'framer-motion';

export const SquadWidget: React.FC = () => {
    // @ts-ignore
    const { data: members, isLoading } = trpc.squad.list.useQuery(undefined, {
        refetchInterval: 3000 // Real-time pulse
    });

    // @ts-ignore
    const spawnMutation = trpc.squad.spawn.useMutation({
        onSuccess: () => {
            // @ts-ignore
            utils.squad.list.invalidate();
            setShowSpawnModal(false);
            setGoal('');
            setBranch('');
        }
    });

    // @ts-ignore
    const killMutation = trpc.squad.kill.useMutation({
        onSuccess: () => {
            // @ts-ignore
            utils.squad.list.invalidate();
        }
    });

    const utils = trpc.useContext();
    const [showSpawnModal, setShowSpawnModal] = useState(false);
    const [branch, setBranch] = useState('');
    const [goal, setGoal] = useState('');

    const handleSpawn = () => {
        if (!branch || !goal) return;
        spawnMutation.mutate({ branch, goal });
    };

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 flex flex-col h-full min-h-[300px]">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold text-white flex items-center">
                    <span className="mr-2">🤖</span> Agent Squad
                </h2>
                <button
                    onClick={() => setShowSpawnModal(true)}
                    className="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded text-sm transition-colors"
                >
                    + Spawn Agent
                </button>
            </div>

            <div className="flex-1 overflow-y-auto space-y-3">
                {isLoading && <div className="text-gray-500 text-center py-4">Scanning worktrees...</div>}

                {members && members.length === 0 && !isLoading && (
                    <div className="text-gray-500 text-center py-8 border-2 border-dashed border-gray-800 rounded-lg">
                        <p>No active agents.</p>
                        <p className="text-sm mt-1">Spawn one to offload tasks.</p>
                    </div>
                )}

                <AnimatePresence>
                    {members && members.map((member: any) => (
                        <motion.div
                            key={member.id}
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, scale: 0.95 }}
                            className="bg-gray-800 rounded p-3 border-l-4 border-blue-500 flex justify-between items-center"
                        >
                            <div>
                                <div className="flex items-center space-x-2">
                                    <span className="font-mono text-blue-300 font-bold">{member.branch}</span>
                                    <span className={`px-2 py-0.5 rounded text-xs ${member.status === 'busy' ? 'bg-yellow-900 text-yellow-200 animate-pulse' :
                                        member.status === 'finished' ? 'bg-green-900 text-green-200' :
                                            'bg-gray-700 text-gray-300'
                                        }`}>
                                        {member.status.toUpperCase()}
                                    </span>
                                </div>
                                <div className="text-xs text-gray-400 mt-1">
                                    ID: {member.id} | Active: {member.active ? 'YES' : 'NO'}
                                </div>
                            </div>

                            <button
                                onClick={() => killMutation.mutate({ branch: member.branch })}
                                className="text-red-500 hover:text-red-400 p-2"
                                title="Terminate Agent"
                            >
                                ✖
                            </button>
                        </motion.div>
                    ))}
                </AnimatePresence>
            </div>

            {/* Spawn Modal */}
            <AnimatePresence>
                {showSpawnModal && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 backdrop-blur-sm"
                    >
                        <motion.div
                            initial={{ scale: 0.9 }}
                            animate={{ scale: 1 }}
                            className="bg-gray-900 border border-gray-700 p-6 rounded-lg w-[500px] shadow-2xl"
                        >
                            <h3 className="text-lg font-bold text-white mb-4">Spawn Autonomous Agent</h3>

                            <div className="space-y-4">
                                <div>
                                    <label className="block text-gray-400 text-sm mb-1">Branch Name (Workspace)</label>
                                    <input
                                        value={branch}
                                        onChange={(e) => setBranch(e.target.value)}
                                        placeholder="e.g., feature/login-page"
                                        className="w-full bg-black border border-gray-700 rounded p-2 text-white font-mono"
                                    />
                                </div>
                                <div>
                                    <label className="block text-gray-400 text-sm mb-1">Mission / Goal</label>
                                    <textarea
                                        value={goal}
                                        onChange={(e) => setGoal(e.target.value)}
                                        placeholder="Detailed instructions for the agent..."
                                        rows={4}
                                        className="w-full bg-black border border-gray-700 rounded p-2 text-white"
                                    />
                                </div>
                            </div>

                            <div className="flex justify-end space-x-3 mt-6">
                                <button
                                    onClick={() => setShowSpawnModal(false)}
                                    className="px-4 py-2 text-gray-400 hover:text-white"
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={handleSpawn}
                                    disabled={!branch || !goal || (spawnMutation as any).isPending || (spawnMutation as any).isLoading}
                                    className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded font-bold disabled:opacity-50"
                                >
                                    {(spawnMutation as any).isPending || (spawnMutation as any).isLoading ? 'Spawning...' : 'Deploy Agent'}
                                </button>
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
};
