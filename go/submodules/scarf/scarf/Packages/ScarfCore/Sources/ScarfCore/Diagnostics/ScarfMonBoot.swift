import Foundation

/// Boot-time wiring for ScarfMon. Both app targets call
/// `ScarfMonBoot.configure(...)` at launch and again whenever the user
/// flips the Diagnostics → Performance toggle.
///
/// Three modes:
/// - `.off` — nothing is recorded. Hot path is one branch + return.
/// - `.signpostOnly` — Instruments-only. Default in the open-source build.
///   Free outside an Instruments session.
/// - `.full` — signpost + ring buffer + os.Logger debug stream. Drives the
///   in-app panel and the "Copy as JSON" button. Opt-in.
public enum ScarfMonBoot {
    public enum Mode: String, Sendable, CaseIterable {
        case off
        case signpostOnly
        case full
    }

    /// User-defaults key for the persisted toggle. Same key on iOS + Mac
    /// so `defaults read com.scarf.app ScarfMonMode` works on either.
    public static let userDefaultsKey = "ScarfMonMode"

    /// Read the persisted mode, defaulting to `.signpostOnly` so users
    /// always get Instruments-visible signposts unless they explicitly
    /// turn them off.
    public static func currentMode(_ defaults: UserDefaults = .standard) -> Mode {
        if let raw = defaults.string(forKey: userDefaultsKey),
           let mode = Mode(rawValue: raw) {
            return mode
        }
        return .signpostOnly
    }

    /// Persist a new mode and reinstall the backend set.
    public static func setMode(_ mode: Mode, _ defaults: UserDefaults = .standard) {
        defaults.set(mode.rawValue, forKey: userDefaultsKey)
        configure(mode: mode)
    }

    /// Install the backend set for a given mode. Returns the active ring
    /// buffer (if any) so the in-app Diagnostics panel can read from it.
    @discardableResult
    public static func configure(mode: Mode) -> ScarfMonRingBuffer? {
        switch mode {
        case .off:
            ScarfMon.install([])
            sharedRingBuffer = nil
            return nil
        case .signpostOnly:
            ScarfMon.install([ScarfMonSignpostBackend()])
            sharedRingBuffer = nil
            return nil
        case .full:
            let ring = ScarfMonRingBuffer()
            sharedRingBuffer = ring
            ScarfMon.install([
                ScarfMonSignpostBackend(),
                ring,
                ScarfMonLoggerBackend()
            ])
            return ring
        }
    }

    /// Process-wide ring buffer when running in `.full` mode. Nil otherwise.
    /// Read by the Diagnostics panel; writes happen through the backend
    /// dispatcher so this property is read-only.
    ///
    /// `nonisolated(unsafe)` because the value is only mutated by
    /// `configure(...)` (which itself runs on whichever actor invokes
    /// the boot helper at app launch — single-writer in practice) and
    /// read from the panel UI on the main actor. Adding a lock here
    /// would just add overhead with no real safety win.
    nonisolated(unsafe) public private(set) static var sharedRingBuffer: ScarfMonRingBuffer?
}
