"""
Borg Self-Healing Verification Engine

Planner-Checker-Revise cycle inspired by:
- Verdent AI: 3-model cross-validation
- GSD Framework: Planner-Checker-Revise loops
- Auggie Wiggum: 14-hour autonomous self-healing
- OpenAI Harness Engineering: verification feedback loops

Modes:
  EXTRACT:  URL -> scrape -> LLM extract -> validate -> (retry if garbage)
  VERIFY:   extraction -> second opinion LLM -> score confidence -> accept/reject
  CORRECT:  failed extraction + error -> correction LLM -> patched result -> validate
  CROSS_VALIDATE: 3-model voting on high-value extractions
"""

import os
import json
import time
import logging
import re
from datetime import datetime, timezone

logger = logging.getLogger(__name__)


# =====================================================================
# EXTRACTION VALIDATION RULES
# =====================================================================
TAXONOMY = [
    "Agent Orchestration & Workflow",
    "Context Engineering & Isolation",
    "Memory & Persistence Architecture",
    "Interface & Developer UX",
    "Connectivity & Interoperability (MCP/A2A)",
    "Infrastructure & Proxy Layers",
    "Guides & Industry Trends",
]

# Known bad patterns (from garbage filter + extended)
REJECTION_PATTERNS = [
    "automated discovery, heuristic mapping",
    "automated discovery (heuristic)",
    "a comprehensive resource detailing",
    "a powerful ai-powered",
    "sign in to continue",
    "unsupported browser",
    "enable javascript",
    "cookie policy",
    "we use cookies",
    "please verify you are a human",
    "access denied",
    "403 forbidden",
    "page not found",
    "sorry, we couldn't find",
]

# Quality scoring dimensions
QUALITY_DIMENSIONS = {
    'specificity': {
        'weight': 0.30,
        'description': 'Are features concrete and specific vs generic?',
        'good_signals': ['implements', 'supports', 'provides', 'enables', 'uses', 'built-in', 'native'],
        'bad_signals': ['various', 'multiple', 'several', 'including but not limited', 'etc'],
    },
    'depth': {
        'weight': 0.25,
        'description': 'Does the description reveal technical depth?',
        'good_signals': ['algorithm', 'architecture', 'protocol', 'framework', 'pipeline', 'heuristic',
                         'embed', 'vector', 'token', 'context window', 'benchmark'],
        'bad_signals': ['interesting', 'useful', 'helpful', 'nice', 'cool'],
    },
    'actionability': {
        'weight': 0.25,
        'description': 'Can a developer act on this information?',
        'good_signals': ['api', 'sdk', 'cli', 'config', 'install', 'deploy', 'integration',
                         'endpoint', 'webhook', 'command'],
        'bad_signals': ['learn more', 'visit', 'check out', 'explore'],
    },
    'completeness': {
        'weight': 0.20,
        'description': 'Are all fields populated meaningfully?',
    },
}


