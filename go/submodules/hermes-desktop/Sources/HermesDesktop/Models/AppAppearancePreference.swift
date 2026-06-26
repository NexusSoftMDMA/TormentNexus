import AppKit
import SwiftUI

enum AppAppearancePreference: String, CaseIterable, Codable, Identifiable {
    case system
    case dark
    case light

    var id: String {
        rawValue
    }

    var title: String {
        switch self {
        case .system:
            return "System"
        case .dark:
            return "Dark"
        case .light:
            return "Light"
        }
    }

    var colorScheme: ColorScheme? {
        switch self {
        case .system:
            return nil
        case .dark:
            return .dark
        case .light:
            return .light
        }
    }

    var nsAppearance: NSAppearance? {
        switch self {
        case .system:
            return nil
        case .dark:
            return NSAppearance(named: .darkAqua)
        case .light:
            return NSAppearance(named: .aqua)
        }
    }

    @MainActor
    func applyToApplication() {
        let appearance = nsAppearance
        NSApplication.shared.appearance = appearance
        for window in NSApplication.shared.windows {
            window.appearance = appearance
            window.viewsNeedDisplay = true
        }
    }
}

enum TerminalFontPreference {
    static let defaultSize: Double = 13
    static let minimumSize: Double = 10
    static let maximumSize: Double = 20

    static func clamped(_ value: Double) -> Double {
        min(max(value, minimumSize), maximumSize)
    }
}

enum TerminalFontFamilyPreference: String, CaseIterable, Codable, Identifiable {
    case systemMonospaced
    case sfMono
    case menlo
    case monaco
    case courier
    case courierNew
    case andaleMono
    case sourceCodePro
    case jetBrainsMono
    case firaCode
    case cascadiaCode
    case hack
    case iosevka

    var id: String {
        rawValue
    }

    var title: String {
        switch self {
        case .systemMonospaced:
            return "System Mono"
        case .sfMono:
            return "SF Mono"
        case .menlo:
            return "Menlo"
        case .monaco:
            return "Monaco"
        case .courier:
            return "Courier"
        case .courierNew:
            return "Courier New"
        case .andaleMono:
            return "Andale Mono"
        case .sourceCodePro:
            return "Source Code Pro"
        case .jetBrainsMono:
            return "JetBrains Mono"
        case .firaCode:
            return "Fira Code"
        case .cascadiaCode:
            return "Cascadia Code"
        case .hack:
            return "Hack"
        case .iosevka:
            return "Iosevka"
        }
    }

    func font(size: Double) -> NSFont {
        let clampedSize = CGFloat(TerminalFontPreference.clamped(size))
        switch self {
        case .systemMonospaced:
            return NSFont.monospacedSystemFont(ofSize: clampedSize, weight: .regular)
        case .sfMono:
            return Self.font(named: ["SFMono-Regular", "SF Mono"], size: clampedSize)
        case .menlo:
            return Self.font(named: ["Menlo-Regular", "Menlo"], size: clampedSize)
        case .monaco:
            return Self.font(named: ["Monaco"], size: clampedSize)
        case .courier:
            return Self.font(named: ["Courier", "CourierNewPSMT"], size: clampedSize)
        case .courierNew:
            return Self.font(named: ["CourierNewPSMT", "Courier New"], size: clampedSize)
        case .andaleMono:
            return Self.font(named: ["AndaleMono", "Andale Mono"], size: clampedSize)
        case .sourceCodePro:
            return Self.font(named: ["SourceCodePro-Regular", "Source Code Pro"], size: clampedSize)
        case .jetBrainsMono:
            return Self.font(named: ["JetBrainsMono-Regular", "JetBrains Mono"], size: clampedSize)
        case .firaCode:
            return Self.font(named: ["FiraCode-Regular", "Fira Code"], size: clampedSize)
        case .cascadiaCode:
            return Self.font(named: ["CascadiaCode-Regular", "Cascadia Code"], size: clampedSize)
        case .hack:
            return Self.font(named: ["Hack-Regular", "Hack"], size: clampedSize)
        case .iosevka:
            return Self.font(named: ["Iosevka-Regular", "Iosevka"], size: clampedSize)
        }
    }

    private static func font(named names: [String], size: CGFloat) -> NSFont {
        for name in names {
            if let font = NSFont(name: name, size: size) {
                return font
            }
        }
        return NSFont.monospacedSystemFont(ofSize: size, weight: .regular)
    }
}

enum AppWindowOpacityPreference {
    static let defaultValue: Double = 1.0
    static let minimumValue: Double = 0.58
    static let maximumValue: Double = 1.0

    static func clamped(_ value: Double) -> Double {
        min(max(value, minimumValue), maximumValue)
    }
}

enum AppWindowMaterialPreference: String, CaseIterable, Codable, Identifiable {
    case solid
    case nativeWindow
    case translucent

    static var allCases: [AppWindowMaterialPreference] {
        [.solid, .translucent]
    }

    var id: String {
        rawValue
    }

    var title: String {
        switch self {
        case .solid:
            return "Solid"
        case .nativeWindow:
            return "Native Window"
        case .translucent:
            return "Translucent"
        }
    }
}

enum AppBackgroundImageFitPreference: String, CaseIterable, Codable, Identifiable {
    case fill
    case fit

    var id: String {
        rawValue
    }

    var title: String {
        switch self {
        case .fill:
            return "Fill"
        case .fit:
            return "Fit"
        }
    }
}

enum AppBackgroundImageBlurPreference {
    static let defaultValue: Double = 0
    static let minimumValue: Double = 0
    static let maximumValue: Double = 24

    static func clamped(_ value: Double) -> Double {
        min(max(value, minimumValue), maximumValue)
    }
}

struct HiddenHermesProfilePreference: Codable, Hashable, Identifiable {
    let hostConnectionFingerprint: String
    let profileName: String

    var id: String {
        "\(hostConnectionFingerprint)|\(profileName)"
    }
}
