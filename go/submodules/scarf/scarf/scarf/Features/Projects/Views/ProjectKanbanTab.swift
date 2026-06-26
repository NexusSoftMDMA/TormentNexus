import SwiftUI
import ScarfCore
import ScarfDesign

/// Per-project Kanban tab. Wraps `KanbanBoardView` with the project's
/// tenant pre-applied + the workspace pre-pinned to the project
/// directory. On first appearance it mints the project's
/// `scarf:<slug>` tenant if one isn't already on disk.
///
/// Capability-gated by `HermesCapabilities.hasKanban` upstream — this
/// view is only added to the project tab list when v0.12+ is detected.
struct ProjectKanbanTab: View {
    @Environment(\.serverContext) private var serverContext
    let project: ProjectEntry

    @State private var resolvedTenant: String?
    @State private var resolveError: String?

    var body: some View {
        Group {
            if let tenant = resolvedTenant {
                KanbanBoardView(
                    context: serverContext,
                    tenantFilter: tenant,
                    projectPath: project.path,
                    projectName: project.name
                )
            } else if let error = resolveError {
                VStack(spacing: ScarfSpace.s3) {
                    Image(systemName: "exclamationmark.triangle")
                        .font(.system(size: 32))
                        .foregroundStyle(ScarfColor.warning)
                    Text("Couldn't set up the project's Kanban tenant.")
                        .scarfStyle(.headline)
                    Text(error)
                        .scarfStyle(.caption)
                        .foregroundStyle(ScarfColor.foregroundMuted)
                        .multilineTextAlignment(.center)
                    Button("Retry") {
                        resolveError = nil
                        resolveTenant()
                    }
                    .buttonStyle(ScarfSecondaryButton())
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .padding()
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task(id: project.id) {
            resolveTenant()
        }
    }

    private func resolveTenant() {
        let resolver = KanbanTenantResolver(context: serverContext)
        let project = self.project
        // Always-mint behaviour: even if the project board is empty
        // and the user hasn't created a task yet, the tenant is
        // pre-allocated so AGENTS.md surfaces it on the next chat.
        //
        // `resolveOrMint` does synchronous FileManager I/O (it walks
        // every project's manifest to mint/read the tenant), so run it
        // off-main — calling it inline on a Kanban tab switch blocked
        // the main thread. `Task {}` inherits this View's @MainActor so
        // the @State writes stay on main; the inner `Task.detached`
        // captures only the Sendable `resolver`/`project` and reports
        // back via a Sendable `Result<String, String>`. The view shows
        // its ProgressView branch until `resolvedTenant` lands. (t-aud03)
        Task {
            do {
                let tenant = try await Task.detached(priority: .userInitiated) {
                    try resolver.resolveOrMint(for: project)
                }.value
                resolvedTenant = tenant
            } catch {
                resolveError = error.localizedDescription
            }
        }
    }
}
