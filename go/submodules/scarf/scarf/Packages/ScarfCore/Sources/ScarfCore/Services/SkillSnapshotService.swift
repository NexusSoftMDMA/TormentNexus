import Foundation
#if canImport(os)
import os
#endif

/// Tracks "what skills did this Scarf instance see last time on this
/// server" so the Skills tab can render a "2 new, 4 updated since you
/// last looked" pill — same pattern Hermes's `hermes skills update`
/// shows in the CLI.
///
/// **Storage shape.** Per server. Mac persists to a JSON file under
/// `~/Library/Application Support/com.scarf/skill-snapshots/<serverID>.json`;
/// iOS persists to `UserDefaults` under
/// `com.scarf.ios.skill-snapshot.<serverID>`. Both stores hold a
/// `[skillId: SkillSignature]` map. `skillId` is `<category>/<name>`
/// from the loader; `SkillSignature` is a compact hash of file count
/// + sorted file names so a skill update (added/removed/renamed file)
/// shows up as a delta even if Hermes-version-pinning isn't in play.
///
/// **Failure model.** Persist errors log + no-op (the diff degrades
/// to "everything looks new" but the user can still mark seen). Read
/// errors return an empty snapshot so the diff treats every skill as
/// new — annoying once but recoverable.
public struct SkillSnapshotService: Sendable {
    #if canImport(os)
    static let logger = Logger(
        subsystem: "com.scarf",
        category: "SkillSnapshotService"
    )
    #endif

    public let serverID: ServerID

    /// Optional Hermes profile discriminator (issue #120). Each profile is
    /// an independent `HERMES_HOME` with its OWN `skills/` dir, so the
    /// last-seen baseline must be namespaced per `(server, profile)` —
    /// otherwise switching profiles diffs the new profile's skills against
    /// the previous profile's baseline and the "What's New" pill lies.
    ///
    /// Normalized through `HermesProfileScope` in `init`, so `nil` /
    /// `"default"` / an invalid name all collapse to the bare per-server
    /// key — byte-for-byte the pre-#120 storage key, so existing
    /// default-profile baselines keep resolving with no migration.
    public let profile: String?
    private let backend: SnapshotBackend

    public init(serverID: ServerID, profile: String? = nil) {
        self.serverID = serverID
        self.profile = HermesProfileScope.normalize(profile)
        #if os(macOS)
        self.backend = .file(MacSnapshotBackend())
        #else
        self.backend = .userDefaults(IOSSnapshotBackend())
        #endif
    }

    /// Public for tests. Production callers use the backend-less
    /// init that picks the right backend per platform.
    public init(serverID: ServerID, profile: String? = nil, backend: SnapshotBackend) {
        self.serverID = serverID
        self.profile = HermesProfileScope.normalize(profile)
        self.backend = backend
    }

    /// Compose the persistence key for a `(server, profile)` pair. A nil /
    /// empty profile yields the bare server UUID — identical to the pre-#120
    /// key, so default-profile baselines survive an upgrade untouched. A
    /// named profile gets a `.`-suffixed namespace (`<uuid>.<name>`); the
    /// name is Hermes-regex-validated (`HermesProfileScope`), so it is a
    /// safe filename / `UserDefaults`-key component.
    static func storageKey(for serverID: ServerID, scope: String?) -> String {
        guard let scope, !scope.isEmpty else { return serverID.uuidString }
        return "\(serverID.uuidString).\(scope)"
    }

    // MARK: - Public API

    /// Compute the delta between the current skill set and the
    /// last snapshot. Returns counts only — the caller renders them
    /// in a pill ("2 new, 4 updated since you last looked").
    public func diff(against current: [HermesSkill]) -> SkillSnapshotDiff {
        let last = backend.read(for: serverID, scope: profile)
        let currentSigs = Self.signatures(for: current)

        var newCount = 0
        var updatedCount = 0
        for (id, sig) in currentSigs {
            if let prev = last[id] {
                if prev != sig { updatedCount += 1 }
            } else {
                newCount += 1
            }
        }
        return SkillSnapshotDiff(
            newCount: newCount,
            updatedCount: updatedCount,
            previousSnapshotEmpty: last.isEmpty,
            changedSkillIds: Set(currentSigs.compactMap { id, sig in
                let prev = last[id]
                if prev == nil { return id }
                return prev != sig ? id : nil
            })
        )
    }

