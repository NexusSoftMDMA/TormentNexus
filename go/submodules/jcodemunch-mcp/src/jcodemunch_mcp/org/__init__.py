"""Org-rollup telemetry (team SKU).

Self-hosted model: a seat records its token-savings into a SQLite store on the
org host; ``org_rollup`` aggregates per org. The store and rollup are
transport-agnostic — local recording works today; a cross-machine HTTP front
door (seats POSTing to the org host) layers on top without changing this core.
"""

from .store import record_seat_report, org_rollup

__all__ = ["record_seat_report", "org_rollup"]
