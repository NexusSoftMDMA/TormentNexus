/**
 * SymbolPinService - Track and manage pinned code symbols for context prioritization
 */

export interface PinnedSymbol {
    id: string;
    name: string;           // e.g. "MCPServer.executeTool"
    file: string;           // absolute path
    type: 'function' | 'class' | 'method' | 'variable' | 'interface';
    lineStart?: number;
    lineEnd?: number;
    priority: number;       // Higher = more important
    addedAt: number;        // timestamp
    notes?: string;         // User notes about why this is pinned
}

export class SymbolPinService {
    private pinnedSymbols: Map<string, PinnedSymbol> = new Map();

    constructor() { }

    /**
     * Pin a symbol to prioritize it in context
     */
    pin(symbol: Omit<PinnedSymbol, 'id' | 'addedAt' | 'priority'>): PinnedSymbol {
        const id = `${symbol.file}#${symbol.name}`;
        const existing = this.pinnedSymbols.get(id);

        const pinned: PinnedSymbol = {
            ...symbol,
            id,
            addedAt: Date.now(),
            priority: existing?.priority ?? this.pinnedSymbols.size + 1
        };

        this.pinnedSymbols.set(id, pinned);
        console.log(`[SymbolPin] 📌 Pinned: ${symbol.name} (${symbol.type})`);
        return pinned;
    }

    /**
     * Unpin a symbol
     */
    unpin(id: string): boolean {
        const existed = this.pinnedSymbols.has(id);
        this.pinnedSymbols.delete(id);
        if (existed) {
            console.log(`[SymbolPin] 🗑️ Unpinned: ${id}`);
        }
        return existed;
    }

    /**
     * Update priority of a pinned symbol
     */
    updatePriority(id: string, priority: number): boolean {
        const symbol = this.pinnedSymbols.get(id);
        if (symbol) {
            symbol.priority = priority;
            return true;
        }
        return false;
    }

    /**
     * Add notes to a pinned symbol
     */
    addNotes(id: string, notes: string): boolean {
        const symbol = this.pinnedSymbols.get(id);
        if (symbol) {
            symbol.notes = notes;
            return true;
        }
        return false;
    }

    /**
     * Get all pinned symbols, sorted by priority (highest first)
     */
    list(): PinnedSymbol[] {
        return Array.from(this.pinnedSymbols.values())
            .sort((a, b) => b.priority - a.priority);
    }

    /**
     * Get pinned symbols for a specific file
     */
    forFile(filePath: string): PinnedSymbol[] {
        return this.list().filter(s => s.file === filePath);
    }

    /**
     * Clear all pinned symbols
     */
    clear(): number {
        const count = this.pinnedSymbols.size;
        this.pinnedSymbols.clear();
        console.log(`[SymbolPin] Cleared ${count} pinned symbols`);
        return count;
    }

    /**
     * Get context prompt with prioritized symbols
     */
    getContextPrompt(): string {
        const symbols = this.list();
        if (symbols.length === 0) return '';

        let context = '📍 **PINNED SYMBOLS (High Priority)**\n\n';
        for (const sym of symbols) {
            context += `### ${sym.type}: \`${sym.name}\`\n`;
            context += `- File: \`${sym.file}\`\n`;
            if (sym.lineStart) {
                context += `- Lines: ${sym.lineStart}${sym.lineEnd ? `-${sym.lineEnd}` : ''}\n`;
            }
            if (sym.notes) {
                context += `- Notes: ${sym.notes}\n`;
            }
            context += '\n';
        }
        return context;
    }
}
