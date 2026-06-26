import Foundation
import os

public struct ProjectDashboardService: Sendable {
    private static let logger = Logger(subsystem: "com.scarf", category: "ProjectDashboardService")

    /// Size ceiling for JSON files we read off SFTP/disk. Both the
    /// registry and individual dashboards are tens-of-KB even on
    /// heavy multi-project setups (one row per project; one widget
    /// list per dashboard). Anything north of 4 MB is either
    /// corrupt or hostile, and decoding it on a memory-pressured
    /// device — the kind that produces the iOS resume-time crashes
    /// in TestFlight feedback AJy1fD58 / AL8Hjm06 (Berlin, iOS 26.5,
    /// 2.87 GB free disk) — risks an OOM kill before the
    /// JSONDecoder can even bail. We treat oversize files as
    /// "missing" so the caller's fallback path runs.
    public static let maxJSONBytes = 4 * 1024 * 1024

    public let context: ServerContext
    public let transport: any ServerTransport

    public nonisolated init(context: ServerContext = .local) {
        self.context = context
        self.transport = context.makeTransport()
    }

    // MARK: - Registry

    public func loadRegistry() -> ProjectRegistry {
        // Tracks time spent reading + decoding projects.json from the transport
        // (local file or SSH). Helps spot slow remote round-trips.
        ScarfMon.measure(.diskIO, "dashboard.loadRegistry") {
            guard let data = try? transport.readFile(context.paths.projectsRegistry) else {
                return ProjectRegistry(projects: [])
            }
            if data.count > Self.maxJSONBytes {
                Self.logger.warning(
                    "Project registry at \(context.paths.projectsRegistry, privacy: .public) is \(data.count) bytes (cap \(Self.maxJSONBytes)); treating as missing"
                )
                return ProjectRegistry(projects: [])
            }
            do {
                return try JSONDecoder().decode(ProjectRegistry.self, from: data)
            } catch {
                Self.logger.error("Failed to decode project registry: \(error.localizedDescription, privacy: .public)")
                return ProjectRegistry(projects: [])
            }
        }
    }

    /// Persist the project registry to `~/.hermes/scarf/projects.json`.
    ///
    /// **Throws** on every non-success path — the previous version of
    /// this method silently swallowed `createDirectory` and `writeFile`
    /// failures with `try?`, which meant the installer could return a
    /// valid-looking `ProjectEntry` while the registry on disk never
    /// received the new row (project would complete install, show a
    /// success screen, then be invisible in the sidebar). Callers that
    /// want fire-and-forget behaviour can still use `try?`, but the
    /// choice is now theirs.
    public func saveRegistry(_ registry: ProjectRegistry) throws {
        let dir = context.paths.scarfDir
        // `createDirectory` is mkdir -p across every transport (Local
        // uses withIntermediateDirectories, SSH/Citadel both ignore
        // "already exists"), so we don't need to fileExists-guard it.
        try transport.createDirectory(dir)
        let data = try JSONEncoder().encode(registry)
        // Pretty-print for readability (agents may read this file).
        let writeData: Data
        if let pretty = try? JSONSerialization.jsonObject(with: data),
           let formatted = try? JSONSerialization.data(withJSONObject: pretty, options: [.prettyPrinted, .sortedKeys]) {
            writeData = formatted
        } else {
            writeData = data
        }
        try transport.writeFile(context.paths.projectsRegistry, data: writeData)
    }

    // MARK: - Dashboard

    public func loadDashboard(for project: ProjectEntry) -> ProjectDashboard? {
        guard let data = try? transport.readFile(project.dashboardPath) else {
            return nil
        }
        if data.count > Self.maxJSONBytes {
            Self.logger.warning(
                "Dashboard for \(project.name, privacy: .public) is \(data.count) bytes (cap \(Self.maxJSONBytes)); treating as missing"
            )
            return nil
        }
        do {
            return try JSONDecoder().decode(ProjectDashboard.self, from: data)
        } catch {
            Self.logger.error("Failed to decode dashboard for \(project.name, privacy: .public): \(error.localizedDescription, privacy: .public)")
            return nil
        }
    }

    public func dashboardExists(for project: ProjectEntry) -> Bool {
        transport.fileExists(project.dashboardPath)
    }

    public func dashboardModificationDate(for project: ProjectEntry) -> Date? {
        transport.stat(project.dashboardPath)?.mtime
    }
}
