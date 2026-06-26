"""Corpus loader — reads benchmark YAML files into structured questions."""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

import yaml


@dataclass
class Question:
    """A single benchmark question with ground-truth."""

    id: str
    repo: str
    question: str
    ground_truth_answer: str
    ground_truth_symbols: list[str] = field(default_factory=list)
    difficulty: str = "medium"  # easy, medium, hard
    category: str = "general"  # general, architecture, debugging, api, refactoring
    tags: list[str] = field(default_factory=list)


@dataclass
class Corpus:
    """Full benchmark corpus."""

    questions: list[Question]
    repos: list[str]

    @classmethod
    def load(cls, corpus_dir: str | Path) -> "Corpus":
        """Load all YAML files from corpus directory."""
        corpus_dir = Path(corpus_dir)
        questions: list[Question] = []
        repos: set[str] = set()

        for yaml_file in sorted(corpus_dir.glob("*.yaml")):
            with open(yaml_file, encoding="utf-8") as f:
                data = yaml.safe_load(f)

            repo = data.get("repo", "")
            repos.add(repo)

            for q in data.get("questions", []):
                questions.append(Question(
                    id=q["id"],
                    repo=repo,
                    question=q["question"],
                    ground_truth_answer=q["ground_truth_answer"],
                    ground_truth_symbols=q.get("ground_truth_symbols", []),
                    difficulty=q.get("difficulty", "medium"),
                    category=q.get("category", "general"),
                    tags=q.get("tags", []),
                ))

        return cls(questions=questions, repos=sorted(repos))

    def filter(
        self,
        repo: Optional[str] = None,
        difficulty: Optional[str] = None,
        category: Optional[str] = None,
    ) -> "Corpus":
        """Return a filtered subset of the corpus."""
        filtered = self.questions
        if repo:
            filtered = [q for q in filtered if q.repo == repo]
        if difficulty:
            filtered = [q for q in filtered if q.difficulty == difficulty]
        if category:
            filtered = [q for q in filtered if q.category == category]
        repos = sorted({q.repo for q in filtered})
        return Corpus(questions=filtered, repos=repos)
