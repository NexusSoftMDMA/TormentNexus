
export interface CommandResult {
    handled: boolean;
    output: string;
    error?: string;
}

export interface ICommand {
    name: string;        // e.g. "git"
    description: string;
    execute(args: string[]): Promise<CommandResult>;
}

export class CommandRegistry {
    private commands: Map<string, ICommand> = new Map();

    constructor() { }

    public register(command: ICommand) {
        this.commands.set(command.name, command);
        console.log(`[CommandRegistry] Registered command: /${command.name}`);
    }

    public async execute(input: string): Promise<CommandResult | null> {
        if (!input.trim().startsWith('/')) return null;

        const parts = input.trim().substring(1).split(' '); // Remove '/'
        const commandName = parts[0].toLowerCase();
        const args = parts.slice(1);

        const command = this.commands.get(commandName);
        if (!command) {
            return {
                handled: true, // We handled it by saying "Unknown command"
                output: `❌ Unknown command: /${commandName}. Type /help for a list.`
            };
        }

        try {
            console.log(`[CommandRegistry] Executing: /${commandName} ${args.join(' ')}`);
            return await command.execute(args);
        } catch (error: any) {
            return {
                handled: true,
                output: `❌ Error executing /${commandName}: ${error.message}`,
                error: error.message
            };
        }
    }

    public getCommands(): ICommand[] {
        return Array.from(this.commands.values());
    }
}
