"""Replayable retrieval benchmark — fixtures, metrics, harness.

Used to detect ranking-quality regressions between releases. A fixture is
a JSON file mapping queries to known-relevant symbol IDs; the harness runs
each query through search_symbols and computes nDCG@k, MRR@k, Recall@k.
"""
