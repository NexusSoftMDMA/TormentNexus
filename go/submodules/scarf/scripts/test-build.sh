#!/usr/bin/env bash
#
# Scarf test-build pipeline — produces a signed (and optionally notarized)
# Universal .app wrapped in a .dmg for testing on a remote Mac.
#
# This is NOT a release. It does not:
#   - bump the marketing/build version
#   - commit, tag, or push anything
#   - update the appcast on gh-pages
#   - create a GitHub release
#   - run the Sparkle EdDSA signing step
#
# It DOES:
#   - archive Release config Universal (arm64 + x86_64) so any remote Mac runs it
#   - export with Developer ID signing (so Gatekeeper accepts it after one allow click)
#   - optionally submit + staple notarization (so Gatekeeper accepts it silently)
#   - wrap the .app in a compact .dmg for easy scp / Drop / AirDrop
#   - write the .dmg to build/test/ with a timestamped + git-hashed filename so
#     multiple test builds don't clobber each other
#
# Usage:
#   ./scripts/test-build.sh                  # signed only — fast (~2 min). On the
#                                             # remote Mac, first launch will show
#                                             # "Apple cannot verify this app";
#                                             # right-click → Open, OR run:
#                                             #   xattr -dr com.apple.quarantine /Applications/Scarf.app
#   ./scripts/test-build.sh --notarize       # add notarytool submit + staple (~5–10 min).
#                                             # Remote Mac launches it cleanly with no
#                                             # Gatekeeper warning.
#   ./scripts/test-build.sh --arm64-only     # arm64-only build (smaller, faster). Skip if
#                                             # the remote might be Intel.
#
# Prerequisites:
#   1. Developer ID Application cert installed in login Keychain (same as release.sh).
#   2. If using --notarize: `xcrun notarytool` profile "scarf-notary" set up.
#   3. ExportOptions.plist at scripts/ExportOptions.plist (already present).
#
set -euo pipefail

# ---------- arg parsing ----------
NOTARIZE=0
ARCHS="arm64 x86_64"
VARIANT_LABEL="Universal"
for arg in "$@"; do
  case "$arg" in
    --notarize)   NOTARIZE=1 ;;
    --arm64-only) ARCHS="arm64"; VARIANT_LABEL="ARM64" ;;
    -h|--help)    sed -n '2,40p' "$0"; exit 0 ;;
    *)            printf '[ERR] unknown arg: %s\n' "$arg" >&2; exit 1 ;;
  esac
done

# ---------- config ----------
TEAM_ID="3Q6X2L86C4"
SCHEME="scarf"
PROJECT="scarf/scarf.xcodeproj"
NOTARY_PROFILE="scarf-notary"
SIGNING_IDENTITY="Developer ID Application"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$REPO_ROOT/build/test"
EXPORT_OPTIONS="$REPO_ROOT/scripts/ExportOptions.plist"

log()  { printf '\033[1;34m==> %s\033[0m\n' "$*"; }
warn() { printf '\033[1;33m[WARN] %s\033[0m\n' "$*" >&2; }
die()  { printf '\033[1;31m[ERR] %s\033[0m\n' "$*" >&2; exit 1; }

require_cmd() { command -v "$1" >/dev/null 2>&1 || die "missing required tool: $1"; }

# ---------- preflight ----------
log "Preflight"
require_cmd xcodebuild
require_cmd xcrun
require_cmd ditto
require_cmd hdiutil

cd "$REPO_ROOT"

# Pick the best available signing identity. Prefer Developer ID Application
# (release-grade — Gatekeeper accepts after notarize/staple). Fall back to
# Apple Development (dev-grade — works for testing on a Mac you own, but the
# remote will need a Gatekeeper bypass on first launch). The fallback exists
# because dev machines without a Developer ID cert still produce useful
# test builds without forcing the user to install one just to ship a DMG
# to another box they own. Notarization can't run on an Apple Development
# build — Apple won't notarize anything signed with a non-Developer-ID cert.
IDENTITIES="$(security find-identity -v -p codesigning 2>/dev/null || true)"
EXPORT_METHOD="developer-id"
USING_DEV_FALLBACK=0
if echo "$IDENTITIES" | grep -q "$SIGNING_IDENTITY"; then
  : # have Developer ID — use it
elif echo "$IDENTITIES" | grep -q "Apple Development"; then
  warn "no 'Developer ID Application' cert — falling back to Apple Development signing."
  warn "the resulting .app will need a Gatekeeper bypass on the remote (right-click → Open)."
  SIGNING_IDENTITY="Apple Development"
  EXPORT_METHOD="development"
  USING_DEV_FALLBACK=1
  if [[ $NOTARIZE -eq 1 ]]; then
    die "Apple can't notarize Apple-Development-signed builds. Install a Developer ID cert, or drop --notarize."
  fi
else
  die "no usable code-signing identity in Keychain (looked for '$SIGNING_IDENTITY' and 'Apple Development')"
fi

if [[ $NOTARIZE -eq 1 ]]; then
  xcrun notarytool history --keychain-profile "$NOTARY_PROFILE" --output-format json >/dev/null 2>&1 \
    || die "notarytool profile '$NOTARY_PROFILE' not set up — run without --notarize, or follow the release.sh header to set it up"
fi

# Note on ExportOptions when falling back to Apple Development: the release
# ExportOptions.plist hardcodes method=developer-id, which xcodebuild rejects
# for development-signed archives. We write a fallback plist below, after
# the clean step nukes BUILD_DIR.

