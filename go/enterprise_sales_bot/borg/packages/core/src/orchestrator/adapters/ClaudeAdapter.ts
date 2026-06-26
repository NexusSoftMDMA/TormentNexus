import { AgentAdapter, AgentConfig } from '../AgentAdapter.js';

export class ClaudeAdapter extends AgentAdapter {
    constructor() {
        const config: AgentConfig = {
            name: 'Claude Code',
            command: 'claude', // Assumes 'claude' is in PATH. Alternatively: 'npx', args: ['@anthropic-ai/claude-code']
            args: ['--non-interactive'], // Attempt to force non-interactive mode if available, or just run
            // Note: Claude Code is highly interactive. Automation might be tricky without a PTY.
            // For now, we assume simple stdio piping works for basic prompts.
        };
        super(config);
    }

    // Claude specific overrides if needed
    // e.g. parsing special output markers
}