class ExtractionValidator:
    """Validates and scores LLM extractions for quality."""

    def __init__(self, llm_pool=None, memory=None):
        self.llm_pool = llm_pool
        self.memory = memory

    def validate(self, rdata, url, html_content=None):
        """
        Full validation pipeline. Returns (is_valid, quality_score, issues).
        quality_score: 0.0 - 1.0
        """
        issues = []
        scores = {}

        # 1. Structural validation
        required_fields = ['CATEGORY', 'SHORT_DESCRIPTION', 'MAIN_FEATURES', 'TAGS']
        for field in required_fields:
            val = str(rdata.get(field, '')).strip()
            if not val or val.lower() in ('none', 'null', 'n/a', 'unknown'):
                issues.append(f"missing_{field.lower()}")
                scores[field] = 0.0
            else:
                scores[field] = 1.0

        # 2. Category validation
        cat = str(rdata.get('CATEGORY', ''))
        if cat not in TAXONOMY:
            # Check partial match (LLMs sometimes add extra text)
            matched = False
            for t in TAXONOMY:
                if t.lower() in cat.lower():
                    rdata['CATEGORY'] = t  # Fix it
                    matched = True
                    break
            if not matched:
                issues.append(f"invalid_category:{cat[:50]}")
                scores['CATEGORY'] = 0.3

        # 3. Rejection pattern check
        all_text = ' '.join(str(rdata.get(k, '')) for k in rdata).lower()
        for pattern in REJECTION_PATTERNS:
            if pattern in all_text:
                issues.append(f"rejection_pattern:{pattern[:30]}")
                return False, 0.0, issues

        # 4. Quality dimension scoring
        desc = str(rdata.get('SHORT_DESCRIPTION', '')).lower()
        features = str(rdata.get('MAIN_FEATURES', '')).lower()
        long_desc = str(rdata.get('LONG_DESCRIPTION', '')).lower()
        combined = f"{desc} {features} {long_desc}"

        quality = 0.0
        for dim_name, dim in QUALITY_DIMENSIONS.items():
            if dim_name == 'completeness':
                # Score based on how many fields are filled
                filled = sum(1 for f in ['CATEGORY', 'SHORT_DESCRIPTION', 'LONG_DESCRIPTION',
                                          'MAIN_FEATURES', 'TAGS', 'INNOVATION_SCORE']
                             if str(rdata.get(f, '')).strip())
                dim_score = filled / 6.0
            else:
                good = sum(1 for s in dim['good_signals'] if s in combined)
                bad = sum(1 for s in dim['bad_signals'] if s in combined)
                dim_score = min(1.0, max(0.0, (good - bad * 2) / max(good + bad, 1) + 0.5))

            quality += dim_score * dim['weight']

        # 5. Feature specificity bonus
        feature_list = [f.strip() for f in features.split(',') if len(f.strip()) > 5]
        if len(feature_list) >= 3:
            quality += 0.1
        avg_feature_len = sum(len(f) for f in feature_list) / max(len(feature_list), 1)
        if avg_feature_len > 30:  # Detailed features
            quality += 0.1

        # 6. Innovation score sanity check
        try:
            innov = int(rdata.get('INNOVATION_SCORE', 0))
            if innov < 1 or innov > 10:
                issues.append(f"innovation_out_of_range:{innov}")
                rdata['INNOVATION_SCORE'] = max(1, min(10, innov))
        except (ValueError, TypeError):
            issues.append("invalid_innovation_score")
            rdata['INNOVATION_SCORE'] = 3

        quality = max(0.0, min(1.0, quality))

        is_valid = quality >= 0.3 and len(issues) == 0
        return is_valid, quality, issues

    def cross_validate(self, rdata, url, html_content):
        """
        Get a second opinion from a different model.
        Returns (agreed, consensus_data, confidence).
        """
        if not self.llm_pool:
            return True, rdata, 0.5  # No pool = auto-accept

        from deep_research import extract_fit_markdown, classify_url_complexity, build_tiered_prompt, BORG_TAXONOMY

        # Build a verification prompt
        fit_text = extract_fit_markdown(html_content, url) if html_content else "N/A"
        verify_prompt = f"""You are a VERIFICATION agent. Compare the proposed extraction against the source content.

URL: {url}
Source Content (excerpt): {fit_text[:2000]}

PROPOSED EXTRACTION:
{json.dumps(rdata, indent=2)}

Verify each field:
1. Is CATEGORY correct? (Must be one of: {', '.join(TAXONOMY)})
2. Is SHORT_DESCRIPTION accurate and specific?
3. Are MAIN_FEATURES concrete and present in the source?
4. Is INNOVATION_SCORE justified?

Return JSON:
{{"verified": true/false, "corrections": {{"field": "corrected_value"}}, "confidence": 0.0-1.0, "reason": "..."}}"""

        try:
            response, model = self.llm_pool.generate_content(verify_prompt, f"verifying {url}")
            if response:
                response = response.strip()
                if "```json" in response:
                    response = response.split("```json")[1].split("```")[0].strip()
                elif "```" in response:
                    response = response.split("```")[1].split("```")[0].strip()
                json_match = re.search(r'\{[^{}]*\}', response, re.DOTALL)
                if json_match:
                    response = json_match.group(0)
                vdata = json.loads(response)

                verified = vdata.get('verified', False)
                confidence = float(vdata.get('confidence', 0.5))
                corrections = vdata.get('corrections', {})

                if verified and not corrections:
                    return True, rdata, confidence

                if corrections:
                    consensus = dict(rdata)
                    for field, value in corrections.items():
                        if field in consensus:
                            consensus[field] = value
                    return verified, consensus, confidence

                return verified, rdata, confidence

        except Exception as e:
            logger.warning("Cross-validation failed for %s: %s", url, e)

        return True, rdata, 0.5


