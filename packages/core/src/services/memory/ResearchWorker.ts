import { LLMService } from "@tormentnexus/ai";
import { JSDOM } from 'jsdom';

export interface Metadata {
    title: string;
    description: string;
    favicon?: string;
    tags: string[];
    researchedAt: string;
}

export class ResearchWorker {
    constructor(private llmService: LLMService) {}

    /**
     * Researches a URL to extract metadata and generate semantic tags.
     */
    async research(url: string): Promise<Metadata> {
        console.log(`[Core:Memory] Researching URL: ${url}`);
        
        try {
            const response = await fetch(url, {
                headers: {
                    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const html = await response.text();
            const { window } = new JSDOM(html);
            const { document } = window;

            const title = this.extractTitle(document);
            const description = this.extractDescription(document);
            const favicon = this.extractFavicon(document, url);

            const tags = await this.generateTags(title, description, url);

            return {
                title,
                description,
                favicon,
                tags,
                researchedAt: new Date().toISOString()
            };
        } catch (error) {
            console.error(`[Core:Memory] Failed to research ${url}:`, error);
            return {
                title: url,
                description: "Failed to extract description.",
                tags: ['failed-research'],
                researchedAt: new Date().toISOString()
            };
        }
    }

    private extractTitle(doc: Document): string {
        return (
            doc.querySelector('meta[property="og:title"]')?.getAttribute('content') ||
            doc.querySelector('meta[name="twitter:title"]')?.getAttribute('content') ||
            doc.querySelector('title')?.textContent ||
            ""
        ).trim();
    }

    private extractDescription(doc: Document): string {
        return (
            doc.querySelector('meta[property="og:description"]')?.getAttribute('content') ||
            doc.querySelector('meta[name="description"]')?.getAttribute('content') ||
            doc.querySelector('meta[name="twitter:description"]')?.getAttribute('content') ||
            ""
        ).trim();
    }

    private extractFavicon(doc: Document, baseUrl: string): string | undefined {
        const icon = doc.querySelector('link[rel*="icon"]')?.getAttribute('href');
        if (icon) {
            return new URL(icon, baseUrl).toString();
        }
        return new URL('/favicon.ico', baseUrl).toString();
    }

    private fallbackTags(title: string, description: string, url: string): string[] {
        const hostname = (() => {
            try {
                return new URL(url).hostname.replace(/^www\./, '');
            } catch {
                return '';
            }
        })();

        const rawTokens = [title, description, hostname]
            .join(' ')
            .toLowerCase()
            .replace(/[^a-z0-9\s-]/g, ' ')
            .split(/\s+/)
            .filter(Boolean);

        const stopWords = new Set([
            'the', 'and', 'for', 'with', 'from', 'that', 'this', 'your', 'into', 'about', 'have', 'http', 'https',
            'www', 'com', 'org', 'net', 'io', 'are', 'was', 'were', 'has', 'had', 'you', 'our', 'their', 'its',
        ]);

        const uniqueTags = Array.from(new Set(rawTokens.filter((token) => token.length > 2 && !stopWords.has(token))));
        return uniqueTags.slice(0, 5).length > 0 ? uniqueTags.slice(0, 5) : ['uncategorized'];
    }

    private async generateTags(title: string, description: string, url: string): Promise<string[]> {
        const prompt = `
            Analyze this web page and generate 3-5 relevant, lowercase, single-word semantic tags.
            TITLE: ${title}
            DESCRIPTION: ${description}
            URL: ${url}
            
            Respond with ONLY a comma-separated list of tags.
        `;

        try {
            const selectedModel = await this.llmService.modelSelector.selectModel({
                taskType: 'worker',
                routingTaskType: 'research',
            });

            const response = await this.llmService.generateText(
                selectedModel.provider,
                selectedModel.modelId,
                'You generate concise semantic tags for web page metadata.',
                prompt,
                {
                taskType: 'worker',
                routingTaskType: 'research',
                }
            );
            
            const tags = response.content
                .split(',')
                .map(t => t.trim().toLowerCase())
                .filter(t => t.length > 0);

            return tags.length > 0 ? tags : this.fallbackTags(title, description, url);
        } catch {
            return this.fallbackTags(title, description, url);
        }
    }
}
