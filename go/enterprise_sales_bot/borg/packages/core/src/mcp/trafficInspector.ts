import type { MCPTrafficEvent } from './types.js';

function formatPrimitive(value: unknown): string {
    if (typeof value === 'string') {
        return value.length > 40 ? `${value.slice(0, 37)}...` : value;
    }

    if (typeof value === 'number' || typeof value === 'boolean' || value === null) {
        return String(value);
    }

    if (Array.isArray(value)) {
        return `[${value.length} items]`;
    }

    if (typeof value === 'object' && value !== null) {
        return `{${Object.keys(value as Record<string, unknown>).slice(0, 4).join(', ')}}`;
    }

    return typeof value;
}

export function summarizeParams(args: unknown): string {
    if (!args || typeof args !== 'object' || Array.isArray(args)) {
        return formatPrimitive(args);
    }

    return Object.entries(args as Record<string, unknown>)
        .slice(0, 5)
        .map(([key, value]) => `${key}=${formatPrimitive(value)}`)
        .join(', ');
}

export class MCPTrafficInspector {
    private readonly events: MCPTrafficEvent[] = [];

    constructor(private readonly maxEvents: number = 200) {}

    public record(event: MCPTrafficEvent): void {
        this.events.push(event);
        if (this.events.length > this.maxEvents) {
            this.events.splice(0, this.events.length - this.maxEvents);
        }
    }

    public getEvents(): MCPTrafficEvent[] {
        return [...this.events];
    }
}
