"use client";

import React, { useState } from "react";

export default function AutopilotDashboardPage() {
    const [loading, setLoading] = useState(false);
    // Assuming Autopilot server serves its web UI at this port/path
    // If it's pure API, we might need a custom UI construction here.
    // Based on README, it has a "Web Dashboard". Let's assume root or specific path.
    const autopilotUrl = "http://localhost:3847";

    return (
        <div className="w-full h-full flex flex-col">
            <div className="p-4 border-b border-gray-800 flex justify-between items-center bg-gray-900">
                <div>
                    <h1 className="text-xl font-bold text-white">OpenCode Autopilot</h1>
                    <p className="text-gray-400 text-sm">Multi-Model AI Council & Governance</p>
                </div>
                <div className="flex gap-2">
                    <a
                        href={autopilotUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="px-3 py-1.5 bg-purple-600 hover:bg-purple-700 text-white rounded text-sm transition-colors"
                    >
                        Open Standalone
                    </a>
                </div>
            </div>
            <div className="flex-1 relative bg-black">
                <iframe
                    src={autopilotUrl}
                    className="w-full h-full border-none"
                    title="Autopilot Dashboard"
                    allow="clipboard-read; clipboard-write"
                />
            </div>
        </div>
    );
}
