"""Online license gate for the org-rollup (team SKU) feature.

**Scope: this gates ONLY org-rollup aggregation** — the per-org savings dashboard
an enterprise buyer signs off on. Individual indexing/search/every other tool is
free and never touches this module. Seat reporting (``org-report`` / the
``/org/report`` ingest) is also ungated, so trial data accrues before purchase.

Verification reuses the shared j*Munch licensing backend (``validate.php``) that
the desktop apps already use — the org host holds no secrets, it GETs the public
validate endpoint and trusts the JSON answer (``product=jcodemunch`` namespaces
our keys; Stripe webhooks populate the backend).

Two resilience rules ported from the desktop client:

* **Sticky offline.** A server-confirmed key stays valid through network
  failures; only an explicit ``{"valid": false}`` (revoked / expired / not found)
  blocks. A server outage never punishes a paying customer. State is cached at
  ``<CODE_INDEX_PATH>/license.json`` and re-confirmed at most every 7 days.
* **Grace window.** A 14-day clock (from the first unlicensed org-rollup attempt)
  lets a new org evaluate the dashboard before paying. After it lapses with no
  qualifying key, org-rollup hard-refuses with a pricing link.

org-rollup is a multi-seat/team feature, so it requires a multi-seat **tier**
(Studio or Platform — see ``ORG_TIERS``); a single-seat Builder license is valid
but does not include it, and lands in the same grace-then-upgrade path with an
"upgrade to Studio/Platform" message.
"""

from __future__ import annotations

import json
import logging
import math
import os
import time
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)

VALIDATE_URL = "https://j.gravelle.us/jCodeMunch/validate.php"
PRODUCT = "jcodemunch"
GET_LICENSE_URL = "https://j.gravelle.us/jCodeMunch/#pricing"

RECHECK_SECONDS = 7 * 24 * 60 * 60   # re-confirm a still-valid key at most this often
GRACE_SECONDS = 14 * 24 * 60 * 60    # evaluation window for a never-licensed org
REQUEST_TIMEOUT = 8.0

# org-rollup is a multi-seat/team feature, so it requires a multi-seat tier.
# Tiers (lowercased, as validate.php returns them): Builder=1 seat, Studio=5,
# Platform=unlimited. Builder is single-seat and does NOT include org-rollup.
ORG_TIERS = {"studio", "platform"}


# --------------------------------------------------------------------------- #
# Cached license state (sticky offline)
# --------------------------------------------------------------------------- #

def _state_path(storage_path: Optional[str] = None) -> Path:
    base = storage_path or os.environ.get("CODE_INDEX_PATH", str(Path.home() / ".code-index"))
    return Path(base) / "license.json"


