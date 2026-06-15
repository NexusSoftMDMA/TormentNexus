# Launch Metrics Dashboard

> Track these numbers daily during the first 2 weeks post-launch. Copy the template below into a spreadsheet or Notion table.

---

## Tracking Cadence

| Phase | Frequency |
|---|---|
| Day 0–3 (Hot zone) | Every 4–6 hours |
| Day 4–7 | Daily |
| Week 2 | Every 2–3 days |
| Ongoing | Weekly |

---

## Metrics Template

### GitHub

| Metric | Baseline | Day 1 | Day 3 | Day 7 | Day 14 |
|---|---|---|---|---|---|
| Stars | | | | | |
| Forks | | | | | |
| Unique clones (14d, via Insights) | | | | | |
| Views / unique visitors (14d) | | | | | |
| Open issues | | | | | |
| Closed issues | | | | | |
| Open PRs | | | | | |
| Merged PRs since launch | | | | | |
| Bug issues opened | | | | | |
| Feature request issues opened | | | | | |
| Bug:feature ratio | | | | | |

> **Note:** GitHub Insights → Traffic tab resets every 14 days. Screenshot it on Day 1 and Day 14.

---

### HN / Reddit

| Metric | Value |
|---|---|
| HN post points | |
| HN comment count | |
| HN rank (peak) | |
| Reddit upvotes | |
| Reddit comment count | |
| Reddit crosspost count | |
| Top repeated ask / question | |
| Most common objection | |
| Time to agent's first HN reply | |
| Time to agent's first Reddit reply | |

---

### Community / Ecosystem

| Metric | Baseline | Day 7 | Day 14 |
|---|---|---|---|
| Discord / Slack members | | | |
| Newsletter subscribers | | | |
| Twitter/X followers | | | |
| Mentions (GitHub, Twitter, LinkedIn) | | | |
| External blog posts / write-ups | | | |
| YouTube / demo video views | | | |

---

### Reliability (Self-hosted users)

| Metric | Target | Actual |
|---|---|---|
| Reported install failures | 0 | |
| P0/P1 bugs filed | 0 | |
| Time-to-first-response on P0 | < 2h | |
| Time-to-fix on P0 | < 24h | |
| Stale issues (> 7 days unanswered) | 0 | |

---

## Signal Interpretation Guide

| Signal | What It Means | Action |
|---|---|---|
| High stars, low clones | Interest but friction in setup | Simplify README / Quick Start |
| High clones, high bug issues | Active adoption + rough edges | Patch and cut a hotfix release |
| Many "how do I…" issues | Docs gap | Expand QUICKSTART or FAQ |
| Many feature requests on same theme | Product-market signal | Add to ROADMAP, acknowledge in issue |
| HN/Reddit silence after post | Off-peak timing or weak hook | Re-post with revised title during prime hours |
| Security question goes viral | Visibility risk or genuine gap | Respond within 1h; link to LAUNCH_SECURITY_FAQ.md |

---

## KPIs — First 30 Days

Set realistic targets before launch to avoid anchoring on arbitrary numbers mid-stream.

| KPI | Conservative | Optimistic |
|---|---|---|
| GitHub stars | 200 | 1 000 |
| GitHub clones (unique) | 80 | 400 |
| Issues filed | 10 | 60 |
| PRs merged from community | 1 | 10 |
| Repeat visitors (14-day window) | 25% | 45% |
| Zero unaddressed P0 bugs | ✅ | ✅ |

---

## Weekly Summary Template

```
## Week N Summary (YYYY-MM-DD → YYYY-MM-DD)

**Stars:** [current] (+[delta])
**Clones (unique):** [current]
**Open issues:** [current] | Closed this week: [delta]
**Merged PRs:** [delta]

**Top asks / themes this week:**
1. [theme]
2. [theme]
3. [theme]

**Blockers / regressions:**
- [none / describe]

**Priority for next week:**
- [ ] [task]
- [ ] [task]
```

---

## References

- [GitHub Traffic Insights](https://github.com/NexusSoftMDMA/TormentNexus/graphs/traffic)
- [GitHub Issues](https://github.com/NexusSoftMDMA/TormentNexus/issues)
- [LAUNCH_POSTS.md](LAUNCH_POSTS.md)
- [LAUNCH_SECURITY_FAQ.md](LAUNCH_SECURITY_FAQ.md)
