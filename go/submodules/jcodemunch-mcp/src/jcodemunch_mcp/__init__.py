"""github-codemunch-mcp - Token-efficient MCP server for GitHub source code exploration."""

from importlib.metadata import PackageNotFoundError, version

try:
    __version__ = version("jcodemunch-mcp")
except PackageNotFoundError:
    __version__ = "unknown"
