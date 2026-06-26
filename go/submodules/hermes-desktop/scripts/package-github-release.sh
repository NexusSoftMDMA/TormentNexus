#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_PATH="$ROOT_DIR/dist/HermesDesktop.app"
ZIP_PATH="$ROOT_DIR/dist/HermesDesktop.app.zip"
SHA256_PATH="$ZIP_PATH.sha256"
MANIFEST_PATH="$ZIP_PATH.manifest.json"

require_clean_release_tree() {
    if [[ "${HERMES_ALLOW_DIRTY_PACKAGE:-}" == "1" ]]; then
        echo "warning: packaging with a dirty working tree because HERMES_ALLOW_DIRTY_PACKAGE=1" >&2
        return
    fi

    local status
    status="$(git -C "$ROOT_DIR" status --porcelain --untracked-files=all)"
    if [[ -n "$status" ]]; then
        echo "error: refusing to package a release from a dirty working tree." >&2
        echo "Commit, stash, or discard local changes before running this script." >&2
        echo >&2
        printf '%s\n' "$status" >&2
        exit 1
    fi
}

plist_read() {
    local plist_path="$1"
    local key="$2"

    /usr/libexec/PlistBuddy -c "Print :$key" "$plist_path"
}

manifest_write_value() {
    local manifest_path="$1"
    local key="$2"
    local value_type="$3"
    local value="$4"

    plutil -insert "$key" "-$value_type" "$value" "$manifest_path"
}

write_release_manifest() {
    local manifest_path="$1"
    local info_plist_path="$APP_PATH/Contents/Info.plist"
    local executable_path="$APP_PATH/Contents/MacOS/$(plist_read "$info_plist_path" CFBundleExecutable)"
    local zip_sha256
    local zip_size_bytes
    local bundle_name
    local bundle_identifier
    local bundle_version
    local bundle_build
    local minimum_system_version
    local executable_name
    local detected_architectures
    local architectures=()
    local arch
    local index

    bundle_name="$(basename "$APP_PATH")"
    bundle_identifier="$(plist_read "$info_plist_path" CFBundleIdentifier)"
    bundle_version="$(plist_read "$info_plist_path" CFBundleShortVersionString)"
    bundle_build="$(plist_read "$info_plist_path" CFBundleVersion)"
    minimum_system_version="$(plist_read "$info_plist_path" LSMinimumSystemVersion)"
    executable_name="$(plist_read "$info_plist_path" CFBundleExecutable)"
    zip_sha256="$(shasum -a 256 "$ZIP_PATH" | awk '{print $1}')"
    zip_size_bytes="$(stat -f '%z' "$ZIP_PATH")"

    read -r -a detected_architectures <<<"$(lipo -archs "$executable_path")"
    if (( ${#detected_architectures[@]} == 0 )); then
        echo "error: no executable architectures detected at $executable_path" >&2
        exit 1
    fi

    IFS=$'\n' architectures=($(printf '%s\n' "${detected_architectures[@]}" | sort))
    unset IFS

    rm -f "$manifest_path"
    plutil -create xml1 "$manifest_path"
    manifest_write_value "$manifest_path" format_version integer 1
    manifest_write_value "$manifest_path" artifact_name string "$(basename "$ZIP_PATH")"
    manifest_write_value "$manifest_path" artifact_sha256 string "$zip_sha256"
    manifest_write_value "$manifest_path" artifact_size_bytes integer "$zip_size_bytes"
    manifest_write_value "$manifest_path" bundle_name string "$bundle_name"
    manifest_write_value "$manifest_path" bundle_identifier string "$bundle_identifier"
    manifest_write_value "$manifest_path" bundle_version string "$bundle_version"
    manifest_write_value "$manifest_path" bundle_build string "$bundle_build"
    manifest_write_value "$manifest_path" minimum_system_version string "$minimum_system_version"
    manifest_write_value "$manifest_path" executable_name string "$executable_name"
    plutil -insert architectures -json '[]' "$manifest_path"

    index=0
    for arch in "${architectures[@]}"; do
        manifest_write_value "$manifest_path" "architectures.$index" string "$arch"
        index=$((index + 1))
    done

    plutil -convert json -r "$manifest_path"
}

require_clean_release_tree

"$ROOT_DIR/scripts/build-macos-app.sh"

PACKAGED_VERSION="$(plist_read "$APP_PATH/Contents/Info.plist" CFBundleShortVersionString)"
if [[ -z "$PACKAGED_VERSION" || "$PACKAGED_VERSION" == "0.0.0" ]]; then
    echo "error: refusing to package a release archive with placeholder version '$PACKAGED_VERSION'." >&2
    echo "Set HERMES_VERSION or package from an exact vX.Y.Z tag." >&2
    exit 1
fi

rm -f "$ZIP_PATH"
rm -f "$SHA256_PATH"
rm -f "$MANIFEST_PATH"
xattr -cr "$APP_PATH" 2>/dev/null || true
ditto -c -k --norsrc --keepParent "$APP_PATH" "$ZIP_PATH"
(
    cd "$ROOT_DIR"
    shasum -a 256 "dist/HermesDesktop.app.zip" > "$SHA256_PATH"
)
write_release_manifest "$MANIFEST_PATH"

echo
echo "Release archive created:"
echo "  $ZIP_PATH"
echo "Checksum:"
echo "  $SHA256_PATH"
echo "Manifest:"
echo "  $MANIFEST_PATH"
