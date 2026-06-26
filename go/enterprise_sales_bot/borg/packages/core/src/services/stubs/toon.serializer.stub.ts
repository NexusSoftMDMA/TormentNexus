import { parse, stringify } from 'yaml';

export const toonSerializer = {
    /**
     * Serializes JSON/JavaScript objects into Tool-Optimized Output Notation (TOON).
     * TOON utilizes YAML for heavy token-reduction (stripping redundant brackets/quotes)
     * and wraps the payload in distinctive `<toon>` tags for reliable LLM parsing.
     */
    serialize: (data: unknown): string => {
        try {
            if (data === undefined || data === null) return "<toon>null</toon>";
            if (typeof data === 'string' && data.trim() === '') return "<toon></toon>";
            const yamlStr = stringify(data, { indent: 2, lineWidth: -1 }).trim();
            return `<toon>\n${yamlStr}\n</toon>`;
        } catch (e) {
            // Fallback to JSON if circular refs or YAML failure occurs
            return JSON.stringify(data);
        }
    },

    /**
     * Parses TOON formatted strings back into JavaScript objects.
     */
    parse: (toonStr: string): unknown => {
        try {
            const match = toonStr.match(/<toon>([\s\S]*?)<\/toon>/);
            if (match) {
                return parse(match[1]);
            }
            // If it doesn't have tags, try standard JSON parse
            return JSON.parse(toonStr);
        } catch (e) {
            return null;
        }
    },
    
    /**
     * Detects if a string is a TOON formatted payload.
     */
    isToon: (str: string): boolean => {
        return typeof str === 'string' && str.trim().startsWith('<toon>') && str.trim().endsWith('</toon>');
    }
};
