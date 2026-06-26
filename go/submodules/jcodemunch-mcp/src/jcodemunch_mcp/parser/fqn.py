"""Bi-directional translation between jcodemunch symbol IDs and PHP FQNs.

Enables interoperability between jcodemunch (``app/Models/User.php::User#class``)
and PhpStorm / Laravel Idea (``App\\Models\\User``).  Uses PSR-4 mappings from
composer.json to perform the conversion.
"""

from typing import Optional

from .imports import resolve_php_namespace


def symbol_to_fqn(symbol_id: str, psr4_map: dict[str, str]) -> Optional[str]:
    """Convert a jcodemunch symbol ID to a PHP fully-qualified name.

    Examples::

        symbol_to_fqn("app/Models/User.php::User#class", {"App\\\\": "app/"})
        → "App\\\\Models\\\\User"

        symbol_to_fqn("app/Models/User.php::User.posts#method", {"App\\\\": "app/"})
        → "App\\\\Models\\\\User::posts"

    Only works for PHP symbols (file must end in ``.php``).
    """
    if "::" not in symbol_id:
        return None

    file_path, rest = symbol_id.split("::", 1)
    if "#" in rest:
        name, kind = rest.rsplit("#", 1)
    else:
        name, kind = rest, ""

    if not file_path.endswith(".php"):
        return None

    if kind == "method" and "." in name:
        class_name, method_name = name.rsplit(".", 1)
        class_fqn = _file_to_namespace(file_path, class_name, psr4_map)
        if class_fqn:
            return f"{class_fqn}::{method_name}"
        return None

    return _file_to_namespace(file_path, name, psr4_map)


def _file_to_namespace(
    file_path: str, class_name: str, psr4_map: dict[str, str]
) -> Optional[str]:
    """Reverse PSR-4: file path + class name → FQN."""
    path_no_ext = file_path[:-4]  # strip .php

    # Sort by longest base_dir first for most-specific match
    for prefix, base_dir in sorted(psr4_map.items(), key=lambda x: -len(x[1])):
        base = base_dir.rstrip("/")
        if path_no_ext.startswith(base + "/") or path_no_ext == base:
            relative = path_no_ext[len(base):].lstrip("/")
            namespace = prefix + relative.replace("/", "\\")
            # Verify the last segment matches class_name
            last = namespace.rsplit("\\", 1)[-1] if "\\" in namespace else namespace
            if last == class_name:
                return namespace
    return None


def fqn_to_symbol(
    fqn: str,
    psr4_map: dict[str, str],
    source_files: set[str],
) -> Optional[str]:
    """Convert a PHP FQN to a jcodemunch symbol ID.

    Examples::

        fqn_to_symbol("App\\\\Models\\\\User", psr4, files)
        → "app/Models/User.php::User#class"

        fqn_to_symbol("App\\\\Models\\\\User::posts", psr4, files)
        → "app/Models/User.php::User.posts#method"

    Returns None if the FQN cannot be resolved to an indexed file.
    """
    method_name: Optional[str] = None
    lookup_fqn = fqn
    if "::" in fqn:
        lookup_fqn, method_name = fqn.rsplit("::", 1)

    file_path = resolve_php_namespace(lookup_fqn, psr4_map, source_files)
    if not file_path:
        return None

    class_name = lookup_fqn.rsplit("\\", 1)[-1]

    if method_name:
        return f"{file_path}::{class_name}.{method_name}#method"
    return f"{file_path}::{class_name}#class"
