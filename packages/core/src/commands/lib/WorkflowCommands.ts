import { ICommand, CommandResult } from "../CommandRegistry.js";
import { spawnAsync } from "../../utils/exec.js";

/**
 * /undo - Undo last commit or changes
 */
export class UndoCommand implements ICommand {
    name = "undo";
    description = "Undo operations. Usage: /undo <commit|changes|staged> [file]";

    async execute(args: string[]): Promise<CommandResult> {
        const subcommand = args[0] || 'help';
        const target = args[1]; // Correctly handle single file target if provided

        try {
            let output = '';
            switch (subcommand) {
                case 'commit':
                    const commitRes = await spawnAsync('git', ['reset', '--soft', 'HEAD~1'], { cwd: process.cwd() });
                    output = `↩️ **Undo Last Commit**\nCommit undone (changes preserved in staging)\n${commitRes.stdout}`;
                    break;

                case 'changes':
                    if (target) {
                        await spawnAsync('git', ['checkout', '--', target], { cwd: process.cwd() });
                        output = `↩️ **Undo Changes**: ${target}\nUnstaged changes discarded.`;
                    } else {
                        await spawnAsync('git', ['checkout', '--', '.'], { cwd: process.cwd() });
                        output = `↩️ **Undo All Changes**\nAll unstaged changes discarded.`;
                    }
                    break;

                case 'staged':
                    if (target) {
                        await spawnAsync('git', ['reset', 'HEAD', target], { cwd: process.cwd() });
                        output = `↩️ **Unstaged**: ${target}`;
                    } else {
                        await spawnAsync('git', ['reset', 'HEAD'], { cwd: process.cwd() });
                        output = `↩️ **Unstaged All Files**`;
                    }
                    break;

                default:
                    output = `**Usage**: /undo <commit|changes|staged> [file]\n- **commit**: Undo last commit (soft reset)\n- **changes [file]**: Discard unstaged changes\n- **staged [file]**: Unstage files`;
            }

            return { handled: true, output };
        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Undo Error:\n```\n${error.message}\n````
            };
        }
    }
}

/**
 * /diff - Show diff with formatting
 */
export class DiffCommand implements ICommand {
    name = "diff";
    description = "Show diff. Usage: /diff [staged|file] [path]";

    async execute(args: string[]): Promise<CommandResult> {
        try {
            let gitArgs = ['diff'];
            let label = 'Working Directory';

            if (args[0] === 'staged') {
                gitArgs = ['diff', '--cached'];
                label = 'Staged Changes';
            } else if (args[0]) {
                // Handle optional multiple file args
                gitArgs = ['diff', '--', ...args];
                label = args.join(' ');
            }

            const result = await spawnAsync('git', gitArgs, { cwd: process.cwd() });
            const stdout = result.stdout;

            if (!stdout.trim()) {
                return { handled: true, output: `📋 **Diff: ${label}**\nNo changes detected.` };
            }

            const maxLen = 3000;
            const truncated = stdout.length > maxLen
                ? stdout.substring(0, maxLen) + '\n... (truncated)'
                : stdout;

            return {
                handled: true,
                output: `📋 **Diff: ${label}**\n```diff\n${truncated}\n````
            };
        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Diff Error:\n```\n${error.message}\n````
            };
        }
    }
}

/**
 * /stash - Quick stash operations
 */
export class StashCommand implements ICommand {
    name = "stash";
    description = "Stash operations. Usage: /stash [push|pop|list|show]";

    async execute(args: string[]): Promise<CommandResult> {
        const subcommand = args[0] || 'push';
        const message = args.slice(1).join(' ');

        try {
            let output = '';
            let gitArgs = ['stash', subcommand];

            switch (subcommand) {
                case 'push':
                    if (message) {
                        gitArgs.push('-m', message);
                    }
                    const pushRes = await spawnAsync('git', gitArgs, { cwd: process.cwd() });
                    output = `📦 **Stash Push**\n${pushRes.stdout || 'Changes stashed.'}`;
                    break;

                case 'pop':
                case 'list':
                    const res = await spawnAsync('git', gitArgs, { cwd: process.cwd() });
                    output = `✅ **Stash ${subcommand}**\n${res.stdout || res.stderr || '(empty)'}`;
                    break;

                case 'show':
                    const showRes = await spawnAsync('git', ['stash', 'show', '-p'], { cwd: process.cwd() });
                    const truncated = showRes.stdout.length > 2000
                        ? showRes.stdout.substring(0, 2000) + '\n... (truncated)'
                        : showRes.stdout;
                    output = `📋 **Stash Show**\n```diff\n${truncated || '(empty)'}\n````
                    break;

                default:
                    output = `**Usage**: /stash <push|pop|list|show> [message]`;
            }

            return { handled: true, output };
        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Stash Error:\n```\n${error.message}\n````
            };
        }
    }
}

/**
 * /fix - Start Auto-Dev Loops (Fix until Pass)
 */
import { AutoDevService } from "../../services/AutoDevService.js";

export class FixCommand implements ICommand {
    name = "fix";
    description = "Start Auto-Dev Loop. Usage: /fix <test|lint|build|status|cancel> [target]";

    constructor(private autoDevGetter: () => AutoDevService | undefined) { }

    async execute(args: string[]): Promise<CommandResult> {
        const autoDev = this.autoDevGetter();
        if (!autoDev) return { handled: true, output: "❌ AutoDevService not initialized." };

        const subcommand = args[0];
        const target = args.slice(1).join(' ');

        if (['test', 'lint', 'build'].includes(subcommand)) {
            const id = await autoDev.startLoop({
                type: subcommand as 'test' | 'lint' | 'build',
                maxAttempts: 5,
                target: target || undefined
            });
            return { handled: true, output: `🔄 **Auto-Dev Loop Started**\nID: \`${id}\`\nType: ${subcommand}\nTarget: ${target || 'All'}\n\nRunning in background... Check status with \`/fix status\`.` };
        }

        if (subcommand === 'status') {
            const loops = autoDev.getLoops();
            if (loops.length === 0) return { handled: true, output: "✅ No active auto-dev loops." };

            let output = "🔄 **Active Loops**\n\n";
            for (const loop of loops) {
                output += `- **${loop.id}**: ${loop.config.type} ${loop.config.target ? `(${loop.config.target})` : ''}\n`;
                output += `  - Status: ${loop.status.toUpperCase()}\n`;
                output += `  - Attempt: ${loop.currentAttempt}/${loop.config.maxAttempts}\n`;
            }
            return { handled: true, output };
        }

        if (subcommand === 'cancel') {
            const id = args[1];
            if (!id) return { handled: true, output: "❌ Usage: /fix cancel <loop-id>" };

            const success = autoDev.cancelLoop(id);
            return {
                handled: true,
                output: success ? `🛑 Loop \`${id}\` cancelled.` : `❌ Loop \`${id}\` not found or not running.`
            };
        }

        return { handled: true, output: "❌ Usage: /fix <test|lint|build|status|cancel> [target]" };
    }
}
