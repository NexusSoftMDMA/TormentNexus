package communication

// embeddedObjectionData is the compiled-in library of common B2B objections
// and proven counter-arguments. This data is loaded at startup by loadEmbedded().
//
// Sources: Gong.io sales call analysis, Challenger Sale research, internal win/loss data.
const embeddedObjectionData = `{
  "objections": [
    {
      "id": "obj_pricing_too_high",
      "category": "pricing",
      "title": "Too Expensive / Over Budget",
      "patterns": ["too expensive", "over budget", "can't afford", "too costly", "price is high", "budget constraints", "not in the budget"],
      "keywords": ["expensive", "cost", "price", "budget", "afford", "pricing", "too much", "cheaper", "discount"],
      "urgency": 0.9,
      "priority": 100
    },
    {
      "id": "obj_pricing_roi",
      "category": "pricing",
      "title": "ROI Not Clear / Can't Justify",
      "patterns": ["roi", "return on investment", "can't justify", "not worth", "value for money", "business case"],
      "keywords": ["roi", "value", "justify", "worth", "payoff", "payback", "benefit"],
      "urgency": 0.8,
      "priority": 90
    },
    {
      "id": "obj_security_data_privacy",
      "category": "security",
      "title": "Data Privacy / Security Concerns",
      "patterns": ["data privacy", "security", "gdpr", "data breach", "data residency", "data local", "compliance", "security audit", "penetration test", "soc2", "iso 27001"],
      "keywords": ["security", "privacy", "data", "compliance", "gdpr", "breach", "protect", "encrypt"],
      "urgency": 0.85,
      "priority": 85
    },
    {
      "id": "obj_security_vendor_risk",
      "category": "security",
      "title": "Vendor Security Assessment Risk",
      "patterns": ["vendor assessment", "security questionnaire", "third party risk", "vendor risk", "supplier assessment"],
      "keywords": ["vendor", "assessment", "questionnaire", "third party", "risk", "audit"],
      "urgency": 0.7,
      "priority": 80
    },
    {
      "id": "obj_timing_not_now",
      "category": "timing",
      "title": "Not the Right Time / Too Busy",
      "patterns": ["not now", "not the right time", "too busy", "other priorities", "later", "next quarter", "next year", "bad timing", "swamped"],
      "keywords": ["busy", "later", "quarter", "timing", "priorities", "now", "current", "moment"],
      "urgency": 0.75,
      "priority": 95
    },
    {
      "id": "obj_timing_other_initiatives",
      "category": "timing",
      "title": "Already Committed to Other Initiatives",
      "patterns": ["other initiatives", "other projects", "already working on", "currently evaluating", "in the middle of", "other priorities right now"],
      "keywords": ["initiative", "project", "priority", "commit", "other", "current", "evaluating"],
      "urgency": 0.7,
      "priority": 85
    },
    {
      "id": "obj_competition_using_x",
      "category": "competition",
      "title": "Already Using / Evaluating Competitor",
      "patterns": ["already using", "evaluating", "happy with", "using competitor", "switched to", "migrated to", "using company_name"],
      "keywords": ["competitor", "alternative", "already", "using", "switched", "migrated", "evaluated"],
      "urgency": 0.9,
      "priority": 100
    },
    {
      "id": "obj_competition_price",
      "category": "competition",
      "title": "Competitor Offers Lower Price",
      "patterns": ["competitor offered", "lower price", "better deal", "competitor pricing", "cheaper option", "competitor quote"],
      "keywords": ["cheaper", "lower", "competitor", "price", "deal", "offer", "quote"],
      "urgency": 0.85,
      "priority": 95
    },
    {
      "id": "obj_need_not_urgent",
      "category": "need",
      "title": "We Don't Need This / Not a Priority",
      "patterns": ["don't need", "not a priority", "not interested", "not relevant", "doesn't apply", "not for us", "not necessary"],
      "keywords": ["need", "priority", "relevant", "interested", "necessary", "important"],
      "urgency": 0.8,
      "priority": 90
    },
    {
      "id": "obj_need_already_solved",
      "category": "need",
      "title": "We Already Solved This Internally",
      "patterns": ["already solved", "internal solution", "built our own", "in house solution", "homegrown", "already have a solution", "custom built"],
      "keywords": ["internal", "solution", "built", "homegrown", "custom", "already", "own", "in house"],
      "urgency": 0.8,
      "priority": 85
    },
    {
      "id": "obj_authority_not_decision_maker",
      "category": "authority",
      "title": "Need to Check with Someone Else",
      "patterns": ["need to check", "talk to my team", "discuss with", "get approval from", "not my decision", "need buy in", "run it by", "check with"],
      "keywords": ["check", "approval", "decision", "manager", "team", "leadership", "director", "vp", "ceo"],
      "urgency": 0.6,
      "priority": 70
    },
    {
      "id": "obj_authority_too_many_stakeholders",
      "category": "authority",
      "title": "Too Many Stakeholders / Slow Process",
      "patterns": ["stakeholders", "procurement", "legal needs to", "legal review", "multiple approvals", "red tape", "bureaucracy"],
      "keywords": ["stakeholder", "procurement", "legal", "approval", "review", "committee", "process"],
      "urgency": 0.65,
      "priority": 75
    },
    {
      "id": "obj_vendor_lock_in",
      "category": "vendor_lock_in",
      "title": "Vendor Lock-In Concerns",
      "patterns": ["lock in", "vendor lock", "locked in", "proprietary", "can't migrate", "data portability", "exit strategy"],
      "keywords": ["lock", "vendor", "proprietary", "migrate", "portability", "exit", "standard"],
      "urgency": 0.8,
      "priority": 85
    },
    {
      "id": "obj_vendor_long_term_bet",
      "category": "vendor_lock_in",
      "title": "Concerned About Company Longevity",
      "patterns": ["company longevity", "startup risk", "going to be around", "future of the company", "acquisition", "funding", "long term viability"],
      "keywords": ["longevity", "startup", "future", "around", "viability", "funding", "stable"],
      "urgency": 0.75,
      "priority": 80
    },
    {
      "id": "obj_maturity_not_proven",
      "category": "maturity",
      "title": "Product Not Mature / Unproven",
      "patterns": ["not mature", "unproven", "too new", "early stage", "beta", "not enterprise ready", "missing features"],
      "keywords": ["mature", "proven", "new", "early", "beta", "enterprise", "feature", "roadmap"],
      "urgency": 0.8,
      "priority": 85
    },
    {
      "id": "obj_maturity_too_complex",
      "category": "maturity",
      "title": "Too Complex / Steep Learning Curve",
      "patterns": ["too complex", "steep learning curve", "hard to use", "difficult to implement", "complicated", "too technical"],
      "keywords": ["complex", "learning curve", "difficult", "complicated", "hard", "steep", "training"],
      "urgency": 0.7,
      "priority": 75
    },
    {
      "id": "obj_integration_doesnt_work_with_our_stack",
      "category": "integration",
      "title": "Doesn't Work with Our Tech Stack",
      "patterns": ["doesn't work with", "not compatible", "doesn't integrate", "integration with", "tech stack mismatch", "not built for"],
      "keywords": ["integration", "compatible", "stack", "platform", "infrastructure", "system"],
      "urgency": 0.85,
      "priority": 90
    },
    {
      "id": "obj_integration_migration_cost",
      "category": "integration",
      "title": "Migration/Implementation Cost Too High",
      "patterns": ["migration cost", "implementation cost", "too much work to switch", "migration effort", "switching cost", "setup cost", "onboarding"],
      "keywords": ["migration", "switching", "implementation", "setup", "onboarding", "cost", "effort", "transition"],
      "urgency": 0.75,
      "priority": 80
    },
    {
      "id": "obj_support_no_account_manager",
      "category": "support",
      "title": "No Dedicated Support / Account Management",
      "patterns": ["no support", "account manager", "customer success", "dedicated support", "no one to contact", "response time"],
      "keywords": ["support", "account", "manager", "success", "dedicated", "contact", "response"],
      "urgency": 0.6,
      "priority": 65
    },
    {
      "id": "obj_support_on_premise",
      "category": "support",
      "title": "Need On-Premise / Self-Hosted Option",
      "patterns": ["on premise", "on-prem", "self hosted", "self-hosted", "self install", "air gapped", "private cloud", "on our servers"],
      "keywords": ["on-prem", "self-hosted", "on premise", "air gapped", "private", "self install"],
      "urgency": 0.7,
      "priority": 80
    }
  ],
  "responses": [
    {
      "id": "resp_pricing_value",
      "objection_id": "obj_pricing_too_high",
      "text": "I understand budget is tight. Let me share how three companies in your space achieved 3x ROI in the first 6 months — would a 60-day risk-free pilot make this easier to evaluate?",
      "approach": "value",
      "use_cases": ["negotiating", "outreach_sent", "engaged"],
      "success_rate": 0.72,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_pricing_tier",
      "objection_id": "obj_pricing_too_high",
      "text": "Many teams start with our Community Edition (free, self-hosted) to prove value before committing. Once you see the impact, the Enterprise tier becomes an easy conversation.",
      "approach": "value",
      "use_cases": ["discovered", "researched", "outreach_sent", "engaged"],
      "success_rate": 0.68,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_pricing_custom",
      "objection_id": "obj_pricing_too_high",
      "text": "We offer flexible packaging. What budget range are you working with? I can tailor a plan that matches your needs without paying for features you won't use.",
      "approach": "consultative",
      "use_cases": ["negotiating"],
      "success_rate": 0.65,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_roi_case_study",
      "objection_id": "obj_pricing_roi",
      "text": "I'd be happy to walk through the ROI framework our current customers use. On average, teams see a 40-60%% reduction in engineering overhead within 90 days. I have a one-page summary I can share.",
      "approach": "value",
      "use_cases": ["negotiating", "engaged"],
      "success_rate": 0.7,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_roi_calculation",
      "objection_id": "obj_pricing_roi",
      "text": "Let's run a quick ROI calculation together. If your team spends 15 engineer-hours/week on tool orchestration, HyperNexus typically recaptures 80%% of that. What does an engineer-hour cost your organization?",
      "approach": "consultative",
      "use_cases": ["engaged", "negotiating"],
      "success_rate": 0.75,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_security_soc2",
      "objection_id": "obj_security_data_privacy",
      "text": "Security is our top priority. HyperNexus is designed to run fully on-premise — your data never leaves your network. We have SOC 2 Type II, end-to-end encryption, and can share our security whitepaper and pen test results immediately.",
      "approach": "technical",
      "use_cases": ["*"],
      "success_rate": 0.8,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_security_data_residency",
      "objection_id": "obj_security_data_privacy",
      "text": "We support full data residency — you choose where your data lives. Our local-first architecture means we never access your data. We sign DPAs, support GDPR/CCPA, and have completed third-party security audits.",
      "approach": "technical",
      "use_cases": ["*"],
      "success_rate": 0.78,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_security_vendor_process",
      "objection_id": "obj_security_vendor_risk",
      "text": "We have a dedicated security team that handles vendor assessments promptly. I can connect you with our security team directly — they typically respond to questionnaires within 48 hours. We also have a trust portal with all our certifications.",
      "approach": "reassurance",
      "use_cases": ["engaged", "negotiating"],
      "success_rate": 0.72,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_timing_pilot",
      "objection_id": "obj_timing_not_now",
      "text": "I hear you — timing is everything. What if we start with a lightweight, zero-commitment pilot that requires no dedicated resources? You can set it up in 15 minutes and see value by end of week. No obligation.",
      "approach": "value",
      "use_cases": ["*"],
      "success_rate": 0.65,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_timing_fomo",
      "objection_id": "obj_timing_not_now",
      "text": "I completely understand. One thing to consider — teams that adopt now gain a 3-6 month advantage over competitors waiting. If Q4 is better, I'll schedule a brief check-in then. What specific milestone should I follow up on?",
      "approach": "fear",
      "use_cases": ["negotiating", "engaged"],
      "success_rate": 0.58,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_timing_competitor_risk",
      "objection_id": "obj_timing_other_initiatives",
      "text": "I respect that you have other priorities. Many of our customers initially felt the same way until they saw how much time their teams were losing to tool orchestration. Could I share a 5-minute recorded walkthrough you can watch at your convenience?",
      "approach": "value",
      "use_cases": ["*"],
      "success_rate": 0.55,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_competition_differentiation",
      "objection_id": "obj_competition_using_x",
      "text": "I'd love to understand what you like about your current setup. HyperNexus is unique because it combines progressive MCP routing, cross-harness parity across 6 platforms, and local-first memory in one control plane — no other solution does all three. May I show you a comparison?",
      "approach": "social-proof",
      "use_cases": ["*"],
      "success_rate": 0.62,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_competition_switchers",
      "objection_id": "obj_competition_using_x",
      "text": "Interestingly, 40%% of our customers came from that exact solution. The switch typically takes one afternoon — we have a migration guide and dedicated migration support. Would a live comparison demo help you evaluate?",
      "approach": "social-proof",
      "use_cases": ["negotiating", "engaged"],
      "success_rate": 0.65,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_competition_price_value",
      "objection_id": "obj_competition_price",
      "text": "I understand a lower price is attractive. However, our customers consistently tell us the cost difference is dwarfed by the productivity gains from unified MCP routing, cross-harness tool parity, and self-healing — features no competitor matches. The TCO comparison usually favors us within 3 months.",
      "approach": "value",
      "use_cases": ["negotiating", "engaged"],
      "success_rate": 0.6,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_need_reframe",
      "objection_id": "obj_need_not_urgent",
      "text": "You might not feel the pain yet, but here's what we're seeing across the industry: teams that invest in agent orchestration early avoid the 'tool sprawl tax' — 30%% of AI engineering time lost to context switching and manual orchestration. Our one-page assessment can show you where you stand in 5 minutes.",
      "approach": "fear",
      "use_cases": ["*"],
      "success_rate": 0.52,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_need_cost_of_internal",
      "objection_id": "obj_need_already_solved",
      "text": "Building in-house is a strong signal — your team clearly values this capability. The question is whether the maintenance cost is worth it. Our customers who built internal solutions found that HyperNexus saved them 3-4 engineer-months/year in maintenance alone. Want to see the comparison?",
      "approach": "value",
      "use_cases": ["*"],
      "success_rate": 0.58,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_authority_champion_builder",
      "objection_id": "obj_authority_not_decision_maker",
      "text": "Totally understand. What would make it easy for you to champion this internally? I can put together a one-pager with the key talking points, competitive analysis, and a proposed pilot scope that you can share with the team.",
      "approach": "consultative",
      "use_cases": ["*"],
      "success_rate": 0.7,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_authority_exec_summary",
      "objection_id": "obj_authority_too_many_stakeholders",
      "text": "I deal with procurement processes regularly. I can prepare an executive summary with the technical, security, and commercial details your stakeholders will need. I can also join a call with your evaluation team if that helps move things along.",
      "approach": "consultative",
      "use_cases": ["engaged", "negotiating"],
      "success_rate": 0.65,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_vendor_open_ecosystem",
      "objection_id": "obj_vendor_lock_in",
      "text": "HyperNexus is built on open standards — MCP, JSON-RPC, SQLite. Your data and configurations are portable. We provide export tools, and you can run our fork or the TormentNexus Community Edition indefinitely with no vendor lock-in. Our business model is value, not captivity.",
      "approach": "technical",
      "use_cases": ["*"],
      "success_rate": 0.75,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_vendor_longevity",
      "objection_id": "obj_vendor_long_term_bet",
      "text": "Valid concern. HyperNexus is built on a stable fork of TormentNexus, which is available under BSL/AGPL — you can self-host forever regardless of our company's fate. We also publish our architecture openly, and the entire fork's code is available on GitHub. Your investment is in the technology, not just our company.",
      "approach": "reassurance",
      "use_cases": ["*"],
      "success_rate": 0.72,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_maturity_roadmap",
      "objection_id": "obj_maturity_not_proven",
      "text": "We're shipping at a rapid pace — our public roadmap shows what's coming next quarter. What specific capabilities are you looking for? I'm happy to arrange a technical deep-dive with our engineering team to address your concerns directly.",
      "approach": "transparency",
      "use_cases": ["*"],
      "success_rate": 0.6,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_maturity_support",
      "objection_id": "obj_maturity_too_complex",
      "text": "We've optimized the onboarding experience. Most teams go from zero to fully operational in under an hour. We provide a guided setup wizard, comprehensive documentation, and dedicated onboarding support. Want to try it yourself with a sandbox environment?",
      "approach": "reassurance",
      "use_cases": ["*"],
      "success_rate": 0.63,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_integration_stack",
      "objection_id": "obj_integration_doesnt_work_with_our_stack",
      "text": "What does your current stack look like? HyperNexus integrates with any MCP-compatible tool, supports all major LLM providers (Google, Anthropic, OpenAI, DeepSeek, OpenRouter, local), and works alongside your existing CI/CD and monitoring. I'm confident we can make it work — let's validate together.",
      "approach": "consultative",
      "use_cases": ["*"],
      "success_rate": 0.68,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_integration_migration_handholding",
      "objection_id": "obj_integration_migration_cost",
      "text": "The migration is designed to be incremental — you can run HyperNexus alongside your existing tools and migrate one workflow at a time. Our engineering team provides direct support during the transition, and most teams complete the process in under a week.",
      "approach": "reassurance",
      "use_cases": ["engaged", "negotiating"],
      "success_rate": 0.65,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_support_dedicated",
      "objection_id": "obj_support_no_account_manager",
      "text": "Enterprise customers receive a dedicated Customer Success Manager, 4-hour SLA support, and direct access to our engineering team. Even Community Edition users get support via our active Discord community and GitHub discussions.",
      "approach": "reassurance",
      "use_cases": ["*"],
      "success_rate": 0.7,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    },
    {
      "id": "resp_support_on_premise",
      "objection_id": "obj_support_on_premise",
      "text": "HyperNexus runs entirely on your infrastructure — we're local-first by design. You deploy the Go sidecar and TypeScript control plane on your own servers with full air-gap support. Our Enterprise edition includes on-prem deployment assistance and custom configuration.",
      "approach": "technical",
      "use_cases": ["*"],
      "success_rate": 0.82,
      "times_used": 0,
      "last_used": "0001-01-01T00:00:00Z"
    }
  ]
}`
