import { ICommand, CommandResult } from "../CommandRegistry.js";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

export class GitCommand implements ICommand {
    name = "git";
    description = "Execute git operations. Usage: /git <status|add|commit|push|pull|log> [args]";

    async execute(args: string[]): Promise<CommandResult> {
        if (args.length === 0) {
            return { handled: true, output: "❌ Usage: /git <subcommand> [args]" };
        }

        const subcommand = args[0];
        const restArgs = args.slice(1).join(" "); // Keep quotes mostly intact? Warning: shell injection risk if not careful.
        // For local tool, we trust the input mostly, but direct exec is risky.
        // Ideally we use a proper git library or array-based spawn.
        // For now, limited subcommands.

        if (!['status', 'add', 'commit', 'push', 'pull', 'log', 'diff'].includes(subcommand)) {
            return { handled: true, output: `❌ Unsupported git subcommand: ${subcommand}. Allowed: status, add, commit, push, pull, log, diff.` };
        }

        try {
            // Safe construction?
            const commandLine = `git ${subcommand} ${restArgs}`;
            const { stdout, stderr } = await execAsync(commandLine, { cwd: process.cwd() });

            let output = stdout || stderr;
            if (subcommand === 'status') {
                output = "📂 **Git Status**\n```\n" + output + "\n```";
            } else {
                output = `✅ **Git ${subcommand}**\n\`\`\`\n${output}\n\`\`\``;
            }

            return {
                handled: true,
                output: output
            };

        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Git Error:\n\`\`\`\n${error.message}\n\`\`\``
            };
        }
    }
}
