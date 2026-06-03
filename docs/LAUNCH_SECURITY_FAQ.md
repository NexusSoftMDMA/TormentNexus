# Launch Security FAQ (HN + Reddit)

Use this as a copy/paste response bank for launch-day comments.

## Short answers (1–3 lines)

### How do you handle secrets?
Secrets are injected through environment/config and should never be committed. tormentnexus is local-first, so secret handling follows your local runtime boundaries and key-scoping practices.

### What if an MCP server/tool is malicious?
Treat every external MCP server as outside your trust boundary. tormentnexus can route calls and improve visibility, but safety still depends on what you connect and enable.

### Does tormentnexus send data to a hosted backend?
Core operation is local-first and does not require a hosted tormentnexus backend. Any outbound data flow depends on the external providers/tools you configure.

### Is this production-ready?
Core workflows are usable now for advanced local/operator use. We are actively hardening policies, docs, and guardrails based on user feedback.

### How is this safer than direct client-to-tool wiring?
tormentnexus centralizes visibility and control (routing, logs, policy surfaces), making behavior easier to audit and constrain than ad-hoc direct wiring.

## Deeper technical follow-ups

### Threat model (concise)
- Trusted: host machine + user account running tormentnexus.
- Untrusted/outside boundary: remote MCP servers, third-party APIs, and any side-effect-capable external tool.
- Security posture: least privilege, explicit allowlists, scoped credentials, and audit visibility.

### Hardening guidance
- Run tormentnexus under least-privilege OS user.
- Scope API keys to minimum permissions; rotate regularly.
- Disable unused MCP servers/tools.
- Separate sensitive workloads by environment/project.
- Review logs for unexpected executions.

### Limitations to acknowledge publicly
- tormentnexus does not make an untrusted external tool trustworthy.
- Security quality depends on connected integrations and operator configuration.
- Some guardrails are still evolving and are part of active hardening.

## Copy/paste long reply

Totally fair question. tormentnexus is local-first and should be treated as a high-privilege operator process. The key security point is trust boundaries: your local host/user is the trusted base, while remote MCP servers and external APIs are outside that boundary. tormentnexus helps by centralizing control and visibility (routing, logs, policy surfaces), but it does not magically make an untrusted external tool trustworthy. In practice, safe operation means least privilege runtime, scoped keys, explicit server/tool allowlists, and regular audit review.

## Skeptic-proof response

Agreed — security claims should be concrete. We’re documenting explicit threat model assumptions and reproducible hardening guidance (least-privilege runtime, scoped credentials, trust boundaries, and audit workflows). If you have a specific failure mode in mind, we’d love to test against it and publish the result.

## High-friction questions (HN-style) + responses

### “This is just another wrapper. Why trust it?”
That’s a valid concern. The value is not “magic security,” it’s operational control: consistent routing, observability, and policy touchpoints across heterogeneous MCP servers.

### “Where are the benchmarks?”
We’re publishing reproducible benchmark scripts and environment details so claims are verifiable.

### “What data leaves my machine exactly?”
Only what your configured integrations send. tormentnexus itself runs local-first; outbound behavior depends on enabled providers/tools and your runtime config.

### “Can I run it fully offline?”
Core control-plane behavior is local. Offline capability for end-to-end workflows depends on whether your selected tools/providers require network access.

## Launch operator checklist (security-specific)

- Confirm no secrets in git (`.env` not committed, `.env.example` present).
- Verify enabled MCP server list is minimal.
- Disable non-essential side-effect tools before demos.
- Keep one quick architecture/trust-boundary diagram ready for replies.
- Log recurring security questions into README/FAQ updates.
