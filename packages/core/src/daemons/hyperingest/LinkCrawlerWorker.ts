import { JSDOM } from 'jsdom';
import { db } from '../../db/index.js';
import { linksBacklogTable } from '../../db/tormentnexus-schema.js';
import { eq, asc } from 'drizzle-orm';
import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@tormentnexus/ai';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../../db/sqliteAvailability.js';

export class LinkCrawlerWorker {
    private isRunning = false;
    private isProcessing = false;
    private interval: NodeJS.Timeout | null = null;
    private readonly llmService: LLMService;
    private sqliteUnavailableLogged = false;

    constructor(llmService: LLMService) {
        this.llmService = llmService;
    }

    public start(intervalMs: number = 60 * 1000): void {
        if (this.isRunning) return;
        this.isRunning = true;
        console.log('[HyperIngest] Starting Link Crawler Worker...');
        
        // Run immediately
        void this.processNextBatch();

        this.interval = setInterval(() => {
            void this.processNextBatch();
        }, intervalMs);
    }

    public stop(): void {
        this.isRunning = false;
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
        console.log('[HyperIngest] Stopped Link Crawler Worker.');
    }

    private async processNextBatch(): Promise<void> {
        if (this.isProcessing) return;
        this.isProcessing = true;

        try {
            // Find up to 5 pending links
            const pendingLinks = await db
                .select()
                .from(linksBacklogTable)
                .where(eq(linksBacklogTable.research_status, 'pending'))
                .orderBy(asc(linksBacklogTable.created_at))
                .limit(5);

            if (pendingLinks.length === 0) {
                this.isProcessing = false;
                return;
            }

            for (const link of pendingLinks) {
                if (!this.isRunning) break;

                console.log(`[HyperIngest] Crawling ${link.url}...`);
                
                // Mark as running
                await db.update(linksBacklogTable)
                    .set({ research_status: 'running', updated_at: new Date() })
                    .where(eq(linksBacklogTable.uuid, link.uuid));

                try {
                    const response = await fetch(link.url, {
                        headers: {
                            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36',
                            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8',
                            'Accept-Language': 'en-US,en;q=0.5',
                        },
                        signal: AbortSignal.timeout(10000)
                    });

                    if (!response.ok) {
                        throw new Error(`HTTP ${response.status}`);
                    }

                    const html = await response.text();
                    const dom = new JSDOM(html);
                    const document = dom.window.document;

                    const pageTitle = document.title?.trim() || document.querySelector('meta[property="og:title"]')?.getAttribute('content') || null;
                    const pageDescription = document.querySelector('meta[name="description"]')?.getAttribute('content') || document.querySelector('meta[property="og:description"]')?.getAttribute('content') || null;
                    
                    let faviconUrl = document.querySelector('link[rel="icon"]')?.getAttribute('href') || document.querySelector('link[rel="shortcut icon"]')?.getAttribute('href');
                    if (faviconUrl) {
                        try {
                            faviconUrl = new URL(faviconUrl, link.url).href;
                        } catch {
                            faviconUrl = null;
                        }
                    }

                    // Clean body text for LLM categorization
                    document.querySelectorAll('script, style, noscript, nav, header, footer').forEach(el => el.remove());
                    const bodyText = document.body?.textContent?.replace(/\s+/g, ' ').trim().slice(0, 5000) || '';

                    // Send to LLM for tags and summarization if needed
                    let tags = link.tags ?? [];
                    if (tags.length === 0 && bodyText.length > 50) {
                        try {
                            const llmResultStr = await this.llmService.generateText('openrouter', DEFAULT_OPENROUTER_FREE_MODEL, 'Extract categories and tags.', 
                                `Extract 3-5 concise, specific tags or categories for this webpage. Focus on the tools, topics, and programming languages mentioned. Return a JSON object with a "tags" array of strings.\n\nTitle: ${pageTitle}\nDesc: ${pageDescription}\nContent: ${bodyText.slice(0, 2000)}`
                            );
                            const jsonStr = llmResultStr.content.replace(/```json/g, '').replace(/```/g, '').trim();
                            const llmResult = JSON.parse(jsonStr);
                            if (llmResult && typeof llmResult === 'object' && 'tags' in llmResult && Array.isArray((llmResult as any).tags)) {
                                tags = (llmResult as any).tags.map(String);
                            }
                        } catch (e: any) {
                            console.warn(`[HyperIngest] LLM tag extraction failed for ${link.url}: ${e.message}`);
                        }
                    }

                    await db.update(linksBacklogTable)
                        .set({
                            page_title: pageTitle,
                            page_description: pageDescription,
                            favicon_url: faviconUrl,
                            tags,
                            research_status: 'done',
                            http_status: response.status,
                            researched_at: new Date(),
                            updated_at: new Date(),
                        })
                        .where(eq(linksBacklogTable.uuid, link.uuid));

                    console.log(`[HyperIngest] Successfully crawled and categorized ${link.url}`);
                } catch (error: any) {
                    console.error(`[HyperIngest] Failed to crawl ${link.url}:`, error.message);
                    await db.update(linksBacklogTable)
                        .set({
                            research_status: 'failed',
                            http_status: error.message.startsWith('HTTP') ? parseInt(error.message.split(' ')[1]) : null,
                            updated_at: new Date(),
                        })
                        .where(eq(linksBacklogTable.uuid, link.uuid));
                }
            }
        } catch (error) {
            if (isSqliteUnavailableError(error)) {
                if (!this.sqliteUnavailableLogged) {
                    console.warn(formatOptionalSqliteFailure(
                        '[HyperIngest] Link crawler backlog access is unavailable',
                        error,
                    ));
                    this.sqliteUnavailableLogged = true;
                }
            } else {
                console.error('[HyperIngest] Error in LinkCrawlerWorker:', error);
            }
        } finally {
            this.isProcessing = false;
        }
    }
}
