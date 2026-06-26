
import puppeteer, { Browser, Page } from 'puppeteer';
import TurndownService from 'turndown';

export class BrowserService {
    private browser: Browser | null = null;
    private pages: Map<string, Page> = new Map();
    private turndownService: TurndownService;

    constructor() {
        this.turndownService = new TurndownService();
    }

    async launch() {
        if (this.browser) return;
        console.log("[BrowserService] Launching Puppeteer...");
        this.browser = await puppeteer.launch({
            headless: true,
            args: ['--no-sandbox', '--disable-setuid-sandbox']
        });
    }

    getStatus(): { active: boolean; pageCount: number; pageIds: string[] } {
        return {
            active: this.browser !== null,
            pageCount: this.pages.size,
            pageIds: Array.from(this.pages.keys()),
        };
    }

    async navigate(url: string): Promise<{ id: string, title: string, content: string }> {
        if (!this.browser) await this.launch();
        const page = await this.browser!.newPage();

        // Block heavy resources
        await page.setRequestInterception(true);
        page.on('request', (req) => {
            if (['image', 'stylesheet', 'font', 'media'].includes(req.resourceType())) {
                req.abort();
            } else {
                req.continue();
            }
        });

        console.log(`[BrowserService] Navigating to ${url}...`);
        await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 });

        const id = Math.random().toString(36).substring(7);
        this.pages.set(id, page);

        const title = await page.title();
        const html = await page.content();
        // Convert a simplified version of the body to markdown
        const bodyContent = await page.evaluate(() => document.body.innerHTML);
        const content = this.turndownService.turndown(bodyContent).substring(0, 10000); // Limit context

        return { id, title, content };
    }

    async click(pageId: string, selector: string): Promise<void> {
        const page = this.pages.get(pageId);
        if (!page) throw new Error(`Page ${pageId} not found`);
        await page.click(selector);
    }

    async type(pageId: string, selector: string, text: string): Promise<void> {
        const page = this.pages.get(pageId);
        if (!page) throw new Error(`Page ${pageId} not found`);
        await page.type(selector, text);
    }

    async screenshot(pageId: string): Promise<string> {
        const page = this.pages.get(pageId);
        if (!page) throw new Error(`Page ${pageId} not found`);
        const buffer = await page.screenshot({ encoding: 'base64' });
        return buffer as string;
    }

    async extract(pageId: string): Promise<string> {
        const page = this.pages.get(pageId);
        if (!page) throw new Error(`Page ${pageId} not found`);
        const bodyContent = await page.evaluate(() => document.body.innerHTML);
        return this.turndownService.turndown(bodyContent);
    }

    async getContent(pageId: string): Promise<string> {
        return this.extract(pageId);
    }

    async close(pageId: string) {
        const page = this.pages.get(pageId);
        if (page) {
            await page.close();
            this.pages.delete(pageId);
        }
    }

    async closeAll() {
        if (this.browser) {
            await this.browser.close();
            this.browser = null;
            this.pages.clear();
        }
    }
}
