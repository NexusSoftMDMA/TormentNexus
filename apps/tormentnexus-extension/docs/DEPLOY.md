# Deployment Instructions

> How to build and deploy TormentNexus Extension for Chrome, Edge, and Firefox.
> How to build and deploy tormentnexus Extension for Chrome, Edge, and Firefox.

## Prerequisites

- **Node.js**: ≥22.12.0
- **pnpm**: ≥9.15.1
- **Git**: For version control

## Build

```bash
# Install dependencies
pnpm install

# Build for Chrome/Edge (production)
pnpm build

# Build for Firefox (production)
pnpm build:firefox

# Package as ZIP for store submission
pnpm zip          # Chrome/Edge
pnpm zip:firefox  # Firefox
```

Both builds output to `dist/`. They overwrite each other — only one target can exist at a time.

## Loading Locally (Development)

| Browser | URL | Steps |
|---------|-----|-------|
| **Chrome** | `chrome://extensions` | Enable Developer mode → Load unpacked → select `dist/` |
| **Edge** | `edge://extensions` | Enable Developer mode → Load unpacked → select `dist/` |
| **Firefox** | `about:debugging` | This Firefox → Load Temporary Add-on → select `dist/manifest.json` |

## Development (Watch Mode)

```bash
# Start dev server (Chrome/Edge, with HMR)
pnpm dev

# Start dev server (Firefox, with HMR)
pnpm dev:firefox
```

Changes to source files will hot-reload the extension automatically.

## Publishing to Stores

### Chrome Web Store
1. Run `pnpm zip` to create the distribution ZIP.
2. Go to [Chrome Developer Dashboard](https://chrome.google.com/webstore/devconsole).
3. Upload the ZIP from the project root.
4. Fill in listing details and submit for review.

### Firefox Add-ons (AMO)
1. Run `pnpm zip:firefox` to create the Firefox-compatible ZIP.
2. Go to [Firefox Add-on Developer Hub](https://addons.mozilla.org/developers/).
3. Upload the ZIP and submit for review.

### Edge Add-ons
1. The Chrome ZIP (`pnpm zip`) works for Edge as well.
2. Go to [Edge Partner Center](https://partner.microsoft.com/en-us/dashboard/microsoftedge/overview).
3. Upload the same Chrome ZIP and submit.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CLI_CEB_FIREFOX` | `false` | Set to `true` for Firefox builds |
| `CLI_CEB_DEV` | `false` | Set to `true` for development mode |
| `NODE_ENV` | `production` | Build environment |

## Version Bumping

```bash
# Update version everywhere (VERSION, package.json files)
pnpm update-version <new_version>

# Example
pnpm update-version 0.8.0
```

Then update `CHANGELOG.md` and commit with version in the message:
```bash
git add -A && git commit -m "feat: <description> (v0.8.0)" && git push origin main
```