# ---------- naming ----------
GIT_HASH="$(git rev-parse --short HEAD 2>/dev/null || echo nohash)"
GIT_DIRTY="$(git diff --quiet 2>/dev/null && git diff --cached --quiet 2>/dev/null || echo -dirty)"
TIMESTAMP="$(date +%Y%m%d-%H%M)"
MARKETING_VERSION="$(awk -F'= ' '/MARKETING_VERSION/ {gsub(/[; ]/,"",$2); print $2; exit}' "$PROJECT/project.pbxproj")"
DMG_NAME="Scarf-test-v${MARKETING_VERSION}-${TIMESTAMP}-${GIT_HASH}${GIT_DIRTY}-${VARIANT_LABEL}.dmg"
DMG_PATH="$BUILD_DIR/$DMG_NAME"

# ---------- archive + export ----------
#
# Repo lives in iCloud Drive; iCloud's daemon sprays `com.apple.FinderInfo`
# xattrs on directories the moment they get touched, which makes codesign
# reject the bundle ("Disallowed xattr"). Even an immediate `xattr -cr`
# loses the race. Work in $TMPDIR instead — non-iCloud — and only copy
# the final DMG back into the repo's build/test/.
WORK_DIR="$(mktemp -d -t scarf-test-build)"
trap 'rm -rf "$WORK_DIR"' EXIT

log "Clean build dir"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

VARIANT_DIR="$WORK_DIR"
ARCHIVE_PATH="$VARIANT_DIR/scarf.xcarchive"
EXPORT_DIR="$VARIANT_DIR/export"
APP_PATH="$EXPORT_DIR/Scarf.app"

# Write the dev-fallback ExportOptions plist into the (non-iCloud) work
# dir. The release ExportOptions.plist hardcodes method=developer-id which
# xcodebuild rejects for development-signed archives.
if [[ $USING_DEV_FALLBACK -eq 1 ]]; then
  EXPORT_OPTIONS="$WORK_DIR/ExportOptions-dev-fallback.plist"
  cat > "$EXPORT_OPTIONS" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>method</key>
  <string>${EXPORT_METHOD}</string>
  <key>teamID</key>
  <string>${TEAM_ID}</string>
  <key>signingStyle</key>
  <string>automatic</string>
  <key>destination</key>
  <string>export</string>
  <key>stripSwiftSymbols</key>
  <true/>
</dict>
</plist>
EOF
fi

log "Archive (archs: $ARCHS)"
xcodebuild \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -configuration Release \
  -archivePath "$ARCHIVE_PATH" \
  -destination "generic/platform=macOS" \
  ONLY_ACTIVE_ARCH=NO \
  ARCHS="$ARCHS" \
  archive

log "Export signed .app"
xcodebuild \
  -exportArchive \
  -archivePath "$ARCHIVE_PATH" \
  -exportPath "$EXPORT_DIR" \
  -exportOptionsPlist "$EXPORT_OPTIONS"

# Xcode exports as scarf.app — rename to Scarf.app (matches release flow).
if [[ -d "$EXPORT_DIR/scarf.app" && ! -d "$APP_PATH" ]]; then
  mv "$EXPORT_DIR/scarf.app" "$APP_PATH"
fi
[[ -d "$APP_PATH" ]] || die "exported app not found at $APP_PATH"

log "Verify signature"
codesign --verify --deep --strict --verbose=2 "$APP_PATH"

# ---------- optional notarize ----------
if [[ $NOTARIZE -eq 1 ]]; then
  NOTARIZE_ZIP="$VARIANT_DIR/Scarf-notarize.zip"
  log "Zip for notarytool"
  ditto -c -k --keepParent "$APP_PATH" "$NOTARIZE_ZIP"

  log "Submit to notarytool (blocking, up to 30m)"
  xcrun notarytool submit "$NOTARIZE_ZIP" \
    --keychain-profile "$NOTARY_PROFILE" \
    --wait \
    --timeout 30m

  log "Staple + validate"
  xcrun stapler staple "$APP_PATH"
  xcrun stapler validate "$APP_PATH"
  spctl --assess --type execute --verbose "$APP_PATH" \
    || warn "spctl --assess returned non-zero — the .app may still launch with right-click → Open"
else
  log "Skipping notarization (pass --notarize to include it)"
fi

# ---------- DMG ----------
# Build a compact UDZO-compressed DMG with /Applications symlink so users can
# just drag the icon over. Stage the contents in a temp dir first so hdiutil
# doesn't accidentally include the .xcarchive / export sibling folders.
log "Stage DMG contents"
STAGE_DIR="$VARIANT_DIR/dmg-stage"
mkdir -p "$STAGE_DIR"
ditto "$APP_PATH" "$STAGE_DIR/Scarf.app"
ln -s /Applications "$STAGE_DIR/Applications"

log "Create $DMG_NAME"
hdiutil create \
  -volname "Scarf $MARKETING_VERSION test" \
  -srcfolder "$STAGE_DIR" \
  -ov \
  -format UDZO \
  "$DMG_PATH"

# ---------- output ----------
DMG_SIZE_MB="$(du -m "$DMG_PATH" | awk '{print $1}')"

cat <<EOF

==========================================================================
Test build ready.

  $DMG_PATH
  ${DMG_SIZE_MB} MB · ${VARIANT_LABEL} · MARKETING_VERSION=${MARKETING_VERSION} · ${GIT_HASH}${GIT_DIRTY}
  Notarized: $([[ $NOTARIZE -eq 1 ]] && echo yes || echo "no (signed only)")

Copy to remote:
  scp "$DMG_PATH" <user>@<remote>:~/

EOF

if [[ $NOTARIZE -eq 0 ]]; then
  cat <<'EOF'
On the remote, first launch will show "Apple cannot verify…". Either:
  - Right-click the app → Open → Open Anyway, OR
  - Strip quarantine after copying the .app into /Applications:
      xattr -dr com.apple.quarantine /Applications/Scarf.app
EOF
fi
