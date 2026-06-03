# Ideas for Improvement: TormentNexus Extension

Creative improvement ideas to evolve the TormentNexus browser extension from a bridge into a universal agent interface.
# Ideas for Improvement: tormentnexus Extension

Creative improvement ideas to evolve the tormentnexus browser extension from a bridge into a universal agent interface.

## 1. Universal Intelligence Integration
- **WASM-Based MCP Server Runner:** Allow the extension to download and run MCP servers directly within the browser using WASM, bypassing the need for a local Node.js process for simple tools.
- **Deep DOM Injection & "Agent Lens":** An overlay that identifies and "labels" elements on any webpage that an agent can interact with, providing a visual map of the page's "API."
- **Contextual "Magic Bar":** A floating command bar that appears only when relevant context is detected (e.g., show "TormentNexus Refactor" when on a GitHub PR page).

## 2. Advanced Browsing & Memory
- **Automatic Knowledge Harvesting:** Silently capture and summarize relevant information as the operator browses (e.g., documentation pages, technical blogs) and sync it to the TormentNexus memory swarm.
- **Contextual "Magic Bar":** A floating command bar that appears only when relevant context is detected (e.g., show "tormentnexus Refactor" when on a GitHub PR page).

## 2. Advanced Browsing & Memory
- **Automatic Knowledge Harvesting:** Silently capture and summarize relevant information as the operator browses (e.g., documentation pages, technical blogs) and sync it to the tormentnexus memory swarm.
- **"Ghost Browser" Sessions:** Spawn headless browser instances from the extension to perform research tasks or verify UI changes without interrupting the operator's active tab.
- **Browser-Native RAG:** A local vector search index running within the extension (via IndexedDB and a small WASM embedding model) for near-instant retrieval of browsed content.

## 3. Decentralized Economy & Mesh
- **Bobcoin-Powered Micro-Payments:** Integrate with the Bobcoin ledger to allow agents to pay for premium MCP tool usage or specific browsing data.
- **Extension-to-Extension P2P:** Allow multiple browser extensions on the same local network to share active tabs and browsed context directly, without going through the orchestrator.
- **Verified Fact Gossip:** When an agent learns a new fact on a webpage, the extension should "gossip" a cryptographically signed proof of that fact to the wider TormentNexus mesh.
- **Verified Fact Gossip:** When an agent learns a new fact on a webpage, the extension should "gossip" a cryptographically signed proof of that fact to the wider tormentnexus mesh.

## 4. Architectural Enhancements
- **WebExt-Bridge Event Bus:** Refactor the Background Worker into a robust, type-safe event bus using `webext-bridge` for more reliable messaging between components.
- **TanStack Query Migration:** Move all MCP tool execution and result caching to `@tanstack/react-query` for automatic retries, background updates, and consistent state management.
- **Shadow DOM Isolation:** Ensure all extension UI elements are perfectly isolated using Shadow DOM to prevent any styling or script interference from the host webpage.
