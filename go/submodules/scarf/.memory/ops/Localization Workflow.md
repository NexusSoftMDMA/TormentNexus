---
title: Localization Workflow
type: note
permalink: scarf/ops/localization-workflow
tags:
- i18n
- localization
---

## Observations
- [catalog] Localizable.xcstrings is the source of truth; per-locale JSONs live in tools/translations/<locale>.json (key = English source string, value = translation). Omitted keys fall back to English at runtime — use that for proper nouns (Scarf, Hermes, Anthropic, OAuth, SSH) and technical terms #catalog
- [merge] tools/merge-translations.py merges JSON files into Localizable.xcstrings; LOCALES list in that script gates which locales are processed #tooling
- [rule] Preserve format specifiers exactly: %@, %lld, %d. Use positional forms (%1$@, %2$lld) when word order needs to change #rule
- [step] Adding a locale: add to knownRegions in project.pbxproj, add JSON in tools/translations/, add to LOCALES in merge script, translate InfoPlist.xcstrings (mic permission), spot-check Dashboard/Chat/Settings in Xcode App Language #howto
- [reference] Deeper context: scarf/docs/I18N.md #docs
