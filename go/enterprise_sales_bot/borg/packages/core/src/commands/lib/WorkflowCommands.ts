import { ICommand, CommandResult } from "../CommandRegistry.js";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

/**
 * /undo - Undo last commit or changes
 */
export class UndoCommand implements ICommand {
    name = "undo";
    description = "Undo operations. Usage: /undo <commit|changes|staged> [file]";

    async execute(args: string[]): Promise<CommandResult> {
        const subcommand = args[0] || 'help';
        const target = args.slice(1).join(' ');

        try {
            let output = '';
            switch (subcommand) {
                case 'commit':
                    // Soft reset last commit
                    const { stdout: undoOut } = await execAsync('git reset --soft HEAD~1', { cwd: process.cwd() });
                    output = `↩️ **Undo Last Commit**\nCommit undone (changes preserved in staging)\n${undoOut}`;
                    break;

                case 'changes':
                    // Discard unstaged changes
                    if (target) {
                        await execAsync(`git checkout -- "${target}"`, { cwd: process.cwd() });
                        output = `↩️ **Undo Changes**: ${target}\nUnstaged changes discarded.`;
                    } else {
                        await execAsync('git checkout -- .', { cwd: process.cwd() });
                        output = `↩️ **Undo All Changes**\nAll unstaged changes discarded.`;
                    }
                    break;

                case 'staged':
                    // Unstage files
                    if (target) {
                        await execAsync(`git reset HEAD "${target}"`, { cwd: process.cwd() });
                        output = `↩️ **Unstaged**: ${target}`;
                    } else {
                        await execAsync('git reset HEAD', { cwd: process.cwd() });
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
                output: `❌ Undo Error:\n\`\`\`\n${error.message}\n\`\`\``
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
            let command = 'git diff';
            let label = 'Working Directory';

            if (args[0] === 'staged') {
                command = 'git diff --cached';
                label = 'Staged Changes';
            } else if (args[0]) {
                command = `git diff -- "${args.join(' ')}"`;
                label = args.join(' ');
            }

            const { stdout } = await execAsync(command, { cwd: process.cwd() });

            if (!stdout.trim()) {
                return { handled: true, output: `📋 **Diff: ${label}**\nNo changes detected.` };
            }

            // Truncate if too long
            const maxLen = 3000;
            const truncated = stdout.length > maxLen
                ? stdout.substring(0, maxLen) + '\n... (truncated)'
                : stdout;

            return {
                handled: true,
                output: `📋 **Diff: ${label}**\n\`\`\`diff\n${truncated}\n\`\`\``
            };
        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Diff Error:\n\`\`\`\n${error.message}\n\`\`\``
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
            switch (subcommand) {
                case 'push':
                    const pushCmd = message ? `git stash push -m "${message}"` : 'git stash push';
                    const { stdout: pushOut } = await execAsync(pushCmd, { cwd: process.cwd() });
                    output = `📦 **Stash Push**\n${pushOut || 'Changes stashed.'}`;
                    break;

                case 'pop':
                    const { stdout: popOut } = await execAsync('git stash pop', { cwd: process.cwd() });
                    output = `📤 **Stash Pop**\n${popOut || 'Stash applied and dropped.'}`;
                    break;

                case 'list':
                    const { stdout: listOut } = await execAsync('git stash list', { cwd: process.cwd() });
                    output = `📋 **Stash List**\n\`\`\`\n${listOut || '(empty)'}\n\`\`\``;
                    break;

                case 'show':
                    const { stdout: showOut } = await execAsync('git stash show -p', { cwd: process.cwd() });
                    const truncated = showOut.length > 2000
                        ? showOut.substring(0, 2000) + '\n... (truncated)'
                        : showOut;
                    output = `📋 **Stash Show**\n\`\`\`diff\n${truncated || '(empty)'}\n\`\`\``;
                    break;

                default:
                    output = `**Usage**: /stash <push|pop|list|show> [message]`;
            }

            return { handled: true, output };
        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Stash Error:\n\`\`\`\n${error.message}\n\`\`\``
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
            // Start Loop
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
