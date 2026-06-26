# Language Contribution Guide

Add a new spoken language to Mnemosyne's MEMORIA extraction engine.

## Overview

The MEMORIA engine extracts structured facts from free-text conversation using regex patterns. Each language needs patterns for 7 extraction categories plus event keywords and date formats. Without a complete set, that language's facts get silently skipped during consolidation.

This template ensures every new language hits the same quality bar as DE, RU, and EN.

## File to Edit

`mnemosyne/core/beam.py` — two places:

1. **`detect_language()` method** (around line 2980) — add character/marker detection
2. **`MULTILINGUAL_PATTERNS` dict** (around line 3015) — add all patterns

## What You Need to Provide

### 1. Language Detection (in `detect_language()`)

```python
if any(c in text_lower for c in '...'):
    return 'xx'
LANG_markers = {'...', '...', ...}
words = set(re.findall(r'\w+', text_lower))
if len(words & LANG_markers) >= 2:
    return 'xx'
```

Requirements:
- Unique script characters (accents, diacritics, non-Latin chars)
- **At least 15 common markers** — everyday words like "I", "you", "not", "and", "for", "the", "this", "my", "have", "want", "like", "with", "from"
- Threshold of 2 matching markers to trigger

### 2. Pattern Categories

Every language must provide all 10 keys. Fill each one based on real usage, not translation of English. Patterns extracted from actual conversation logs are far more reliable than direct translations.

#### `negation`
Regex catching "I never/not ..." statements. Must capture 15-120 characters after the negation phrase.

```
EN: r'(I(?: have|\'ve)?\s*(?:never|not)\s+[^.,;!?\n]{15,120})'
DE: r'(Ich(?: habe|\'ve)?\s+(?:nie|niemals|nicht)\s+[^.,;!?\n]{15,120})'
```

**Provide:** negation phrases unique to your language (e.g. "non", "mai", "nessuno")

#### `decision`
Regex catching decision phrases like "decided to", "switched to", "chose".

```
EN: r'(?:decided to|chose to|opted for|selected|picked|switching to)\s+([^.,;!?\n]{10,120})'
```

**Provide:** 5+ decision trigger phrases

#### `entity`
Regex for "my X needs Y" patterns. Complex, needs both entity nouns and action verbs.

```
EN: r'(?:the|my|our|your)\s+([a-z_]+(?:\s+(?:table|model|schema|...)))\s+(?:needs?|requires?|should|...)\s+([^.,;!?\n]{10,80})'
```

**Provide:**
- Possessive pronouns (my, your, our, etc.)
- **At least 20 technical entity nouns** relevant to software/infra workflows (table, model, API, endpoint, function, module, route, handler, tool, plugin, script, config, setting, workflow, pipeline, process, system, server, client, service, database, query, file, repo, branch, PR, issue, task, job)
- **At least 15 action verbs** in your language (needs, requires, should, could, will, uses, runs, handles, etc.)

#### `sequence`
Regex for sequence markers like "first", "then", "finally".

```
EN: r'((?:first|second|third|fourth|fifth|finally|next|then|after that)[^.,;!?\n]{15,120})'
```

**Provide:** 6+ sequence words

#### `instruction_false_positives`
List of phrases that look like instructions but aren't. English example:
```python
['i think you should leave', 'should behave', 'their work style']
```

**Provide:** known false-positive phrases. At minimum test these patterns against real conversation to identify at least 3 false positives.

#### `instruction_imperative`
Verbs that indicate imperative mood (commands/instructions).

```
EN: 'always|never|remember|use|keep|avoid|ensure|check|verify|run|test|...'
```

**Provide:** 20+ imperative verbs. Must match your language's imperative conjugation.

#### `instruction`
Regex combining imperative markers with the imperative verbs.

```
EN: r'(?:always|never|must|must not|should(?: not)?(?=\s+(?:you|we|i|one)\s+(?:IMPVERBS))|need(?:s)? to(?: not)?|...)\s+([^.,;!?\n]{10,200})'
```

Replace `IMPVERBS` with an `(?:verb1|verb2|...)` group.

**Tip:** In languages where instructions use a different grammatical structure (e.g. subjunctive, infinitive), adapt the regex accordingly.

#### `preference`
Regex for "I like/prefer/hate/want X" statements.

```
EN: r'(?:I(?: |\')?(?:like|love|prefer|hate|dislike|enjoy|use|stick with|switched to|moved to|changed to|want|need|tend to|usually|would rather|don\'t like|don\'t want|not a fan of|am okay with|am comfortable with|am used to|am happy with|am tired of|am sick of|prefer not to|try to avoid|find it easier to|find it better to|find it useful to))\s+([^.,;!?\n]{10,200})'
```

**Provide:** 15+ preference/opinion verb phrases. Include positive, negative, and neutral expressions.

#### `event_keywords`
List of words that indicate events happened.

```
EN: ['meeting', 'call', 'scheduled', 'happened', 'occurred', 'plan to', 'will be on', 'due on', 'release', 'deadline', 'launched', 'deployed', 'released', 'published', 'posted', 'started', 'began', 'finished', 'completed', 'ended', 'event', 'conference', 'workshop', 'appointment']
```

**Provide:** 15+ event-related words and phrases.

#### `named_months`
Regex matching named month + day patterns in your language.

```
EN: r'((?:January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+\d{1,2}(?:st|nd|rd|th)?,?\s*(?:\d{4})?)'
```

**Provide:** Full month names + common abbreviations, with date format used in your language (day before month? month before day? dot separators?)

## 3. Testing Requirements

Before submitting:

1. **Create test content** — write 10-15 realistic sentences in your language covering each category (2 negations, 2 decisions, 2 entities, 2 sequences, 2 instructions, 2 preferences, 2 events, 2 dates)

2. **Run extraction manually:**
   ```python
   from mnemosyne.core.beam import BeamMemory
   bm = BeamMemory.__new__(BeamMemory)
   for sentence in test_sentences:
       result = bm.extract_and_store_facts(sentence, message_idx=0)
       print(sentence, '->', result)
   ```

3. **Check for false positives** — run the same test through the `instruction_false_positives` list. If any sentence matches that should not, add it to the list.

4. **Run the full test suite** to confirm your patterns don't break English extraction:
   ```bash
   python -m pytest tests/ -q --timeout=120
   ```
   Minimum: no regressions on English benchmarks.

5. **Verify detection works** — test `detect_language()` on your test sentences and on English sentences. Your language should trigger only when the content is actually in your language.

## PR Checklist

Before opening the PR, confirm each item:

- [ ] `detect_language()` — character detection markers added
- [ ] `detect_language()` — 15+ common word markers added
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `negation` pattern
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `decision` pattern with 5+ triggers
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `entity` pattern with 20+ nouns + 15+ verbs
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `sequence` pattern with 6+ markers
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `instruction_false_positives` with 3+ entries
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `instruction_imperative` with 20+ verbs
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `instruction` pattern (with IMPVERBS substitution or adapted grammar)
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `preference` pattern with 15+ phrases
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `event_keywords` with 15+ entries
- [ ] `MULTILINGUAL_PATTERNS['xx']` — `named_months` regex
- [ ] 10-15 test sentences extracted correctly (all categories covered)
- [ ] False positive list catches known non-instruction phrases
- [ ] Full test suite passes with no regressions
- [ ] `LANGUAGE_CODES` list updated in `extract_and_store_facts` (if applicable)
