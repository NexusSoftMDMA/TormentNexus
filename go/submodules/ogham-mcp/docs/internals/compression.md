# Ogham MCP — Memory Compression

`compression.py` implements tiered degradation of memory content over time.

## Compression Levels

| Level | Name | When | Result |
|-------|------|------|--------|
| 0 | Full | Recent, active, important | Full text preserved |
| 1 | Gist | 7+ days old, low activity | Key sentences (~30% of original) |
| 2 | Tags | 30+ days old, low activity | One-line summary + tags (<200 chars) |

Original content is always preserved in `original_content` column for restoration.

## Target Selection (`get_compression_target`)

Based on age in hours with a **resistance multiplier**:

| Factor | Multiplier |
|--------|-----------|
| `importance > 0.7` | 2.0x |
| `confidence > 0.8` | 1.3x |
| `access_count > 10` | 1.5x |

Thresholds: `gist = 168h * resistance`, `tags = 720h * resistance`

So an important, confident, frequently-accessed memory would need to be `168 * 2.0 * 1.3 * 1.5 = 655 hours ≈ 27 days` old before gist compression, vs 7 days for a default memory.

## Gist Compression (`compress_to_gist`)

Extracts the most information-dense ~30% of content:

1. Extract and preserve code blocks verbatim
2. Split remaining text into sentences
3. Score each sentence by information density:
   - File paths: +3.0
   - Error/Exception mentions: +4.0
   - Decision words: +3.0
   - Version numbers: +2.0
   - Inline code: +2.0
   - First sentence (primacy): +10.0
   - Last sentence (recency): +8.0
4. Select highest-scoring sentences until ~30% of original length is reached
5. Always include first and last sentences
6. Reconstruct in original order + append code blocks

## Tags Compression (`compress_to_tags`)

Reduces to `<200 chars`: first sentence + top-5 tags.

Format: `First sentence of content. | Tags: tag1, tag2, tag3`
