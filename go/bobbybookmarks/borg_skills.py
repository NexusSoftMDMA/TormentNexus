"""
Borg Skill Evolution Engine

Learns from extraction patterns and auto-promotes successful strategies.
Inspired by:
- everything-claude-code: /evolve command
- Agno: Behavioral Learning Protocol + Learning Stores
- VoltAgent: Audited SKILL.md runbooks

A "skill" is a reusable extraction template that knows:
  - What domain/pattern it works for
  - What prompt modifications yield best results
  - What quality score it typically achieves
  - When to evolve into a better version
"""

import os
import json
import time
import sqlite3
import logging
from datetime import datetime, timezone
from collections import defaultdict

logger = logging.getLogger(__name__)


# =====================================================================
# SKILL TEMPLATES — Pre-built extraction strategies
# =====================================================================
SKILL_TEMPLATES = {
    'github_repo': {
        'name': 'GitHub Repository Analysis',
        'triggers': ['github.com'],
        'prompt_modifiers': [
            'Focus on: stars, language, last commit date, README quality',
            'Extract: primary use case, key dependencies, API surface',
            'Prioritize: how this differs from alternatives in the same space',
        ],
        'expected_quality': 0.85,
        'min_innovation': 5,
    },
    'documentation_site': {
        'name': 'Documentation/API Site Analysis',
        'triggers': ['docs.', '/docs/', '/api/', 'readme', '/guide/'],
        'prompt_modifiers': [
            'Focus on: API endpoints, authentication model, rate limits',
            'Extract: key abstractions, data models, extension points',
            'Prioritize: what makes this API/service unique',
        ],
        'expected_quality': 0.75,
        'min_innovation': 4,
    },
    'research_paper': {
        'name': 'Research Paper Analysis',
        'triggers': ['arxiv.org', 'paperswithcode.com', '/paper/', 'huggingface.co/papers'],
        'prompt_modifiers': [
            'Focus on: novel technique, benchmark results, key insight',
            'Extract: methodology, evaluation metrics, limitations',
            'Prioritize: what breakthrough or improvement this represents',
        ],
        'expected_quality': 0.80,
        'min_innovation': 7,
    },
    'blog_article': {
        'name': 'Technical Blog Analysis',
        'triggers': ['medium.com', 'substack.com', '/blog/', 'dev.to', 'hashnode'],
        'prompt_modifiers': [
            'Focus on: key takeaways, practical code examples, architecture decisions',
            'Extract: techniques described, tools mentioned, lessons learned',
            'Prioritize: actionable insights a developer can apply today',
        ],
        'expected_quality': 0.70,
        'min_innovation': 4,
    },
    'reddit_discussion': {
        'name': 'Reddit/Forum Discussion Analysis',
        'triggers': ['reddit.com', 'news.ycombinator.com', 'discourse'],
        'prompt_modifiers': [
            'Focus on: consensus opinions, novel approaches mentioned, real-world experiences',
            'Extract: tools recommended, patterns discussed, warnings given',
            'Prioritize: community-validated insights over individual claims',
        ],
        'expected_quality': 0.55,
        'min_innovation': 3,
    },
    'product_landing': {
        'name': 'Product/Tool Landing Page',
        'triggers': [],  # Default fallback
        'prompt_modifiers': [
            'Focus on: core value proposition, pricing model, integrations',
            'Extract: key features, target audience, competitive advantages',
            'Prioritize: what specific problem this solves and how',
        ],
        'expected_quality': 0.60,
        'min_innovation': 3,
    },
}


