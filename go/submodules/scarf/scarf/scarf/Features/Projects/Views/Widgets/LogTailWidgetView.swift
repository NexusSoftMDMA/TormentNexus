import SwiftUI
import ScarfCore
import ScarfDesign

/// Tails the last N lines of a file under the project root, monospaced.
/// Best paired with cron jobs that write atomically (write-temp + rename)
/// — the project-wide `.scarf/` directory watch (v2.7) refreshes the
/// widget when a new file lands. In-place appends to an existing file
/// won't tick `lastChangeDate`; the cron job should `touch dashboard.json`
/// after each run if it appends in place.
struct LogTailWidgetView: View {
    let widget: DashboardWidget

    @Environment(\.serverContext) private var serverContext
    @Environment(\.selectedProjectRoot) private var projectRoot
    @Environment(HermesFileWatcher.self) private var fileWatcher

    @State private var loadedTail: String?
    @State private var loadError: WidgetPathResolver.ResolveError?
    @State private var ioError: String?
    @State private var isLoading = false

    private var lineCount: Int { max(1, min(200, widget.lines ?? 20)) }

    var body: some View {
        Group {
            switch WidgetPathResolver.resolve(widget.path, projectRoot: projectRoot) {
            case .failure(let err):
                WidgetErrorCard(
                    title: widget.title,
                    reason: err.userMessage,
                    hint: "Set `path` to a file relative to the project root, e.g. `reports/uptime.log`."
                )
            case .success(let resolved):
                content(for: resolved)
                    .task(id: refreshKey(resolved)) {
                        await reload(absPath: resolved)
                    }
            }
        }
    }

    private func refreshKey(_ resolved: String) -> String {
        // Force a reload whenever either the widget config or any project
        // file changes (the latter via fileWatcher.lastChangeDate).
        "\(resolved)|\(lineCount)|\(fileWatcher.lastChangeDate.timeIntervalSince1970)"
    }

    @ViewBuilder
    private func content(for absPath: String) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 6) {
                Image(systemName: "doc.text.below.ecg")
                    .foregroundStyle(.secondary)
                    .scarfStyle(.caption)
                Text(widget.title)
                    .scarfStyle(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Text("last \(lineCount)")
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
                if isLoading {
                    ProgressView().controlSize(.mini)
                }
            }
            if let ioError {
                Text(ioError)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            } else if let loadedTail {
                tailBody(loadedTail)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(ScarfColor.backgroundSecondary)
        .clipShape(RoundedRectangle(cornerRadius: ScarfRadius.lg))
    }

    @ViewBuilder
    private func tailBody(_ tail: String) -> some View {
        let lines = tail.split(separator: "\n", omittingEmptySubsequences: false)
        if lines.isEmpty {
            Text("(empty)").font(.caption2).foregroundStyle(.secondary)
        } else {
            VStack(alignment: .leading, spacing: 1) {
                ForEach(Array(lines.enumerated()), id: \.offset) { _, line in
                    Text(String(line))
                        .font(.system(.caption2, design: .monospaced))
                        .foregroundStyle(.primary)
                        .lineLimit(1)
                        .truncationMode(.tail)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(8)
            .background(.quaternary.opacity(0.4))
            .clipShape(RoundedRectangle(cornerRadius: ScarfRadius.sm))
        }
    }

    private func reload(absPath: String) async {
        let context = serverContext
        let n = lineCount
        isLoading = true
        defer { isLoading = false }
        let outcome: WidgetIOResult<String> = await Task.detached {
            let transport = context.makeTransport()
            do {
                // Measures disk/transport latency for reading the log file.
                let data = try ScarfMon.measure(.diskIO, "widget.log_tail.load") {
                    try transport.readFile(absPath)
                }
                guard let text = String(data: data, encoding: .utf8) else {
                    return .failure("File is not UTF-8 — log_tail expects text.")
                }
                let stripped = AnsiStripper.strip(text)
                let parts = stripped.split(separator: "\n", omittingEmptySubsequences: false)
                return .success(parts.suffix(n).joined(separator: "\n"))
            } catch {
                return .failure("Could not read file: \(error.localizedDescription)")
            }
        }.value
        switch outcome {
        case .success(let s):
            self.loadedTail = s
            self.ioError = nil
        case .failure(let err):
            self.loadedTail = nil
            self.ioError = err.message
        }
    }
}
