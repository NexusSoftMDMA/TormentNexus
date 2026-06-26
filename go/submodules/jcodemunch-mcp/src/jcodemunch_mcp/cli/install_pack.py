"""install-pack subcommand -- download and install Starter Pack indexes."""
from __future__ import annotations

import json
import os
import sys
import tempfile
import zipfile
from pathlib import Path
from typing import Optional

import httpx

STARTER_PACK_API = "https://j.gravelle.us/jCodeMunch/starter-packs-system/api/index.php"

# ANSI helpers
_BOLD = "\033[1m"
_GREEN = "\033[32m"
_YELLOW = "\033[33m"
_RED = "\033[31m"
_DIM = "\033[2m"
_RESET = "\033[0m"

# Unicode with ASCII fallback for Windows cp1252
try:
    "\u2714\u2718\u00b7".encode(sys.stdout.encoding or "utf-8")
    _CHECK = "\u2714"
    _CROSS = "\u2718"
    _DOT = " \u00b7 "
except (UnicodeEncodeError, LookupError):
    _CHECK = "+"
    _CROSS = "x"
    _DOT = " - "


def _storage_path() -> Path:
    """Return the global index directory."""
    return Path(os.environ.get("CODE_INDEX_PATH", str(Path.home() / ".code-index")))


def _mask_license(key: str) -> str:
    """Mask a license key for display: first 4 + last 4 chars."""
    if len(key) <= 8:
        return key[:4] + "****"
    return key[:4] + "****" + key[-4:]


def _list_packs() -> int:
    """Fetch and display the starter pack catalog."""
    try:
        resp = httpx.get(f"{STARTER_PACK_API}?action=catalog", timeout=15)
        resp.raise_for_status()
    except httpx.HTTPError:
        print(
            f"  {_RED}{_CROSS} Could not reach the starter packs server. "
            f"Check your network connection.{_RESET}",
            file=sys.stderr,
        )
        return 1

    catalog = resp.json()
    packs = catalog.get("packs", [])
    if not packs:
        print("  No starter packs available yet.")
        return 0

    print()
    print(f"  {_BOLD}jCodeMunch Starter Packs{_RESET}")
    print(f"  {'=' * 50}")
    print()

    for pack in packs:
        pack_id = pack.get("id", "unknown")
        name = pack.get("name", pack_id)
        description = pack.get("description", "")
        symbols = pack.get("symbols", 0)
        size = pack.get("size", "")
        free = pack.get("free", False)
        download_url = pack.get("download_url", "")

        tag = f"{_GREEN}  FREE  {_RESET}" if free else f"{_YELLOW} LICENSE {_RESET}"
        print(f"  {tag}  {pack_id:<20s}  {_BOLD}{name}{_RESET}")
        details = []
        if symbols:
            details.append(f"{symbols:,} symbols")
        if size:
            details.append(size)
        if details:
            print(f"           {_DOT.join(details)}")
        if description:
            print(f"           {_DIM}{description}{_RESET}")
        print(f"           jcodemunch-mcp install-pack {pack_id}")
        if download_url:
            print(f"           {_DIM}{download_url}{_RESET}")
        print()

    print(f"  {_DIM}Free packs require no license. Licensed packs require a jCodeMunch license.{_RESET}")
    print(f"  {_DIM}Get a license: https://j.gravelle.us/jCodeMunch/#pricing{_RESET}")
    print()
    return 0


