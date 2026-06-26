# Phase 3: Fleet Orchestration

## 1. TormentNexus Fleet Manager
The Go sidecar will act as the "Ground Control" for multiple concurrent TormentNexus sessions.

### Functional Requirements
- **Session Registry**: Track all active CLI/TIU instances across the host.
- **PID Monitoring**: Real-time health checks of child processes.
- **Auto-Restart**: Automatically respawn crashed sessions with context restoration.
- **Resource Balancing**: Manage CPU/Memory allocation across the fleet.

## 2. Shared Organizational Memory
Facts learned in "Session A" should be immediately available to "Session B" via the L2 Vault.

### Sync Logic
- **Passive Harvesting**: The `TrafficObserver` in Go extracts facts from A2A signals.
- **Global Broadcast**: High-value facts are emitted via `EventBus` to all active sessions.
- **Heat Injection**: When a session hits a similar problem, "hot" memories from previous sessions are injected into the L1 Scratchpad.

## 3. Remote Fleet Control
Control the entire local fleet from the mobile or web dashboard via a unified API surface.
