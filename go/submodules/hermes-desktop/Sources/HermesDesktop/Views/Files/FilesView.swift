import SwiftUI

struct FilesView: View {
    @EnvironmentObject private var appState: AppState
    @Binding var splitLayout: HermesSplitLayout
    @State private var pendingWorkspaceFileID: String?
    @State private var bookmarkPendingRemoval: UUID?
    @State private var collapsedBookmarkGroupIDs: Set<String> = []
    @State private var showBrowserSheet = false
    @State private var showDiscardFileAlert = false
    @State private var showReloadDiscardAlert = false
    @State private var showRemoveBookmarkAlert = false

    var body: some View {
        HermesCollapsibleHSplitView(layout: $splitLayout, detailMinWidth: HermesSplitMetrics.WorkbenchDetail.minWidth) {
            VStack(alignment: .leading, spacing: 18) {
                HermesPageHeader(
                    title: "Files",
                    subtitle: "Read and edit files on the active Hermes machine."
                )

                filesToolbar
                libraryPanel
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .padding(.horizontal, 20)
            .padding(.vertical, 20)
        } detail: {
            editorPane
                .hermesSplitDetailColumn(
                    minWidth: HermesSplitMetrics.WorkbenchDetail.minWidth,
                    idealWidth: HermesSplitMetrics.WorkbenchDetail.editorIdealWidth
                )
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .task(id: selectedFileLoadTaskID) {
            await appState.loadSelectedWorkspaceFile()
        }
        .sheet(isPresented: $showBrowserSheet) {
            WorkspaceFileBrowserSheet()
                .environmentObject(appState)
        }
        .alert(L10n.string("Discard unsaved edits in this file?"), isPresented: $showDiscardFileAlert) {
            Button(L10n.string("Discard"), role: .destructive) {
                if let currentFileID {
                    appState.discardWorkspaceFile(currentFileID)
                }
                if let pendingWorkspaceFileID {
                    appState.selectWorkspaceFile(pendingWorkspaceFileID)
                }
                pendingWorkspaceFileID = nil
            }
            Button(L10n.string("Stay"), role: .cancel) {
                pendingWorkspaceFileID = nil
            }
        } message: {
            Text(L10n.string("Switching away will drop the unsaved edits in the current file."))
        }
        .alert(L10n.string("Reload saved file and discard edits?"), isPresented: $showReloadDiscardAlert) {
            Button(L10n.string("Reload"), role: .destructive) {
                if let selectedReference {
                    Task {
                        await appState.loadWorkspaceFile(selectedReference, forceReload: true)
                    }
                }
            }
            Button(L10n.string("Cancel"), role: .cancel) {}
        } message: {
            Text(L10n.string("This will replace your unsaved changes with the currently saved file content."))
        }
        .alert(L10n.string("Remove this bookmark?"), isPresented: $showRemoveBookmarkAlert) {
            Button(L10n.string("Remove"), role: .destructive) {
                if let bookmarkPendingRemoval {
                    appState.removeWorkspaceFileBookmark(id: bookmarkPendingRemoval)
                }
                bookmarkPendingRemoval = nil
            }
            Button(L10n.string("Cancel"), role: .cancel) {
                bookmarkPendingRemoval = nil
            }
        } message: {
            Text(L10n.string("The saved file stays untouched."))
        }
    }

    private var filesToolbar: some View {
        HStack(spacing: 10) {
            HermesCreateActionButton("Add File") {
                showBrowserSheet = true
            }
        }
        .fixedSize(horizontal: true, vertical: false)
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var libraryPanel: some View {
        HermesSurfacePanel(
            title: "Library",
            subtitle: "Pinned Hermes files and your bookmarks."
        ) {
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 16) {
                    fileGroup(title: "Canonical", references: appState.canonicalWorkspaceFileReferences)

                    if appState.bookmarkedWorkspaceFileGroups.isEmpty {
                        emptyBookmarks
                    } else {
                        bookmarkGroups(appState.bookmarkedWorkspaceFileGroups)
                    }
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        }
    }

    private func fileGroup(title: String, references: [WorkspaceFileReference]) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(L10n.string(title))
                .font(.caption.weight(.semibold))
                .foregroundStyle(.secondary)
                .textCase(.uppercase)

            LazyVStack(alignment: .leading, spacing: 10) {
                ForEach(references) { reference in
                    WorkspaceFileCardRow(
                        reference: reference,
                        subtitle: reference.subtitle,
                        isSelected: reference.id == currentFileID,
                        isDirty: appState.workspaceFileDocument(for: reference.id)?.isDirty == true,
                        onSelect: {
                            select(reference)
                        },
                        onRemove: reference.bookmarkID.map { bookmarkID in
                            {
                                bookmarkPendingRemoval = bookmarkID
                                showRemoveBookmarkAlert = true
                            }
                        }
                    )
                }
            }
        }
    }

    private func bookmarkGroups(_ groups: [WorkspaceFileBookmarkGroup]) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(L10n.string("Bookmarks"))
                .font(.caption.weight(.semibold))
                .foregroundStyle(.secondary)
                .textCase(.uppercase)

            LazyVStack(alignment: .leading, spacing: 10) {
                ForEach(groups) { group in
                    bookmarkFolderGroup(group)
                }
            }
        }
    }

    private func bookmarkFolderGroup(_ group: WorkspaceFileBookmarkGroup) -> some View {
        DisclosureGroup(isExpanded: bookmarkGroupExpansionBinding(for: group.id)) {
            LazyVStack(alignment: .leading, spacing: 8) {
                ForEach(group.references) { reference in
                    WorkspaceFileCardRow(
                        reference: reference,
                        subtitle: groupedBookmarkSubtitle(for: reference),
                        isSelected: reference.id == currentFileID,
                        isDirty: appState.workspaceFileDocument(for: reference.id)?.isDirty == true,
                        onSelect: {
                            select(reference)
                        },
                        onRemove: reference.bookmarkID.map { bookmarkID in
                            {
                                bookmarkPendingRemoval = bookmarkID
                                showRemoveBookmarkAlert = true
                            }
                        }
                    )
                }
            }
            .padding(.top, 10)
        } label: {
            HStack(spacing: 9) {
                Image(systemName: "folder")
                    .foregroundStyle(.secondary)
                    .frame(width: 18)

                VStack(alignment: .leading, spacing: 3) {
                    Text(group.title)
                        .font(.subheadline.weight(.semibold))
                        .foregroundStyle(.primary)
                        .lineLimit(1)

                    Text(group.directoryPath)
                        .font(.system(.caption, design: .monospaced))
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }

                Spacer(minLength: 8)

                Text("\(group.references.count)")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
                    .padding(.horizontal, 7)
                    .padding(.vertical, 3)
                    .background(HermesTheme.rowFill, in: Capsule())
            }
            .contentShape(Rectangle())
        }
        .tint(.secondary)
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .fill(HermesTheme.rowFill)
        )
        .overlay {
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
        }
    }

    private func bookmarkGroupExpansionBinding(for groupID: String) -> Binding<Bool> {
        Binding {
            !collapsedBookmarkGroupIDs.contains(groupID)
        } set: { isExpanded in
            if isExpanded {
                collapsedBookmarkGroupIDs.remove(groupID)
            } else {
                collapsedBookmarkGroupIDs.insert(groupID)
            }
        }
    }

    private func groupedBookmarkSubtitle(for reference: WorkspaceFileReference) -> String? {
        let filename = WorkspaceFileBookmark.displayTitle(for: reference.remotePath)
        return filename == reference.title ? nil : filename
    }

    private var emptyBookmarks: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(L10n.string("Bookmarks"))
                .font(.caption.weight(.semibold))
                .foregroundStyle(.secondary)
                .textCase(.uppercase)

            HermesInsetSurface {
                HStack(alignment: .top, spacing: 10) {
                    Image(systemName: "bookmark")
                        .foregroundStyle(.secondary)

                    VStack(alignment: .leading, spacing: 5) {
                        Text(L10n.string("No files added yet"))
                            .font(.subheadline.weight(.semibold))

                        Text(L10n.string("Add files you want to revisit."))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
    }

    private var editorPane: some View {
        Group {
            if let selectedReference {
                WorkspaceFileEditorPane(
                    reference: selectedReference,
                    document: currentDocument,
                    text: editorBinding,
                    onReload: {
                        if currentDocument?.isDirty == true {
                            showReloadDiscardAlert = true
                        } else {
                            Task {
                                await appState.loadWorkspaceFile(selectedReference, forceReload: true)
                            }
                        }
                    },
                    onSave: {
                        Task {
                            await appState.saveWorkspaceFile(fileID: selectedReference.id)
                        }
                    },
                    onRemove: selectedReference.bookmarkID.map { bookmarkID in
                        {
                            bookmarkPendingRemoval = bookmarkID
                            showRemoveBookmarkAlert = true
                        }
                    }
                )
            } else {
                ScrollView {
                    HermesSurfacePanel {
                        ContentUnavailableView(
                            L10n.string("No File Selected"),
                            systemImage: "doc.text.magnifyingglass",
                            description: Text(L10n.string("Choose a file from the library."))
                        )
                        .frame(maxWidth: .infinity, minHeight: 320)
                    }
                    .padding(.horizontal, 24)
                    .padding(.vertical, 22)
                }
            }
        }
    }

    private func select(_ reference: WorkspaceFileReference) {
        guard reference.id != currentFileID else { return }

        if currentDocument?.isDirty == true {
            pendingWorkspaceFileID = reference.id
            showDiscardFileAlert = true
        } else {
            appState.selectWorkspaceFile(reference.id)
        }
    }

    private var editorBinding: Binding<String> {
        Binding {
            guard let currentFileID else { return "" }
            return appState.workspaceFileDocument(for: currentFileID)?.content ?? ""
        } set: { newValue in
            guard let currentFileID else { return }
            appState.updateWorkspaceFile(currentFileID, content: newValue)
        }
    }

    private var selectedReference: WorkspaceFileReference? {
        appState.selectedWorkspaceFileReference
    }

    private var currentFileID: String? {
        selectedReference?.id
    }

    private var currentDocument: FileEditorDocument? {
        guard let currentFileID else { return nil }
        return appState.workspaceFileDocument(for: currentFileID)
    }

    private var selectedFileLoadTaskID: String {
        "\(appState.activeConnectionID?.uuidString ?? "none")|\(appState.selectedWorkspaceFileID)"
    }
}

private struct WorkspaceFileCardRow: View {
    let reference: WorkspaceFileReference
    let subtitle: String?
    let isSelected: Bool
    let isDirty: Bool
    let onSelect: () -> Void
    let onRemove: (() -> Void)?

    var body: some View {
        Button(action: onSelect) {
            VStack(alignment: .leading, spacing: 9) {
                HStack(alignment: .top, spacing: 10) {
                    Image(systemName: reference.systemImage)
                        .foregroundStyle(reference.isRemovable ? Color.accentColor : Color.secondary)
                        .frame(width: 18)

                    VStack(alignment: .leading, spacing: 5) {
                        HStack(spacing: 7) {
                            Text(reference.title)
                                .font(.headline)
                                .foregroundStyle(.primary)
                                .lineLimit(1)

                            if isDirty {
                                HermesBadge(text: "Unsaved", tint: .orange)
                            }
                        }

                        if let subtitle, !subtitle.isEmpty {
                            Text(subtitle)
                                .font(.system(.caption, design: .monospaced))
                                .foregroundStyle(.secondary)
                                .lineLimit(1)
                                .truncationMode(.middle)
                        }
                    }

                    Spacer(minLength: 10)

                    if onRemove != nil {
                        Image(systemName: "bookmark.fill")
                            .font(.caption.weight(.semibold))
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .fill(isSelected ? HermesTheme.selectedFill : HermesTheme.rowFill)
            )
            .overlay {
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .strokeBorder(isSelected ? HermesTheme.selectedStroke : HermesTheme.subtleStroke, lineWidth: 1)
            }
        }
        .buttonStyle(.plain)
        .contextMenu {
            if let onRemove {
                Button(L10n.string("Remove Bookmark"), role: .destructive, action: onRemove)
            }
        }
    }
}

private struct WorkspaceFileEditorPane: View {
    let reference: WorkspaceFileReference
    let document: FileEditorDocument?
    @Binding var text: String
    let onReload: () -> Void
    let onSave: () -> Void
    let onRemove: (() -> Void)?

    private var isDirty: Bool {
        document?.isDirty == true
    }

    private var isLoading: Bool {
        document?.isLoading == true
    }

    private var hasLoaded: Bool {
        document?.hasLoaded == true
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            headerPanel

            if let errorMessage = document?.errorMessage {
                HermesSurfacePanel {
                    Text(errorMessage)
                        .font(.subheadline)
                        .foregroundStyle(.red)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .textSelection(.enabled)
                }
            }

            editorPanel
        }
        .padding(.horizontal, 24)
        .padding(.vertical, 22)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }

    private var headerPanel: some View {
        HermesSurfacePanel {
            VStack(alignment: .leading, spacing: 16) {
                HStack(alignment: .top, spacing: 12) {
                    VStack(alignment: .leading, spacing: 7) {
                        HStack(spacing: 8) {
                            Image(systemName: reference.systemImage)
                                .foregroundStyle(.secondary)

                            Text(reference.title)
                                .font(.title2)
                                .fontWeight(.semibold)
                                .lineLimit(1)
                        }

                        Text(reference.remotePath)
                            .font(.system(.caption, design: .monospaced))
                            .foregroundStyle(.secondary)
                            .lineLimit(2)
                            .truncationMode(.middle)
                            .textSelection(.enabled)
                    }

                    Spacer(minLength: 12)

                    if isDirty {
                        HermesBadge(text: "Unsaved", tint: .orange)
                    } else if let lastSavedAt = document?.lastSavedAt {
                        Text(L10n.string("Saved %@", DateFormatters.relativeFormatter().localizedString(for: lastSavedAt, relativeTo: .now)))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                ViewThatFits(in: .horizontal) {
                    HStack(spacing: 10) {
                        actionButtons
                    }

                    VStack(alignment: .leading, spacing: 10) {
                        actionButtons
                    }
                }
            }
        }
    }

    private var actionButtons: some View {
        Group {
            Button(L10n.string("Reload"), action: onReload)
            .disabled(isLoading)

            Button(L10n.string("Save"), action: onSave)
            .buttonStyle(.borderedProminent)
            .disabled(!isDirty || isLoading || !hasLoaded)

            if let onRemove {
                Button(L10n.string("Remove Bookmark"), role: .destructive, action: onRemove)
                .disabled(isLoading)
            }

            if isLoading {
                ProgressView()
                    .controlSize(.small)
            }
        }
    }

    private var editorPanel: some View {
        HermesSurfacePanel(
            title: "Content",
            subtitle: "Loaded from the active host."
        ) {
            ZStack {
                TextEditor(text: $text)
                    .font(.system(.body, design: .monospaced))
                    .disabled(isLoading || !hasLoaded)
                    .scrollContentBackground(.hidden)
                    .padding(10)
                    .background(Color(NSColor.textBackgroundColor))
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    .overlay {
                        RoundedRectangle(cornerRadius: 12, style: .continuous)
                            .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)

                if isLoading {
                    HermesLoadingOverlay()
                } else if !hasLoaded {
                    ContentUnavailableView(
                        L10n.string("Loading file"),
                        systemImage: "doc.text",
                        description: Text(L10n.string("Reading from the active Hermes machine."))
                    )
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                }
            }
            .frame(maxWidth: .infinity, minHeight: 360, maxHeight: .infinity)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }
}

private struct WorkspaceFileBrowserSheet: View {
    @EnvironmentObject private var appState: AppState
    @Environment(\.dismiss) private var dismiss
    @State private var pathText = ""
    @State private var didLoadInitialDirectory = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .firstTextBaseline) {
                VStack(alignment: .leading, spacing: 5) {
                    Text(L10n.string("Add Remote File"))
                        .font(.title2)
                        .fontWeight(.semibold)

                    Text(L10n.string("Browse the active host."))
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                Button(L10n.string("Done")) {
                    dismiss()
                }
                .keyboardShortcut(.cancelAction)
            }

            HStack(spacing: 8) {
                TextField(L10n.string("Remote path"), text: $pathText)
                    .font(.system(.body, design: .monospaced))
                    .textFieldStyle(.roundedBorder)
                    .onSubmit {
                        browse(pathText)
                    }

                Button {
                    browse(pathText)
                } label: {
                    Label(L10n.string("Go"), systemImage: "arrow.forward")
                }
            }

            HStack(spacing: 8) {
                Button {
                    browse(appState.workspaceFileBrowserDefaultPath)
                } label: {
                    Label(L10n.string("Hermes Home"), systemImage: "house")
                }

                Button {
                    browse("~")
                } label: {
                    Label(L10n.string("Home"), systemImage: "person.crop.circle")
                }

                if let parentDisplayPath = appState.workspaceFileBrowserListing?.parentDisplayPath {
                    Button {
                        browse(parentDisplayPath)
                    } label: {
                        Label(L10n.string("Up"), systemImage: "arrow.up")
                    }
                }

                Spacer()
            }
            .controlSize(.small)

            if let errorMessage = appState.workspaceFileBrowserError {
                Text(errorMessage)
                    .font(.subheadline)
                    .foregroundStyle(.red)
                    .padding(10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color.red.opacity(0.08), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
            }

            browserContent

            HStack {
                if let listing = appState.workspaceFileBrowserListing {
                    let visibleCount = listing.entries.count
                    let total = listing.totalEntryCount
                    Text(
                        listing.isTruncated
                            ? L10n.string("Showing %@ of %@ items", "\(visibleCount)", "\(total)")
                            : L10n.string("%@ items", "\(total)")
                    )
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                Button {
                    addTypedPath()
                } label: {
                    Label(L10n.string("Add Path"), systemImage: "plus")
                }
                .disabled(pathText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
            }
        }
        .padding(22)
        .frame(width: 760, height: 560)
        .task {
            guard !didLoadInitialDirectory else { return }
            didLoadInitialDirectory = true
            let initialPath = appState.workspaceFileBrowserDefaultPath
            pathText = initialPath
            await appState.browseWorkspaceDirectory(path: initialPath)
            if let displayPath = appState.workspaceFileBrowserListing?.displayPath {
                pathText = displayPath
            }
        }
    }

    private var browserContent: some View {
        Group {
            if appState.isLoadingWorkspaceFileBrowser && appState.workspaceFileBrowserListing == nil {
                HermesLoadingState(label: "Loading files...", minHeight: 300)
            } else if let listing = appState.workspaceFileBrowserListing {
                List {
                    ForEach(listing.entries) { entry in
                        browserRow(entry)
                    }
                }
                .listStyle(.inset)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                .overlay {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
                }
            } else {
                ContentUnavailableView(
                    L10n.string("No Directory Loaded"),
                    systemImage: "folder",
                    description: Text(L10n.string("Enter a path to browse files on the active Hermes machine."))
                )
                .frame(maxWidth: .infinity, minHeight: 300)
            }
        }
    }

    private func browserRow(_ entry: RemoteDirectoryEntry) -> some View {
        HStack(spacing: 10) {
            Image(systemName: entryIcon(for: entry))
                .foregroundStyle(entry.kind == .directory ? .blue : .secondary)
                .frame(width: 18)

            VStack(alignment: .leading, spacing: 3) {
                Text(entry.name)
                    .lineLimit(1)

                Text(entryMetadata(for: entry))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }

            Spacer()

            if entry.canOpenDirectory {
                Button {
                    browse(entry.displayPath)
                } label: {
                    Label(L10n.string("Open"), systemImage: "folder")
                }
                .controlSize(.small)
            } else if entry.isTooLargeToEdit {
                Text(L10n.string("Too large to edit"))
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
            } else if isBookmarked(entry) {
                Button {
                } label: {
                    Label(L10n.string("Added"), systemImage: "checkmark")
                }
                .controlSize(.small)
                .disabled(true)
            } else if entry.canBookmark {
                Button {
                    addBookmark(entry)
                } label: {
                    Label(L10n.string("Add"), systemImage: "plus")
                }
                .controlSize(.small)
            }
        }
        .contentShape(Rectangle())
        .onTapGesture(count: 2) {
            if entry.canOpenDirectory {
                browse(entry.displayPath)
            } else if entry.canBookmark, !isBookmarked(entry) {
                addBookmark(entry)
            }
        }
    }

    private func browse(_ path: String) {
        let trimmed = path.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }

        pathText = trimmed
        Task {
            await appState.browseWorkspaceDirectory(path: trimmed)
            if let displayPath = appState.workspaceFileBrowserListing?.displayPath {
                pathText = displayPath
            }
        }
    }

    private func addTypedPath() {
        let trimmed = pathText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        appState.addWorkspaceFileBookmark(remotePath: trimmed, selectAfterAdd: false)
    }

    private func addBookmark(_ entry: RemoteDirectoryEntry) {
        appState.addWorkspaceFileBookmark(remotePath: entry.displayPath, selectAfterAdd: false)
    }

    private func isBookmarked(_ entry: RemoteDirectoryEntry) -> Bool {
        appState.bookmarkedWorkspaceFileReferences.contains { reference in
            reference.remotePath == entry.displayPath
        }
    }

    private func entryIcon(for entry: RemoteDirectoryEntry) -> String {
        switch entry.kind {
        case .directory:
            return "folder"
        case .file:
            return "doc.text"
        case .symlink:
            return "link"
        case .other:
            return "questionmark.square"
        }
    }

    private func entryMetadata(for entry: RemoteDirectoryEntry) -> String {
        var parts: [String] = []

        switch entry.kind {
        case .directory:
            parts.append(L10n.string("Folder"))
        case .file:
            if let size = entry.size {
                parts.append(ByteCountFormatter.string(fromByteCount: size, countStyle: .file))
            } else {
                parts.append(L10n.string("File"))
            }
        case .symlink:
            parts.append(L10n.string("Link"))
        case .other:
            parts.append(L10n.string("Other"))
        }

        if entry.isSymlink, entry.kind != .symlink {
            parts.append(L10n.string("Link"))
        }

        if let modifiedDate = entry.modifiedDate {
            parts.append(DateFormatters.relativeFormatter().localizedString(for: modifiedDate, relativeTo: .now))
        }

        if !entry.isReadable {
            parts.append(L10n.string("No read access"))
        }

        return parts.joined(separator: " / ")
    }
}
