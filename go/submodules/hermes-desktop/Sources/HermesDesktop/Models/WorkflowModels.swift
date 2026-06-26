import Foundation

enum WorkflowPromptFormatter {
    static func normalizeForLaunch(_ prompt: String) -> String {
        prompt
            .components(separatedBy: .newlines)
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
            .joined(separator: " ")
    }
}

struct WorkflowSkillReference: Codable, Hashable, Sendable, Identifiable {
    let relativePath: String
    let slug: String
    let name: String?

    var id: String {
        relativePath
    }

    init(relativePath: String, slug: String, name: String?) {
        self.relativePath = relativePath
        self.slug = slug
        self.name = name?.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    init(skill: SkillSummary) {
        self.init(
            relativePath: skill.relativePath,
            slug: skill.slug,
            name: skill.name
        )
    }

    var resolvedName: String {
        if let name,
           !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return name
        }

        return slug
    }
}

struct WorkflowPreset: Codable, Identifiable, Equatable, Hashable, Sendable {
    var id: UUID
    var workspaceScopeFingerprint: String
    var name: String
    var prompt: String
    var assignedSkills: [WorkflowSkillReference]
    var createdAt: Date
    var updatedAt: Date

    init(
        id: UUID = UUID(),
        workspaceScopeFingerprint: String,
        name: String,
        prompt: String,
        assignedSkills: [WorkflowSkillReference],
        createdAt: Date = Date(),
        updatedAt: Date = Date()
    ) {
        self.id = id
        self.workspaceScopeFingerprint = workspaceScopeFingerprint
        self.name = name.trimmingCharacters(in: .whitespacesAndNewlines)
        self.prompt = prompt.trimmingCharacters(in: .whitespacesAndNewlines)
        self.assignedSkills = Self.normalizedSkillReferences(assignedSkills)
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }

    var promptPreview: String {
        let compactPrompt = prompt
            .components(separatedBy: .whitespacesAndNewlines)
            .filter { !$0.isEmpty }
            .joined(separator: " ")

        guard compactPrompt.count > 140 else {
            return compactPrompt
        }

        let index = compactPrompt.index(compactPrompt.startIndex, offsetBy: 140)
        return compactPrompt[..<index].trimmingCharacters(in: .whitespacesAndNewlines) + "…"
    }

    var searchableText: [String] {
        [name, prompt] + assignedSkills.map(\.relativePath) + assignedSkills.map(\.resolvedName)
    }

    func matchesSearch(_ query: String) -> Bool {
        let trimmedQuery = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedQuery.isEmpty else { return true }

        let normalizedQuery = trimmedQuery.folding(options: [.diacriticInsensitive, .caseInsensitive], locale: .current)
        return searchableText.contains { value in
            value.folding(options: [.diacriticInsensitive, .caseInsensitive], locale: .current)
                .localizedStandardContains(normalizedQuery)
        }
    }

    func updated(
        name: String,
        prompt: String,
        assignedSkills: [WorkflowSkillReference]
    ) -> WorkflowPreset {
        WorkflowPreset(
            id: id,
            workspaceScopeFingerprint: workspaceScopeFingerprint,
            name: name,
            prompt: prompt,
            assignedSkills: assignedSkills,
            createdAt: createdAt,
            updatedAt: Date()
        )
    }

    private static func normalizedSkillReferences(_ references: [WorkflowSkillReference]) -> [WorkflowSkillReference] {
        let deduped = Dictionary(references.map { ($0.relativePath, $0) }, uniquingKeysWith: { _, latest in latest })
        return deduped.values.sorted { lhs, rhs in
            let comparison = lhs.slug.localizedCaseInsensitiveCompare(rhs.slug)
            if comparison != .orderedSame {
                return comparison == .orderedAscending
            }

            return lhs.relativePath.localizedCaseInsensitiveCompare(rhs.relativePath) == .orderedAscending
        }
    }
}

enum WorkflowRunDestination: Equatable, Sendable {
    case terminal
    case chat
}

struct WorkflowDraft: Equatable {
    var name = ""
    var prompt = ""
    var selectedSkills: [WorkflowSkillReference] = []

