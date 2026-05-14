import { ICommand, CommandResult } from "../CommandRegistry.js";
import { spawnAsync } from "../../utils/exec.js";

export class GitCommand implements ICommand {
    name = "git";
    description = "Execute git operations. Usage: /git <status|add|commit|push|pull|log|diff> [args]";

    async execute(args: string[]): Promise<CommandResult> {
        if (args.length === 0) {
            return { handled: true, output: "❌ Usage: /git <subcommand> [args]" };
        }

        const subcommand = args[0];
        const restArgs = args.slice(1);

        if (!['status', 'add', 'commit', 'push', 'pull', 'log', 'diff'].includes(subcommand)) {
            return { handled: true, output: `❌ Unsupported git subcommand: ${subcommand}. Allowed: status, add, commit, push, pull, log, diff.` };
        }

        try {
            const result = await spawnAsync("git", [subcommand, ...restArgs], { cwd: process.cwd() });

            let output = result.stdout || result.stderr;
            if (subcommand === 'status') {
                output = "📂 **Git Status**\n```\n" + output + "\n```";
            } else if (result.exitCode !== 0) {
                 output = `❌ Git Error (Exit ${result.exitCode}):\n```\n${result.stderr}\n``\*`;
            } else {
                output = `✅ **Git ${subcommand}**\n```\n${output}\n``\*`;
            }

            return {
                handled: true,
                output: output
            };

        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Execution Error:\n```\n${error.message}\n``\*`
            };
        }
    }
}
