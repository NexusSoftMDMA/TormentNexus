
import React, { useState, useEffect } from 'react';
import { Box, Text } from 'ink';

export interface LogEntry {
    id: string;
    timestamp: string;
    level: 'info' | 'warn' | 'error' | 'debug';
    message: string;
    source?: string;
}

export const EventLog = ({ logs }: { logs: LogEntry[] }) => {
    // Show last 5 logs
    const visibleLogs = logs.slice(-5);

    return (
        <Box flexDirection="column" borderStyle="single" borderColor="gray" minHeight={7}>
            <Box marginBottom={1}><Text bold>The Pulse</Text></Box>
            {visibleLogs.map((log) => (
                <Box key={log.id}>
                    <Text color="gray">[{log.timestamp.split('T')[1].split('.')[0]}] </Text>
                    <Text color={log.level === 'error' ? 'red' : log.level === 'warn' ? 'yellow' : 'white'}>
                        {log.message}
                    </Text>
                </Box>
            ))}
            {visibleLogs.length === 0 && <Text color="gray">No activity...</Text>}
        </Box>
    );
};
