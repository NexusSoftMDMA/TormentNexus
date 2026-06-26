import SwiftUI

struct WorkflowsView: View {
    @EnvironmentObject private var appState: AppState
    @Binding var splitLayout: HermesSplitLayout

    @State private var searchText = ""
    @State private var editorMode: WorkflowEditorMode?
    @State private var editorDraft = WorkflowDraft()
    @State private var workflowToDelete: WorkflowPreset?
    @State private var showDeleteConfirmation = false

    var body: some View {
        HermesCollapsibleHSplitView(layout: $splitLayout, detailMinWidth: HermesSplitMetrics.WorkbenchDetail.minWidth) {
            VStack(alignment: .leading, spacing: 18) {
                HermesPageHeader(
                    title: "Workflows",
                    subtitle: "Create reusable prompt presets that open in a fresh Terminal tab on the active connection."
                ) {
                    HermesExpandableSearchField(
                        text: $searchText,
                        prompt: L10n.string("Search workflows"),
                        expandedWidth: 220,
                        focusRequestID: appState.searchFocusRequestID
                    )
                    .fixedSize(horizontal: true, vertical: false)
                }

                workflowsToolbar
                workflowsContent
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .padding(.horizontal, 20)
            .padding(.vertical, 20)
        } detail: {
            detailContent
                .hermesSplitDetailColumn(
                    minWidth: HermesSplitMetrics.WorkbenchDetail.minWidth,
                    idealWidth: HermesSplitMetrics.WorkbenchDetail.formIdealWidth
                )
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .task(id: appState.activeConnectionID) {
            appState.loadWorkflows(reset: true)
            await appState.loadSkills(reset: false)
            appState.loadWorkflows(reset: true)
        }
        .onChange(of: appState.skills) { _, newValue in
            guard editorMode != nil else { return }
            editorDraft.refreshSelectedSkills(using: newValue)
        }
        .alert(L10n.string("Remove workflow?"), isPresented: $showDeleteConfirmation) {
            Button(L10n.string("Remove"), role: .destructive) {
                guard let workflowToDelete else { return }
                appState.deleteWorkflow(workflowToDelete)
            }
            Button(L10n.string("Cancel"), role: .cancel) {}
        } message: {
            if let workflowToDelete {
                Text(L10n.string(
                    "“%@” will be removed from this Mac for the active host/profile. This cannot be undone.",
                    workflowToDelete.name
                ))
            }
        }
    }

    private var workflowsToolbar: some View {
        HStack(spacing: 10) {
            HermesCreateActionButton("New Workflow", help: "Create a reusable terminal workflow") {
                startCreating()
            }
        }
        .fixedSize(horizontal: true, vertical: false)
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private var workflowsContent: some View {
        if appState.workflows.isEmpty {
            HermesSurfacePanel {
                ContentUnavailableView(
                    L10n.string("No workflows saved"),
                    systemImage: "bolt.horizontal.circle",
                    description: Text(L10n.string("Create a reusable prompt preset for this host/profile, with optional preloaded skills."))
                )
                .frame(maxWidth: .infinity, minHeight: 300)
            }
        } else {
            HermesSurfacePanel(
                title: panelTitle,
                subtitle: "Select a workflow to inspect its prompt, assigned skills and launch readiness."
            ) {
                if filteredWorkflows.isEmpty {
                    ContentUnavailableView(
                        L10n.string("No matching workflows"),
                        systemImage: "magnifyingglass",
                        description: Text(L10n.string("Try searching by workflow name, prompt text, or skill path."))
                    )
                    .frame(maxWidth: .infinity, minHeight: 300)
                } else {
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 10) {
                            ForEach(filteredWorkflows) { workflow in
                                WorkflowCardRow(
                                    workflow: workflow,
                                    resolvedSkills: resolvedSkills(for: workflow),
                                    missingSkills: missingSkills(for: workflow),
                                    isSelected: workflow.id == appState.selectedWorkflowID && editorMode == nil
                                ) {
                                    editorMode = nil
                                    appState.selectedWorkflowID = workflow.id
                                }
                            }
                        }
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                }
            }
        }
    }

    @ViewBuilder
    private var detailContent: some View {
        if let editorMode {
            WorkflowEditorView(
                connectionKind: appState.activeConnection?.kind ?? .ssh,
                mode: editorMode,
                draft: $editorDraft,
                availableSkills: appState.skills,
                selectedMissingSkills: selectedMissingSkills,
                catalogErrorMessage: appState.skillsError,
                onCancel: {
                    self.editorMode = nil
                },
                onRemoveMissingSkill: { relativePath in
                    editorDraft.removeSkill(relativePath: relativePath)
                },
                onSave: {
                    saveEditor()
                }
            )
        } else {
            WorkflowDetailView(
                workflow: selectedWorkflow,
                resolvedSkills: selectedWorkflow.map(resolvedSkills(for:)) ?? [],
                missingSkills: selectedWorkflow.map(missingSkills(for:)) ?? [],
                catalogErrorMessage: appState.skillsError,
                onCreate: {
                    startCreating()
                },
                onEdit: {
                    startEditing()
                },
                onDelete: {
                    guard let selectedWorkflow else { return }
                    workflowToDelete = selectedWorkflow
                    showDeleteConfirmation = true
                },
                onRun: { destination in
                    guard let selectedWorkflow else { return }
                    Task { await appState.runWorkflow(selectedWorkflow, destination: destination) }
                }
            )
        }
    }

    private var filteredWorkflows: [WorkflowPreset] {
        appState.workflows.filter { $0.matchesSearch(searchText) }
    }

    private var panelTitle: String {
        let total = appState.workflows.count
        let filtered = filteredWorkflows.count

        if searchText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return L10n.string("Saved Workflows (%@)", "\(total)")
        }

        return L10n.string("Saved Workflows (%@ of %@)", "\(filtered)", "\(total)")
    }

    private var selectedWorkflow: WorkflowPreset? {
        appState.workflow(id: appState.selectedWorkflowID)
    }

    private var selectedMissingSkills: [WorkflowSkillReference] {
        let availableRelativePaths = Set(appState.skills.map(\.relativePath))
        return editorDraft.normalizedSelectedSkills.filter { !availableRelativePaths.contains($0.relativePath) }
    }

    private func resolvedSkills(for workflow: WorkflowPreset) -> [SkillSummary] {
        let availableByRelativePath = Dictionary(uniqueKeysWithValues: appState.skills.map { ($0.relativePath, $0) })
        return workflow.assignedSkills.compactMap { availableByRelativePath[$0.relativePath] }
    }

    private func missingSkills(for workflow: WorkflowPreset) -> [WorkflowSkillReference] {
        let availableRelativePaths = Set(appState.skills.map(\.relativePath))
        return workflow.assignedSkills.filter { !availableRelativePaths.contains($0.relativePath) }
    }

    private func startCreating() {
        editorDraft = WorkflowDraft()
        editorMode = .create
    }

    private func startEditing() {
        guard let selectedWorkflow else { return }
        var draft = WorkflowDraft.from(workflow: selectedWorkflow)
        draft.refreshSelectedSkills(using: appState.skills)
        editorDraft = draft
        editorMode = .edit(selectedWorkflow)
    }

    private func saveEditor() {
        let didSave: Bool

        switch editorMode {
        case .create:
            didSave = appState.createWorkflow(editorDraft)
        case .edit(let workflow):
            didSave = appState.updateWorkflow(workflow, draft: editorDraft)
        case .none:
            didSave = false
        }

        if didSave {
            editorMode = nil
        }
    }
}

private enum WorkflowEditorMode {
    case create
    case edit(WorkflowPreset)

