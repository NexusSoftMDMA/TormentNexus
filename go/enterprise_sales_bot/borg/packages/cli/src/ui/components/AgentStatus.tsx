
import React from 'react';
import { Box, Text } from 'ink';

export interface AgentState {
    name: string;
    status: 'idle' | 'thinking' | 'working' | 'error';
    task?: string;
}

export const AgentStatus = ({ agents }: { agents: AgentState[] }) => {
    return (
        <Box flexDirection="column" borderStyle="single" borderColor="magenta" minHeight={5}>
            <Box marginBottom={1}><Text bold>Agents</Text></Box>
            {agents.map((agent) => (
                <Box key={agent.name} justifyContent="space-between">
                    <Text bold color="cyan">{agent.name}</Text>
                    <Text color={
                        agent.status === 'working' ? 'green' :
                            agent.status === 'thinking' ? 'yellow' :
                                agent.status === 'error' ? 'red' : 'gray'
                    }>
                        {agent.status.toUpperCase()}
                    </Text>
                </Box>
            ))}
            {agents.length === 0 && <Text color="gray">No active agents.</Text>}
        </Box>
    );
};
