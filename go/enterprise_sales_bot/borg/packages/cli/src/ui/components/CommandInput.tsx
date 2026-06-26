
import React, { useState } from 'react';
import { Box, Text } from 'ink';
// @ts-ignore
import TextInput from 'ink-text-input';

export const CommandInput = ({ onSubmit }: { onSubmit: (cmd: string) => void }) => {
    const [query, setQuery] = useState('');

    const InkTextInput = TextInput as any;

    return (
        <Box borderStyle="round" borderColor="green">
            <Text bold color="green">➜ </Text>
            <InkTextInput
                value={query}
                onChange={setQuery}
                onSubmit={(val: string) => {
                    onSubmit(val);
                    setQuery('');
                }}
                placeholder="Enter command for Director..."
            />
        </Box>
    );
};
