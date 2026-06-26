"""Context providers for enriching code indexes with business metadata.

Context providers detect ecosystem tools (dbt, Terraform, OpenAPI, etc.)
and inject business context into symbols and file summaries during indexing.
"""

from .base import ContextProvider, FileContext, discover_providers, enrich_symbols, collect_metadata, collect_extra_imports

from . import dbt  # noqa: F401
from . import git_blame  # noqa: F401
from . import laravel  # noqa: F401
from . import nuxt  # noqa: F401
from . import nextjs  # noqa: F401
from . import decorator_routes  # noqa: F401  (Flask, FastAPI, Spring Boot, NestJS, ASP.NET)
from . import express  # noqa: F401  (Express, Fastify, Hono, Koa)
from . import go_routers  # noqa: F401  (Gin, Chi, Echo, Fiber)
from . import django  # noqa: F401
from . import rails  # noqa: F401

__all__ = [
    "ContextProvider",
    "FileContext",
    "collect_extra_imports",
    "collect_metadata",
    "discover_providers",
    "enrich_symbols",
]
