#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SWIFTPM_HOME="$ROOT_DIR/.swiftpm-home"
SCRATCH_PATH="$ROOT_DIR/.build-tests"

mkdir -p \
    "$SWIFTPM_HOME/cache" \
    "$SWIFTPM_HOME/configuration" \
    "$SWIFTPM_HOME/security" \
    "$SWIFTPM_HOME/module-cache"

pick_sdk() {
    if [[ -n "${SDKROOT:-}" && -d "${SDKROOT}" ]]; then
        printf '%s\n' "$SDKROOT"
        return
    fi

    local developer_dir
    developer_dir="$(xcode-select -p 2>/dev/null || true)"
    if [[ -n "$developer_dir" && "$developer_dir" != "/Library/Developer/CommandLineTools" ]]; then
        xcrun --show-sdk-path
        return
    fi

    local clt_sdks="/Library/Developer/CommandLineTools/SDKs"
    local selected=""

    if [[ -d "$clt_sdks" ]]; then
        selected="$(ls -d "$clt_sdks"/MacOSX15.[0-9]*.sdk 2>/dev/null | sort | tail -n 1 || true)"
        if [[ -z "$selected" ]]; then
            selected="$(ls -d "$clt_sdks"/MacOSX15*.sdk 2>/dev/null | grep -v 'MacOSX15.sdk$' | sort | tail -n 1 || true)"
        fi
        if [[ -z "$selected" ]]; then
            selected="$(ls -d "$clt_sdks"/MacOSX*.sdk 2>/dev/null | grep -v 'MacOSX26' | sort | tail -n 1 || true)"
        fi
    fi

    if [[ -z "$selected" ]]; then
        selected="$(xcrun --show-sdk-path)"
    fi

    printf '%s\n' "$selected"
}

pick_frameworks_dir() {
    local developer_dir
    developer_dir="$(xcode-select -p 2>/dev/null || true)"

    if [[ -n "$developer_dir" && "$developer_dir" != "/Library/Developer/CommandLineTools" ]]; then
        local xcode_frameworks_dir="$developer_dir/Library/Developer/Frameworks"
        if [[ -d "$xcode_frameworks_dir" ]]; then
            printf '%s\n' "$xcode_frameworks_dir"
        fi
        return
    fi

    local clt_frameworks_dir="/Library/Developer/CommandLineTools/Library/Developer/Frameworks"
    if [[ -d "$clt_frameworks_dir" ]]; then
        printf '%s\n' "$clt_frameworks_dir"
    fi
}

BUILD_SDK="$(pick_sdk)"
FRAMEWORKS_DIR="$(pick_frameworks_dir)"
TEST_ENV=(
    "CLANG_MODULE_CACHE_PATH=$SWIFTPM_HOME/module-cache"
    "SDKROOT=$BUILD_SDK"
)
SWIFT_TEST_ARGS=(
    test
    --disable-sandbox
    --manifest-cache local
    --cache-path "$SWIFTPM_HOME/cache"
    --config-path "$SWIFTPM_HOME/configuration"
    --security-path "$SWIFTPM_HOME/security"
    --scratch-path "$SCRATCH_PATH"
)

if [[ -n "$FRAMEWORKS_DIR" ]]; then
    TEST_ENV+=("DYLD_FRAMEWORK_PATH=$FRAMEWORKS_DIR")
    SWIFT_TEST_ARGS+=(
        -Xswiftc -F
        -Xswiftc "$FRAMEWORKS_DIR"
        -Xlinker -F
        -Xlinker "$FRAMEWORKS_DIR"
        -Xlinker -rpath
        -Xlinker "$FRAMEWORKS_DIR"
    )
fi

env "${TEST_ENV[@]}" swift "${SWIFT_TEST_ARGS[@]}"
