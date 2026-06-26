import Foundation
import Observation
import ScarfCore
import os

/// Drives `RestoreServerSheet`. Mirrors `BackupServerViewModel`: the
/// flow is pickArchive → inspect → confirm → run → done | failed.
@Observable
@MainActor
final class RestoreServerViewModel {
    enum Phase: Equatable {
        case awaitingFile
        case inspecting
        case ready(RemoteRestoreService.InspectionResult)
        case running(RemoteRestoreService.Progress)
        case done(RemoteRestoreService.RestoreResult)
        case failed(String)

        static func == (lhs: Phase, rhs: Phase) -> Bool {
            switch (lhs, rhs) {
            case (.awaitingFile, .awaitingFile): return true
            case (.inspecting, .inspecting): return true
            case (.ready, .ready): return true
            case (.running(let a), .running(let b)): return a == b
            case (.done, .done): return true
            case (.failed(let a), .failed(let b)): return a == b
            default: return false
            }
        }
    }

    private static let logger = Logger(subsystem: "com.scarf", category: "RestoreServerViewModel")

    let context: ServerContext
    var phase: Phase = .awaitingFile
    var pauseCronJobs = true
    var targetProjectsRoot: String = ""

    private var workTask: Task<Void, Never>?

    init(context: ServerContext) {
        self.context = context
    }

    func inspect(archiveURL: URL) async {
        phase = .inspecting
        let service = RemoteRestoreService(context: context)
        do {
            let result = try await service.inspect(archiveURL: archiveURL)
            // Default the projects root to `<targetHome>/projects`.
            if targetProjectsRoot.isEmpty {
                let home = result.targetHomeResolved ?? (result.manifest.hermes.homePath as NSString).deletingLastPathComponent
                targetProjectsRoot = home + "/projects"
            }
            phase = .ready(result)
        } catch {
            phase = .failed(error.localizedDescription)
            Self.logger.error("Restore inspect failed: \(error.localizedDescription, privacy: .public)")
        }
    }

    func runRestore(inspection: RemoteRestoreService.InspectionResult) {
        let opts = RemoteRestoreService.RestoreOptions(
            targetProjectsRoot: targetProjectsRoot.isEmpty ? nil : targetProjectsRoot,
            pauseCronJobs: pauseCronJobs
        )
        phase = .running(.planning)
        // Same two-step capture pattern as BackupServerViewModel:
        // weak handle in the outer Task, strong promotion inside, so
        // the @Sendable progress callback hops back via the box
        // without the Swift 6 var-self warning.
        let weakSelf = WeakBox(self)
        workTask = Task { @MainActor in
            guard let viewModel = weakSelf.value else { return }
            let service = RemoteRestoreService(context: viewModel.context)
            do {
                let result = try await service.run(
                    inspection: inspection,
                    options: opts,
                    progress: { step in
                        Task { @MainActor in
                            weakSelf.value?.phase = .running(step)
                        }
                    }
                )
                viewModel.phase = .done(result)
            } catch is CancellationError {
                viewModel.phase = .failed("Cancelled.")
            } catch {
                viewModel.phase = .failed(error.localizedDescription)
                Self.logger.error("Restore run failed: \(error.localizedDescription, privacy: .public)")
            }
        }
    }

    private final class WeakBox: @unchecked Sendable {
        weak var value: RestoreServerViewModel?
        init(_ v: RestoreServerViewModel) { self.value = v }
    }

    func cancel() {
        workTask?.cancel()
        workTask = nil
    }
}