class SelfHealingEngine:
    """
    Orchestrates the Planner-Checker-Revise loop.
    
    Flow:
      1. EXTRACT: Initial LLM extraction
      2. VALIDATE: Check extraction quality
      3. If low quality -> CORRECT: Send error context to LLM for fix
      4. If still low -> CROSS_VALIDATE: Second model opinion
      5. If still low -> REJECT with detailed reason
    """

    def __init__(self, llm_pool=None, memory=None):
        self.llm_pool = llm_pool
        self.memory = memory
        self.validator = ExtractionValidator(llm_pool, memory)

        # Stats
        self.stats = {
            'extracted': 0,
            'validated_first_pass': 0,
            'corrected': 0,
            'cross_validated': 0,
            'rejected_after_all': 0,
            'avg_quality': 0.0,
            'quality_history': [],
        }

    def process_extraction(self, url, rdata, html_content=None, raw_response=None):
        """
        Full self-healing pipeline for an extraction result.
        Returns (final_rdata, quality_score, decision_path).
        """
        decision_path = []

        # Phase 1: Validate
        is_valid, quality, issues = self.validator.validate(rdata, url, html_content)
        decision_path.append(f"validate:quality={quality:.2f},issues={len(issues)}")

        if is_valid and quality >= 0.5:
            self.stats['validated_first_pass'] += 1
            self._record_quality(quality)
            if self.memory:
                self.memory.record_skill_outcome('extraction_first_pass', True,
                                                  metadata={'quality': quality, 'url': url})
            return rdata, quality, decision_path

        # Phase 2: Attempt correction if quality is salvageable (0.15 - 0.49)
        if 0.15 <= quality < 0.5 and self.llm_pool and raw_response:
            decision_path.append(f"correct:attempting_fix")
            corrected = self._attempt_correction(url, rdata, issues, html_content, raw_response)
            if corrected:
                is_valid2, quality2, issues2 = self.validator.validate(corrected, url, html_content)
                decision_path.append(f"correct:quality={quality2:.2f}")
                if is_valid2 and quality2 > quality:
                    self.stats['corrected'] += 1
                    self._record_quality(quality2)
                    if self.memory:
                        self.memory.record_skill_outcome('self_healing_correction', True,
                                                          metadata={'before': quality, 'after': quality2})
                    return corrected, quality2, decision_path

        # Phase 3: Cross-validate for high-value targets (innovation >= 8)
        innovation = 0
        try:
            innovation = int(rdata.get('INNOVATION_SCORE', 0))
        except (ValueError, TypeError):
            pass

        if innovation >= 8 and html_content:
            decision_path.append("cross_validate:high_value")
            agreed, consensus, confidence = self.validator.cross_validate(
                rdata, url, html_content)
            if agreed and confidence > 0.6:
                self.stats['cross_validated'] += 1
                _, quality3, _ = self.validator.validate(consensus, url, html_content)
                self._record_quality(quality3)
                if self.memory:
                    self.memory.record_skill_outcome('cross_validation', True)
                return consensus, quality3, decision_path

        # Phase 4: Final rejection
        self.stats['rejected_after_all'] += 1
        self._record_quality(quality)
        if self.memory:
            self.memory.record_skill_outcome('extraction_failed', False,
                                              metadata={'quality': quality, 'issues': issues})
        decision_path.append(f"rejected:quality={quality:.2f}")
        return None, quality, decision_path

    def _attempt_correction(self, url, rdata, issues, html_content, raw_response):
        """Ask the LLM to fix a bad extraction."""
        from deep_research import extract_fit_markdown

        correction_prompt = f"""The previous extraction had quality issues. Fix them.

URL: {url}
Issues found: {json.dumps(issues)}
Original extraction: {json.dumps(rdata, indent=2)}

Source content excerpt: {(extract_fit_markdown(html_content, url) if html_content else '')[:2000]}

Return a CORRECTED JSON extraction with all fields properly filled:
- CATEGORY: one of {', '.join(TAXONOMY)}
- SHORT_DESCRIPTION: specific, factual, one sentence
- LONG_DESCRIPTION: detailed technical breakdown
- MAIN_FEATURES: 3-5 SPECIFIC concrete features (NOT generic)
- INNOVATION_SCORE: 1-10
- TAGS: 8-12 lowercase technical tags"""

        try:
            response, model = self.llm_pool.generate_content(correction_prompt, f"correcting {url}")
            if response:
                response = response.strip()
                if "```json" in response:
                    response = response.split("```json")[1].split("```")[0].strip()
                elif "```" in response:
                    response = response.split("```")[1].split("```")[0].strip()
                json_match = re.search(r'\{[^{}]*\}', response, re.DOTALL)
                if json_match:
                    response = json_match.group(0)
                return json.loads(response)
        except Exception as e:
            logger.warning("Correction attempt failed for %s: %s", url, e)
        return None

    def _record_quality(self, quality):
        self.stats['quality_history'].append(quality)
        if self.stats['quality_history']:
            self.stats['avg_quality'] = sum(self.stats['quality_history']) / len(self.stats['quality_history'])

    def get_stats(self):
        return dict(self.stats)
