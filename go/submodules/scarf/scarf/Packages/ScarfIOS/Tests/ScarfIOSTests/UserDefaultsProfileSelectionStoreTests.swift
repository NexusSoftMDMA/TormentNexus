import Testing
import Foundation
import ScarfCore
@testable import ScarfIOS

/// Real `UserDefaults` round-trip coverage for the ScarfGo profile
/// selection store (issue #120, Design B). The protocol behavior is
/// covered Linux-side via `InMemoryProfileSelectionStore` in
/// `HermesProfileScopeTests`; this guards the encode/decode/key wiring of
/// the persistent implementation, which only builds on Apple targets.
@Suite struct UserDefaultsProfileSelectionStoreTests {

    /// A throwaway `UserDefaults` suite so tests never touch standard
    /// defaults and never collide with each other.
    private func makeDefaults() -> UserDefaults {
        let suite = "test.profile-selection.\(UUID().uuidString)"
        return UserDefaults(suiteName: suite)!
    }

    @Test func absentSelectionIsDefault() {
        let store = UserDefaultsProfileSelectionStore(defaults: makeDefaults())
        #expect(store.selectedProfile(for: ServerID()) == nil)
    }

    @Test func roundTripAndIsolation() {
        let store = UserDefaultsProfileSelectionStore(defaults: makeDefaults())
        let a = ServerID()
        let b = ServerID()
        store.setSelectedProfile("admin", for: a)
        store.setSelectedProfile("gateway", for: b)
        #expect(store.selectedProfile(for: a) == "admin")
        #expect(store.selectedProfile(for: b) == "gateway")
    }

    @Test func clearingResetsToDefault() {
        let store = UserDefaultsProfileSelectionStore(defaults: makeDefaults())
        let id = ServerID()
        store.setSelectedProfile("gateway", for: id)
        store.setSelectedProfile(nil, for: id)
        #expect(store.selectedProfile(for: id) == nil)
    }

    @Test func defaultAndInvalidNamesNormalizeToDefault() {
        let store = UserDefaultsProfileSelectionStore(defaults: makeDefaults())
        let id = ServerID()
        store.setSelectedProfile("default", for: id)
        #expect(store.selectedProfile(for: id) == nil)
        store.setSelectedProfile("Bad Name", for: id)
        #expect(store.selectedProfile(for: id) == nil)
    }

    /// The persistence guarantee: a fresh store instance backed by the
    /// same defaults reads what a prior instance wrote (survives app
    /// relaunch).
    @Test func persistsAcrossStoreInstances() {
        let defaults = makeDefaults()
        let id = ServerID()
        UserDefaultsProfileSelectionStore(defaults: defaults)
            .setSelectedProfile("gateway", for: id)
        #expect(UserDefaultsProfileSelectionStore(defaults: defaults)
            .selectedProfile(for: id) == "gateway")
    }

    /// Clearing the last entry removes the backing key entirely rather
    /// than leaving an empty blob behind.
    @Test func emptyMapRemovesBackingKey() {
        let defaults = makeDefaults()
        let id = ServerID()
        let store = UserDefaultsProfileSelectionStore(defaults: defaults)
        store.setSelectedProfile("gateway", for: id)
        #expect(defaults.data(forKey: UserDefaultsProfileSelectionStore.defaultDefaultsKey) != nil)
        store.setSelectedProfile(nil, for: id)
        #expect(defaults.data(forKey: UserDefaultsProfileSelectionStore.defaultDefaultsKey) == nil)
    }
}
