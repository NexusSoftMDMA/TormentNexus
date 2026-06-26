export class AutoTagger {

    // Simulates an LLM-based categorization of content
    public generateTags(content: string, title: string): string[] {
        const text = `${title} ${content}`.toLowerCase();
        const tags = new Set<string>();

        if (text.includes('react') || text.includes('frontend')) tags.add('ui');
        if (text.includes('database') || text.includes('sql') || text.includes('postgres')) tags.add('db');
        if (text.includes('mcp') || text.includes('protocol')) tags.add('mcp');
        if (text.includes('llm') || text.includes('ai') || text.includes('agent')) tags.add('ai');
        if (text.includes('go ') || text.includes('golang')) tags.add('golang');

        if (tags.size === 0) tags.add('uncategorized');

        return Array.from(tags);
    }
}
