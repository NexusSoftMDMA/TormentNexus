import Testing
@testable import HermesDesktop

struct TerminalThemeTests {
    @Test
    func presetSelectionProducesStablePresetAppearance() {
        let preference = TerminalThemePreference().selectingPreset(.dusk)
        let appearance = preference.resolvedAppearance

        #expect(appearance.style == .dusk)
        #expect(appearance.name == "Dusk")
        #expect(appearance.paletteStyle == .dusk)
        #expect(!appearance.isCustom)
        #expect(appearance.ansiPalette.count == 16)
    }

    @Test
    func quickPresetsExposeSixStableChoices() {
        let presetStyles = TerminalThemePreference.quickPresets.map(\.style)
        let presetNames = TerminalThemePreference.quickPresets.map(\.name)

        #expect(presetStyles == [.graphite, .evergreen, .dusk, .paper, .harbor, .ember])
        #expect(presetNames == ["Graphite", "Evergreen", "Dusk", "Paper", "Aubergine", "Porcelain"])
        #expect(TerminalThemePreference.quickPresets.allSatisfy { $0.ansiPalette.count == 16 })
    }

    @Test
    func hexColorRoundTripsForInlineEditor() throws {
        let color = try #require(TerminalThemeColor(hexString: "#0B171A"))

        #expect(color == TerminalThemeColor(hex: 0x0B171A))
        #expect(color.hexString == "#0B171A")
        #expect(TerminalThemeColor(hexString: "not-a-color") == nil)
    }

    @Test
    func hsbColorSupportsGeneratedMatrixPalette() {
        let color = TerminalThemeColor(hue: 0.72, saturation: 0.64, brightness: 0.88)
        let hsb = color.hsb

        #expect(abs(hsb.hue - 0.72) < 0.01)
        #expect(abs(hsb.saturation - 0.64) < 0.01)
        #expect(abs(hsb.brightness - 0.88) < 0.01)
    }

    @Test
    func customBackgroundPreservesPresetPaletteAndForeground() {
        let customBackground = TerminalThemeColor(hex: 0x010203)
        let preference = TerminalThemePreference(style: .evergreen)
            .updatingBackgroundColor(customBackground)
        let appearance = preference.resolvedAppearance

        #expect(preference.style == .custom)
        #expect(appearance.style == .custom)
        #expect(appearance.isCustom)
        #expect(appearance.backgroundColor == customBackground)
        #expect(appearance.paletteStyle == .evergreen)
        #expect(appearance.foregroundColor == TerminalThemePreference(style: .evergreen).resolvedAppearance.foregroundColor)
        #expect(appearance.ansiPalette.count == 16)
    }

    @Test
    func customForegroundKeepsExistingCustomBackground() {
        let customBackground = TerminalThemeColor(hex: 0x112233)
        let customForeground = TerminalThemeColor(hex: 0xF0E0D0)
        let preference = TerminalThemePreference(style: .paper)
            .updatingBackgroundColor(customBackground)
            .updatingForegroundColor(customForeground)
        let appearance = preference.resolvedAppearance

        #expect(preference.style == .custom)
        #expect(appearance.backgroundColor == customBackground)
        #expect(appearance.foregroundColor == customForeground)
        #expect(appearance.paletteStyle == .paper)
        #expect(appearance.isCustom)
    }

    @Test
    func customAccentPaletteCanChangeWithoutLosingCustomColors() {
        let customBackground = TerminalThemeColor(hex: 0x112233)
        let customForeground = TerminalThemeColor(hex: 0xF0E0D0)
        let preference = TerminalThemePreference(style: .paper)
            .updatingBackgroundColor(customBackground)
            .updatingForegroundColor(customForeground)
            .updatingPaletteStyle(.harbor)
        let appearance = preference.resolvedAppearance

        #expect(preference.style == .custom)
        #expect(appearance.backgroundColor == customBackground)
        #expect(appearance.foregroundColor == customForeground)
        #expect(appearance.paletteStyle == .harbor)
        #expect(appearance.ansiPalette == TerminalThemePreference(style: .harbor).resolvedAppearance.ansiPalette)
    }
}
