---
title: iOS chat: dismiss the keyboard with an inline composer button, never a .toolbar(.keyboard) accessory
type: note
permalink: scarf/conventions/ios-chat-dismiss-the-keyboard-with-an-inline-composer
created: 2026-06-24
updated: 2026-06-24
tags:
- ios
- chat
- keyboard
- swiftui
- gotcha
---

## Observations

- [gotcha] The iOS chat composer (`scarf/Scarf iOS/Chat/ChatView.swift`) is the last child of the body `VStack`, so SwiftUI keyboard-avoidance pins it directly above the keyboard. A `ToolbarItemGroup(placement: .keyboard)` accessory bar ALSO sits directly above the keyboard — the two occupy the same strip and collide. #keyboard
- [symptom] On iOS 26.5 the collision surfaced as the accessory's `keyboard.chevron.compact.down` drawn on top of the composer's leading `paperclip`, and the paperclip stole the chevron's taps → users reported "hide-keyboard button overlays attachment" + "button does nothing" + "no way to hide keyboard" (v2.12.0/build 44 TestFlight reports, 2026-06). #bug
- [history] The `.keyboard` accessory was added in `8023097` (gh#107, "no button to hide keyboard") and shipped since v2.10.1. It was latent until iOS 26.5's keyboard-toolbar behavior change exposed it — NOT a 2.12.0 code regression. #regression
- [rule] Dismiss the iOS chat keyboard with a PLAIN inline `Button { composerFocused = false }` inside `composerRow` (rendered `if composerFocused`), NOT a `.toolbar(placement: .keyboard)` accessory. A normal sibling button can't overlap the composer and always receives its own tap; `composerFocused` is the `@FocusState` bound via `.focused(...)`. #rule
- [keep] `.scrollDismissesKeyboard(.interactively)` on the message list stays — it covers swipe-down on a scrollable transcript; the inline button covers the empty / resumed-empty state where there's nothing to scroll. #keyboard
- [apple-apps-caveat] Mail/Notes/Messages use `.keyboard` accessory bars because their text view fills the screen — they have no separate composer pinned above the keyboard. A chat composer does, so the Apple-apps precedent does NOT transfer here. #design

## Relations
- relates_to [[iOS Platform Rules]]
