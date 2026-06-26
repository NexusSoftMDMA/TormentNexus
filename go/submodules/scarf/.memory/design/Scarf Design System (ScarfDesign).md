---
title: Scarf Design System (ScarfDesign)
type: note
permalink: scarf/design/scarf-design-system-scarf-design
tags:
- design
- ui
---

## Observations
- [package] All app UI uses the typed token bundle at scarf/Packages/ScarfDesign/. Both `scarf` and `scarf mobile` targets `import ScarfDesign` #package
- [colors] Tokens: ScarfColor.accent, .foregroundPrimary/Muted/Faint, .backgroundPrimary/Secondary/Tertiary, .border/.borderStrong, .success/.danger/.warning/.info, .Tool.{bash,edit,search,web,think}. Resolve from ScarfBrand.xcassets; auto light/dark #color
- [typography] Use .scarfStyle(.title2/.body/.captionUppercase/…) — eleven preset styles. Never .font(.system(size: 13.5)) #typography
- [spacing] ScarfSpace.s1…s10 (4/8/12/16/20/24/32/40), ScarfRadius.sm/md/lg/xl/xxl/pill, .scarfShadow(.sm/.md/.lg/.xl). Hardcoded .padding(12) or cornerRadius:8 is a code smell #spacing
- [components] ScarfPageHeader, ScarfCard, ScarfBadge, ScarfTextField, ScarfSectionHeader, ScarfDivider, ScarfPrimary/Secondary/Ghost/DestructiveButton (apply with .buttonStyle(...)) #components
- [branding] Rust accent palette. AccentColor.colorset resolves Color.accentColor to rust so unmigrated SwiftUI controls still tint correctly. Don't introduce purple/violet (legacy) #branding
- [anti-pattern] Don't use yellow #F0AD4E for success — that's .warning; .success is green. Don't ship terminal/syntax-highlight palettes through ScarfColor — keep content semantics inline #pitfalls
- [reference] Full screen mockups live at design/static-site/ui-kit/*.jsx (open design/static-site/index.html). ScarfChatView.ChatRootView is a 3-pane chat redesign target — preview only, not yet swapped into live chat (RichChatView owns the real ACP pipeline) #reference

## Relations
- applies_to [[Scarf Architecture Rules]]
- extended_by [[iOS Platform Rules]]