    var title: String {
        switch self {
        case .create:
            return "New Workflow"
        case .edit:
            return "Edit Workflow"
        }
    }

    var actionTitle: String {
        switch self {
        case .create:
            return "Create Workflow"
        case .edit:
            return "Save Changes"
        }
    }
}

private struct WorkflowCardRow: View {
    let workflow: WorkflowPreset
    let resolvedSkills: [SkillSummary]
    let missingSkills: [WorkflowSkillReference]
    let isSelected: Bool
    let onSelect: () -> Void

    private var cardFillColor: Color {
        isSelected ? HermesTheme.selectedFill : HermesTheme.rowFill
    }

    private var cardStrokeColor: Color {
        isSelected ? HermesTheme.selectedStroke : HermesTheme.subtleStroke
    }

    private var previewSkills: [WorkflowSkillReference] {
        Array(workflow.assignedSkills.prefix(2))
    }

    private var overflowSkillCount: Int {
        max(0, workflow.assignedSkills.count - previewSkills.count)
    }

    private var missingRelativePaths: Set<String> {
        Set(missingSkills.map(\.relativePath))
    }

    var body: some View {
        Button(action: onSelect) {
            VStack(alignment: .leading, spacing: 10) {
                HStack(alignment: .top, spacing: 10) {
                    VStack(alignment: .leading, spacing: 5) {
                        Text(workflow.name)
                            .font(.headline)
                            .foregroundStyle(.primary)
                            .multilineTextAlignment(.leading)

                        Text(workflow.promptPreview)
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .multilineTextAlignment(.leading)
                            .lineLimit(3)
                    }

                    Spacer(minLength: 12)

                    if missingSkills.isEmpty {
                        HermesBadge(text: "Runnable", tint: .green)
                    } else {
                        HermesBadge(text: "Missing skills", tint: .orange)
                    }
                }

                HStack(spacing: 6) {
                    WorkflowMetricBadge(text: L10n.string("%@ skills", "\(workflow.assignedSkills.count)"))

                    ForEach(previewSkills) { skillReference in
                        HermesBadge(
                            text: skillReference.resolvedName,
                            tint: missingRelativePaths.contains(skillReference.relativePath) ? .orange : .secondary
                        )
                    }

                    if overflowSkillCount > 0 {
                        WorkflowMetricBadge(text: "+\(overflowSkillCount)")
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(14)
            .background(cardFillColor, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .stroke(cardStrokeColor, lineWidth: 1)
            }
        }
        .buttonStyle(.plain)
    }
}

private struct WorkflowDetailView: View {
    let workflow: WorkflowPreset?
    let resolvedSkills: [SkillSummary]
    let missingSkills: [WorkflowSkillReference]
    let catalogErrorMessage: String?
    let onCreate: () -> Void
    let onEdit: () -> Void
    let onDelete: () -> Void
    let onRun: (WorkflowRunDestination) -> Void

    var body: some View {
        if let workflow {
            ScrollView {
                VStack(alignment: .leading, spacing: 18) {
                    HermesSurfacePanel {
                        VStack(alignment: .leading, spacing: 16) {
                            HStack(alignment: .top, spacing: 12) {
                                VStack(alignment: .leading, spacing: 6) {
                                    Text(workflow.name)
                                        .font(.title2.weight(.semibold))

                                    Text(L10n.string("Launch this preset in a fresh Terminal tab on the active host/profile."))
                                        .font(.callout)
                                        .foregroundStyle(.secondary)
                                }

                                Spacer(minLength: 16)

                                if missingSkills.isEmpty {
                                    HermesBadge(text: "Runnable", tint: .green)
                                } else {
                                    HermesBadge(text: "Missing skills", tint: .orange)
                                }
                            }

                            if let catalogErrorMessage,
                               resolvedSkills.isEmpty,
                               missingSkills.count == workflow.assignedSkills.count {
                                WorkflowInlineNotice(
                                    title: "Unable to refresh the skill catalog",
                                    message: catalogErrorMessage
                                )
                            } else if !missingSkills.isEmpty {
                                WorkflowInlineNotice(
                                    title: "Some saved skills are unavailable",
                                    message: L10n.string(
                                        "This workflow references skills that are not available on the active host/profile: %@",
                                        missingSkills.map(\.relativePath).joined(separator: ", ")
                                    )
                                )
                            }

                            HStack(spacing: 10) {
                                Menu {
                                    Button {
                                        onRun(.terminal)
                                    } label: {
                                        Label(L10n.string("Run in terminal"), systemImage: "terminal")
                                    }

                                    Button {
                                        onRun(.chat)
                                    } label: {
                                        Label(L10n.string("Run in chat"), systemImage: "bubble.left.and.bubble.right")
                                    }
                                } label: {
                                    Label(L10n.string("Run Workflow"), systemImage: "play.fill")
                                }
                                .buttonStyle(.borderedProminent)
                                .disabled(!missingSkills.isEmpty)

                                Button(L10n.string("Edit Workflow")) {
                                    onEdit()
                                }
                                .buttonStyle(.bordered)

                                Button(L10n.string("Remove")) {
                                    onDelete()
                                }
                                .buttonStyle(.bordered)
                            }
                        }
                    }

                    HermesSurfacePanel(
                        title: "Prompt",
                        subtitle: "This prompt is sent as the first turn when the workflow starts in Terminal."
                    ) {
                        Text(workflow.prompt)
                            .font(.body.monospaced())
                            .textSelection(.enabled)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(14)
                            .background(HermesTheme.insetFill, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                            .overlay {
                                RoundedRectangle(cornerRadius: 12, style: .continuous)
                                    .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
                            }
                    }

                    HermesSurfacePanel(
                        title: "Assigned Skills",
                        subtitle: "Skills are preloaded at launch using the Hermes CLI skill flags."
                    ) {
                        VStack(alignment: .leading, spacing: 12) {
                            if !resolvedSkills.isEmpty {
                                WorkflowSkillList(
                                    title: "Available on this host",
                                    references: resolvedSkills.map { WorkflowSkillReference(skill: $0) }
                                )
                            }

                            if !missingSkills.isEmpty {
                                WorkflowSkillList(title: "Unavailable on this host", references: missingSkills, tint: .orange)
                            }
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                    }
                }
                .padding(20)
            }
        } else {
            HermesSurfacePanel {
                ContentUnavailableView(
                    L10n.string("No workflow selected"),
                    systemImage: "bolt.horizontal.circle",
                    description: Text(L10n.string("Choose a workflow from the list or create a new reusable preset for this host/profile."))
                )
                .frame(maxWidth: .infinity, minHeight: 320)
            }
        }
    }
}

private struct WorkflowEditorView: View {
    let connectionKind: ConnectionKind
    let mode: WorkflowEditorMode
    @Binding var draft: WorkflowDraft
    let availableSkills: [SkillSummary]
    let selectedMissingSkills: [WorkflowSkillReference]
    let catalogErrorMessage: String?
    let onCancel: () -> Void
    let onRemoveMissingSkill: (String) -> Void
    let onSave: () -> Void

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                HermesSurfacePanel {
                    VStack(alignment: .leading, spacing: 16) {
                        Text(L10n.string(mode.title))
                            .font(.title2.weight(.semibold))

                        Text(L10n.string(
                            connectionKind == .local
                                ? "Workflows stay local to this Mac and launch a fresh local Terminal tab when run."
                                : "Workflows stay local to this Mac and launch a fresh remote Terminal tab when run."
                        ))
                            .font(.callout)
                            .foregroundStyle(.secondary)

                        if let validationError = draft.validationError {
                            WorkflowInlineNotice(
                                title: "Finish the required fields",
                                message: L10n.string(validationError)
                            )
                        } else if let catalogErrorMessage,
                                  availableSkills.isEmpty {
                            WorkflowInlineNotice(
                                title: "Unable to load the skill catalog",
                                message: catalogErrorMessage
                            )
                        }

                        HStack(spacing: 10) {
                            Button(L10n.string(mode.actionTitle)) {
                                onSave()
                            }
                            .buttonStyle(.borderedProminent)

                            Button(L10n.string("Cancel")) {
                                onCancel()
                            }
                            .buttonStyle(.bordered)
                        }
                    }
                }

                HermesSurfacePanel(
                    title: "Workflow Details",
                    subtitle: "Give this preset a stable name and the prompt that should seed the terminal session."
                ) {
                    VStack(alignment: .leading, spacing: 14) {
                        WorkflowFormField(label: "Name") {
                            TextField(L10n.string("Nightly release audit"), text: $draft.name)
                                .textFieldStyle(.roundedBorder)
                        }

                        WorkflowFormField(label: "Prompt") {
                            TextEditor(text: $draft.prompt)
                                .font(.body.monospaced())
                                .frame(minHeight: 180)
                                .padding(10)
                                .background(HermesTheme.insetFill, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                                .overlay {
                                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                                        .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
                                }
                        }
                    }
                }

                HermesSurfacePanel(
                title: "Assigned Skills",
                    subtitle: "Select any skills you want to preload for the whole terminal session."
                ) {
                    VStack(alignment: .leading, spacing: 14) {
                        if !selectedMissingSkills.isEmpty {
                            WorkflowInlineNotice(
                                title: "Unavailable saved skills",
                                message: L10n.string("These saved skills are not available on the active host/profile anymore. Remove them or keep editing while they stay unavailable.")
                            )

                            WorkflowMissingSkillEditorList(
                                references: selectedMissingSkills,
                                onRemove: onRemoveMissingSkill
                            )
                        }

                        if availableSkills.isEmpty {
                            Text(L10n.string("No discovered skills are available for selection right now."))
                                .foregroundStyle(.secondary)
                        } else {
                            LazyVStack(alignment: .leading, spacing: 10) {
                                ForEach(availableSkills) { skill in
                                    WorkflowSkillToggleRow(
                                        skill: skill,
                                        isSelected: Binding(
                                            get: { draft.containsSkill(relativePath: skill.relativePath) },
                                            set: { draft.setSkillSelected($0, skill: skill) }
                                        )
                                    )
                                }
                            }
                        }
                    }
                }
            }
            .padding(20)
        }
    }
}

private struct WorkflowSkillToggleRow: View {
    let skill: SkillSummary
    @Binding var isSelected: Bool

    var body: some View {
        Toggle(isOn: $isSelected) {
            VStack(alignment: .leading, spacing: 8) {
                Text(skill.resolvedName)
                    .font(.headline)
                    .lineLimit(1)

                Text(skill.relativePath)
                    .font(.caption.monospaced())
                    .foregroundStyle(.secondary)
                    .lineLimit(1)

                if let description = skill.trimmedDescription {
                    Text(description)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                        .fixedSize(horizontal: false, vertical: true)
                } else {
                    Color.clear
                        .frame(height: 38)
                }

                Spacer(minLength: 0)
            }
            .frame(maxWidth: .infinity, minHeight: 92, alignment: .topLeading)
        }
        .toggleStyle(.checkbox)
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(14)
        .background(HermesTheme.rowFill, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
        }
    }
}

private struct WorkflowSkillList: View {
    let title: String
    let references: [WorkflowSkillReference]
    var tint: Color = .secondary

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(L10n.string(title))
                .font(.headline)

            ForEach(references) { reference in
                HStack(alignment: .top, spacing: 10) {
                    Image(systemName: tint == .orange ? "exclamationmark.triangle.fill" : "checkmark.circle.fill")
                        .foregroundStyle(tint)
                        .padding(.top, 2)

                    VStack(alignment: .leading, spacing: 4) {
                        Text(reference.resolvedName)
                            .font(.subheadline.weight(.semibold))

                        Text(reference.relativePath)
                            .font(.caption.monospaced())
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

private struct WorkflowMissingSkillEditorList: View {
    let references: [WorkflowSkillReference]
    let onRemove: (String) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            ForEach(references) { reference in
                HStack(alignment: .top, spacing: 10) {
                    VStack(alignment: .leading, spacing: 4) {
                        Text(reference.resolvedName)
                            .font(.subheadline.weight(.semibold))

                        Text(reference.relativePath)
                            .font(.caption.monospaced())
                            .foregroundStyle(.secondary)
                    }

                    Spacer(minLength: 12)

                    Button(L10n.string("Remove")) {
                        onRemove(reference.relativePath)
                    }
                    .buttonStyle(.bordered)
                    .controlSize(.small)
                }
                .padding(12)
                .background(HermesTheme.warningFill, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                .overlay {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .strokeBorder(HermesTheme.warningStroke, lineWidth: 1)
                }
            }
        }
    }
}

private struct WorkflowInlineNotice: View {
    let title: String
    let message: String

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(L10n.string(title))
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(HermesTheme.warningForeground)

            Text(message)
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(HermesTheme.warningFill, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay {
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(HermesTheme.warningStroke, lineWidth: 1)
        }
    }
}

private struct WorkflowFormField<Content: View>: View {
    let label: String
    let content: Content

    init(label: String, @ViewBuilder content: () -> Content) {
        self.label = label
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(L10n.string(label))
                .font(.headline)

            content
        }
    }
}

private struct WorkflowMetricBadge: View {
    let text: String

    var body: some View {
        Text(text)
            .font(.caption2.weight(.semibold))
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(HermesTheme.rowFill, in: Capsule())
            .overlay {
                Capsule()
                    .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
            }
    }
}