def _install_pack(
    pack_id: str,
    license_key: Optional[str] = None,
    force: bool = False,
    base_path: Optional[Path] = None,
) -> int:
    """Download and install a starter pack."""
    base = base_path if base_path is not None else _storage_path()
    marker = base / f".pack-{pack_id}.json"

    # Already installed?
    if marker.exists() and not force:
        print(f"  Pack '{pack_id}' is already installed. Use --force to re-download.")
        return 0

    # Build download URL
    url = f"{STARTER_PACK_API}?action=download&pack={pack_id}"
    if license_key:
        url += f"&license={license_key}"

    print(f"  Downloading starter pack '{pack_id}'...", flush=True)
    try:
        resp = httpx.get(url, timeout=120, follow_redirects=True)
        resp.raise_for_status()
    except httpx.HTTPError:
        print(
            f"  {_RED}{_CROSS} Could not reach the starter packs server. "
            f"Check your network connection.{_RESET}",
        )
        return 1

    content_type = resp.headers.get("content-type", "")

    # Error response (JSON)
    if "application/json" in content_type:
        data = resp.json()
        error = data.get("error", "Unknown error")
        print(f"  {_RED}{_CROSS} {error}{_RESET}")
        print()
        free_packs = data.get("free_packs")
        if free_packs:
            print(f"  Free packs available: {', '.join(free_packs)}")
        get_license = data.get("get_license")
        if get_license:
            print(f"  Get a license: {get_license}")
        hint = data.get("hint")
        if hint:
            print(f"  Hint: {hint}")
        if license_key:
            print(f"  License: {_mask_license(license_key)}")
        return 1

    # Expect a zip
    if "application/zip" not in content_type and "application/octet-stream" not in content_type:
        print(
            f"  {_RED}{_CROSS} Unexpected response (content-type: {content_type}){_RESET}",
            file=sys.stderr,
        )
        return 1

    pack_version = resp.headers.get("X-Pack-Version", "unknown")

    # Save to temp file, extract
    base.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(suffix=".zip", delete=False) as tmp:
        tmp.write(resp.content)
        tmp_path = Path(tmp.name)

    try:
        with zipfile.ZipFile(tmp_path, "r") as zf:
            # Validate -- no path traversal
            for info in zf.infolist():
                if info.filename.startswith("/") or ".." in info.filename:
                    print(
                        f"  {_RED}{_CROSS} Archive contains unsafe paths. Aborting.{_RESET}",
                        file=sys.stderr,
                    )
                    return 1

            manifest_data = None
            symbol_count = 0

            for info in zf.infolist():
                if info.is_dir():
                    continue

                # Strip top-level <pack-id>/ prefix
                parts = info.filename.split("/", 1)
                if len(parts) < 2 or not parts[1]:
                    continue
                relative = parts[1]

                # Capture manifest
                if relative == "manifest.json":
                    manifest_data = json.loads(zf.read(info.filename))
                    symbol_count = manifest_data.get("total_symbols", 0)
                    continue

                dest = base / relative
                dest.parent.mkdir(parents=True, exist_ok=True)
                with zf.open(info.filename) as src, open(dest, "wb") as dst:
                    dst.write(src.read())

            # Write install marker
            if manifest_data is None:
                manifest_data = {"pack": pack_id, "version": pack_version}
            manifest_data["installed_version"] = pack_version
            marker.write_text(json.dumps(manifest_data), encoding="utf-8")

    except zipfile.BadZipFile:
        print(
            f"  {_RED}{_CROSS} Downloaded file is corrupted. Try again with --force.{_RESET}",
        )
        return 1
    finally:
        tmp_path.unlink(missing_ok=True)

    # Success
    pack_name = (manifest_data or {}).get("name", pack_id)
    repos = (manifest_data or {}).get("repos", [])

    print()
    print(f"  {_GREEN}{_CHECK} Installed starter pack: {pack_name}{_RESET}")
    if symbol_count:
        print(f"  {_GREEN}{_CHECK} {symbol_count:,} symbols ready for retrieval{_RESET}")
    print(f"  {_GREEN}{_CHECK} Index location: {base}{_RESET}")
    print()

    if repos:
        print("  Repos included:")
        for repo in repos:
            repo_name = repo if isinstance(repo, str) else repo.get("repo", "")
            if repo_name:
                print(f"    - {repo_name}")
        print()

    # Opt-in telemetry
    if os.environ.get("JCODEMUNCH_SHARE_SAVINGS", "1") != "0":
        try:
            httpx.post(
                f"{STARTER_PACK_API}?action=telemetry",
                json={
                    "pack": pack_id,
                    "version": pack_version,
                    "platform": sys.platform,
                },
                timeout=5,
            )
        except Exception:
            pass

    return 0


def run_install_pack(
    pack_id: Optional[str] = None,
    license_key: Optional[str] = None,
    list_packs: bool = False,
    force: bool = False,
) -> int:
    """Entry point for the install-pack subcommand."""
    if list_packs or not pack_id:
        return _list_packs()
    return _install_pack(pack_id, license_key=license_key, force=force)