    /// Record the current skills as seen. Called on user action
    /// ("Mark all as seen") or on the first-ever load (when
    /// `previousSnapshotEmpty` is true and we don't want to show
    /// every skill as new on next launch).
    public func markSeen(_ current: [HermesSkill]) {
        backend.write(Self.signatures(for: current), for: serverID, scope: profile)
    }

    // MARK: - Private

    /// Compute a stable per-skill signature: `<fileCount>:<joined-files>`.
    /// Hash via Foundation's `String.hashValue` would be unstable
    /// across launches, so we keep the raw concatenation; payload is
    /// small enough that storage size doesn't matter.
    static func signatures(for skills: [HermesSkill]) -> [String: SkillSignature] {
        var result: [String: SkillSignature] = [:]
        for skill in skills {
            let files = skill.files.sorted().joined(separator: "|")
            result[skill.id] = "\(skill.files.count):\(files)"
        }
        return result
    }
}

public typealias SkillSignature = String

/// Result of comparing the current skills to the last snapshot.
public struct SkillSnapshotDiff: Sendable, Equatable {
    public let newCount: Int
    public let updatedCount: Int
    /// True when this is the first time we've ever seen the skills
    /// for this server (no prior snapshot on disk). The view should
    /// silently mark everything as seen rather than rendering a
    /// "5 new!" pill on a fresh install.
    public let previousSnapshotEmpty: Bool
    /// Skill ids (`<category>/<name>`) that count as new or updated.
    /// Used by the Mac UI to filter the tree to only changed entries
    /// when the user taps the pill.
    public let changedSkillIds: Set<String>

    public var hasChanges: Bool { newCount + updatedCount > 0 }

    public init(
        newCount: Int,
        updatedCount: Int,
        previousSnapshotEmpty: Bool,
        changedSkillIds: Set<String>
    ) {
        self.newCount = newCount
        self.updatedCount = updatedCount
        self.previousSnapshotEmpty = previousSnapshotEmpty
        self.changedSkillIds = changedSkillIds
    }

    /// Compact label for the "What's New" pill, e.g.
    /// "2 new, 4 changed since you last looked" or "1 new skill".
    ///
    /// Wording note (issue #78): we used to say "X updated since you
    /// last looked" but the same screen also surfaces an "Updates"
    /// sub-tab driven by `hermes skills check` (skills with newer
    /// **upstream** versions available). Two surfaces with the word
    /// "update" meaning two different things read as a contradiction
    /// to the user. "Changed" describes the local file delta without
    /// colliding with upstream-update vocabulary.
    public var label: String {
        switch (newCount, updatedCount) {
        case (let n, 0): return n == 1 ? "1 new skill since you last looked" : "\(n) new skills since you last looked"
        case (0, let u): return u == 1 ? "1 changed skill since you last looked" : "\(u) changed skills since you last looked"
        default: return "\(newCount) new, \(updatedCount) changed since you last looked"
        }
    }
}

// MARK: - Backend abstraction

/// Per-platform persistence for the skill-snapshot store. Both
/// implementations encode `[skillId: signature]` as JSON and key by
/// `ServerID`. Tests use the in-memory variant.
public enum SnapshotBackend: Sendable {
    case file(MacSnapshotBackend)
    case userDefaults(IOSSnapshotBackend)
    case inMemory(InMemorySnapshotBackend)

    func read(for serverID: ServerID, scope: String? = nil) -> [String: SkillSignature] {
        switch self {
        case .file(let b): return b.read(for: serverID, scope: scope)
        case .userDefaults(let b): return b.read(for: serverID, scope: scope)
        case .inMemory(let b): return b.read(for: serverID, scope: scope)
        }
    }

    func write(_ snapshot: [String: SkillSignature], for serverID: ServerID, scope: String? = nil) {
        switch self {
        case .file(let b): b.write(snapshot, for: serverID, scope: scope)
        case .userDefaults(let b): b.write(snapshot, for: serverID, scope: scope)
        case .inMemory(let b): b.write(snapshot, for: serverID, scope: scope)
        }
    }
}