class SkillEvolutionEngine:
    """
    Manages skill lifecycle: match -> apply -> evaluate -> evolve.
    
    Evolution criteria:
    - Skill with >70% success rate after 20+ uses = "mature"
    - Skill with <30% success rate after 10+ uses = "needs evolution"
    - Evolved skill beats original by >10% = "promoted"
    """

    def __init__(self, memory=None):
        self.memory = memory
        self._load_adaptations()

    def _load_adaptations(self):
        """Load any user-adapted skills from the skills directory."""
        self.adaptations = {}
        skills_dir = os.path.join('skills')
        if os.path.isdir(skills_dir):
            for fname in os.listdir(skills_dir):
                if fname.endswith('.json'):
                    try:
                        with open(os.path.join(skills_dir, fname)) as f:
                            skill = json.load(f)
                            self.adaptations[skill.get('name', fname)] = skill
                    except Exception:
                        pass

    def match_skill(self, url, html_content=None):
        """
        Match a URL (and optionally HTML) to the best skill template.
        Returns (skill_name, skill_config).
        """
        url_lower = url.lower()

        # Check adaptations first (user-defined overrides)
        for name, skill in self.adaptations.items():
            for trigger in skill.get('triggers', []):
                if trigger in url_lower:
                    return name, skill

        # Check built-in templates
        best_match = None
        best_match_count = 0

        for skill_name, config in SKILL_TEMPLATES.items():
            match_count = sum(1 for t in config['triggers'] if t in url_lower)
            if match_count > best_match_count:
                best_match = skill_name
                best_match_count = match_count

        if best_match:
            return best_match, SKILL_TEMPLATES[best_match]

        # Default fallback
        return 'product_landing', SKILL_TEMPLATES['product_landing']

    def build_skill_prompt(self, url, skill_name, base_content, skill_config):
        """
        Build a skill-enhanced extraction prompt.
        Adds domain-specific modifiers to the base prompt.
        """
        prompt = f"Analyze this resource using the '{skill_config.get('name', skill_name)}' extraction skill.\n\n"
        prompt += f"URL: {url}\n\n"

        # Add skill-specific modifiers
        modifiers = skill_config.get('prompt_modifiers', [])
        if modifiers:
            prompt += "EXTRACTION PRIORITIES:\n"
            for mod in modifiers:
                prompt += f"  - {mod}\n"
            prompt += "\n"

        prompt += f"Content:\n{base_content}\n\n"
        prompt += "Categorize into EXACTLY ONE: Agent Orchestration & Workflow, "
        prompt += "Context Engineering & Isolation, Memory & Persistence Architecture, "
        prompt += "Interface & Developer UX, Connectivity & Interoperability (MCP/A2A), "
        prompt += "Infrastructure & Proxy Layers, Guides & Industry Trends\n\n"
        prompt += "Return strict JSON:\n"
        prompt += "- CATEGORY: one category from above\n"
        prompt += "- SHORT_DESCRIPTION: 1 specific sentence about what this DOES\n"
        prompt += "- LONG_DESCRIPTION: detailed technical breakdown\n"
        prompt += "- MAIN_FEATURES: 3-5 SPECIFIC concrete features (comma separated)\n"
        prompt += f"- INNOVATION_SCORE: 1-10 (minimum {skill_config.get('min_innovation', 3)} for this type)\n"
        prompt += "- TAGS: 8-12 lowercase technical tags\n\n"
        prompt += "CRITICAL: MAIN_FEATURES must be SPECIFIC capabilities.\n"
        prompt += "NEVER use: 'automated discovery', 'heuristic mapping', or generic phrases.\n"

        return prompt

    def evaluate_skill(self, skill_name, quality_score, was_accepted):
        """
        Record skill outcome and check if evolution is needed.
        Returns (should_evolve, reason).
        """
        if self.memory:
            self.memory.record_skill_outcome(
                f"skill:{skill_name}",
                success=was_accepted,
                metadata={'quality': quality_score}
            )

        # Check evolution criteria
        if not self.memory:
            return False, None

        skills = self.memory.get_best_skills(limit=50)
        for skill in skills:
            if skill['skill_name'] == f"skill:{skill_name}":
                total = skill['success_count'] + skill['fail_count']
                win_rate = skill['success_count'] / max(total, 1)

                if total >= 10 and win_rate < 0.3:
                    return True, f"low_win_rate:{win_rate:.2f}_after_{total}_attempts"
                if total >= 20 and win_rate > 0.7:
                    return False, "mature_skill"
                break

        return False, None

    def evolve_skill_prompt(self, skill_name, current_config, failures_sample):
        """
        Generate an evolved prompt modification for a struggling skill.
        Returns new prompt_modifiers list.
        """
        evolution_prompt = f"""The extraction skill '{skill_name}' is underperforming.
        
Current prompt modifiers: {json.dumps(current_config.get('prompt_modifiers', []))}

Recent failure patterns: {json.dumps(failures_sample[:5])}

Generate 3 NEW prompt modifier sentences that would improve extraction quality.
These should be more specific and actionable than the current ones.

Return JSON: {{"new_modifiers": ["modifier1", "modifier2", "modifier3"], "reasoning": "..."}}"""

        # In a real implementation, this would call the LLM
        # For now, return enhanced defaults
        new_modifiers = list(current_config.get('prompt_modifiers', []))
        new_modifiers.append("EXTRA: If content is thin or gated, set INNOVATION_SCORE to 1-3")
        new_modifiers.append("EXTRA: Focus on what makes this DIFFERENT from similar tools")
        return new_modifiers

    def get_skill_report(self):
        """Generate a skill performance report."""
        report = {
            'templates': list(SKILL_TEMPLATES.keys()),
            'adaptations': list(self.adaptations.keys()),
        }

        if self.memory:
            skills = self.memory.get_best_skills(limit=20)
            report['learned_skills'] = []
            for s in skills:
                if s['skill_name'].startswith('skill:'):
                    name = s['skill_name'].replace('skill:', '')
                    total = s['success_count'] + s['fail_count']
                    report['learned_skills'].append({
                        'name': name,
                        'win_rate': s['win_rate'],
                        'total_uses': total,
                        'status': 'mature' if s['win_rate'] > 0.7 and total > 20 else
                                  'evolving' if s['win_rate'] > 0.3 else
                                  'struggling',
                    })

        return report
