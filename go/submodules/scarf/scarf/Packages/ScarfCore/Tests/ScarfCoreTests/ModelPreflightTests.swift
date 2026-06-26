import Testing
import Foundation
@testable import ScarfCore

/// Pure tests for `ModelPreflight` — both the `check(_:)` configured-vs-
/// missing classifier and the v2.8 `detectMismatch(_:)` provider/prefix
/// reconciliation. The mismatch path is what surfaces the orange
/// "Model/provider mismatch in config.yaml" banner in ChatView when the
/// user switches OAuth providers via Credential Pools and `model.default`
/// is left carrying the old provider's prefix.
@Suite struct ModelPreflightTests {

    // MARK: - check(_:) — missing-field classifier

    @Test func bothModelAndProviderEmptyReportsMissingBoth() {
        var cfg = HermesConfig.empty
        cfg.model = ""
        cfg.provider = ""
        #expect(ModelPreflight.check(cfg) == .missingBoth)
    }

    @Test func bothModelAndProviderUnknownReportsMissingBoth() {
        // `HermesConfig.empty` defaults model/provider to the literal
        // "unknown" — the classifier must treat that the same as "".
        let cfg = HermesConfig.empty
        #expect(ModelPreflight.check(cfg) == .missingBoth)
    }

    @Test func providerSetButModelEmptyReportsMissingModel() {
        var cfg = HermesConfig.empty
        cfg.model = ""
        cfg.provider = "anthropic"
        #expect(ModelPreflight.check(cfg) == .missingModel)
    }

    @Test func modelSetButProviderEmptyReportsMissingProvider() {
        var cfg = HermesConfig.empty
        cfg.model = "claude-sonnet-4.6"
        cfg.provider = ""
        #expect(ModelPreflight.check(cfg) == .missingProvider)
    }

    @Test func bothSetReportsConfigured() {
        var cfg = HermesConfig.empty
        cfg.model = "claude-sonnet-4.6"
        cfg.provider = "anthropic"
        #expect(ModelPreflight.check(cfg) == .configured)
    }

    @Test func whitespaceTreatedAsUnsetForBothFields() {
        var cfg = HermesConfig.empty
        cfg.model = "  "
        cfg.provider = "\n"
        #expect(ModelPreflight.check(cfg) == .missingBoth)
    }

    @Test func resultIsConfiguredOnlyForConfiguredCase() {
        #expect(ModelPreflight.Result.configured.isConfigured)
        #expect(!ModelPreflight.Result.missingBoth.isConfigured)
        #expect(!ModelPreflight.Result.missingModel.isConfigured)
        #expect(!ModelPreflight.Result.missingProvider.isConfigured)
    }

    // MARK: - detectMismatch(_:)

    @Test func detectMismatchReturnsNilWhenNoPrefixOnModelDefault() {
        var cfg = HermesConfig.empty
        cfg.model = "claude-sonnet-4.6"
        cfg.provider = "anthropic"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchReturnsNilWhenPrefixMatchesProvider() {
        var cfg = HermesConfig.empty
        cfg.model = "anthropic/claude-sonnet-4.6"
        cfg.provider = "anthropic"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchReturnsNilWhenModelDefaultIsUnset() {
        var cfg = HermesConfig.empty
        cfg.model = ""
        cfg.provider = "nous"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchReturnsNilWhenProviderIsUnset() {
        var cfg = HermesConfig.empty
        cfg.model = "anthropic/claude-sonnet-4.6"
        cfg.provider = ""
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchReturnsNilWhenBothUnknown() {
        // The literal "unknown" sentinel from the YAML parser fallback
        // counts as unset on both sides — no mismatch to report.
        let cfg = HermesConfig.empty // model + provider both "unknown"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchSurfacesPrefixVsActiveProvider() {
        // The dogfooding scenario: Anthropic-prefixed model still sitting
        // in config.yaml after the user OAuth'd into Nous via Credential
        // Pools. Hermes can't reconcile and chats die with -32603 at
        // first prompt. The banner offers a one-click fix in either
        // direction; this test pins the data the banner reads.
        var cfg = HermesConfig.empty
        cfg.model = "anthropic/claude-sonnet-4.6"
        cfg.provider = "nous"
        let mismatch = ModelPreflight.detectMismatch(cfg)
        #expect(mismatch != nil)
        #expect(mismatch?.prefixProvider == "anthropic")
        #expect(mismatch?.activeProvider == "nous")
        #expect(mismatch?.modelDefault == "anthropic/claude-sonnet-4.6")
        #expect(mismatch?.bareModel == "claude-sonnet-4.6")
    }

    @Test func detectMismatchIsCaseInsensitiveOnPrefixMatch() {
        // Hermes accepts both `Anthropic/...` and `anthropic/...` casings
        // in the wild — case-only differences must NOT surface as a
        // mismatch (would be a false-positive banner).
        var cfg = HermesConfig.empty
        cfg.model = "Anthropic/claude-sonnet-4.6"
        cfg.provider = "anthropic"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchHandlesNonAnthropicProviders() {
        // The mismatch banner needs to work for any provider pair —
        // not just the dogfooding case. Pin the openai+nous shape.
        var cfg = HermesConfig.empty
        cfg.model = "openai/gpt-5"
        cfg.provider = "nous"
        let mismatch = ModelPreflight.detectMismatch(cfg)
        #expect(mismatch?.prefixProvider == "openai")
        #expect(mismatch?.activeProvider == "nous")
        #expect(mismatch?.bareModel == "gpt-5")
    }

    @Test func detectMismatchReturnsNilForEmptyBareModel() {
        // A pathological "anthropic/" with no model name after the
        // slash isn't a valid mismatch — caller has no bare model to
        // write back. The classifier should refuse to surface it
        // rather than emit a useless fix button.
        var cfg = HermesConfig.empty
        cfg.model = "anthropic/"
        cfg.provider = "nous"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchReturnsNilForEmptyPrefix() {
        // Symmetric pathological case — leading slash, no provider
        // prefix. Don't fire.
        var cfg = HermesConfig.empty
        cfg.model = "/claude-sonnet-4.6"
        cfg.provider = "nous"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }

    @Test func detectMismatchHandlesModelsWithMultipleSlashes() {
        // Some provider/model strings carry path-style segments after
        // the first slash (e.g. an OpenRouter style path). The first
        // slash separates prefix from bare model; the rest of the
        // string is the bare model verbatim.
        var cfg = HermesConfig.empty
        cfg.model = "openrouter/anthropic/claude-sonnet-4.6"
        cfg.provider = "anthropic"
        let mismatch = ModelPreflight.detectMismatch(cfg)
        #expect(mismatch?.prefixProvider == "openrouter")
        #expect(mismatch?.activeProvider == "anthropic")
        #expect(mismatch?.bareModel == "anthropic/claude-sonnet-4.6")
    }

    @Test func detectMismatchTrimsWhitespaceBeforeComparing() {
        // A stray newline in a hand-edited config.yaml shouldn't read
        // as a mismatch when the trimmed values agree.
        var cfg = HermesConfig.empty
        cfg.model = "anthropic/claude-sonnet-4.6  "
        cfg.provider = " anthropic\n"
        #expect(ModelPreflight.detectMismatch(cfg) == nil)
    }
}
