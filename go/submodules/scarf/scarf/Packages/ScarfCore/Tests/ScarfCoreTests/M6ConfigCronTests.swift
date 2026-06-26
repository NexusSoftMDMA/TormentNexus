import Testing
import Foundation
@testable import ScarfCore

/// M6: YAML parser port + HermesConfig loader. Pure functions — no
/// `ServerContext.sshTransportFactory` races, so this suite can run
/// in parallel with everything else.
///
/// The write-path tests for Cron editing + Settings-from-yaml live
/// in `M5FeatureVMTests` (the serialized suite that already owns
/// the factory-install pattern) to avoid cross-suite parallel
/// collisions on the shared factory static.
@Suite struct M6ConfigCronTests {

    // MARK: - YAML parser

    @Test func parsesScalarKeyValues() {
        let yaml = """
        model:
          default: gpt-4o
          provider: openai
        """
        let p = HermesYAML.parseNestedYAML(yaml)
        #expect(p.values["model.default"] == "gpt-4o")
        #expect(p.values["model.provider"] == "openai")
    }

    @Test func parsesBulletLists() {
        let yaml = """
        permanent_allowlist:
          - ls
          - pwd
          - 'cat /etc/hostname'
        """
        let p = HermesYAML.parseNestedYAML(yaml)
        #expect(p.lists["permanent_allowlist"] == ["ls", "pwd", "cat /etc/hostname"])
    }

    @Test func parsesNestedMaps() {
        let yaml = """
        terminal:
          docker_env:
            PATH: /usr/local/bin
            HOME: /home/hermes
        """
        let p = HermesYAML.parseNestedYAML(yaml)
        #expect(p.maps["terminal.docker_env"]?["PATH"] == "/usr/local/bin")
        #expect(p.maps["terminal.docker_env"]?["HOME"] == "/home/hermes")
        #expect(p.values["terminal.docker_env.PATH"] == "/usr/local/bin")
    }

    @Test func ignoresCommentsAndBlankLines() {
        let yaml = """
        # Top-level comment
        model:
          # inline comment
          default: gpt-4o

          provider: openai
        """
        let p = HermesYAML.parseNestedYAML(yaml)
        #expect(p.values["model.default"] == "gpt-4o")
        #expect(p.values["model.provider"] == "openai")
    }

    @Test func stripsQuotes() {
        #expect(HermesYAML.stripYAMLQuotes("'quoted'") == "quoted")
        #expect(HermesYAML.stripYAMLQuotes("\"quoted\"") == "quoted")
        #expect(HermesYAML.stripYAMLQuotes("plain") == "plain")
        #expect(HermesYAML.stripYAMLQuotes("'unbalanced") == "'unbalanced")
        #expect(HermesYAML.stripYAMLQuotes("") == "")
    }

    @Test func handlesInlineLiterals() {
        let yaml = """
        empty_map: {}
        empty_list: []
        """
        let p = HermesYAML.parseNestedYAML(yaml)
        #expect(p.maps["empty_map"] != nil)
        #expect(p.lists["empty_list"] != nil)
    }

    // MARK: - HermesConfig from YAML

    @Test func emptyYAMLProducesDefaults() {
        let c = HermesConfig(yaml: "")
        #expect(c.model == "unknown")
        #expect(c.provider == "unknown")
        #expect(c.display.skin == "default")
        #expect(c.streaming == true)
        #expect(c.security.redactSecrets == true)
        #expect(c.compression.enabled == true)
        #expect(c.voice.ttsProvider == "edge")
        // v0.13 additions default to empty / off when the YAML omits
        // them — pre-v0.13 hosts produce this exact shape.
        #expect(c.imageGenModel == "")
        #expect(c.openrouterResponseCacheEnabled == false)
    }

