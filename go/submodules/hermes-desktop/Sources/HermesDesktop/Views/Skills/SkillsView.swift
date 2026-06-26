import SwiftUI

struct SkillsView: View {
    @EnvironmentObject private var appState: AppState
    @Binding var splitLayout: HermesSplitLayout
    @State private var searchText = ""
    @State private var editorMode: SkillEditorMode?
    @State private var editorDraft = SkillDraft()
    @State private var rawMarkdownContent = ""

    var body: some View {
        HermesCollapsibleHSplitView(layout: $splitLayout, detailMinWidth: HermesSplitMetrics.WorkbenchDetail.minWidth) {
            VStack(alignment: .leading, spacing: 18) {
                HermesPageHeader(
                    title: "Skills",
                    subtitle: "Browse the Hermes skill library discovered on the active host."
                ) {
                    HermesExpandableSearchField(
                        text: $searchText,
                        prompt: L10n.string("Search skills"),
                        expandedWidth: 220,
                        focusRequestID: appState.searchFocusRequestID
                    )
                    .fixedSize(horizontal: true, vertical: false)
                }

                skillsToolbar
                skillsContent
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .padding(.horizontal, 20)
            .padding(.vertical, 20)
        } detail: {
            detailContent
                .hermesSplitDetailColumn(
                    minWidth: HermesSplitMetrics.WorkbenchDetail.minWidth,
                    idealWidth: HermesSplitMetrics.WorkbenchDetail.standardIdealWidth
                )
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .task(id: appState.activeConnectionID) {
            if appState.skills.isEmpty {
                await appState.loadSkills(reset: true)
            }
        }
    }

    @ViewBuilder
    private var skillsContent: some View {
        skillsPanel
    }

    @ViewBuilder
    private var skillsPanel: some View {
        if appState.isLoadingSkills && appState.skills.isEmpty {
            HermesSurfacePanel {
                HermesLoadingState(
                    label: "Loading skills…",
                    minHeight: 300
                )
            }
        } else if let error = appState.skillsError, appState.skills.isEmpty {
            HermesSurfacePanel {
                ContentUnavailableView(
                    L10n.string("Unable to load skills"),
                    systemImage: "exclamationmark.triangle",
                    description: Text(error)
                )
                .frame(maxWidth: .infinity, minHeight: 300)
            }
        } else if appState.skills.isEmpty {
            HermesSurfacePanel {
                ContentUnavailableView(
                    L10n.string("No skills found"),
                    systemImage: "book.closed",
                    description: Text(noSkillsDescription)
                )
                .frame(maxWidth: .infinity, minHeight: 300)
            }
        } else {
            HermesSurfacePanel(
                title: panelTitle,
                subtitle: "Select a skill to inspect its metadata, related assets and full SKILL.md content."
            ) {
                if filteredSkills.isEmpty {
                    ContentUnavailableView(
                        L10n.string("No matching skills"),
                        systemImage: "magnifyingglass",
                        description: Text(L10n.string("Try searching by skill name or category."))
                    )
                    .frame(maxWidth: .infinity, minHeight: 300)
                } else {
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 10) {
                            ForEach(filteredSkills) { skill in
                                SkillCardRow(
                                    skill: skill,
                                    isSelected: skill.id == appState.selectedSkillID
                                ) {
                                    Task {
                                        await appState.loadSkillDetail(summary: skill)
                                    }
                                }
                            }
                        }
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                }
            }
            .overlay(alignment: .topTrailing) {
                if appState.isLoadingSkills && !appState.isRefreshingSkills && !appState.skills.isEmpty {
                    HermesLoadingOverlay()
                        .padding(18)
                }
            }
        }
    }

    private var noSkillsDescription: String {
        if appState.activeConnection?.kind == .local {
            return L10n.string("No readable SKILL.md files were discovered in this Mac’s real Hermes skill roots.")
        }
        return L10n.string("No readable SKILL.md files were discovered in the Hermes skill roots for this SSH target.")
    }

    private var skillsToolbar: some View {
        HStack(spacing: 10) {
            HermesCreateActionButton("New Skill") {
                startCreating()
            }
            .disabled(appState.isSavingSkillDraft || appState.isLoadingSkills)
        }
        .fixedSize(horizontal: true, vertical: false)
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var panelTitle: String {
        let total = appState.skills.count
        let filtered = filteredSkills.count

        if searchText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return L10n.string("Discovered Skills (%@)", "\(total)")
        }

        return L10n.string("Discovered Skills (%@ of %@)", "\(filtered)", "\(total)")
    }

    private var filteredSkills: [SkillSummary] {
        appState.skills.filter { $0.matchesSearch(searchText) }
    }

    private var selectedSkill: SkillSummary? {
        guard let selectedSkillID = appState.selectedSkillID else { return nil }
        return appState.skills.first(where: { $0.id == selectedSkillID })
    }

    @ViewBuilder
    private var detailContent: some View {
        if let editorMode {
            SkillEditorView(
                mode: editorMode,
                draft: $editorDraft,
                rawMarkdownContent: $rawMarkdownContent,
                detail: appState.selectedSkillDetail,
                errorMessage: appState.skillsError,
                isSaving: appState.isSavingSkillDraft,
                onCancel: {
                    self.editorMode = nil
                },
                onSave: {
                    await saveEditor()
                }
            )
        } else {
            SkillDetailView(
                summary: selectedSkill,
                detail: appState.selectedSkillDetail,
                errorMessage: appState.skillsError,
                isLoading: appState.isLoadingSkillDetail,
                onCreate: {
                    startCreating()
                },
                onEdit: {
                    startEditing()
                }
            )
        }
    }
    private func startCreating() {
        var draft = SkillDraft()
        draft.refreshSuggestedSlug()
        editorDraft = draft
        rawMarkdownContent = draft.generatedMarkdown
        editorMode = .create
    }

    private func startEditing() {
        guard let detail = appState.selectedSkillDetail, !detail.isReadOnly else { return }
        editorDraft = SkillDraft.from(detail: detail)
        rawMarkdownContent = detail.markdownContent
        editorMode = .edit
    }

    private func saveEditor() async {
        switch editorMode {
        case .create:
            let didSave = await appState.createSkill(editorDraft)
            if didSave {
                editorMode = nil
            }
        case .edit:
            guard let detail = appState.selectedSkillDetail else { return }
            let didSave = await appState.updateSkill(
                detail,
                markdownContent: rawMarkdownContent,
                ensureReferencesFolder: editorDraft.includeReferencesFolder,
                ensureScriptsFolder: editorDraft.includeScriptsFolder,
                ensureTemplatesFolder: editorDraft.includeTemplatesFolder
            )
            if didSave {
                editorMode = nil
            }
        case nil:
            break
        }
    }
}

private struct SkillCardRow: View {
    let skill: SkillSummary
    let isSelected: Bool
    let onSelect: () -> Void

