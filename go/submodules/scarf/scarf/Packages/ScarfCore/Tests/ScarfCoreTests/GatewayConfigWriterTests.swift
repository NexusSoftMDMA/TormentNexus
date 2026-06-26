import Testing
import Foundation
@testable import ScarfCore

/// Round-trip + idempotence tests for `GatewayConfigWriter.setList`. Pure
/// `String` operations only — runs cleanly on Linux SwiftPM.
///
/// Hermes v0.16 reads allowlists from top-level `<platform>.allowed_*`
/// sections, so every assertion below targets that path (NOT the old
/// `gateway.platforms.<platform>.*` lineage).
@Suite struct GatewayConfigWriterTests {

    // MARK: - Insert

    @Test func setListInsertsBlockOnEmpty() {
        let yaml = ""
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: ["C0123ABCD", "C0456EFGH"]
        )
        // Top-level `slack:` section, NOT `gateway.platforms.slack`.
        #expect(updated.contains("slack:"))
        #expect(!updated.contains("gateway:"))
        #expect(!updated.contains("platforms:"))
        #expect(updated.contains("  allowed_channels:"))
        #expect(updated.contains("    - C0123ABCD"))
        #expect(updated.contains("    - C0456EFGH"))
    }

    @Test func setListAppendsScaffoldPreservingPriorContent() {
        let yaml = """
        model:
          default: gpt-4o
          provider: openai
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: ["C01"]
        )
        // Original content preserved verbatim at the top.
        #expect(updated.contains("model:"))
        #expect(updated.contains("  default: gpt-4o"))
        #expect(updated.contains("  provider: openai"))
        // New top-level `slack:` scaffold appended.
        #expect(updated.contains("slack:"))
        #expect(updated.contains("  allowed_channels:"))
        #expect(updated.contains("    - C01"))
    }

    // MARK: - Replace

    @Test func setListReplacesExistingBlock() {
        let yaml = """
        slack:
          allowed_channels:
            - C_OLD_1
            - C_OLD_2
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: ["C_NEW_1"]
        )
        #expect(updated.contains("- C_NEW_1"))
        #expect(!updated.contains("- C_OLD_1"))
        #expect(!updated.contains("- C_OLD_2"))
    }

    @Test func setListPreservesScalarSiblings() {
        // The `gateway_restart_notification` scalar sibling of
        // `allowed_channels` (plus any Scarf-managed key like
        // `reply_to_mode`) must stay byte-for-byte after a list-write to the
        // same platform.
        let yaml = """
        slack:
          allowed_channels:
            - C_OLD
          reply_to_mode: thread
          gateway_restart_notification: true
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: ["C_NEW"]
        )
        #expect(updated.contains("- C_NEW"))
        #expect(!updated.contains("- C_OLD"))
        // Scalars at the same indent must survive.
        #expect(updated.contains("reply_to_mode: thread"))
        #expect(updated.contains("gateway_restart_notification: true"))
    }

    @Test func setListMergesIntoExistingPlatformSectionPreservingSiblingKey() {
        // CRITICAL merge case: `slack:` already holds a Scarf-managed
        // sibling (`reply_to_mode`) but no `allowed_channels` yet. Writing
        // the allowlist must splice the key into the EXISTING section, not
        // clobber it or append a duplicate `slack:` block.
        let yaml = """
        slack:
          reply_to_mode: thread
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: ["C99"]
        )
        // Sibling key preserved.
        #expect(updated.contains("reply_to_mode: thread"))
        // Allowlist landed under the same section.
        #expect(updated.contains("  allowed_channels:"))
        #expect(updated.contains("    - C99"))
        // Exactly one `slack:` header — no duplicate section appended.
        let slackHeaders = updated
            .components(separatedBy: "\n")
            .filter { $0 == "slack:" }
        #expect(slackHeaders.count == 1)
    }

    @Test func setListPreservesOtherPlatformsBlocks() {
        // Editing slack must not touch matrix.
        let yaml = """
        slack:
          allowed_channels:
            - C_SLACK
        matrix:
          allowed_rooms:
            - '!room1:matrix.org'
            - '!room2:matrix.org'
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: ["C_SLACK_NEW"]
        )
        #expect(updated.contains("- C_SLACK_NEW"))
        // Matrix block intact.
        #expect(updated.contains("matrix:"))
        #expect(updated.contains("'!room1:matrix.org'"))
        #expect(updated.contains("'!room2:matrix.org'"))
    }

    // MARK: - Remove

    @Test func setListWithEmptyItemsRemovesBlock() {
        let yaml = """
        slack:
          allowed_channels:
            - C01
            - C02
          gateway_restart_notification: true
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: []
        )
        // Block removed; sibling scalar preserved.
        #expect(!updated.contains("allowed_channels:"))
        #expect(!updated.contains("- C01"))
        #expect(!updated.contains("- C02"))
        #expect(updated.contains("gateway_restart_notification: true"))
    }

    @Test func setListWithEmptyItemsOnAbsentBlockIsNoOp() {
        let yaml = """
        model:
          default: gpt-4o
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml,
            platform: "slack",
            key: "allowed_channels",
            items: []
        )
        #expect(updated == yaml)
    }

    // MARK: - Idempotence

    @Test func setListIsIdempotent() {
        let yaml = """
        model:
          default: gpt-4o
        """
        let once = GatewayConfigWriter.setList(
            in: yaml,
            platform: "telegram",
            key: "allowed_chats",
            items: ["@alice", "@bob"]
        )
        let twice = GatewayConfigWriter.setList(
            in: once,
            platform: "telegram",
            key: "allowed_chats",
            items: ["@alice", "@bob"]
        )
        #expect(once == twice)
    }

    @Test func setListReplaceThenReplaceIsStable() {
        let yaml = ""
        let a = GatewayConfigWriter.setList(
            in: yaml, platform: "matrix", key: "allowed_rooms",
            items: ["!a:m", "!b:m"]
        )
        let b = GatewayConfigWriter.setList(
            in: a, platform: "matrix", key: "allowed_rooms",
            items: ["!c:m"]
        )
        #expect(b.contains("- '!c:m'"))
        #expect(!b.contains("'!a:m'"))
        #expect(!b.contains("'!b:m'"))
    }

    // MARK: - Quoting

    @Test func setListQuotesItemsContainingColons() {
        // Matrix room IDs contain `:` — must be single-quoted.
        let yaml = ""
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "matrix", key: "allowed_rooms",
            items: ["!RoomId:matrix.org"]
        )
        #expect(updated.contains("'!RoomId:matrix.org'"))
    }

    @Test func setListQuotesItemsStartingWithAt() {
        // Telegram usernames `@alice`.
        let yaml = ""
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "telegram", key: "allowed_chats",
            items: ["@alice"]
        )
        #expect(updated.contains("'@alice'"))
    }

    @Test func setListLeavesPlainAlphanumericUnquoted() {
        // Slack channel IDs are A-Z0-9 — emit unquoted for readability.
        let yaml = ""
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "slack", key: "allowed_channels",
            items: ["C0123ABCD"]
        )
        #expect(updated.contains("- C0123ABCD"))
        #expect(!updated.contains("'C0123ABCD'"))
    }

    @Test func setListEscapesEmbeddedSingleQuotes() {
        let yaml = ""
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "slack", key: "allowed_channels",
            items: ["weird:'name"]
        )
        // Embedded single quote doubled per YAML spec.
        #expect(updated.contains("'weird:''name'"))
    }

    // MARK: - Insertion when section exists but key is absent

    @Test func setListInsertsKeyUnderExistingPlatformBlock() {
        // Top-level `slack:` exists with a `gateway_restart_notification`
        // scalar; `allowed_channels` is missing. Add it without disturbing
        // the scalar sibling.
        let yaml = """
        slack:
          gateway_restart_notification: false
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "slack", key: "allowed_channels",
            items: ["C42"]
        )
        #expect(updated.contains("gateway_restart_notification: false"))
        #expect(updated.contains("allowed_channels:"))
        #expect(updated.contains("- C42"))
    }

    // MARK: - Round-trip with the YAML loader

    @Test func roundTripsThroughHermesConfigYAMLLoader() {
        // Write a list, then parse the result through HermesConfig+YAML and
        // confirm we read back what we wrote — proving the writer lands the
        // allowlist exactly where the reader looks (top-level
        // `slack.allowed_channels`).
        var yaml = ""
        yaml = GatewayConfigWriter.setList(
            in: yaml, platform: "slack", key: "allowed_channels",
            items: ["C01", "C02"]
        )
        let cfg = HermesConfig(yaml: yaml)
        let block = cfg.gatewayPlatforms["slack"]
        #expect(block?.allowedChannels == ["C01", "C02"])
    }

    @Test func roundTripMergesAndPreservesSiblingThroughLoader() {
        // Full merge round-trip: an existing `slack:` section with a sibling
        // key, write an allowlist, confirm via the loader that the allowlist
        // lands at `slack.allowed_channels` (NOT
        // `gateway.platforms.slack.allowed_channels`) and the sibling
        // survives.
        let yaml = """
        slack:
          reply_to_mode: thread
        """
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "slack", key: "allowed_channels",
            items: ["C7", "C8"]
        )
        let cfg = HermesConfig(yaml: updated)
        #expect(cfg.gatewayPlatforms["slack"]?.allowedChannels == ["C7", "C8"])
        // Sibling key still present in the raw YAML.
        #expect(updated.contains("reply_to_mode: thread"))
    }
}