    @Test func parsesImageGenAndOpenRouterCache() {
        // WS-6 / v0.16: round-trip the two new top-level keys. Hermes
        // v0.16 reads `openrouter.response_cache` as a SCALAR bool
        // directly under `openrouter:`. This test pins the parser line +
        // setter key + UI binding to that single shape.
        let yaml = """
        image_gen:
          model: openai/gpt-image-1
        openrouter:
          response_cache: true
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.imageGenModel == "openai/gpt-image-1")
        #expect(c.openrouterResponseCacheEnabled == true)
    }

    @Test func openRouterResponseCacheScalarFalseDecodes() {
        // The scalar `false` round-trips honestly (the v0.16 bug was that
        // a disable wrote a nested dict that Hermes read as truthy).
        let c = HermesConfig(yaml: """
        openrouter:
          response_cache: false
        """)
        #expect(c.openrouterResponseCacheEnabled == false)
    }

    @Test func openRouterResponseCacheLegacyNestedDecodesToFalse() {
        // Defensive read: a legacy nested value flattens to a different
        // dotted key, so the scalar lookup misses and we fall to the
        // `false` default. The next save writes the scalar, healing it.
        let c = HermesConfig(yaml: """
        openrouter:
          response_cache:
            enabled: true
        """)
        #expect(c.openrouterResponseCacheEnabled == false)
    }

    @Test func parsesBitwardenSecretsBlock() {
        // WS-F (v0.15): round-trip the `secrets.bitwarden.*` block. Pins
        // the parser line + setter key shapes to a single source of truth.
        let yaml = """
        secrets:
          bitwarden:
            enabled: true
            access_token_env: MY_BWS_TOKEN
            project_id: proj-123
            override_existing: true
            server_url: https://vault.bitwarden.eu
            cache_ttl_seconds: 600
            auto_install: false
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.bitwarden.enabled == true)
        #expect(c.bitwarden.accessTokenEnv == "MY_BWS_TOKEN")
        #expect(c.bitwarden.projectID == "proj-123")
        #expect(c.bitwarden.overrideExisting == true)
        #expect(c.bitwarden.serverURL == "https://vault.bitwarden.eu")
        #expect(c.bitwarden.cacheTTLSeconds == 600)
        #expect(c.bitwarden.autoInstall == false)
    }

    @Test func bitwardenAbsentYieldsDefaults() {
        // An absent block must produce the v0.15 server-side defaults so a
        // pre-v0.15 host looks identical to a freshly-installed one.
        let c = HermesConfig(yaml: "")
        #expect(c.bitwarden.enabled == false)
        #expect(c.bitwarden.accessTokenEnv == "BWS_ACCESS_TOKEN")
        #expect(c.bitwarden.projectID == "")
        #expect(c.bitwarden.overrideExisting == false)
        #expect(c.bitwarden.serverURL == "")
        #expect(c.bitwarden.cacheTTLSeconds == 300)
        #expect(c.bitwarden.autoInstall == true)
    }

    @Test func parsesTopLevelModel() {
        let yaml = """
        model:
          default: claude-4-opus
          provider: anthropic
        agent:
          reasoning_effort: high
          service_tier: pro
          max_turns: 50
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.model == "claude-4-opus")
        #expect(c.provider == "anthropic")
        #expect(c.reasoningEffort == "high")
        #expect(c.serviceTier == "pro")
        #expect(c.maxTurns == 50)
    }

    @Test func parsesDisplaySection() {
        let yaml = """
        display:
          skin: dark
          compact: true
          streaming: false
          show_reasoning: true
          show_cost: true
          personality: professional
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.display.skin == "dark")
        #expect(c.display.compact == true)
        #expect(c.streaming == false)
        #expect(c.showReasoning == true)
        #expect(c.showCost == true)
        #expect(c.personality == "professional")
    }

    @Test func parsesSecuritySection() {
        let yaml = """
        security:
          redact_secrets: false
          tirith_enabled: false
          tirith_timeout: 15
          website_blocklist:
            enabled: true
            domains:
              - example.com
              - evil.org
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.security.redactSecrets == false)
        #expect(c.security.tirithEnabled == false)
        #expect(c.security.tirithTimeout == 15)
        #expect(c.security.blocklistEnabled == true)
        #expect(c.security.blocklistDomains == ["example.com", "evil.org"])
    }

    @Test func parsesSlackWithLegacyAndNewerPaths() {
        // Newer path wins when both present.
        let newerWins = HermesConfig(yaml: """
        platforms:
          slack:
            reply_to_mode: all
        slack:
          reply_to_mode: first
        """)
        #expect(newerWins.slack.replyToMode == "all")

        // Legacy-only path used when newer is absent.
        let legacyFallback = HermesConfig(yaml: """
        slack:
          reply_to_mode: first
        """)
        #expect(legacyFallback.slack.replyToMode == "first")

        // Default when neither present.
        let defaulted = HermesConfig(yaml: "")
        #expect(defaulted.slack.replyToMode == "first")
    }

    @Test func parsesAuxiliarySection() {
        let yaml = """
        auxiliary:
          vision:
            provider: openai
            model: gpt-4-vision
            timeout: 60
          compression:
            provider: anthropic
            model: claude-3-haiku
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.auxiliary.vision.provider == "openai")
        #expect(c.auxiliary.vision.model == "gpt-4-vision")
        #expect(c.auxiliary.vision.timeout == 60)
        #expect(c.auxiliary.compression.provider == "anthropic")
        // Not-configured aux blocks default to "auto" / empty.
        #expect(c.auxiliary.sessionSearch.provider == "auto")
        #expect(c.auxiliary.mcp.provider == "auto")
    }

    @Test func parsesPermanentAllowlist() {
        let yaml = """
        permanent_allowlist:
          - ls
          - pwd
          - stat
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.commandAllowlist == ["ls", "pwd", "stat"])
    }

    @Test func parsesCommandAllowlistLegacyName() {
        // Fall back to `command_allowlist` when `permanent_allowlist` absent.
        let yaml = """
        command_allowlist:
          - whoami
          - id
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.commandAllowlist == ["whoami", "id"])
    }

    @Test func preservesQuotedStrings() {
        let yaml = """
        model:
          default: "gpt-4o with spaces"
        timezone: 'America/New_York'
        """
        let c = HermesConfig(yaml: yaml)
        #expect(c.model == "gpt-4o with spaces")
        #expect(c.timezone == "America/New_York")
    }

    // MARK: - v0.16 top-level <platform>.allowed_* allowlists

    @Test func gatewayPlatformsEmptyByDefault() {
        let c = HermesConfig(yaml: "")
        #expect(c.gatewayPlatforms.isEmpty)
    }

    @Test func parsesGatewayAllowlistsForSlack() {
        // v0.16: allowlists live at top-level `slack.allowed_*`.
        let yaml = """
        slack:
          allowed_channels:
            - C01
            - C02
          busy_ack_enabled: false
          gateway_restart_notification: true
          slash_command_notice_ttl_seconds: 120
        """
        let cfg = HermesConfig(yaml: yaml)
        let block = cfg.gatewayPlatforms["slack"]
        #expect(block?.allowedChannels == ["C01", "C02"])
        #expect(block?.busyAckEnabled == false)
        #expect(block?.gatewayRestartNotification == true)
        #expect(block?.slashCommandNoticeTTLSeconds == 120)
    }

    @Test func parsesGatewayAllowlistsForTelegramAndMatrix() {
        let yaml = """
        telegram:
          allowed_chats:
            - '@alice'
            - '12345'
        matrix:
          allowed_rooms:
            - '!room:matrix.org'
        """
        let cfg = HermesConfig(yaml: yaml)
        #expect(cfg.gatewayPlatforms["telegram"]?.allowedChats == ["@alice", "12345"])
        #expect(cfg.gatewayPlatforms["matrix"]?.allowedRooms == ["!room:matrix.org"])
    }

    @Test func parsesGatewayAllowlistForDingtalkAsChats() {
        // v0.16: Hermes reads `dingtalk.allowed_chats` (NOT allowed_rooms).
        let yaml = """
        dingtalk:
          allowed_chats:
            - cidABC123
        """
        let cfg = HermesConfig(yaml: yaml)
        #expect(cfg.gatewayPlatforms["dingtalk"]?.allowedChats == ["cidABC123"])
    }

    @Test func gatewayAllowlistCoexistsWithLegacyPlatformKeys() {
        // Regression: the legacy `slack.reply_to_mode` /
        // `matrix.require_mention` keys live in the SAME top-level section as
        // the v0.16 allowlist keys — both must keep parsing, no collisions.
        let yaml = """
        slack:
          reply_to_mode: all
          allowed_channels:
            - C01
        matrix:
          require_mention: false
          allowed_rooms:
            - '!room:matrix.org'
        """
        let cfg = HermesConfig(yaml: yaml)
        #expect(cfg.slack.replyToMode == "all")
        #expect(cfg.matrix.requireMention == false)
        #expect(cfg.gatewayPlatforms["slack"]?.allowedChannels == ["C01"])
        #expect(cfg.gatewayPlatforms["matrix"]?.allowedRooms == ["!room:matrix.org"])
    }

    @Test func gatewayPlatformsSkipsPlatformsWithoutGatewayKeys() {
        // Only Slack carries a gateway key — platforms without one must NOT
        // appear in `gatewayPlatforms`.
        let yaml = """
        slack:
          busy_ack_enabled: true
        """
        let cfg = HermesConfig(yaml: yaml)
        #expect(cfg.gatewayPlatforms["slack"] != nil)
        #expect(cfg.gatewayPlatforms["mattermost"] == nil)
        #expect(cfg.gatewayPlatforms["telegram"] == nil)
    }

    @Test func cronScheduleMemberwise() {
        let s = CronSchedule(
            kind: "cron",
            runAt: nil,
            display: "9am weekdays",
            expression: "0 9 * * 1-5"
        )
        #expect(s.kind == "cron")
        #expect(s.display == "9am weekdays")
    }

    @Test func hermesCronJobMemberwiseAndWithEnabled() {
        let job = HermesCronJob(
            id: "j1",
            name: "Brief",
            prompt: "summarize",
            skills: ["cal"],
            schedule: CronSchedule(kind: "cron"),
            enabled: true,
            state: "scheduled",
            deliver: "discord:general"
        )
        #expect(job.enabled)
        let toggled = job.withEnabled(false)
        #expect(toggled.enabled == false)
        // Every other field round-trips.
        #expect(toggled.id == job.id)
        #expect(toggled.name == job.name)
        #expect(toggled.prompt == job.prompt)
        #expect(toggled.skills == job.skills)
        #expect(toggled.deliver == job.deliver)
    }

    @Test func cronJobsFileMemberwise() {
        let jobs = [
            HermesCronJob(
                id: "a", name: "A", prompt: "p",
                schedule: CronSchedule(kind: "cron"),
                enabled: true, state: "scheduled"
            )
        ]
        let file = CronJobsFile(jobs: jobs, updatedAt: "2026-04-23T00:00:00Z")
        #expect(file.jobs.count == 1)
        #expect(file.updatedAt == "2026-04-23T00:00:00Z")
        // Codable round-trip should survive.
        let data = try! JSONEncoder().encode(file)
        let decoded = try! JSONDecoder().decode(CronJobsFile.self, from: data)
        #expect(decoded.jobs.count == 1)
        #expect(decoded.jobs[0].name == "A")
        #expect(decoded.updatedAt == file.updatedAt)
    }
}