#if os(macOS)
/// Mac persistence: one JSON file per server under
/// `~/Library/Application Support/com.scarf/skill-snapshots/`.
public struct MacSnapshotBackend: Sendable {
    public init() {}

    public func read(for serverID: ServerID, scope: String? = nil) -> [String: SkillSignature] {
        guard let url = Self.fileURL(for: serverID, scope: scope),
              let data = try? Data(contentsOf: url),
              let map = try? JSONDecoder().decode([String: SkillSignature].self, from: data)
        else { return [:] }
        return map
    }

    public func write(_ snapshot: [String: SkillSignature], for serverID: ServerID, scope: String? = nil) {
        guard let url = Self.fileURL(for: serverID, scope: scope) else { return }
        do {
            try FileManager.default.createDirectory(
                at: url.deletingLastPathComponent(),
                withIntermediateDirectories: true
            )
            let encoder = JSONEncoder()
            encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
            let data = try encoder.encode(snapshot)
            try data.write(to: url, options: .atomic)
        } catch {
            #if canImport(os)
            SkillSnapshotService.logger.warning(
                "couldn't persist skill snapshot for \(serverID.uuidString, privacy: .public): \(error.localizedDescription, privacy: .public)"
            )
            #endif
        }
    }

    private static func fileURL(for serverID: ServerID, scope: String?) -> URL? {
        guard let appSupport = FileManager.default.urls(
            for: .applicationSupportDirectory,
            in: .userDomainMask
        ).first else { return nil }
        return appSupport
            .appendingPathComponent("com.scarf", isDirectory: true)
            .appendingPathComponent("skill-snapshots", isDirectory: true)
            .appendingPathComponent("\(SkillSnapshotService.storageKey(for: serverID, scope: scope)).json")
    }
}
#else
public struct MacSnapshotBackend: Sendable {
    public init() {}
    public func read(for serverID: ServerID, scope: String? = nil) -> [String: SkillSignature] { [:] }
    public func write(_ snapshot: [String: SkillSignature], for serverID: ServerID, scope: String? = nil) {}
}
#endif

/// iOS persistence: one UserDefaults key per server.
/// `@unchecked Sendable` because `UserDefaults` itself doesn't conform
/// — it is in fact thread-safe per Apple docs, the conformance just
/// hasn't been added to the headers.
public final class IOSSnapshotBackend: @unchecked Sendable {
    public static let keyPrefix = "com.scarf.ios.skill-snapshot."

    private let defaults: UserDefaults

    public init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    public func read(for serverID: ServerID, scope: String? = nil) -> [String: SkillSignature] {
        let key = Self.keyPrefix + SkillSnapshotService.storageKey(for: serverID, scope: scope)
        guard let data = defaults.data(forKey: key),
              let map = try? JSONDecoder().decode([String: SkillSignature].self, from: data)
        else { return [:] }
        return map
    }

    public func write(_ snapshot: [String: SkillSignature], for serverID: ServerID, scope: String? = nil) {
        let key = Self.keyPrefix + SkillSnapshotService.storageKey(for: serverID, scope: scope)
        guard let data = try? JSONEncoder().encode(snapshot) else { return }
        defaults.set(data, forKey: key)
    }
}

/// In-memory backend for tests. Per-instance map; not shared. Keyed by the
/// composed `(server, scope)` storage key — same namespacing as the
/// production backends — so profile-scoping behavior is exercised exactly.
public final class InMemorySnapshotBackend: @unchecked Sendable {
    private var store: [String: [String: SkillSignature]] = [:]

    public init() {}

    public func read(for serverID: ServerID, scope: String? = nil) -> [String: SkillSignature] {
        store[SkillSnapshotService.storageKey(for: serverID, scope: scope)] ?? [:]
    }

    public func write(_ snapshot: [String: SkillSignature], for serverID: ServerID, scope: String? = nil) {
        store[SkillSnapshotService.storageKey(for: serverID, scope: scope)] = snapshot
    }
}
