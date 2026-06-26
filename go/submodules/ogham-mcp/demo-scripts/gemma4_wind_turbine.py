#!/usr/bin/env python3
"""Gemma 4 wind turbine multimodal test (Ollama).

Reads three artefacts about a fictional wind turbine product line:
  - power_curves.png  (a chart of power output vs wind speed for 3 models)
  - specs.txt          (technical spec sheet for TX-100, TX-200, TX-300)
  - maintenance_costs.csv (annual maintenance costs and MTBF per model)

Then asks Gemma 4 questions that require cross-referencing the chart,
the spec sheet, and the cost table at the same time.

Usage:
    .venv/bin/python extras/gemma4_wind_turbine.py
    .venv/bin/python extras/gemma4_wind_turbine.py --model gemma4:e2b
    .venv/bin/python extras/gemma4_wind_turbine.py --model gemma4:e4b

Test data lives in a sibling repo because it's shared with the
multimodal-embedding-test project. Override with --data-dir if you've
moved it.

Notes from earlier runs:
  - gemma4:e2b -> ~2B effective params, 5.1B total Q4_K_M, 7.2GB
  - gemma4:e4b -> ~4.5B effective params, 8.4B total Q4_K_M, 9.6GB
  - Must use /api/chat with messages format. /api/generate returns empty.
  - Pass image as a file path in the user message's `images` field.
  - System role for the framing, user role for the data + question.
"""

import argparse
import time
from pathlib import Path

DEFAULT_DATA_DIR = Path(
    "/Users/kevinburns/Developer/web-projects/multimodal-embedding-test/test_data"
)
DEFAULT_MODEL = "gemma4:e2b"


SYSTEM_PROMPT = (
    "You are a wind turbine engineering assistant. You will be given a "
    "technical spec sheet, a maintenance cost table, and a power curve chart. "
    "Answer the user's questions by combining information from all three sources. "
    "Be specific. Show your reasoning briefly."
)

QUESTIONS = [
    "Which model has the best capacity factor at moderate wind speeds (10-15 m/s)? Why?",
    (
        "If I'm operating at a site with average wind speed 14 m/s, which model gives me "
        "the lowest cost per MWh over 25 years? Show the calculation."
    ),
    (
        "Looking at the power curve chart, at what wind speed does the TX-300 start "
        "producing more power than the TX-200?"
    ),
    (
        "Which model has the lowest unplanned downtime risk based on the spec sheet, and "
        "what does the cost table say about its maintenance burden?"
    ),
]


def run(model: str, data_dir: Path) -> None:
    from ollama import chat

    specs = (data_dir / "specs.txt").read_text()
    costs = (data_dir / "maintenance_costs.csv").read_text()
    chart = str(data_dir / "power_curves.png")

    base_user = (
        f"Here is the technical spec sheet:\n{specs}\n\n"
        f"Here is the maintenance cost table (CSV):\n{costs}\n\n"
        f"Attached is a power curve chart for the same models."
    )

    print(f"=== {model} wind turbine test ===\n")
    total_start = time.time()

    for i, question in enumerate(QUESTIONS, 1):
        print(f"--- Q{i}: {question}")
        t0 = time.time()
        resp = chat(
            model=model,
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": base_user, "images": [chart]},
                {"role": "user", "content": question},
            ],
            options={"temperature": 0.2, "num_ctx": 32768},
        )
        elapsed = time.time() - t0
        answer = resp["message"]["content"].strip()
        print(f"({elapsed:.1f}s)")
        print(answer)
        print()

    print(f"=== Total: {time.time() - total_start:.1f}s ===")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--model", default=DEFAULT_MODEL, help="Ollama model tag")
    parser.add_argument(
        "--data-dir",
        type=Path,
        default=DEFAULT_DATA_DIR,
        help="Directory containing power_curves.png, specs.txt, maintenance_costs.csv",
    )
    args = parser.parse_args()

    if not args.data_dir.is_dir():
        raise SystemExit(f"Data directory not found: {args.data_dir}")
    for filename in ("power_curves.png", "specs.txt", "maintenance_costs.csv"):
        if not (args.data_dir / filename).exists():
            raise SystemExit(f"Missing test file: {args.data_dir / filename}")

    run(args.model, args.data_dir)


if __name__ == "__main__":
    main()
