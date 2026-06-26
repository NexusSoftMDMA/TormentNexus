import json
import time

from fastmcp import FastMCP
from starlette.requests import Request
from starlette.responses import JSONResponse

mcp = FastMCP("Ogham")

_health_cache: dict = {"result": None, "expires": 0.0}
HEALTH_CACHE_TTL = 30  # seconds


@mcp.custom_route("/health", methods=["GET"])
async def health(request: Request) -> JSONResponse:
    """Health check endpoint -- served on the SSE port in SSE mode.
    Cached for 30 seconds to avoid hitting the database on every call."""
    now = time.monotonic()
    if _health_cache["result"] and now < _health_cache["expires"]:
        return JSONResponse(_health_cache["result"])

    from ogham.health import full_health_check

    result = json.loads(json.dumps(full_health_check(), default=str))
    _health_cache["result"] = result
    _health_cache["expires"] = now + HEALTH_CACHE_TTL
    return JSONResponse(result)
