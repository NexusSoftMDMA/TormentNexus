---
title: iOS Platform Rules
type: note
permalink: scarf/design/i-os-platform-rules
tags:
- ios
- design
- accessibility
---

## Observations
- [rule] Dynamic Type policy: ScarfFont has fixed point sizes — using it blanket on iOS would regress accessibility. Use ScarfFont only for chrome/badges/intentional-display elements (status badges, chip labels, header chrome). Keep semantic .font(.headline)/.body/.caption for content (list rows, body copy, error messages, chat content) #a11y
- [decision-tree] Per text element: 'is this read for content?' → semantic token; 'is this chrome/label/badge?' → ScarfFont #a11y
- [clamp] iOS clamps Dynamic Type at scene root (ScarfIOSApp.swift): .dynamicTypeSize(.xSmall ... .accessibility2) — keep that clamp #a11y
- [rule] Mac has no Dynamic Type constraint — adopts ScarfFont everywhere #platform
- [page-chrome] Don't retrofit ScarfPageHeader over iOS tab roots — iOS uses .navigationTitle + .navigationBarTitleDisplayMode(.large). ScarfPageHeader on iOS only for sub-views without a native large-title bar #navigation
- [buttons] iOS: only swap .borderedProminent → ScarfPrimaryButton. Leave .bordered and .plain native — they're iOS conventions and inherit rust via AccentColor.colorset #buttons

## Relations
- refines [[Scarf Design System (ScarfDesign)]]