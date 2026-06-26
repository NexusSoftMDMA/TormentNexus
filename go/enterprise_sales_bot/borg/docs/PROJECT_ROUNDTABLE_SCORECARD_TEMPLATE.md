# tormentnexus Frontier-Model Scorecard Template

_Last updated: 2026-03-19_

Use this template to compare multiple model responses to the current tormentnexus roundtable prompt.

Recommended inputs per model:

- `docs/PROJECT_ROUNDTABLE_BRIEF.md`
- `docs/PROJECT_ROUNDTABLE_DEBATE_PROMPT.md` or `docs/PROJECT_ROUNDTABLE_EXECUTIVE_PROMPT.md`

---

## Roundtable metadata

- Date:
- Reviewer:
- Prompt used:
- Brief version/date:
- Models compared:
  -
  -
  -

---

## Quick ranking table

| Model | Overall usefulness (1-10) | Strategic clarity (1-10) | 1.0 realism (1-10) | Architecture quality (1-10) | Sequencing quality (1-10) | Anti-scope discipline (1-10) | Best insight | Biggest flaw |
|---|---:|---:|---:|---:|---:|---:|---|---|
| Model A |  |  |  |  |  |  |  |  |
| Model B |  |  |  |  |  |  |  |  |
| Model C |  |  |  |  |  |  |  |  |
| Model D |  |  |  |  |  |  |  |  |

---

## Per-model evaluation template

### Model: [name]

#### 1. Verdict snapshot
- One-sentence summary of the model’s judgment:
- Overall quality score (1-10):
- Would I use this as planning input? Yes / No / Partially

#### 2. What it got right
-
-
-

#### 3. What it got wrong
-
-
-

#### 4. Kernel judgment
- Score (1-10):
- Notes:

#### 5. 1.0 realism
- Score (1-10):
- Notes:

#### 6. Sequencing quality
- Score (1-10):
- Notes:

#### 7. Anti-scope discipline
- Score (1-10):
- Notes:

#### 8. Strongest recommendation
-

#### 9. Weakest recommendation
-

#### 10. Proposed next slices

| Slice | Good idea? | Priority | Notes |
|---|---|---|---|
| 1 | Yes / No / Mixed | High / Medium / Low |  |
| 2 | Yes / No / Mixed | High / Medium / Low |  |
| 3 | Yes / No / Mixed | High / Medium / Low |  |
| 4 | Yes / No / Mixed | High / Medium / Low |  |
| 5 | Yes / No / Mixed | High / Medium / Low |  |
| 6 | Yes / No / Mixed | High / Medium / Low |  |

#### 11. Keep / reject / postpone table

| Recommendation | Keep | Reject | Postpone | Reason |
|---|:---:|:---:|:---:|---|
|  |  |  |  |  |
|  |  |  |  |  |
|  |  |  |  |  |
|  |  |  |  |  |

#### 12. Net takeaway
- Best 3 ideas worth carrying forward:
  -
  -
  -
- 3 ideas to ignore:
  -
  -
  -

---

## Cross-model synthesis

### Points of strong agreement
-
-
-

### Points of disagreement worth debating
-
-
-

### Best combined 1.0 definition
-

### Best combined next 6 slices
1.
2.
3.
4.
5.
6.

---

## Decision rubric

| Criterion | Weight | Question |
|---|---:|---|
| Kernel clarity | 20% | Did the model correctly define what tormentnexus fundamentally is? |
| 1.0 realism | 20% | Did it produce a believable near-term ship target? |
| Architectural coherence | 15% | Did it preserve a control-plane architecture rather than sprawl? |
| Sequencing quality | 15% | Did it recommend the right order of work? |
| Anti-scope discipline | 10% | Did it avoid parity-for-parity’s-sake? |
| Risk awareness | 10% | Did it correctly identify failure modes? |
| Actionability | 10% | Could a maintainer actually use the recommendations tomorrow? |

### Weighted total formula

$$
\text{weighted total} = \sum (\text{score out of 10} \times \text{weight})
$$

---

## Final maintainer call

1. Which model produced the best recommendation overall?
2. Which model best understood tormentnexus’s actual identity?
3. Which model was most dangerous because it encouraged scope inflation?
4. What should be adopted immediately?
5. What should be explicitly rejected for now?
6. What should become the next task file?

### Final answer
- Best overall model:
- Best architecture model:
- Best product model:
- Most overreaching model:
- Immediate adoption:
- Explicit deferrals:
- Next task file candidate:
