import { describe, expect, it } from "vitest";
import { extractBookmarksFromPayload, normalizeBookmarkUrl } from "./bobby-bookmarks-adapter.js";

describe("BobbyBookmarksBacklogAdapter helpers", () => {
    it("normalizes bookmark URLs by stripping tracking params and hashes", () => {
        expect(
            normalizeBookmarkUrl("https://example.com/page?utm_source=test&q=tormentnexus&fbclid=abc#section")
        ).toBe("https://example.com/page?q=tormentnexus");
    });

    it("matches the shared BobbyBookmarks canonicalization rules for host, path, ports, and tracking params", () => {
        expect(
            normalizeBookmarkUrl("HTTPS://WWW.EXAMPLE.COM:443/Page/?utm_source=test&_ga=1&a=1#section")
        ).toBe("https://www.example.com/page?a=1");
    });

    it("adds an https scheme for bare URLs", () => {
        expect(normalizeBookmarkUrl("example.com/Page/")).toBe("https://example.com/page");
    });

    it("extracts bookmarks from common API payload shapes", () => {
        const bookmark = { id: 1, url: "https://example.com" };

        expect(extractBookmarksFromPayload([bookmark])).toEqual([bookmark]);
        expect(extractBookmarksFromPayload({ items: [bookmark] })).toEqual([bookmark]);
        expect(extractBookmarksFromPayload({ bookmarks: [bookmark] })).toEqual([bookmark]);
        expect(extractBookmarksFromPayload({ data: [bookmark] })).toEqual([bookmark]);
        expect(extractBookmarksFromPayload({ nope: [] })).toEqual([]);
    });
});
