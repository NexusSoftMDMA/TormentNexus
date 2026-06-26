"""Base classes for context providers.

A ContextProvider detects an ecosystem tool in a project folder, loads its
metadata, and enriches symbols with business context. The abstraction is
intentionally minimal — providers only need to implement detect/load/lookup.
"""

import logging
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

logger = logging.getLogger(__name__)


@dataclass
class FileContext:
    """Business context metadata for a single file from any ecosystem tool.

    This is the common currency between context providers and the indexing
    pipeline. Providers translate their tool-specific metadata into this
    structure, and the pipeline consumes it generically.
    """

    description: str = ""
    tags: list[str] = field(default_factory=list)
    properties: dict[str, str] = field(default_factory=dict)

    def summary_context(self, max_properties: int = 10) -> str:
        """Build a concise context string for AI summarization prompts."""
        parts = []
        if self.description:
            parts.append(self.description)
        if self.tags:
            parts.append(f"Tags: {', '.join(self.tags)}")
        if self.properties:
            items = list(self.properties.items())[:max_properties]
            prop_strs = []
            for key, val in items:
                if val:
                    prop_strs.append(f"{key} ({val})")
                else:
                    prop_strs.append(key)
            if len(self.properties) > max_properties:
                prop_strs.append(f"... and {len(self.properties) - max_properties} more")
            parts.append(f"Properties: {', '.join(prop_strs)}")
        return ". ".join(parts)

    def file_summary(self) -> str:
        """Build an enriched file-level summary."""
        parts = []
        if self.description:
            desc = self.description.strip()
            if len(desc) > 200:
                desc = desc[:197] + "..."
            parts.append(desc)
        if self.tags:
            parts.append(f"Tags: {', '.join(self.tags)}")
        if self.properties:
            parts.append(f"{len(self.properties)} properties")
        return ". ".join(parts)

    def search_keywords(self) -> list[str]:
        """Extract keywords for search indexing."""
        keywords = []
        if self.tags:
            keywords.extend(self.tags)
        if self.properties:
            keywords.extend(self.properties.keys())
        return keywords


class ContextProvider(ABC):
    """Base class for ecosystem context providers.

    Subclasses detect a specific tool (dbt, Terraform, etc.), load its
    metadata, and provide per-file context lookups. Providers are
    auto-discovered and run during indexing — no configuration needed.
    """

    @property
    @abstractmethod
    def name(self) -> str:
        """Short identifier for this provider (e.g., 'dbt', 'terraform')."""
        ...

    @abstractmethod
    def detect(self, folder_path: Path) -> bool:
        """Return True if this provider's ecosystem is present in the folder."""
        ...

    @abstractmethod
    def load(self, folder_path: Path) -> None:
        """Load metadata from the project. Called only if detect() returned True."""
        ...

    @abstractmethod
    def get_file_context(self, file_path: str) -> Optional[FileContext]:
        """Look up context for a file by its relative path (stem match is fine)."""
        ...

    @abstractmethod
    def stats(self) -> dict:
        """Return provider-specific stats for the index response."""
        ...

    def get_metadata(self) -> dict:
        """Return structured metadata to persist in the index.

        Override in subclasses to expose searchable metadata.
        Keys should be namespaced by provider (e.g., 'dbt_columns').

        Column metadata convention:
            To expose columns for ``search_columns``, emit a key ending in
            ``_columns`` whose value is ``{model_name: {col_name: col_desc}}``.
            Example::

                {"dbt_columns": {"fact_orders": {"order_id": "Primary key", ...}}}

            Any provider following this convention will be automatically
            discovered by ``search_columns`` — no tool-side changes needed.
        """
        return {}

    def get_extra_imports(self) -> dict[str, list[dict]]:
        """Return additional import edges discovered by this provider.

        Override in subclasses to inject framework-specific import edges
        into the dependency graph (e.g., Blade template references,
        route-to-controller mappings).

        Returns:
            Dict mapping relative file paths to lists of import dicts::

                {"routes/web.php": [{"specifier": "App\\\\Http\\\\Controllers\\\\UserController", "names": ["UserController"]}]}
        """
        return {}


# -- Registry of known providers --

_PROVIDER_CLASSES: list[type[ContextProvider]] = []


def register_provider(cls: type[ContextProvider]) -> type[ContextProvider]:
    """Class decorator to register a context provider."""
    _PROVIDER_CLASSES.append(cls)
    return cls


def discover_providers(folder_path: Path) -> list[ContextProvider]:
    """Instantiate and detect all registered context providers.

    Returns only providers whose ecosystem was detected in the folder.
    """
    active = []
    for cls in _PROVIDER_CLASSES:
        try:
            provider = cls()
            if provider.detect(folder_path):
                provider.load(folder_path)
                logger.info("Context provider '%s' activated", provider.name)
                active.append(provider)
        except Exception as e:
            logger.warning("Context provider '%s' failed: %s", cls.__name__, e)
    return active


def collect_metadata(providers: list[ContextProvider]) -> dict:
    """Collect structured metadata from all active providers for index persistence."""
    metadata: dict = {}
    for provider in providers:
        try:
            provider_meta = provider.get_metadata()
            if provider_meta:
                for key, value in provider_meta.items():
                    if key in metadata:
                        logger.warning(
                            "Metadata key '%s' from provider '%s' overwrites "
                            "existing key (previous value dropped)",
                            key,
                            provider.name,
                        )
                    metadata[key] = value
        except Exception as e:
            logger.warning("Metadata collection from '%s' failed: %s", provider.name, e)
    return metadata


def enrich_symbols(symbols: list, providers: list[ContextProvider]) -> None:
    """Attach context from all active providers to symbols in-place."""
    if not providers:
        return

    # Pre-compute file contexts once per unique file per provider
    # to avoid O(symbols × providers) redundant lookups.
    unique_files = {sym.file for sym in symbols}
    file_ctx_cache: dict[tuple[str, str], tuple[str, list[str]]] = {}
    for provider in providers:
        pname = provider.name
        for fpath in unique_files:
            ctx = provider.get_file_context(fpath)
            if ctx is not None:
                summary = ctx.summary_context(max_properties=8)
                kw = ctx.search_keywords()
                if summary or kw:
                    file_ctx_cache[(pname, fpath)] = (summary, kw)

    for sym in symbols:
        context_parts = []
        for provider in providers:
            cached = file_ctx_cache.get((provider.name, sym.file))
            if cached is None:
                continue
            summary, kw = cached
            if summary:
                context_parts.append(f'{provider.name}: {summary}')
            if kw:
                existing = set(sym.keywords)
                sym.keywords.extend(k for k in kw if k not in existing)

        if context_parts:
            sym.ecosystem_context = "; ".join(context_parts)


def collect_extra_imports(
    providers: list[ContextProvider],
    file_imports: dict[str, list[dict]],
) -> None:
    """Merge extra import edges from context providers into file_imports in-place.

    Providers can inject framework-specific import edges (e.g., Blade template
    references, route→controller mappings) that the standard language-level
    import extractor cannot detect.
    """
    for provider in providers:
        try:
            extras = provider.get_extra_imports()
            for file_path, imports in extras.items():
                if file_path in file_imports:
                    # Deduplicate: only add specifiers not already present
                    existing = {imp["specifier"] for imp in file_imports[file_path]}
                    for imp in imports:
                        if imp["specifier"] not in existing:
                            file_imports[file_path].append(imp)
                            existing.add(imp["specifier"])
                else:
                    file_imports[file_path] = list(imports)
        except Exception as e:
            logger.warning(
                "Extra imports from '%s' failed: %s", provider.name, e
            )