    var normalizedName: String {
        name.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    var normalizedPrompt: String {
        prompt.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    var normalizedSelectedSkills: [WorkflowSkillReference] {
        let deduped = Dictionary(selectedSkills.map { ($0.relativePath, $0) }, uniquingKeysWith: { _, latest in latest })
        return deduped.values.sorted { lhs, rhs in
            let comparison = lhs.slug.localizedCaseInsensitiveCompare(rhs.slug)
            if comparison != .orderedSame {
                return comparison == .orderedAscending
            }

            return lhs.relativePath.localizedCaseInsensitiveCompare(rhs.relativePath) == .orderedAscending
        }
    }

    var validationError: String? {
        guard !normalizedName.isEmpty else {
            return "Workflow name is required."
        }

        guard !normalizedPrompt.isEmpty else {
            return "Workflow prompt is required."
        }

        return nil
    }

    func containsSkill(relativePath: String) -> Bool {
        normalizedSelectedSkills.contains { $0.relativePath == relativePath }
    }

    mutating func setSkillSelected(_ isSelected: Bool, skill: SkillSummary) {
        let reference = WorkflowSkillReference(skill: skill)

        if isSelected {
            if let index = selectedSkills.firstIndex(where: { $0.relativePath == reference.relativePath }) {
                selectedSkills[index] = reference
            } else {
                selectedSkills.append(reference)
            }
        } else {
            selectedSkills.removeAll { $0.relativePath == reference.relativePath }
        }
    }

    mutating func removeSkill(relativePath: String) {
        selectedSkills.removeAll { $0.relativePath == relativePath }
    }

    mutating func refreshSelectedSkills(using catalog: [SkillSummary]) {
        let catalogByRelativePath = Dictionary(uniqueKeysWithValues: catalog.map { ($0.relativePath, WorkflowSkillReference(skill: $0)) })
        selectedSkills = normalizedSelectedSkills.map { reference in
            catalogByRelativePath[reference.relativePath] ?? reference
        }
    }

    static func from(workflow: WorkflowPreset) -> WorkflowDraft {
        WorkflowDraft(
            name: workflow.name,
            prompt: workflow.prompt,
            selectedSkills: workflow.assignedSkills
        )
    }
}

struct WorkflowLaunchInvocation: Equatable, Sendable {
    let prompt: String
    let hermesProfileName: String?
    let skillRelativePaths: [String]
    let startupCommandLine: String

    init(workflow: WorkflowPreset, connection: ConnectionProfile) {
        self.prompt = workflow.prompt
        self.hermesProfileName = connection.cliHermesProfileName
        self.skillRelativePaths = workflow.assignedSkills.map(\.relativePath)
        self.startupCommandLine = connection.remoteHermesCommandLine(arguments: Self.buildArguments(
            hermesProfileName: connection.cliHermesProfileName,
            skillRelativePaths: workflow.assignedSkills.map(\.relativePath)
        ))
    }

    var arguments: [String] {
        Self.buildArguments(
            hermesProfileName: hermesProfileName,
            skillRelativePaths: skillRelativePaths
        )
    }

    var initialInput: String {
        WorkflowPromptFormatter.normalizeForLaunch(prompt)
    }

    var commandLine: String {
        (["hermes"] + arguments)
            .map(\.shellQuotedForTerminalCommand)
            .joined(separator: " ")
    }

    private static func buildArguments(
        hermesProfileName: String?,
        skillRelativePaths: [String]
    ) -> [String] {
        var values = [String]()

        if let hermesProfileName {
            values.append(contentsOf: ["--profile", hermesProfileName])
        }

        for skillRelativePath in skillRelativePaths {
            values.append(contentsOf: ["--skills", skillRelativePath])
        }

        values.append("chat")
        return values
    }
}

struct WorkflowChatLaunchInvocation: Equatable, Sendable {
    let prompt: String
    let skillCommands: [String]
    let initialInput: String
    let tuiInvocation: HermesTUIInvocation

    init(workflow: WorkflowPreset, connection: ConnectionProfile) {
        self.prompt = workflow.prompt
        let skillCommands = workflow.assignedSkills.map { "/\($0.slug)" }
        self.skillCommands = skillCommands
        self.tuiInvocation = HermesTUIInvocation(sessionID: nil, connection: connection)

        let normalizedPrompt = WorkflowPromptFormatter.normalizeForLaunch(workflow.prompt)
        self.initialInput = (skillCommands + [normalizedPrompt])
            .filter { !$0.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty }
            .joined(separator: "\n")
    }
}
