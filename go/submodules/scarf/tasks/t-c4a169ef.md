---
id: t-c4a169ef
title: Fix iOS chat keyboard-dismiss button (overlap + no-op, iOS 26.5)
status: done
added: 2026-06-24
priority: high
---

## Description

v2.12.0 (44) iOS reports: the "hide keyboard" button (1) is overlaid on top of the paperclip/attachment button and (2) does nothing when tapped; users say there's no way to dismiss the keyboard in chat.

Root cause: `ChatView.body` (scarf/Scarf iOS/Chat/ChatView.swift:157) declares a `ToolbarItemGroup(placement: .keyboard)` dismiss button. That keyboard accessory bar sits directly above the keyboard — but the chat composer is ALSO pinned above the keyboard (last child of the body VStack). Both occupy the same strip, so the accessory's leading "hide keyboard" chevron renders on top of the composer's leading paperclip and the two fight for hit-testing (taps land on the paperclip, not the dismiss button). The `.keyboard` accessory was added in 8023097 (gh#107) and shipped since v2.10.1; iOS 26.5's keyboard-toolbar behavior change is what newly exposed the collision.

Fix: remove the `.keyboard` ToolbarItemGroup and place the dismiss affordance as a normal inline button in `composerRow` (after the optional paperclip, before the TextField), shown only while `composerFocused`. Normal view = no system-accessory mystery, no hit-test conflict, no overlap. Keep `.scrollDismissesKeyboard(.interactively)` for swipe-down on non-empty transcripts (the inline button covers the empty/resumed-empty state the screenshots show). Update the now-stale comments at 149-156 and 756-762.

Scope: single file, iOS only. No ScarfCore/shared changes. Verify: iOS target compiles clean; visually confirm composer in simulator.

## Plan



## Artifacts



