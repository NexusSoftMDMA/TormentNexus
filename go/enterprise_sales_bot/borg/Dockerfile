# syntax=docker/dockerfile:1.7

# Stage 1: Builder
FROM node:20-slim AS builder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN apt-get update && apt-get install -y python3 make g++ && rm -rf /var/lib/apt/lists/*
RUN corepack enable

WORKDIR /app

# Copy root config
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.base.json turbo.json ./

# Copy package manifests for the minimal runtime workspace graph.
COPY packages/adk/package.json packages/adk/
COPY packages/agents/package.json packages/agents/
COPY packages/ai/package.json packages/ai/
COPY packages/core/package.json packages/core/
COPY packages/mcp-registry/package.json packages/mcp-registry/
COPY packages/memory/package.json packages/memory/
COPY packages/search/package.json packages/search/
COPY packages/tools/package.json packages/tools/
COPY packages/tsconfig/package.json packages/tsconfig/
COPY packages/types/package.json packages/types/
COPY packages/ui/package.json packages/ui/
COPY apps/web/package.json apps/web/

# Install dependencies (frozen lockfile)
RUN pnpm config set store-dir /pnpm/store \
 && pnpm config set fetch-retries 5 \
 && pnpm config set fetch-retry-mintimeout 10000 \
 && pnpm config set fetch-retry-maxtimeout 120000 \
 && pnpm config set network-concurrency 8

RUN --mount=type=cache,id=tormentnexus-pnpm-store-v2,target=/pnpm/store \
	pnpm install --frozen-lockfile

# Copy source code required to build the Core API and Web dashboard.
COPY packages/adk packages/adk
COPY packages/agents packages/agents
COPY packages/ai packages/ai
COPY packages/core packages/core
COPY packages/mcp-registry packages/mcp-registry
COPY packages/memory packages/memory
COPY packages/search packages/search
COPY packages/tools packages/tools
COPY packages/tsconfig packages/tsconfig
COPY packages/types packages/types
COPY packages/ui packages/ui
COPY apps/web apps/web
COPY scripts scripts

# Ignore any stale incremental TypeScript caches from the host checkout.
RUN find packages apps -name "*.tsbuildinfo" -delete

# Build only the packages required for runtime.
RUN pnpm -C packages/adk build \
 && pnpm -C packages/ai build \
 && pnpm -C packages/agents build \
 && pnpm -C packages/mcp-registry build \
 && pnpm -C packages/memory build \
 && pnpm -C packages/search build \
 && pnpm -C packages/tools build \
 && pnpm -C packages/types build \
 && pnpm -C packages/core build \
 && pnpm -C packages/ui build \
 && pnpm -C apps/web build --webpack

# Stage 2: Core Runner
FROM node:20-slim AS core
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app

# Copy built artifacts from builder
COPY --from=builder /app ./

# Expose Core port
EXPOSE 3000

# Start Core
CMD ["node", "packages/core/dist/index.js"]

# Stage 3: Web Runner
FROM node:20-slim AS web
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app

# Copy built artifacts from builder
COPY --from=builder /app ./

# Expose the internal Next.js listener port.
EXPOSE 3000

# Start Web
WORKDIR /app/apps/web
CMD ["pnpm", "start"]
