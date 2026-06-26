import { linksBacklogRepository } from "../db/repositories/links-backlog.repo.js";

export interface BobbyBookmarksBookmark {
    id?: number;
    url?: string;
    normalized_url?: string;
    title?: string | null;
    description?: string | null;
    tags?: unknown;
    source?: string | null;
    is_duplicate?: boolean;
    duplicate_of?: number | string | null;
    research_status?: string | null;
    http_status?: number | null;
    page_title?: string | null;
    page_description?: string | null;
    favicon_url?: string | null;
    researched_at?: string | null;
    cluster_id?: number | string | null;
    import_session_id?: number | null;
    [key: string]: unknown;
}

export interface BobbyBookmarksSyncReport {
    source: "bobbybookmarks";
    fetched: number;
    upserted: number;
    pages: number;
    errors: string[];
    baseUrl: string;
}

export interface BobbyBookmarksAdapterOptions {
    baseUrl: string;
    perPage?: number;
    includeDuplicates?: boolean;
    includeResearched?: boolean;
}

function normalizeBaseUrl(baseUrl: string): string {
    return baseUrl.replace(/\/+$/, "");
}

function coerceTags(tags: unknown): string[] {
    if (Array.isArray(tags)) {
        return tags.filter((tag): tag is string => typeof tag === "string" && tag.trim().length > 0);
    }
    if (typeof tags === "string" && tags.trim().length > 0) {
        return tags.split(",").map((tag) => tag.trim()).filter(Boolean);
    }
    return [];
}

export function normalizeBookmarkUrl(rawUrl: string): string {
    const trimmed = rawUrl.trim();
    if (!trimmed) {
        return trimmed;
    }

    const candidate = /^[a-zA-Z][a-zA-Z0-9+\-.]*:\/\//.test(trimmed) ? trimmed : `https://${trimmed}`;

    try {
        const url = new URL(candidate);
        const blockedParams = new Set([
            "utm_source",
            "utm_medium",
            "utm_campaign",
            "utm_term",
            "utm_content",
            "utm_id",
            "utm_reader",
            "utm_name",
            "utm_cid",
            "fbclid",
            "gclid",
            "gclsrc",
            "dclid",
            "msclkid",
            "adrefer",
            "ref",
            "source",
            "mc_cid",
            "mc_eid",
            "zanpid",
            "openid",
            "_ga",
            "_gid",
            "igshid",
            "yclid",
            "twclid",
            "li_fat_id",
            "epik",
            "rdid",
            "ttclid",
            "wbraid",
            "gbraid",
            "srsltid",
        ]);
        const defaultPorts = new Map([
            ["http:", "80"],
            ["https:", "443"],
            ["ftp:", "21"],
        ]);

        const kept = Array.from(url.searchParams.entries())
            .filter(([key]) => {
                const normalizedKey = key.toLowerCase();
                return !blockedParams.has(normalizedKey) && !normalizedKey.startsWith("utm_");
            })
            .map(([key, value]) => [key.toLowerCase(), value] as const)
            .sort(([left], [right]) => left.localeCompare(right));

        url.protocol = url.protocol.toLowerCase();
        url.hostname = url.hostname.toLowerCase();
        if (url.port !== "" && defaultPorts.get(url.protocol) === url.port) {
            url.port = "";
        }
        url.pathname = url.pathname.toLowerCase();
        if (url.pathname !== "/" && url.pathname.endsWith("/")) {
            url.pathname = url.pathname.replace(/\/+$/, "");
        }
        if (!url.pathname) {
            url.pathname = "/";
        }
        url.search = "";
        for (const [key, value] of kept) {
            url.searchParams.append(key, value);
        }
        url.hash = "";
        return url.toString();
    } catch {
        return trimmed.toLowerCase();
    }
}

export function extractBookmarksFromPayload(payload: unknown): BobbyBookmarksBookmark[] {
    if (Array.isArray(payload)) {
        return payload as BobbyBookmarksBookmark[];
    }
    if (!payload || typeof payload !== "object") {
        return [];
    }

    const record = payload as Record<string, unknown>;
    if (Array.isArray(record.items)) {
        return record.items as BobbyBookmarksBookmark[];
    }
    if (Array.isArray(record.bookmarks)) {
        return record.bookmarks as BobbyBookmarksBookmark[];
    }
    if (Array.isArray(record.data)) {
        return record.data as BobbyBookmarksBookmark[];
    }

    return [];
}

