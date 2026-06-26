import SwiftUI
import ScarfCore
import ScarfIOS

/// Cross-tab signalling for ScarfGo. Mirrors the Mac app's
/// `AppCoordinator` pattern: an `@Observable` carrier injected via
/// `.environment(_:)` that any view in the tab tree can reach.
///
/// v2.5 expands the surface to include project handoff: tapping
/// "New Chat" inside a Project Detail view sets `pendingProjectChat`
/// and routes to the Chat tab, where ChatController consumes it and
/// dispatches `resetAndStartInProject(_:)` (same wiring the in-Chat
/// project picker sheet already uses).
@Observable
@MainActor
final class ScarfGoCoordinator {

    /// Which tab ScarfGoTabRoot should present. Changing this from
    /// anywhere in the tree re-selects the tab. Bound as `selection:`
    /// on the root TabView.
    var selectedTab: Tab = .chat

    /// The server whose profile selection this coordinator owns. Used
    /// to key the per-server selection store.
    private let serverID: ServerID

    /// Persistence for the selected profile (issue #120, Design B).
    /// Injected so tests can substitute `InMemoryProfileSelectionStore`.
    private let selectionStore: any IOSProfileSelectionStore

    /// Selected Hermes profile to scope this session's reads/writes/CLI
    /// (issue #120, Design B). `nil` = the default (root) profile.
    /// ScarfGoTabRoot derives the effective `HERMES_HOME` from this and
    /// rebuilds the tab subtree when it changes — the iOS analogue of
    /// the Mac app's relaunch-to-flush-state, but WITHOUT mutating the
    /// host's `active_profile`. Set via `setSelectedProfile(_:)`.
    private(set) var selectedProfile: String?

    init(
        serverID: ServerID,
        selectionStore: any IOSProfileSelectionStore = UserDefaultsProfileSelectionStore()
    ) {
        self.serverID = serverID
        self.selectionStore = selectionStore
        self.selectedProfile = selectionStore.selectedProfile(for: serverID)
    }

    /// Set and persist the selected profile, scoping every subsequent
    /// read/write/CLI for this session to it. Pass `nil`/`"default"`/an
    /// invalid name to return to the root profile. A no-op when the
    /// (normalized) value is unchanged, so callers can wire it directly
    /// to a picker without churning the view tree.
    func setSelectedProfile(_ name: String?) {
        let normalized = HermesProfileScope.normalize(name)
        guard normalized != selectedProfile else { return }
        selectionStore.setSelectedProfile(normalized, for: serverID)
        selectedProfile = normalized
    }

    /// If non-nil, ChatController should resume this session on next
    /// appear instead of starting a fresh one. Consumed (cleared) by
    /// ChatController after it honours the request.
    var pendingResumeSessionID: String?

    /// If non-nil, the Chat tab should start an in-project session at
    /// this absolute remote path on next appear instead of a quick
    /// chat. Consumed (cleared) by ChatController after it kicks off
    /// `resetAndStartInProject(_:)`. Mirrors Mac's
    /// `AppCoordinator.pendingProjectChat`.
    var pendingProjectChat: String?

    /// Most-recent scene-phase value observed at the WindowGroup
    /// level. Tab-specific view models (e.g. `ChatController`)
    /// observe `scenePhaseTick` to react to transitions even when
    /// they're on a non-foreground tab — `.onChange(of: ScenePhase)`
    /// alone wouldn't fire for views that aren't on screen.
    private(set) var scenePhase: ScenePhase = .active
    private(set) var scenePhaseTick: Int = 0
    /// Wallclock when we last observed `.background`. Used by tab
    /// view-models to decide whether a quick `.active` transition is
    /// worth a full re-verify (long suspensions warrant it; brief
    /// notification-center peeks don't). `nil` until the first
    /// background transition.
    private(set) var lastBackgroundedAt: Date?

    func setScenePhase(_ phase: ScenePhase) {
        if phase == .background, scenePhase != .background {
            lastBackgroundedAt = Date()
            // Close this server's pooled SSH connection on background. iOS
            // suspends the socket regardless; evicting guarantees a clean
            // warm reconnect on return instead of racing a half-dead
            // session. The ACP chat channel manages its own scene-phase
            // lifecycle separately and is unaffected.
            let id = serverID
            Task { await CitadelTransportPool.shared.evict(id) }
        }
        scenePhase = phase
        scenePhaseTick &+= 1
    }

    enum Tab: Hashable {
        case dashboard, projects, chat, skills, system
    }

    /// Convenience: route to Chat and queue a resume. Dashboard rows
    /// call this on tap. Clearing `pendingResumeSessionID` is the
    /// consumer's responsibility — in ChatController's case, right
    /// after the resume flow wins (success or failure).
    func resumeSession(_ id: String) {
        pendingResumeSessionID = id
        selectedTab = .chat
    }

    /// Convenience: route to Chat and queue a project-scoped session
    /// start at `path`. Project Detail's "New Chat" toolbar button
    /// calls this. Clearing `pendingProjectChat` is the consumer's
    /// responsibility (ChatController) once `resetAndStartInProject`
    /// has been dispatched.
    func startChatInProject(path: String) {
        pendingProjectChat = path
        selectedTab = .chat
    }
}

/// Environment key so subviews can pull the coordinator without
/// explicit threading.
private struct ScarfGoCoordinatorKey: EnvironmentKey {
    static let defaultValue: ScarfGoCoordinator? = nil
}

extension EnvironmentValues {
    var scarfGoCoordinator: ScarfGoCoordinator? {
        get { self[ScarfGoCoordinatorKey.self] }
        set { self[ScarfGoCoordinatorKey.self] = newValue }
    }
}
