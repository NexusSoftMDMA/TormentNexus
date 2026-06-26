
## [v1.0.0-alpha.37] - 2026-03-07

### Added
- **Fit Markdown Scraper Filter:** Modified `internal/sync/linkcrawler.go` `extractVisibleText` to aggressively strip out `<nav>`, `<header>`, `<footer>`, and `<aside>` HTML tags before converting text. This removes substantial HTML boilerplate, enabling token optimizations mapped to the "Progressive Context Optimization" architectural goal.
