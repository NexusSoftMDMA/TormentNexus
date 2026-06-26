# Contributing to Chorus

Thanks for taking an interest. Chorus is small and intentional — we're
trying to make it easy to ship a useful contribution without rewriting
half the codebase.

## Quick start

```bash
git clone https://github.com/chorus-codes/chorus.git
cd chorus
pnpm install
pnpm dev:daemon       # daemon on :7707 (tsx watch — hot reload on edit)
pnpm dev              # cockpit on :5050 (Next.js)
pnpm test             # full vitest suite (~15s)
pnpm typecheck        # tsc --noEmit
pnpm lint             # eslint
```

Node ≥ 20 required (Node 22 is the daily-driver target). We pin pnpm
to v9 in CI; please use the same locally.

## Where to start

- **Bug fix** — open an issue first if it's not obvious. Reference the
  issue in your PR.
- **Small feature** — open an issue describing the use case before
  writing code. We'd rather say "yes, that fits" early than ask you
  to rework after.
- **Big feature** — drop a draft in [`planning/`](./planning/) (a
  short markdown brief: problem, approach, alternatives, risks). Once
  we've agreed direction in the issue/PR thread, write the code.
- **Docs** — README, planning notes, code comments — always welcome.

## House rules

1. **Tests live alongside code.** `tests/<feature>.test.ts` for
   anything that's not pure types. Aim for 80% coverage on new logic.
2. **TypeScript strict, no `any`.** Use `unknown` + narrowing.
   Explicit `Promise<T>` return types on async exports.
3. **One concern per PR.** A bug fix doesn't need surrounding cleanup;
   a feature doesn't need a refactor for "while we're in there."
4. **Conventional commits.** `feat:`, `fix:`, `refactor:`, `docs:`,
   `chore:`, `test:`. The PR title becomes the squash-merge subject.
5. **Don't break `pnpm test`.** CI gates on the full suite green.
6. **Run Chorus on your PR.** We dogfood — `chorus init` + the
   `code-review` template will fan-out your diff to ≥3 lineages
   before the human review starts.

## What we don't ship

- Speculative abstractions for "future flexibility"
- Error handling for scenarios the type system already rules out
- Comments explaining *what* the code does (the code does that)
- Backwards-compat shims when we can just change the call sites

## Reporting issues

Use the issue templates in [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/).
For security issues, see [`SECURITY.md`](./SECURITY.md) — please don't
open a public issue.

## Code of Conduct

This project follows the [Contributor Covenant](./CODE_OF_CONDUCT.md).
By participating you agree to its terms.

## License

By contributing you agree your work is licensed under
[Apache-2.0](./LICENSE), the same license as the project.
