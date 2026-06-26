import CryptoKit
import Darwin
import Foundation

struct AppPaths {
    let fileManager: FileManager
    let applicationSupportURL: URL
    let connectionsURL: URL
    let preferencesURL: URL
    let appearanceAssetsURL: URL
    let controlSocketDirectoryURL: URL

    private static let privateDirectoryPermissions = NSNumber(value: Int16(0o700))

    init(fileManager: FileManager = .default) {
        let baseSupport = fileManager.urls(
            for: .applicationSupportDirectory,
            in: .userDomainMask
        ).first ?? URL(fileURLWithPath: NSTemporaryDirectory(), isDirectory: true)
        self.init(
            fileManager: fileManager,
            applicationSupportURL: baseSupport.appendingPathComponent("HermesDesktop", isDirectory: true),
            controlSocketDirectoryURL: URL(
                fileURLWithPath: "/tmp/hd-\(getuid())",
                isDirectory: true
            )
        )
    }

    init(
        fileManager: FileManager = .default,
        applicationSupportURL: URL,
        controlSocketDirectoryURL: URL
    ) {
        self.fileManager = fileManager
        self.applicationSupportURL = applicationSupportURL
        self.connectionsURL = applicationSupportURL.appendingPathComponent("connections.json")
        self.preferencesURL = applicationSupportURL.appendingPathComponent("preferences.json")
        self.appearanceAssetsURL = applicationSupportURL.appendingPathComponent("Appearance", isDirectory: true)
        self.controlSocketDirectoryURL = controlSocketDirectoryURL

        ensureApplicationSupportDirectory()
        ensureControlSocketDirectory()
    }

    func ensureApplicationSupportDirectory() {
        createPrivateDirectoryIfNeeded(at: applicationSupportURL)
    }

    func ensureControlSocketDirectory() {
        createPrivateDirectoryIfNeeded(at: controlSocketDirectoryURL)
    }

    func ensureAppearanceAssetsDirectory() {
        createPrivateDirectoryIfNeeded(at: appearanceAssetsURL)
    }

    func appearanceBackgroundImageURL(fileName: String) -> URL {
        appearanceAssetsURL.appendingPathComponent(fileName, isDirectory: false)
    }

    func controlPath(for connection: ConnectionProfile) -> String {
        ensureControlSocketDirectory()

        return controlSocketDirectoryURL
            .appendingPathComponent(controlSocketIdentifier(for: connection))
            .path
    }

    private func createPrivateDirectoryIfNeeded(at url: URL) {
        let attributes: [FileAttributeKey: Any] = [
            .posixPermissions: Self.privateDirectoryPermissions
        ]

        if !fileManager.fileExists(atPath: url.path) {
            try? fileManager.createDirectory(at: url, withIntermediateDirectories: true, attributes: attributes)
        } else {
            var isDirectory: ObjCBool = false
            if fileManager.fileExists(atPath: url.path, isDirectory: &isDirectory),
               !isDirectory.boolValue {
                return
            }
        }

        try? fileManager.setAttributes(attributes, ofItemAtPath: url.path)
    }

    private func controlSocketIdentifier(for connection: ConnectionProfile) -> String {
        // Scope SSH control sockets to the workspace so profiles on the same host stay isolated.
        let digest = SHA256.hash(data: Data(connection.workspaceScopeFingerprint.utf8))
        let hexDigest = digest.map { String(format: "%02x", $0) }.joined()
        return String(hexDigest.prefix(24))
    }
}
