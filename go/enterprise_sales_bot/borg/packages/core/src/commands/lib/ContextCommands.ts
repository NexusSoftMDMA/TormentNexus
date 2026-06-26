import { ICommand, CommandResult } from "../CommandRegistry.js";
import { ContextManager } from "../../context/ContextManager.js";
import * as path from 'path';

export class ContextCommand implements ICommand {
    name = "context";
    description = "Manage pinned context files. Usage: /context <add|remove|list|clear> [file]";

    constructor(private contextManager: ContextManager) { }

    async execute(args: string[]): Promise<CommandResult> {
        if (args.length === 0) {
            return { handled: true, output: "Usage: /context <add|remove|list|clear> [file]" };
        }

        const subcommand = args[0];
        const target = args.slice(1).join(" ");

        try {
            if (subcommand === 'add') {
                if (!target) return { handled: true, output: "❌ Please specify a file path." };
                const msg = this.contextManager.add(target);
                return { handled: true, output: `✅ ${msg}` };
            }
            else if (subcommand === 'remove') {
                if (!target) return { handled: true, output: "❌ Please specify a file path or name." };
                const msg = this.contextManager.remove(target);
                return { handled: true, output: `✅ ${msg}` };
            }
            else if (subcommand === 'list') {
                const files = this.contextManager.list();
                if (files.length === 0) return { handled: true, output: "📭 No pinned files." };

                const list = files.map(f => `- ${path.relative(process.cwd(), f)}`).join('\n');
                return { handled: true, output: `📌 **Pinned Files**\n${list}` };
            }
            else if (subcommand === 'clear') {
                const msg = this.contextManager.clear();
                return { handled: true, output: `✅ ${msg}` };
            }
            else {
                return { handled: true, output: "Unknown subcommand. Use add, remove, list, or clear." };
            }
        } catch (error: any) {
            return { handled: true, output: `❌ Error: ${error.message}` };
        }
    }
}
