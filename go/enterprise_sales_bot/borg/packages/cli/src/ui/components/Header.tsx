
import React from 'react';
import { Box, Text } from 'ink';

export const Header = ({ version = "0.0.1", status = "offline" }: { version?: string, status?: string }) => {
    return (
        <Box borderStyle="round" borderColor="cyan" paddingX={1} justifyContent="space-between">
            <Text bold color="cyan">TORMENTNEXUS HIVE MIND</Text>
            <Text>v{version}</Text>
            <Text color={status === 'online' ? 'green' : 'red'}>● {status.toUpperCase()}</Text>
        </Box>
    );
};