    private var cardFillColor: Color {
        isSelected ? HermesTheme.selectedFill : HermesTheme.rowFill
    }

    private var cardStrokeColor: Color {
        isSelected ? HermesTheme.selectedStroke : HermesTheme.subtleStroke
    }

    var body: some View {
        Button(action: onSelect) {
            VStack(alignment: .leading, spacing: 10) {
                HStack(alignment: .top, spacing: 10) {
                    VStack(alignment: .leading, spacing: 5) {
                        HStack(spacing: 8) {
                            Text(skill.resolvedName)
                                .font(.headline)
                                .foregroundStyle(.primary)
                                .multilineTextAlignment(.leading)

                            if !skill.source.isLocal {
                                HermesBadge(text: skill.sourceLabel, tint: .secondary)
                            }
                        }

                        Text(skill.relativePath)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }

                    Spacer(minLength: 12)

                    if let category = skill.category {
                        HermesBadge(text: category, tint: .secondary)
                    }
                }

                if let description = skill.trimmedDescription {
                    Text(description)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                } else {
                    Text(L10n.string("No description in frontmatter"))
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .italic()
                }

                if !skill.previewBadges.isEmpty {
                    SkillCardBadgeScroller(
                        badges: skill.previewBadges,
                        backgroundColor: cardFillColor
                    )
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .fill(cardFillColor)
            )
            .overlay {
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .strokeBorder(cardStrokeColor, lineWidth: 1)
            }
        }
        .buttonStyle(.plain)
    }
}

private struct SkillCardBadgeScroller: View {
    let badges: [SkillPreviewBadge]
    let backgroundColor: Color

    @State private var contentWidth: CGFloat = 0
    @State private var viewportWidth: CGFloat = 0

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(badges) { badge in
                    HermesBadge(
                        text: badge.text,
                        tint: badge.tint,
                        isMonospaced: badge.isMonospaced
                    )
                }
            }
            .background(
                GeometryReader { proxy in
                    Color.clear
                        .preference(key: SkillBadgeContentWidthKey.self, value: proxy.size.width)
                }
            )
        }
        .background(
            GeometryReader { proxy in
                Color.clear
                    .preference(key: SkillBadgeViewportWidthKey.self, value: proxy.size.width)
            }
        )
        .onPreferenceChange(SkillBadgeContentWidthKey.self) { contentWidth = $0 }
        .onPreferenceChange(SkillBadgeViewportWidthKey.self) { viewportWidth = $0 }
        .overlay(alignment: .trailing) {
            if contentWidth > viewportWidth + 1 {
                LinearGradient(
                    colors: [.clear, backgroundColor],
                    startPoint: .leading,
                    endPoint: .trailing
                )
                .frame(width: 34)
                .allowsHitTesting(false)
            }
        }
    }
}

private struct SkillPreviewBadge: Identifiable {
    let id: String
    let text: String
    let tint: Color
    var isMonospaced = false
}

private struct SkillBadgeContentWidthKey: PreferenceKey {
    static let defaultValue: CGFloat = 0

    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

private struct SkillBadgeViewportWidthKey: PreferenceKey {
    static let defaultValue: CGFloat = 0

    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

private extension SkillSummary {
    var previewBadges: [SkillPreviewBadge] {
        var badges: [SkillPreviewBadge] = []

        if let version, !version.isEmpty {
            badges.append(
                SkillPreviewBadge(
                    id: "version-\(version)",
                    text: version,
                    tint: .secondary,
                    isMonospaced: true
                )
            )
        }

        for tag in tags {
            badges.append(
                SkillPreviewBadge(
                    id: "tag-\(tag)",
                    text: tag,
                    tint: .accentColor
                )
            )
        }

        for relatedSkill in relatedSkills {
            badges.append(
                SkillPreviewBadge(
                    id: "related-\(relatedSkill)",
                    text: relatedSkill,
                    tint: .secondary,
                    isMonospaced: true
                )
            )
        }

        for feature in featureBadges {
            badges.append(
                SkillPreviewBadge(
                    id: "feature-\(feature.id)",
                    text: feature.title,
                    tint: feature.color
                )
            )
        }

        return badges
    }
}