async function fetchJson(url: string): Promise<unknown> {
    const response = await fetch(url, {
        headers: {
            Accept: "application/json",
            "User-Agent": "TormentNexus/BobbyBookmarks-Adapter",
        },
    });
    if (!response.ok) {
        throw new Error(`HTTP ${response.status} ${response.statusText}`);
    }
    return response.json();
}

export class BobbyBookmarksBacklogAdapter {
    private readonly baseUrl: string;
    private readonly perPage: number;
    private readonly includeDuplicates: boolean;
    private readonly includeResearched: boolean;

    constructor(options: BobbyBookmarksAdapterOptions) {
        this.baseUrl = normalizeBaseUrl(options.baseUrl);
        this.perPage = Math.min(Math.max(options.perPage ?? 100, 1), 250);
        this.includeDuplicates = options.includeDuplicates ?? true;
        this.includeResearched = options.includeResearched ?? true;
    }

    async sync(): Promise<BobbyBookmarksSyncReport> {
        const report: BobbyBookmarksSyncReport = {
            source: "bobbybookmarks",
            fetched: 0,
            upserted: 0,
            pages: 0,
            errors: [],
            baseUrl: this.baseUrl,
        };

        for (let page = 1; page <= 100; page += 1) {
            const url = new URL(`${this.baseUrl}/api/bookmarks`);
            url.searchParams.set("page", String(page));
            url.searchParams.set("per_page", String(this.perPage));
            if (this.includeDuplicates) {
                url.searchParams.set("show_duplicates", "true");
            }
            if (!this.includeResearched) {
                url.searchParams.set("research_status", "pending");
            }

            let payload: unknown;
            try {
                payload = await fetchJson(url.toString());
            } catch (error) {
                const message = error instanceof Error ? error.message : String(error);
                report.errors.push(`page ${page}: ${message}`);
                break;
            }

            const bookmarks = extractBookmarksFromPayload(payload);
            if (bookmarks.length === 0) {
                break;
            }

            report.pages += 1;
            report.fetched += bookmarks.length;

            for (const bookmark of bookmarks) {
                if (!bookmark || typeof bookmark.url !== "string" || bookmark.url.trim().length === 0) {
                    continue;
                }

                const normalizedUrl = typeof bookmark.normalized_url === "string" && bookmark.normalized_url.trim().length > 0
                    ? bookmark.normalized_url.trim()
                    : normalizeBookmarkUrl(bookmark.url);

                await linksBacklogRepository.upsertLink({
                    url: bookmark.url.trim(),
                    normalized_url: normalizedUrl,
                    title: typeof bookmark.title === "string" ? bookmark.title : null,
                    description: typeof bookmark.description === "string" ? bookmark.description : null,
                    tags: coerceTags(bookmark.tags),
                    source: "bobbybookmarks",
                    is_duplicate: Boolean(bookmark.is_duplicate),
                    duplicate_of: bookmark.duplicate_of != null ? String(bookmark.duplicate_of) : null,
                    research_status: typeof bookmark.research_status === "string" ? bookmark.research_status : "pending",
                    http_status: typeof bookmark.http_status === "number" ? bookmark.http_status : null,
                    page_title: typeof bookmark.page_title === "string" ? bookmark.page_title : null,
                    page_description: typeof bookmark.page_description === "string" ? bookmark.page_description : null,
                    favicon_url: typeof bookmark.favicon_url === "string" ? bookmark.favicon_url : null,
                    researched_at: typeof bookmark.researched_at === "string" ? new Date(bookmark.researched_at) : null,
                    cluster_id: bookmark.cluster_id != null ? String(bookmark.cluster_id) : null,
                    bobbybookmarks_bookmark_id: typeof bookmark.id === "number" ? bookmark.id : null,
                    import_session_id: typeof bookmark.import_session_id === "number" ? bookmark.import_session_id : null,
                    raw_payload: bookmark,
                    synced_at: new Date(),
                });
                report.upserted += 1;
            }

            if (bookmarks.length < this.perPage) {
                break;
            }
        }

        return report;
    }
}
