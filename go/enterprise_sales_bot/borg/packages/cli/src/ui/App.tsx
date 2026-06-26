import React, { useState, useEffect } from 'react';
import { Box, Text, useApp } from 'ink';
import { Header } from './components/Header.js';
import { EventLog, LogEntry } from './components/EventLog.js';
import { AgentStatus, AgentState } from './components/AgentStatus.js';
import { CommandInput } from './components/CommandInput.js';
import { LogEmitter } from '../logger-patch.js';

// We'll import these dynamically in useEffect to ensure they run in the same process context if needed,
// but for the TUI, we are the host.
// @ts-ignore -- path alias not resolved under NodeNext
import { startOrchestrator } from '@tormentnexus/core/orchestrator';

export const App = () => {
    const { exit } = useApp();
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [agents, setAgents] = useState<AgentState[]>([]);
    const [systemStatus, setSystemStatus] = useState('initializing');
    const [orchestratorStarted, setOrchestratorStarted] = useState(false);

    // 1. Capture Logs
    useEffect(() => {
        const handleLog = (entry: LogEntry) => {
            setLogs(prev => [...prev.slice(-49), entry]); // Keep last 50
        };
        LogEmitter.on('log', handleLog);
        return () => {
            LogEmitter.off('log', handleLog);
        };
    }, []);

    // 2. Start Orchestrator (Once)
    useEffect(() => {
        if (!orchestratorStarted) {
            setOrchestratorStarted(true);
            startOrchestrator().then(() => {
                setSystemStatus('online');
                LogEmitter.emit('log', {
                    id: 'sys', timestamp: new Date().toISOString(), level: 'info', message: 'Orchestrator Started via TUI'
                });
            }).catch(e => {
                setSystemStatus('error');
                LogEmitter.emit('log', {
                    id: 'sys', timestamp: new Date().toISOString(), level: 'error', message: `Startup Failed: ${e.message}`
                });
            });
        }
    }, [orchestratorStarted]);

    // 3. Poll for Agent Status (Use the API since Orchestrator exposes it)
    useEffect(() => {
        const poll = async () => {
            if (systemStatus !== 'online') return;
            try {
                // @ts-ignore
                const res = await fetch('http://localhost:3000/trpc/director.status');
                const data = await res.json();
                const status = data.result?.data;

                if (status) {
                    setAgents([
                        {
                            name: 'Director',
                            status: status.active ? 'working' : 'idle',
                            task: status.goal
                        },
                        // We could add more agents here if the API exposed them
                        { name: 'Council', status: 'idle' },
                        { name: 'Healer', status: 'idle' }
                    ]);
                }
            } catch (e) {
                // Ignore poll errors
            }
        };
        const interval = setInterval(poll, 1000);
        return () => clearInterval(interval);
    }, [systemStatus]);

    const handleCommand = async (cmd: string) => {
        if (cmd === '/exit') {
            exit();
            process.exit(0);
        }

        LogEmitter.emit('log', {
            id: 'usr', timestamp: new Date().toISOString(), level: 'info', message: `> ${cmd}`
        });

        // Send to Director via API
        try {
            // @ts-ignore
            await fetch('http://localhost:3000/trpc/director.ask', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ prompt: cmd })
            });
        } catch (e: any) {
            LogEmitter.emit('log', {
                id: 'err', timestamp: new Date().toISOString(), level: 'error', message: `Failed to send command: ${e.message}`
            });
        }
    };

    return (
        <Box flexDirection="column" padding={1} height="100%">
            <Header version="2.3.0" status={systemStatus} />
            <Box flexDirection="row" marginTop={1} flexGrow={1}>
                <Box width="70%" marginRight={1}>
                    <EventLog logs={logs} />
                </Box>
                <Box width="30%">
                    <AgentStatus agents={agents} />
                </Box>
            </Box>
            <Box marginTop={1}>
                <CommandInput onSubmit={handleCommand} />
            </Box>
        </Box>
    );
};
