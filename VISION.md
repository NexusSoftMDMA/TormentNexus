# Vision: AI Hypervisor (Nexus & HyperCode)

## North Star: The AI Hypervisor
**Nexus** is the underlying coordination kernel and "AI Hypervisor." **HyperCode** is the flagship, developer-facing autonomous coding runtime powered by Nexus.

## Architectural Layers
- **Nexus Runtime (/kernel/runtime)**: The deterministic execution kernel.
- **Nexus Memory (/kernel/memory)**: Active L1/L2 memory with Heat-based promotion.
- **Nexus Router (/kernel/router)**: Semantic tool selection and waterfall provider fallback.
- **Nexus Control Plane (/control-plane)**: UI and observability dashboards.

## Implemented Foundations
### Phase 1: Neural OS Memory (v1.0.0-alpha.56)
Implemented the **Active Memory Layer**. Memories are scored, promoted, and demoted based on real-world utility ("Heat") and temporal relevance.