def _load_state(storage_path: Optional[str] = None) -> dict:
    try:
        return json.loads(_state_path(storage_path).read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return {}


def _save_state(state: dict, storage_path: Optional[str] = None) -> None:
    try:
        p = _state_path(storage_path)
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(json.dumps(state), encoding="utf-8")
    except OSError as exc:
        logger.debug("could not persist license state (%s)", exc)


def mask_key(key: str) -> str:
    """First 4 + last 4 for display; never log a full key."""
    key = (key or "").strip()
    if len(key) <= 8:
        return "*" * len(key)
    return f"{key[:4]}…{key[-4:]}"


def _license_key() -> str:
    """Key from env (wins) or config; empty string when unset."""
    key = (os.environ.get("JCODEMUNCH_LICENSE_KEY") or "").strip()
    if key:
        return key
    try:
        from ..config import get as _config_get
        return (_config_get("license_key") or "").strip()
    except Exception:
        return ""


# --------------------------------------------------------------------------- #
# Server check
# --------------------------------------------------------------------------- #

def _check_server(key: str) -> Optional[dict]:
    """Return ``{"valid": bool, "tier"?, "error"?}`` from validate.php, or None
    when the server can't be reached (network error / unparseable body). None
    means "leave cached state alone" — the sticky-offline rule.

    The endpoint carries ``valid`` in the JSON body for both its 200 and 400
    responses, so the body is trusted over the status code."""
    try:
        import httpx
    except Exception:  # httpx is a core dep, but never let its absence crash the gate
        return None
    try:
        resp = httpx.get(
            VALIDATE_URL,
            params={"product": PRODUCT, "license": key},  # params auto-url-encodes the key
            timeout=REQUEST_TIMEOUT,
            follow_redirects=True,
        )
        data = resp.json()
    except Exception:
        return None
    if isinstance(data, dict) and isinstance(data.get("valid"), bool):
        return {"valid": data["valid"], "tier": data.get("tier"), "error": data.get("error")}
    return None


def _is_validated(key: str, storage_path: Optional[str] = None) -> dict:
    """Resolve a key to ``{valid: bool, tier, error, confirmed: bool}`` using the
    cache + server, honoring sticky-offline. ``confirmed`` is True only when the
    server has affirmatively answered (ever) for this key."""
    if not key:
        return {"valid": False, "tier": None, "error": "no license key", "confirmed": False}

    state = _load_state(storage_path)
    now = time.time()
    cached_same = state.get("key") == key
    fresh = cached_same and (now - float(state.get("checked_at") or 0)) < RECHECK_SECONDS

    # A fresh, server-confirmed key needs no network call.
    if fresh and state.get("valid") is True:
        return {"valid": True, "tier": state.get("tier"), "error": None, "confirmed": True}

    answer = _check_server(key)
    if answer is None:
        # Unreachable. Sticky: keep a prior confirmed-valid; otherwise unconfirmed.
        if cached_same and state.get("valid") is True:
            return {"valid": True, "tier": state.get("tier"), "error": None, "confirmed": True}
        return {"valid": False, "tier": None,
                "error": "could not reach the license server", "confirmed": False}

    # Definitive answer — cache it.
    new_state = dict(state)
    new_state.update({
        "key": key,
        "valid": answer["valid"],
        "tier": answer.get("tier") if answer["valid"] else None,
        "checked_at": now,
        "last_error": None if answer["valid"] else (answer.get("error") or "invalid license key"),
    })
    _save_state(new_state, storage_path)
    return {"valid": answer["valid"], "tier": new_state["tier"],
            "error": new_state["last_error"], "confirmed": True}


# --------------------------------------------------------------------------- #
# The gate
# --------------------------------------------------------------------------- #

def check_gate(*, storage_path: Optional[str] = None) -> dict:
    """Decide whether org-rollup may run. Returns a dict:

        {allowed, mode, reason, tier, grace_days_left, get_license, key_masked}

    ``mode`` ∈ {"licensed", "grace", "blocked"}. Starts the grace clock on the
    first unlicensed call (persisted), so the evaluation window is real-time, not
    install-time."""
    key = _license_key()
    res = _is_validated(key, storage_path)
    tier = (res.get("tier") or "").lower()

    if res["valid"] and tier in ORG_TIERS:
        return {
            "allowed": True,
            "mode": "licensed",
            "reason": f"{tier} license valid",
            "tier": res.get("tier"),
            "grace_days_left": None,
            "get_license": None,
            "key_masked": mask_key(key) if key else "",
        }

    # Not entitled. Two shapes: (a) no/invalid key, or (b) a VALID license whose
    # tier (e.g. Builder, single-seat) does not include the org-rollup feature.
    # Both still get the grace window — a paying Builder customer evaluating an
    # upgrade deserves the trial too — but the messaging differs.
    insufficient_tier = res["valid"] and tier not in ORG_TIERS
    if insufficient_tier:
        base_reason = (f"the {tier or 'current'} tier does not include org-rollup "
                       "(requires Studio or Platform)")
    else:
        base_reason = res.get("error") or "no license key set"

    state = _load_state(storage_path)
    now = time.time()
    first_seen = state.get("grace_started_at")
    if not first_seen:
        first_seen = now
        state["grace_started_at"] = first_seen
        _save_state(state, storage_path)
    elapsed = now - float(first_seen)
    grace_left = GRACE_SECONDS - elapsed
    days_left = max(0, math.ceil(grace_left / 86400)) if grace_left > 0 else 0

    # Surface the real tier when they hold a valid (but insufficient) license.
    shown_tier = res.get("tier") if insufficient_tier else None

    if grace_left > 0:
        lede = "tier upgrade needed" if insufficient_tier else "unlicensed evaluation"
        return {
            "allowed": True,
            "mode": "grace",
            "reason": f"{lede} ({base_reason}); {days_left} day(s) left in trial",
            "tier": shown_tier,
            "grace_days_left": days_left,
            "get_license": GET_LICENSE_URL,
            "key_masked": mask_key(key) if key else "",
        }

    requirement = ("a Studio or Platform license" if insufficient_tier else "a license")
    return {
        "allowed": False,
        "mode": "blocked",
        "reason": f"org-rollup requires {requirement} ({base_reason}); evaluation period ended",
        "tier": shown_tier,
        "grace_days_left": 0,
        "get_license": GET_LICENSE_URL,
        "key_masked": mask_key(key) if key else "",
    }
