import Foundation
import ScarfCore

@Observable
final class MemoryViewModel {
    let context: ServerContext
    private let fileService: HermesFileService

    init(context: ServerContext = .local) {
        self.context = context
        self.fileService = HermesFileService(context: context)
    }


    var memoryContent = ""
    var userContent = ""
    var memoryProvider = ""
    var isEditing = false
    var editingFile: EditTarget = .memory
    var editText = ""
    var profiles: [String] = []
    var activeProfile = ""
    var isLoading = false

    enum EditTarget {
        case memory, user
    }

    var memoryCharCount: Int { memoryContent.count }
    var userCharCount: Int { userContent.count }

    var hasExternalProvider: Bool {
        let stripped = memoryProvider
            .trimmingCharacters(in: .whitespaces)
            .trimmingCharacters(in: CharacterSet(charactersIn: "'\""))
        return !stripped.isEmpty && stripped != "file"
    }

    var hasMultipleProfiles: Bool { !profiles.isEmpty }

    func load() {
        isLoading = true
        let svc = fileService
        let currentProfile = activeProfile
        // Sync transport calls would beach-ball the UI on remote — dispatch
        // off main, then commit results back on MainActor. v2.8: wrapped
        // in ScarfMon so we can see how many SSH RTTs this load actually
        // costs (4 sequential SFTP reads on the slow path).
        Task.detached { [weak self] in
            await ScarfMon.measureAsync(.diskIO, "memory.load") {
                let config = svc.loadConfig()
                let profiles = svc.loadMemoryProfiles()
                let profile = currentProfile.isEmpty ? config.memoryProfile : currentProfile
                let memory = svc.loadMemory(profile: profile)
                let user = svc.loadUserProfile(profile: profile)
                ScarfMon.event(.diskIO, "memory.load.bytes", count: 0, bytes: memory.utf8.count + user.utf8.count)
                await MainActor.run { [weak self] in
                    guard let self else { return }
                    self.memoryProvider = config.memoryProvider
                    self.profiles = profiles
                    self.activeProfile = profile
                    self.memoryContent = memory
                    self.userContent = user
                    self.isLoading = false
                }
            }
        }
    }

    func switchProfile(_ profile: String) {
        activeProfile = profile
        let svc = fileService
        Task.detached { [weak self] in
            let memory = svc.loadMemory(profile: profile)
            let user = svc.loadUserProfile(profile: profile)
            await MainActor.run { [weak self] in
                self?.memoryContent = memory
                self?.userContent = user
            }
        }
    }

    func startEditing(_ target: EditTarget) {
        editingFile = target
        editText = target == .memory ? memoryContent : userContent
        isEditing = true
    }

    func save() {
        let svc = fileService
        let target = editingFile
        let text = editText
        let profile = activeProfile
        Task.detached { [weak self] in
            switch target {
            case .memory: svc.saveMemory(text, profile: profile)
            case .user:   svc.saveUserProfile(text, profile: profile)
            }
            await MainActor.run { [weak self] in
                guard let self else { return }
                switch target {
                case .memory: self.memoryContent = text
                case .user:   self.userContent = text
                }
                self.isEditing = false
            }
        }
    }

    func cancelEditing() {
        isEditing = false
    }
}
