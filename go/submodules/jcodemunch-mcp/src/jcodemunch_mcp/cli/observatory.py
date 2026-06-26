"""``jcodemunch-mcp observatory`` — public OSS code-health observatory pipeline.

Reads a JSON config listing OSS repos, clones each, indexes, runs
``get_repo_health``, and writes static HTML + JSON + RSS artifacts to
the configured output directory. Designed for weekly cron / GitHub
Actions / launchd; static output is hosting-agnostic.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Optional


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Run the public OSS health observatory pipeline.",
    )
    sub = parser.add_subparsers(dest="action")

    build = sub.add_parser("build", help="Run the full pipeline against a config file.")
    build.add_argument("--config", type=Path, required=True,
        help="Path to the observatory config JSON.")
    build.add_argument("--output-dir", type=Path, default=None,
        help="Override config's output_dir.")
    build.add_argument("--workdir", type=Path, default=None,
        help="Override config's workdir (where checkouts + index live).")

    init_p = sub.add_parser("init", help="Write a starter config file.")
    init_p.add_argument("--out", type=Path, default=Path("observatory.config.json"),
        help="Where to write the starter config.")

    args = parser.parse_args(argv)

    if args.action == "init":
        return _run_init(args.out)
    if args.action == "build":
        return _run_build(args.config, args.output_dir, args.workdir)

    parser.print_help()
    return 1


def _run_init(out: Path) -> int:
    if out.exists():
        print(f"error: {out} already exists. Refusing to overwrite.", file=sys.stderr)
        return 2
    starter = {
        "output_dir": "./observatory_out",
        "workdir": "./.observatory_work",
        "history_cap": 52,
        "repos": [
            {"url": "https://github.com/expressjs/express", "label": "Express",  "blurb": "Classic Node.js web framework."},
            {"url": "https://github.com/tiangolo/fastapi",   "label": "FastAPI",  "blurb": "Modern Python async web framework."},
            {"url": "https://github.com/gin-gonic/gin",      "label": "Gin",      "blurb": "High-perf Go HTTP framework."},
            {"url": "https://github.com/pydantic/pydantic",  "label": "Pydantic", "blurb": "Python data validation via type hints."},
            {"url": "https://github.com/jgravelle/jcodemunch-mcp", "label": "jcodemunch-mcp", "blurb": "(self-audit)"},
        ],
    }
    out.write_text(json.dumps(starter, indent=2) + "\n", encoding="utf-8")
    print(f"wrote starter config -> {out}")
    print("Next: jcodemunch-mcp observatory build --config", out)
    return 0


def _run_build(config_path: Path, output_dir: Optional[Path], workdir: Optional[Path]) -> int:
    if not config_path.exists():
        print(f"error: config not found: {config_path}", file=sys.stderr)
        return 2

    from ..tools.observatory import load_config, run_pipeline
    cfg = load_config(config_path)
    if output_dir:
        cfg.output_dir = Path(output_dir).expanduser().resolve()
    if workdir:
        cfg.workdir = Path(workdir).expanduser().resolve()

    print(f"observatory build: {len(cfg.repos)} repo(s) -> {cfg.output_dir}")
    summary = run_pipeline(cfg)
    print(json.dumps(summary, indent=2))
    return 0 if summary["failed"] == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
