---
title: Synthesized memberwise init drops @Sendable from closure params — write an explicit init
type: note
permalink: scarf/conventions/synthesized-memberwise-init-drops-sendable-from-closure-params-write-an-explicit-init
tags:
- swift6
- sendable
- swiftui
- concurrency
- gotcha
---

When a SwiftUI `View` (or any struct) stores a closure property typed `@Sendable` (e.g. `let onForget: @MainActor @Sendable () async -> Void`), the **synthesized memberwise initializer drops the `@Sendable`** from the corresponding parameter. The struct then converts a non-Sendable param into a `@Sendable` stored property *inside* the synthesized init — and the Swift-6 strict-concurrency diagnostic surfaces at the CALL SITE as "converting non-Sendable function value to '@MainActor @Sendable …' may introduce data races", not where you'd expect.

## Observations

- [symptom] The warning lands on the argument at the construction site (`Child(onFoo: onFoo)`) even though BOTH the caller's property and the child's property are declared `@Sendable`. Annotating the types alone does NOT silence it. #swift6
- [fix] Give the child an **explicit `init`** whose closure params keep the annotation: `init(..., onFoo: @escaping @MainActor @Sendable () async -> Void) { self.onFoo = onFoo }`. The explicit init preserves `@Sendable` where the synthesized one didn't. #fix
- [safe] Marking `@MainActor`-isolated closures `@Sendable` is always safe — they're invoked on the main actor regardless of where they're referenced. Thread the `@MainActor @Sendable` annotation through the whole chain (property + init param at every hop); passing `@Sendable → non-@Sendable` downstream is a safe widening and needs no change. #safe
- [seen] t-aud28: `SystemTab.onSoftDisconnect`/`onForget` in `Scarf iOS/App/ScarfGoTabRoot.swift`. Clean iOS build went 6→0 warnings only after adding the explicit init. #example

## Relations

- relates_to [[Scarf Architecture Rules]]
