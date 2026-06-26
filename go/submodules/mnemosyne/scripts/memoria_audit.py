#!/usr/bin/env python3
"""
MEMORIA Health Audit Tool
Queries the 5 memoria tables and generates a health report:
  - Table counts
  - Usefulness sampling (random 50 entries per table classified into categories)
  - False positive analysis
  - Top noisy patterns identified

Run: python3 memoria_audit.py [--db-path ~/.hermes/mnemosyne/data/mnemosyne.db]
"""

import argparse
import sqlite3
import re
from collections import Counter
from datetime import datetime

# Classification rules for usefulness
def classify_fact(key, value, fact_type, importance):
    """Classify a fact as useful, noisy, or ambiguous."""
    val_lower = value.lower() if value else ""
    key_lower = key.lower() if key else ""

    # Clearly noisy patterns
    if any(kw in val_lower or kw in key_lower for kw in [
        'forecast', 'weather', 'temperature', 'rain', 'snow', 'wind', 'humidity',
        'regenrisiko', 'zum', 'heute', 'morgen', 'gestern',
    ]):
        return "noise_weather"

    if fact_type == 'sequence':
        # Sequence fragments are often noise
        short_words = [w for w in key_lower.split('_') if len(w) <= 3]
        if len(short_words) > len(key_lower.split('_')) / 2:
            return "noise_fragment"
        return "ambiguous_sequence"

    if fact_type == 'metric':
        # Check if value looks like a personal metric vs system metric
        if any(kw in key_lower for kw in ['latency', 'response_time', 'api_', 'pct',
                                            'version', 'count', 'total', 'rate',
                                            'kb', 'mb', 'gb', 'timeout']):
            return "noise_system_metric"
        if importance < 0.6:
            return "noise_low_importance"
        return "ambiguous_metric"

    if fact_type == 'date':
        return "noise_date_mention"

    if fact_type == 'version':
        return "useful_version"

    return "unknown"


def classify_instruction(instruction, topic):
    """Classify an instruction as useful, noisy, or ambiguous."""
    instr_lower = instruction.lower() if instruction else ""

    # Known false positive patterns
    if any(fp in instr_lower for fp in [
        'i think you should leave',
        'should behave',
        'their work style',
        'should i', 'should we', 'should it',
    ]):
        return "noise_false_positive"

    # Check if it's a real imperative/constraint
    if any(w in instr_lower for w in ['always remember', 'never', 'must', 'need to',
                                        'required to', 'prefer', 'make sure',
                                        'remember to', 'don\'t forget']):
        return "useful_imperative"

    # "should" that passes filters — check if it seems useful
    if 'should' in instr_lower:
        if any(w in instr_lower for w in ['check', 'verify', 'use', 'keep', 'avoid',
                                            'ensure', 'run', 'test', 'build',
                                            'deploy', 'push', 'merge', 'commit',
                                            'update', 'install', 'configure']):
            return "useful_should_technical"
        return "noise_should_conversational"

    return "noise_other"


def classify_preference(pref, topic):
    """Classify a preference as useful, noisy, or ambiguous."""
    pref_lower = pref.lower() if pref else ""

    # "I need to know" is not a preference
    if any(w in pref_lower for w in ['need to know', 'need your', 'need the',
                                       'want to know', 'want your', 'want the',
                                       'want to make sure', 'want to check']):
        return "noise_informational"

    # Real preference signals
    if any(w in pref_lower for w in ['prefer', 'like working', 'hate', 'dislike',
                                       'stick with', 'switched to', 'moved to',
                                       'am okay with', 'comfortable with',
                                       'used to', 'happy with', 'tired of']):
        return "useful_preference"

    # More ambiguous "I want/need" — check for concrete context
    if any(w in pref_lower for w in ['want to', 'need to', 'use', 'find it']):
        # Must have specific context to be useful
        if len(topic) > 20 and not any(w in pref_lower for w in ['what', 'how', 'why']):
            return "ambiguous_want"
        return "noise_vague_want"

    return "noise_other"


def classify_timeline(date, description, source):
    """Classify a timeline entry as useful, noisy, or ambiguous."""
    desc_lower = description.lower() if description else ""

    # Event context signals
    if any(kw in desc_lower for kw in [
        'meeting on', 'scheduled for', 'happened on', 'occurred', 'plan to',
        'will be on', 'due on', 'release', 'deadline', 'launched', 'deployed',
        'released', 'conference', 'workshop', 'appointment',
    ]):
        return "useful_event"

    # File paths and report references — noise
    if any(kw in desc_lower for kw in [
        '.md`', 'reports/', 'file', 'path', 'report-',
    ]):
        return "noise_filepath"

    return "noise_no_event_context"


def classify_kg(subject, predicate, obj):
    """Classify a KG triple as useful, noisy, or ambiguous."""
    # Negation triples — check if coherent
    if predicate == 'negation':
        if len(obj) > 20 and 'said' not in obj.lower() and 'wrote' not in obj.lower():
            return "useful_negation"
        return "noise_negation_fragment"

    # Decision triples
    if predicate == 'decision':
        if len(obj) > 15 and '|' not in obj and '```' not in obj:
            return "useful_decision"
        return "noise_decision_fragment"

    # Entity-action (requires)
    if predicate == 'requires':
        return "useful_entity_action"

    return "unknown"


