"""Markdown file IO with YAML frontmatter splitting."""

from pathlib import Path

import yaml

_FM_DELIMITER = "---"


def write_concept(path: Path, frontmatter: dict, body: str) -> None:
    """Write a concept file: YAML frontmatter block + markdown body.

    Frontmatter is delimited by --- lines per OKF spec §4. Uses sort_keys=False
    so our deterministic key order from memory_to_frontmatter survives.
    """
    yaml_text = yaml.safe_dump(frontmatter, sort_keys=False, allow_unicode=True)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        f"{_FM_DELIMITER}\n{yaml_text}{_FM_DELIMITER}\n{body}\n",
        encoding="utf-8",
    )


def read_concept(path: Path) -> tuple[dict, str]:
    """Parse a concept file into (frontmatter dict, body string).

    Raises ValueError if frontmatter is missing or not parseable as YAML.
    """
    text = path.read_text(encoding="utf-8")
    if not text.startswith(f"{_FM_DELIMITER}\n") and not text.startswith(f"{_FM_DELIMITER}\r\n"):
        raise ValueError(f"{path}: no frontmatter block (file must start with '---')")

    # Split on the second --- delimiter (after the opening one)
    parts = text.split(f"{_FM_DELIMITER}\n", 2)
    if len(parts) < 3:
        raise ValueError(f"{path}: frontmatter block not closed (expected closing '---')")
    _, yaml_block, body = parts
    try:
        fm = yaml.safe_load(yaml_block) or {}
    except yaml.YAMLError as e:
        raise ValueError(f"{path}: could not parse frontmatter YAML: {e}") from e
    if not isinstance(fm, dict):
        raise ValueError(f"{path}: frontmatter must be a YAML mapping, got {type(fm).__name__}")
    # write_concept appends exactly one trailing '\n' after the body. Strip it so
    # that repeated export→import cycles don't accumulate extra newlines. If the
    # file was externally modified and has no trailing newline, this is a no-op.
    if body.endswith("\n"):
        body = body[:-1]
    return fm, body
