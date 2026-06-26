import Foundation
import Testing

@testable import HermesDesktop

@MainActor
struct AppStateUpdateCheckTests {
    @Test
    func launchUpdateCheckPersistsTimestampAfterSuccessfulCheck() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let service = UpdateCheckService(
            fetch: { _ in
                let payload = try JSONSerialization.data(withJSONObject: [
                    "tag_name": "v9.9.9",
                    "html_url": "https://github.com/dodo-reach/hermes-desktop/releases/tag/v9.9.9"
                ])
                return HTTPResult(statusCode: 200, data: payload)
            }
        )
        let appState = AppState(
            paths: makeTestAppPaths(root: root),
            updateCheckService: service
        )

        #expect(appState.connectionStore.lastAutomaticUpdateCheckAt == nil)

        await appState.checkForUpdatesAtLaunch()

        #expect(appState.connectionStore.lastAutomaticUpdateCheckAt != nil)
        #expect(appState.availableUpdate?.latestVersion == "9.9.9")
    }

    @Test
    func launchUpdateCheckDoesNotPersistTimestampAfterFailure() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let service = UpdateCheckService(
            fetch: { _ in
                HTTPResult(statusCode: 503, data: Data())
            }
        )
        let appState = AppState(
            paths: makeTestAppPaths(root: root),
            updateCheckService: service
        )

        #expect(appState.connectionStore.lastAutomaticUpdateCheckAt == nil)

        await appState.checkForUpdatesAtLaunch()

        #expect(appState.connectionStore.lastAutomaticUpdateCheckAt == nil)
        #expect(appState.availableUpdate == nil)
    }
}
