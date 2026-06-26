# Distribution And Verification

This document describes the current Hermes Desktop release model in this
repository.

It is intentionally narrow: what the public release zip is, how it is produced,
what you can verify yourself, and what the current limitations are.

## Current Release Shape

Hermes Desktop is currently distributed as:

- a universal macOS app bundle for Apple Silicon and Intel
- zipped as `HermesDesktop.app.zip`
- ad-hoc signed with `codesign --sign -`
- not notarized by Apple

The release packaging flow in this repo is script-based:

- `scripts/build-macos-app.sh` builds the app bundle, ad-hoc signs it, and
  verifies the resulting bundle with `codesign --verify --deep --strict`
- `scripts/package-github-release.sh` builds the app, zips it, creates a
  SHA-256 checksum file for the zip, and emits a small JSON manifest for the
  release artifact
- `scripts/verify-release.sh` checks the packaged zip against that manifest,
  verifies the checksum, extracts the app bundle, and validates basic bundle
  expectations plus `codesign --verify --deep --strict`

## Release Manifest

Each packaged release now includes `HermesDesktop.app.zip.manifest.json`.

The manifest is intentionally small and stable. It records:

- the release zip file name
- the SHA-256 of the zip
- the zip size in bytes
- the bundled app name
- the bundle identifier
- the bundle version and build number
- the minimum supported macOS version
- the executable name
- the executable architectures present in the shipped app

This manifest is meant to make the package easier to inspect and verify. It is
not a signature, attestation, or notarization substitute.

## What Ad-Hoc Signing Means Here

Ad-hoc signing is still useful, but it is a limited signal.

In this repository's current flow, ad-hoc signing means the bundle is signed in
a way macOS can validate for internal integrity. It does not mean:

- the app is notarized by Apple
- the app is associated with a named Apple Developer identity
- Apple has reviewed or scanned the release as part of notarization

That is why first launch may require right-click `Open`, and why macOS may warn
that Apple cannot verify the app for malware.

## First Launch And macOS 26.5.1

On most supported macOS versions, right-clicking `HermesDesktop.app`, choosing
`Open`, and confirming the warning is the normal first-launch path. If needed,
macOS also provides `Open Anyway` under `System Settings` > `Privacy &
Security` shortly after a blocked launch.

On some Macs running macOS 26.5.1 (build 25F80), we have observed Gatekeeper
showing a stronger "is damaged and can't be opened" alert for quarantined,
ad-hoc signed apps without offering the normal override. The following
workaround is only for that macOS version and build when that exact alert
appears.

Before changing the quarantine attribute, verify the downloaded zip against
`HermesDesktop.app.zip.sha256` attached to the same GitHub Release:

```bash
shasum -a 256 ~/Downloads/HermesDesktop.app.zip
```

Do not continue if the result differs from the published checksum. After
extracting the verified zip and moving the app into Applications, run:

```bash
xattr -dr com.apple.quarantine "/Applications/HermesDesktop.app"
open "/Applications/HermesDesktop.app"
```

This removes the browser quarantine from that app bundle only. It does not
disable Gatekeeper globally and does not require `sudo`.

## What The Published Checksum Proves

Each release zip includes a SHA-256 checksum and a small JSON manifest.

Comparing your local download against that checksum is useful because it lets
you confirm your copy matches the release asset you downloaded.

It does not, by itself, prove:

- who created the release
- that the release contents are benign
- that the GitHub account or release page you trusted was uncompromised

Checksums help you detect mismatch or corruption. They are not a substitute for
reviewing the source or understanding the distribution model.

## How To Verify A Release Zip

After downloading `HermesDesktop.app.zip`:

```bash
shasum -a 256 HermesDesktop.app.zip
```

Compare the output with the checksum published in the GitHub Release.

If you want the repo to perform the same local verification flow used in CI,
run this from a checkout of the repository:

```bash
./scripts/verify-release.sh \
  /path/to/HermesDesktop.app.zip \
  /path/to/HermesDesktop.app.zip.manifest.json
```

After extracting and moving the app into place:

```bash
codesign --verify --deep --strict /Applications/HermesDesktop.app
```

If you want more visibility into what macOS sees in the bundle:

```bash
codesign -dv --verbose=4 /Applications/HermesDesktop.app
```

For network behavior, you can observe live connections with Little Snitch,
LuLu, `lsof`, or `nettop`.

## Strongest Trust Path In This Repo

The strongest trust path available in this repository today is to inspect the
source and build the app yourself:

```bash
./scripts/build-macos-app.sh
```

That produces a local app bundle in `dist/HermesDesktop.app`.

This is still an ad-hoc signed, non-notarized bundle, because that is the
current build and release model in the repo. Building locally does not turn it
into a notarized distribution, but it does let you trust your own build inputs
instead of a downloaded zip. Because the resulting app is produced locally, it
also does not inherit browser download quarantine. You can launch it directly:

```bash
open "dist/HermesDesktop.app"
```

## What The App Depends On At Runtime

Trusting the release zip is only part of the picture. Hermes Desktop also
depends on the environment it connects to.

The app assumes:

- `python3` is available on the machine where Hermes runs
- Hermes data lives under `~/.hermes`, a named profile, or the configured
  custom Hermes home
- `hermes` is available through the app's prepared PATH for in-app chat,
  terminal resume, and workflow launches

For direct-local mode, commands run on this Mac with the current macOS user's
permissions. For SSH mode, your Mac must already have a working
non-interactive SSH path to the target host. See [../SECURITY.md](../SECURITY.md)
for the runtime security model.

## Honest Limitation

The current release model is intentionally documented without stronger claims
than the code and scripts support today.

If the project later adopts Developer ID signing, notarization, or additional
release verification, this document should be updated at the same time as the
implementation. Until then, the correct description is simple:

- public releases are ad-hoc signed
- public releases are not notarized
- published checksums help verify the downloaded zip
- the release manifest makes the artifact metadata easier to inspect and
  compare
- building from source is the clearest trust path for cautious users