def run_audit(db_path):
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()

    print("=" * 70)
    print(f"MEMORIA HEALTH AUDIT — {datetime.now().strftime('%Y-%m-%d %H:%M')}")
    print(f"Database: {db_path}")
    print("=" * 70)

    tables = {
        'memoria_facts': classify_fact,
        'memoria_instructions': classify_instruction,
        'memoria_preferences': classify_preference,
        'memoria_timelines': classify_timeline,
        'memoria_kg': classify_kg,
    }

    all_results = {}

    for table, classifier_fn in tables.items():
        print(f"\n{'─' * 70}")
        print(f"  {table}")
        print(f"{'─' * 70}")

        # Get count
        cursor.execute(f"SELECT COUNT(*) FROM {table}")
        count = cursor.fetchone()[0]
        print(f"  Total entries: {count}")

        if count == 0:
            all_results[table] = {"count": 0, "categories": {}}
            continue

        # Get column names
        cursor.execute(f"SELECT * FROM {table} LIMIT 1")
        col_names = [desc[0] for desc in cursor.description]

        # Sample random 100 entries (or all if less than 100)
        sample_size = min(100, count)
        cursor.execute(f"""
            SELECT * FROM {table}
            ORDER BY RANDOM()
            LIMIT {sample_size}
        """)
        sample = cursor.fetchall()

        # Classify each entry
        categories = Counter()
        examples = {cat: [] for cat in set()}
        examples = {}

        for row in sample:
            row_dict = dict(row)
            if table == 'memoria_facts':
                cat = classifier_fn(
                    row_dict.get('key', ''),
                    row_dict.get('value', ''),
                    row_dict.get('fact_type', ''),
                    row_dict.get('importance', 0.5)
                )
            elif table == 'memoria_instructions':
                cat = classifier_fn(
                    row_dict.get('instruction', ''),
                    row_dict.get('topic', '')
                )
            elif table == 'memoria_preferences':
                cat = classifier_fn(
                    row_dict.get('preference', ''),
                    row_dict.get('topic', '')
                )
            elif table == 'memoria_timelines':
                cat = classifier_fn(
                    row_dict.get('date', ''),
                    row_dict.get('description', ''),
                    row_dict.get('source', '')
                )
            elif table == 'memoria_kg':
                cat = classifier_fn(
                    row_dict.get('subject', ''),
                    row_dict.get('predicate', ''),
                    row_dict.get('object', '')
                )

            categories[cat] += 1
            if cat not in examples:
                examples[cat] = []
            # Store first 3 examples per category
            if len(examples[cat]) < 3:
                if table == 'memoria_facts':
                    examples[cat].append(f"  {row_dict.get('key', '?')} = {str(row_dict.get('value', ''))[:80]}")
                elif table == 'memoria_instructions':
                    examples[cat].append(f"  {str(row_dict.get('instruction', ''))[:100]}")
                elif table == 'memoria_preferences':
                    examples[cat].append(f"  {str(row_dict.get('preference', ''))[:100]}")
                elif table == 'memoria_timelines':
                    examples[cat].append(f"  {str(row_dict.get('description', ''))[:100]}")
                elif table == 'memoria_kg':
                    examples[cat].append(f"  {row_dict.get('subject', '?')} -- {row_dict.get('predicate', '')} -> {str(row_dict.get('object', ''))[:60]}")

        # Print breakdown
        useful_count = sum(v for k, v in categories.items() if k.startswith('useful'))
        noise_count = sum(v for k, v in categories.items() if k.startswith('noise'))
        ambiguous_count = sum(v for k, v in categories.items() if k.startswith('ambiguous'))
        total_classified = useful_count + noise_count + ambiguous_count

        print(f"  Sampled: {sample_size} entries")
        print(f"  Useful:  {useful_count}/{sample_size} ({useful_count*100//max(sample_size,1)}%)")
        print(f"  Noise:   {noise_count}/{sample_size} ({noise_count*100//max(sample_size,1)}%)")
        print(f"  Ambiguous: {ambiguous_count}/{sample_size} ({ambiguous_count*100//max(sample_size,1)}%)")
        print()

        for cat, count in sorted(categories.items(), key=lambda x: -x[1]):
            pct = count * 100 // max(sample_size, 1)
            bar = "█" * (pct // 5) + "░" * (20 - pct // 5)
            print(f"  {bar} {count:3d} ({pct:2d}%) {cat}")
            if cat in examples:
                for ex in examples[cat][:2]:
                    print(f"          {ex}")

        all_results[table] = {
            "count": count,
            "categories": dict(categories),
            "useful_pct": useful_count * 100 // max(total_classified, 1),
        }

    # Summary
    print(f"\n{'=' * 70}")
    print("  SUMMARY")
    print(f"{'=' * 70}")
    for table, results in all_results.items():
        useful_pct = results.get("useful_pct", 0)
        bar = "█" * (useful_pct // 5) + "░" * (20 - useful_pct // 5)
        print(f"  {bar}  {table:30s}  {results['count']:6d} entries  {useful_pct:2d}% useful")

    conn.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="MEMORIA Health Audit")
    parser.add_argument("--db-path", default=None,
                        help="Path to mnemosyne.db")
    args = parser.parse_args()

    if args.db_path:
        run_audit(args.db_path)
    else:
        import os
        default_paths = [
            os.path.expanduser("~/.hermes/mnemosyne/data/mnemosyne.db"),
            os.path.expanduser("~/.hermes/projects/mnemosyne/mnemosyne.db"),
            "/root/.hermes/mnemosyne/data/mnemosyne.db",
        ]
        for p in default_paths:
            if os.path.exists(p):
                run_audit(p)
                break
        else:
            print("No database found. Specify --db-path")
