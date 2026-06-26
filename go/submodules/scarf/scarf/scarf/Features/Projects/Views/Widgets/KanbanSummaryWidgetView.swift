import SwiftUI
import ScarfCore
import ScarfDesign

/// `kanban_summary` dashboard widget. Renders a compact 3-row list of
/// the most-pressing tasks (running + blocked + todo, by priority)
/// for the active project's tenant, plus a glance string footer.
///
/// Looks up the project's tenant from `<project>/.scarf/manifest.json`
/// at first render (cheap; cached). Falls back to "no tasks" copy when
/// no tenant is minted yet (i.e. the user hasn't opened the Kanban
/// tab yet).
struct KanbanSummaryWidgetView: View {
    let widget: DashboardWidget

    @Environment(\.serverContext) private var serverContext
    @Environment(\.selectedProjectRoot) private var projectRoot

    @State private var tenant: String?
    @State private var tasks: [HermesKanbanTask] = []
    @State private var stats: HermesKanbanStats = .empty
    @State private var isLoading = false
    @State private var error: String?
    @State private var pollTask: Task<Void, Never>?

    private var maxRows: Int {
        if case .number(let n) = widget.value { return max(1, Int(n)) }
        return 3
    }

    var body: some View {
        ScarfCard {
            VStack(alignment: .leading, spacing: ScarfSpace.s2) {
                header
                if let error {
                    errorRow(error)
                } else if tasks.isEmpty && !isLoading {
                    emptyRow
                } else {
                    ForEach(tasks.prefix(maxRows)) { task in
                        taskRow(task)
                    }
                }
                if !stats.glanceString.isEmpty {
                    Text(stats.glanceString)
                        .scarfStyle(.caption)
                        .foregroundStyle(ScarfColor.foregroundFaint)
                        .padding(.top, 4)
                }
            }
        }
        .onAppear { startPolling() }
        .onDisappear { pollTask?.cancel() }
    }

    private var header: some View {
        HStack {
            Text(widget.title.isEmpty ? "Kanban" : widget.title)
                .scarfStyle(.headline)
                .foregroundStyle(ScarfColor.foregroundPrimary)
            Spacer()
            Image(systemName: "rectangle.split.3x1")
                .foregroundStyle(ScarfColor.foregroundMuted)
        }
    }

    private func taskRow(_ task: HermesKanbanTask) -> some View {
        HStack(spacing: ScarfSpace.s2) {
            statusDot(for: task.status)
            Text(task.title)
                .scarfStyle(.body)
                .foregroundStyle(ScarfColor.foregroundPrimary)
                .lineLimit(1)
            Spacer(minLength: 0)
            if let assignee = task.assignee, !assignee.isEmpty {
                Text(initials(of: assignee))
                    .font(.system(size: 9, weight: .semibold))
                    .foregroundStyle(ScarfColor.accentActive)
                    .frame(width: 16, height: 16)
                    .background(ScarfColor.accentTint)
                    .clipShape(Circle())
            }
        }
        .padding(.vertical, 2)
    }

    private func statusDot(for status: String) -> some View {
        let color: Color
        switch KanbanStatus.from(status) {
        case .running:           color = ScarfColor.info
        case .blocked:           color = ScarfColor.warning
        case .done:              color = ScarfColor.success
        default:                 color = ScarfColor.foregroundMuted
        }
        return Circle()
            .fill(color)
            .frame(width: 8, height: 8)
    }

    private var emptyRow: some View {
        Text("No active tasks for this project.")
            .scarfStyle(.caption)
            .foregroundStyle(ScarfColor.foregroundFaint)
            .padding(.vertical, ScarfSpace.s2)
    }

    private func errorRow(_ message: String) -> some View {
        HStack(spacing: 4) {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(ScarfColor.warning)
                .font(.caption)
            Text(message)
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundMuted)
                .lineLimit(2)
        }
    }

    // MARK: - Loading

    private func startPolling() {
        pollTask?.cancel()
        pollTask = Task {
            while !Task.isCancelled {
                await loadOnce()
                try? await Task.sleep(nanoseconds: 10_000_000_000)
            }
        }
    }

    private func loadOnce() async {
        guard let projectRoot, !projectRoot.isEmpty else { return }
        if tenant == nil {
            tenant = readTenant(at: projectRoot, context: serverContext)
        }
        guard let tenant, !tenant.isEmpty else {
            tasks = []
            return
        }
        isLoading = true
        defer { isLoading = false }
        let svc = KanbanService(context: serverContext)
        let filter = KanbanListFilter(tenant: tenant)
        do {
            let polled = try await svc.list(filter)
            // Sort by priority DESC, status preference (running > blocked > todo).
            tasks = polled
                .filter {
                    let status = KanbanStatus.from($0.status)
                    return status != .done && status != .archived
                }
                .sorted { lhs, rhs in
                    let lp = lhs.priority ?? 0
                    let rp = rhs.priority ?? 0
                    if lp != rp { return lp > rp }
                    return statusRank(lhs.status) < statusRank(rhs.status)
                }
            stats = (try? await svc.stats()) ?? .empty
            error = nil
        } catch let err as KanbanError {
            error = err.errorDescription
        } catch {
            self.error = error.localizedDescription
        }
    }

    private nonisolated func statusRank(_ status: String) -> Int {
        switch KanbanStatus.from(status) {
        case .running:  return 0
        case .blocked:  return 1
        case .ready:    return 2
        case .todo:     return 3
        default:        return 4
        }
    }

    // `context` passed in (read on MainActor by the caller) instead of
    // touching the View's MainActor-isolated `serverContext` from this
    // nonisolated method. (t-aud23)
    private nonisolated func readTenant(at projectPath: String, context serverContext: ServerContext) -> String? {
        let manifestPath = projectPath + "/.scarf/manifest.json"
        let transport = serverContext.makeTransport()
        guard transport.fileExists(manifestPath),
              let data = try? transport.readFile(manifestPath),
              let manifest = try? JSONDecoder().decode(ProjectTemplateManifest.self, from: data)
        else {
            return nil
        }
        return manifest.kanbanTenant
    }

    private func initials(of name: String) -> String {
        let parts = name.split(whereSeparator: { !$0.isLetter && !$0.isNumber })
        let letters = parts.prefix(2).compactMap { $0.first.map(String.init) }
        return letters.joined().uppercased()
    }
}
