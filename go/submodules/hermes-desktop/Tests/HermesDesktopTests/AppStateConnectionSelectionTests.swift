import Foundation
import Testing

@testable import HermesDesktop

@MainActor
struct AppStateConnectionSelectionTests {
    @Test
    func launchSelectsSavedHostWhenPreferenceIsMissing() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let paths = makeTestAppPaths(root: root)
        let savedConnection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com"
        )
        let store = ConnectionStore(paths: paths)
        store.upsert(savedConnection)

        let appState = AppState(paths: paths)

        #expect(appState.activeConnectionID == savedConnection.id)
        #expect(appState.activeConnection?.id == savedConnection.id)
        #expect(appState.connectionStore.lastConnectionID == savedConnection.id)
    }

    @Test
    func launchReplacesStalePreferenceWithSavedHost() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let paths = makeTestAppPaths(root: root)
        let savedConnection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com"
        )
        let store = ConnectionStore(paths: paths)
        store.upsert(savedConnection)
        store.lastConnectionID = UUID()

        let appState = AppState(paths: paths)

        #expect(appState.activeConnectionID == savedConnection.id)
        #expect(appState.activeConnection?.id == savedConnection.id)
        #expect(appState.connectionStore.lastConnectionID == savedConnection.id)
    }

    @Test
    func savingFirstHostMakesItActive() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let appState = AppState(paths: makeTestAppPaths(root: root))
        let savedConnection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com"
        )

        appState.saveConnection(savedConnection)

        #expect(appState.activeConnectionID == savedConnection.id)
        #expect(appState.activeConnection?.id == savedConnection.id)
        #expect(appState.connectionStore.lastConnectionID == savedConnection.id)
    }

    @Test
    func deletingActiveHostSelectsRemainingHost() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let appState = AppState(paths: makeTestAppPaths(root: root))
        let firstConnection = ConnectionProfile(
            label: "Alpha",
            sshHost: "alpha.example.com"
        )
        let secondConnection = ConnectionProfile(
            label: "Beta",
            sshHost: "beta.example.com"
        )
        appState.saveConnection(firstConnection)
        appState.saveConnection(secondConnection)

        appState.deleteConnection(firstConnection)

        #expect(appState.activeConnectionID == secondConnection.id)
        #expect(appState.activeConnection?.id == secondConnection.id)
        #expect(appState.connectionStore.lastConnectionID == secondConnection.id)
    }
}
