---
title: Project Dashboards Feature
type: note
permalink: scarf/features/project-dashboards-feature
tags:
- dashboards
- feature
source_sha: 427321d742d63298100f9e444f96fd1524d7a46c
source_paths: scarf/scarf/Features, README.md
---

## Observations
- [feature] Project Dashboards are custom, agent-generated visualizations per project. Schema supports stat boxes, charts, tables, progress bars, checklists, rich text, and embedded web views — all defined in a simple JSON file. #schema
- [design] Dashboards are intended to be authored and maintained by the Hermes agent itself (agent writes the JSON; Scarf renders with live refresh). #agent-authored

## Relations
- documented_in [[Scarf Project Overview]]