import re
from urllib.parse import urlparse, urlunparse, parse_qs, urlencode, ParseResult

# Params that are purely tracking and should always be stripped
TRACKING_PARAMS = {
    "utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
    "utm_id", "utm_reader", "utm_name", "utm_cid",
    "fbclid", "gclid", "gclsrc", "dclid", "msclkid",
    "adrefer", "ref", "source",
    "mc_cid", "mc_eid",
    "zanpid",
    "openid",
    "_ga", "_gid",
    "igshid",
    "yclid",
    "twclid",
    "li_fat_id",
    "epik",
    "rdid",
    "ttclid",
    "wbraid", "gbraid",
    "srsltid",
}

# Params that carry meaningful content and must never be stripped
MEANINGFUL_PARAMS = {
    "q", "query", "search", "term", "s", "p",
    "id", "page", "category", "tag", "keywords",
    "v",       # YouTube video id
    "t",       # timestamp / thread id
    "n",       # page number on some sites
    "sort", "order", "type", "filter",
    "lang", "locale",
}

DEFAULT_PORTS = {
    "http": 80,
    "https": 443,
    "ftp": 21,
}


def normalize_url(url: str) -> str:
    """
    Normalize a URL for deduplication purposes:
    - Lowercase scheme and host
    - Remove default port
    - Remove tracking params
    - Keep meaningful params
    - Sort remaining query params
    - Strip trailing slash from path (keep bare '/')
    - Remove fragment
    """
    url = url.strip()
    if not url:
        return url

    # Ensure scheme
    if not re.match(r"^[a-zA-Z][a-zA-Z0-9+\-.]*://", url):
        url = "https://" + url

    try:
        parsed = urlparse(url)
    except Exception:
        return url.lower()

    scheme = parsed.scheme.lower()
    host = parsed.hostname or ""
    host = host.lower()

    # Remove default ports
    try:
        port = parsed.port
    except ValueError:
        # If port is invalid (e.g. not an integer), we just ignore it
        port = None
    
    if port and DEFAULT_PORTS.get(scheme) == port:
        port = None

    netloc = host
    if port:
        netloc = f"{host}:{port}"

    # Handle path: strip trailing slash but keep bare '/', and lowercase
    path = parsed.path.lower()
    if path != "/" and path.endswith("/"):
        path = path.rstrip("/")
    if not path:
        path = "/"

    # Process query params
    query_string = parsed.query
    if query_string:
        params = parse_qs(query_string, keep_blank_values=True)

        # Separate tracking, meaningful and unknown params
        tracking = {}
        meaningful = {}
        unknown = {}

        for k, v in params.items():
            k_lower = k.lower()
            # Check for wildcard tracking patterns (e.g. utm_*)
            if k_lower in TRACKING_PARAMS or k_lower.startswith("utm_"):
                tracking[k_lower] = v
            elif k_lower in MEANINGFUL_PARAMS:
                meaningful[k_lower] = v
            else:
                unknown[k_lower] = v

        # If ALL params are tracking, strip everything
        if not meaningful and not unknown:
            final_params = {}
        else:
            # Keep meaningful + unknown, strip tracking
            final_params = {**meaningful, **unknown}

        # Sort and rebuild
        if final_params:
            sorted_items = sorted(final_params.items())
            new_query = urlencode(
                {k: v[0] if len(v) == 1 else v for k, v in sorted_items},
                doseq=True,
            )
        else:
            new_query = ""
    else:
        new_query = ""

    # Reconstruct without fragment
    normalized = urlunparse(ParseResult(
        scheme=scheme,
        netloc=netloc,
        path=path,
        params="",
        query=new_query,
        fragment="",
    ))
    return normalized


def get_project_url(url: str) -> str:
    """
    Identify the core project URL for a given link.
    Example: github.com/owner/repo/issues -> github.com/owner/repo
    """
    url = url.strip()
    if not url:
        return url

    try:
        parsed = urlparse(url)
    except Exception:
        return url.lower()

    host = (parsed.hostname or "").lower()
    path = parsed.path.lower()

    if "github.com" in host:
        path_parts = [part for part in path.split('/') if part]
        if len(path_parts) >= 2:
            # Keep only owner/repo
            project_path = f"/{path_parts[0]}/{path_parts[1]}"
            return urlunparse(ParseResult(
                scheme=parsed.scheme.lower(),
                netloc=host,
                path=project_path,
                params="",
                query="",
                fragment="",
            ))
    
    # For others, we could be more aggressive, but let's stick to standard normalization
    # unless it's a known documentation site where we might want the root.
    if "docs." in host or "documentation" in path:
        # Try to find common doc roots
        return urlunparse(ParseResult(
            scheme=parsed.scheme.lower(),
            netloc=host,
            path=path.split('/')[0] if path.startswith('/') else "/",
            params="",
            query="",
            fragment="",
        ))

    return normalize_url(url)


def deduplicate_bookmarks(bookmarks_list: list) -> tuple[list, list]:
    """
    Given a list of bookmark dicts (each with at least 'url'),
    return (unique_list, duplicate_list).
    First occurrence of each normalized_url wins.
    Each dict gets a 'normalized_url' key added.
    """
    seen: dict[str, dict] = {}
    unique: list = []
    duplicates: list = []

    for bm in bookmarks_list:
        url = bm.get("url", "")
        norm = normalize_url(url)
        bm["normalized_url"] = norm
        if norm in seen:
            bm["is_duplicate"] = True
            bm["duplicate_of_url"] = seen[norm].get("url", "")
            duplicates.append(bm)
        else:
            bm["is_duplicate"] = False
            seen[norm] = bm
            unique.append(bm)

    return unique, duplicates
