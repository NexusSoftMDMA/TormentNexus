# Building Scarf

Scarf is a native macOS app built with Xcode. For contributor builds, use the local script:

```bash
./scripts/local-build.sh
```

Requirements:

- macOS 14.6 (Sonoma) or newer at runtime — that's the app's `MACOSX_DEPLOYMENT_TARGET`. Sonoma support is intentional and load-bearing; do not raise this without an explicit decision to drop Sonoma users
- Xcode 16.0 or newer, selected by `xcode-select` (needed for Swift 6 strict-concurrency features the project uses)
- Metal toolchain installed
- Hermes installed at `~/.hermes/` (see the project README for setup)

If the Metal toolchain is missing, the script will offer to install it in interactive shells. You can also install it manually:

```bash
xcodebuild -downloadComponent MetalToolchain
```

`scripts/local-build.sh` resolves Swift package dependencies, detects `arm64` vs `x86_64`, and builds the Debug app unsigned. Signing is intentionally disabled for local Debug builds so contributors do not need the maintainer's Apple Developer account.

Release signing is separate from contributor builds. Maintainers should continue using the existing release process for signed distributable builds.
