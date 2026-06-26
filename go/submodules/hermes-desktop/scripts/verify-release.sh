#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_ZIP_PATH="$ROOT_DIR/dist/HermesDesktop.app.zip"
ZIP_PATH="${1:-$DEFAULT_ZIP_PATH}"
MANIFEST_PATH="${2:-$ZIP_PATH.manifest.json}"
SHA256_PATH="$ZIP_PATH.sha256"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hermes-release-verify.XXXXXX")"

cleanup() {
    rm -rf "$TMP_DIR"
}

trap cleanup EXIT

manifest_read_raw() {
    local key="$1"

    plutil -extract "$key" raw -o - "$MANIFEST_PATH"
}

info_read() {
    local plist_path="$1"
    local key="$2"

    /usr/libexec/PlistBuddy -c "Print :$key" "$plist_path"
}

info_read_optional() {
    local plist_path="$1"
    local key="$2"

    /usr/libexec/PlistBuddy -c "Print :$key" "$plist_path" 2>/dev/null || true
}

fail() {
    echo "error: $*" >&2
    exit 1
}

assert_file_exists() {
    local path="$1"

    [[ -e "$path" ]] || fail "expected path not found: $path"
}

assert_equals() {
    local label="$1"
    local expected="$2"
    local actual="$3"

    if [[ "$expected" != "$actual" ]]; then
        fail "$label mismatch: expected '$expected' but found '$actual'"
    fi
}

normalize_arches() {
    printf '%s\n' "$@" | sort | paste -sd ' ' -
}

assert_file_exists "$ZIP_PATH"
assert_file_exists "$MANIFEST_PATH"

EXPECTED_ARTIFACT_NAME="$(manifest_read_raw artifact_name)"
EXPECTED_SHA256="$(manifest_read_raw artifact_sha256)"
EXPECTED_SIZE_BYTES="$(manifest_read_raw artifact_size_bytes)"
EXPECTED_BUNDLE_NAME="$(manifest_read_raw bundle_name)"
EXPECTED_BUNDLE_IDENTIFIER="$(manifest_read_raw bundle_identifier)"
EXPECTED_BUNDLE_VERSION="$(manifest_read_raw bundle_version)"
EXPECTED_BUNDLE_BUILD="$(manifest_read_raw bundle_build)"
EXPECTED_MINIMUM_SYSTEM_VERSION="$(manifest_read_raw minimum_system_version)"
EXPECTED_EXECUTABLE_NAME="$(manifest_read_raw executable_name)"
EXPECTED_ARCH_COUNT="$(plutil -extract architectures raw -expect array -o - "$MANIFEST_PATH")"
EXPECTED_ARCHES=()

for (( index=0; index<EXPECTED_ARCH_COUNT; index++ )); do
    EXPECTED_ARCHES+=("$(manifest_read_raw "architectures.$index")")
done

assert_equals "artifact name" "$EXPECTED_ARTIFACT_NAME" "$(basename "$ZIP_PATH")"
assert_equals "zip size" "$EXPECTED_SIZE_BYTES" "$(stat -f '%z' "$ZIP_PATH")"
assert_equals "zip sha256" "$EXPECTED_SHA256" "$(shasum -a 256 "$ZIP_PATH" | awk '{print $1}')"

if [[ -f "$SHA256_PATH" ]]; then
    CHECKSUM_FILE_HASH="$(awk '{print $1}' "$SHA256_PATH")"
    assert_equals "sha256 sidecar" "$EXPECTED_SHA256" "$CHECKSUM_FILE_HASH"
fi

ditto -x -k "$ZIP_PATH" "$TMP_DIR"

APP_PATH="$TMP_DIR/$EXPECTED_BUNDLE_NAME"
INFO_PLIST_PATH="$APP_PATH/Contents/Info.plist"
EXECUTABLE_PATH="$APP_PATH/Contents/MacOS/$EXPECTED_EXECUTABLE_NAME"
RESOURCE_BUNDLE_PATH="$APP_PATH/Contents/Resources/HermesDesktop_HermesDesktop.bundle"
ICON_PATH="$APP_PATH/Contents/Resources/AppIcon.icns"

assert_file_exists "$APP_PATH"
assert_file_exists "$INFO_PLIST_PATH"
assert_file_exists "$EXECUTABLE_PATH"
assert_file_exists "$RESOURCE_BUNDLE_PATH"
assert_file_exists "$ICON_PATH"

assert_equals "bundle identifier" "$EXPECTED_BUNDLE_IDENTIFIER" "$(info_read "$INFO_PLIST_PATH" CFBundleIdentifier)"
assert_equals "bundle version" "$EXPECTED_BUNDLE_VERSION" "$(info_read "$INFO_PLIST_PATH" CFBundleShortVersionString)"
assert_equals "bundle build" "$EXPECTED_BUNDLE_BUILD" "$(info_read "$INFO_PLIST_PATH" CFBundleVersion)"
assert_equals "minimum system version" "$EXPECTED_MINIMUM_SYSTEM_VERSION" "$(info_read "$INFO_PLIST_PATH" LSMinimumSystemVersion)"
assert_equals "bundle executable" "$EXPECTED_EXECUTABLE_NAME" "$(info_read "$INFO_PLIST_PATH" CFBundleExecutable)"

LOCAL_NETWORK_USAGE_DESCRIPTION="$(info_read_optional "$INFO_PLIST_PATH" NSLocalNetworkUsageDescription)"
if [[ -z "$LOCAL_NETWORK_USAGE_DESCRIPTION" ]]; then
    fail "missing NSLocalNetworkUsageDescription; local LAN SSH hosts may be blocked by macOS Local Network privacy"
fi

read -r -a ACTUAL_ARCHES <<<"$(lipo -archs "$EXECUTABLE_PATH")"
assert_equals \
    "executable architectures" \
    "$(normalize_arches "${EXPECTED_ARCHES[@]}")" \
    "$(normalize_arches "${ACTUAL_ARCHES[@]}")"

codesign --verify --deep --strict "$APP_PATH" >/dev/null

echo "Release verification succeeded:"
echo "  Zip: $ZIP_PATH"
echo "  Manifest: $MANIFEST_PATH"
echo "  Bundle: $APP_PATH"
echo "  Version: $EXPECTED_BUNDLE_VERSION ($EXPECTED_BUNDLE_BUILD)"
echo "  Architectures: $(normalize_arches "${ACTUAL_ARCHES[@]}")"
