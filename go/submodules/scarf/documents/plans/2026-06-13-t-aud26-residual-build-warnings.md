# Plan — t-aud26: Clear the 5 residual build warnings (toward zero-warning)

> Status: NOT STARTED · Created 2026-06-13 · Owner: next session · Risk: LOW ·
> Source: t-aud23 clean build. These are PRE-EXISTING and a different class from
> t-aud23's main-actor-isolation backlog (which is already 0). The project rule is a
> zero-warning build (`Scarf Architecture Rules`), so close these out.

## The 5 warnings (from a clean macOS build)

1. `scarf/scarf/Core/Services/OAuthKeepaliveCronService.swift:93` —
   "expression is 'async' but is not marked with 'await'; this is an error in the Swift 6
   language mode." **Swift-6 error-class.** Inspect line 93: an `async` call is being made
   without `await`. Fix = add `await` (confirm the enclosing context is `async`; if not,
   wrap appropriately). Verify behavior unchanged.

2. `scarf/scarf/Features/Templates/ViewModels/CatalogViewModel.swift:121` —
   "no 'async' operations occur within 'await' expression." The `await applyLoad(...)` (or
   adjacent) awaits something that doesn't actually suspend. Inspect `applyLoad`: if it's a
   synchronous/`@MainActor`-non-async call, drop the redundant `await`; if it *should* be
   async, leave it and the warning is benign. Prefer removing the redundant `await`.

3. `scarf/scarf/Features/Chat/Views/RichChatInputBar.swift:502` —
   "result of call to 'loadObject(ofClass:completionHandler:)' is unused." Fix =
   `_ = provider.loadObject(...)` (the call is a fire-and-forget drag-load; the discarded
   `NSProgress` is intentional). Confirm we don't need the progress handle.

4. `scarf/scarf/Features/Health/ViewModels/HealthViewModel.swift:686` —
   "result of call to 'run(resultType:body:)' is unused." Fix = `_ = …run(resultType:body:)`
   (or `@discardableResult` if that API is ours and the result is genuinely optional).

5. `scarf/scarf/Features/Templates/Views/TemplateMarkdown.swift:62` —
   "variable 'lines' was never mutated; consider changing to 'let' constant." Fix =
   `var lines` → `let lines`. Trivial.

## Approach

- 3, 4, 5 are trivial one-liners (`_ =` / `let`). Do them first.
- 1 and 2 are async/await — read the surrounding function, apply the right fix (add `await`
  for #1; remove the redundant `await` for #2), and make sure no behavior changes.
- Each edited file needs Read-before-Edit; they're small regions.

## Verification

- `xcodebuild -project scarf/scarf.xcodeproj -scheme scarf -configuration Debug
  -destination 'platform=macOS' clean build` → **0 warnings** (grep `": warning:"`,
  excluding the benign `appintentsmetadataprocessor` line).
- App still builds + runs; OAuth keepalive cron + catalog refresh behave unchanged.
- Also build `scarf mobile` (RichChatInputBar / shared) to confirm no iOS fallout.

## Notes

- This finishes the zero-warning goal that t-aud23 started (t-aud23 = the 74 main-actor
  warnings; this = the last 5 of a different kind).
