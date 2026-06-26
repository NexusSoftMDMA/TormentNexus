import Foundation
import ScarfCore

/// `UserDefaults`-backed `IOSProfileSelectionStore` (issue #120, Design B).
///
/// The selection is not sensitive (SSH keys live in the Keychain), so
/// `UserDefaults` is the right home — same low-ceremony call as
/// `UserDefaultsIOSServerConfigStore`. The protocol and the in-memory
/// implementation live in ScarfCore.
///
/// Data shape: JSON `[ServerID.uuidString: String]` under
/// `com.scarf.ios.profile-selections.v1`. An absent key, an absent entry,
/// or an entry that fails normalization all read back as `nil` (default
/// profile). Writing `nil`/`"default"`/an invalid name removes the entry.
///
/// **Threading.** `setSelectedProfile` is a read-modify-write over a single
/// JSON blob, which is NOT atomic across concurrent writers. Production
/// drives this only through the `@MainActor` `ScarfGoCoordinator`, so writes
/// are serialized; if a future caller writes off the main actor, add
/// synchronization here.
public struct UserDefaultsProfileSelectionStore: IOSProfileSelectionStore {
    public static let defaultDefaultsKey = "com.scarf.ios.profile-selections.v1"

    private let defaults: UserDefaults
    private let key: String

    public init(
        defaults: UserDefaults = .standard,
        key: String = defaultDefaultsKey
    ) {
        self.defaults = defaults
        self.key = key
    }

    public func selectedProfile(for id: ServerID) -> String? {
        HermesProfileScope.normalize(read()[id.uuidString])
    }

    public func setSelectedProfile(_ name: String?, for id: ServerID) {
        var all = read()
        if let normalized = HermesProfileScope.normalize(name) {
            all[id.uuidString] = normalized
        } else {
            all.removeValue(forKey: id.uuidString)
        }
        write(all)
    }

    private func read() -> [String: String] {
        guard let data = defaults.data(forKey: key),
              let raw = try? JSONDecoder().decode([String: String].self, from: data)
        else { return [:] }
        return raw
    }

    private func write(_ all: [String: String]) {
        guard !all.isEmpty else {
            defaults.removeObject(forKey: key)
            return
        }
        if let data = try? JSONEncoder().encode(all) {
            defaults.set(data, forKey: key)
        }
    }
}
