import Foundation
import Testing
@testable import HermesDesktop

struct WorkspaceFileModelsTests {
    @Test
    func bookmarkUsesReadableTitleAndStableFileID() {
        let bookmark = WorkspaceFileBookmark(
            id: UUID(uuidString: "3C2E63A9-6A5B-4F10-8AF5-58434840904E")!,
            workspaceScopeFingerprint: "host|alice||~/.hermes",
            remotePath: "  ~/.hermes/memories/NOTES.md  "
        )

        #expect(bookmark.remotePath == "~/.hermes/memories/NOTES.md")
        #expect(bookmark.displayTitle == "NOTES.md")
        #expect(bookmark.fileID == "bookmark:3C2E63A9-6A5B-4F10-8AF5-58434840904E")
    }

    @Test
    func customBookmarkTitleIsTrimmedAndCanonicalFilesKeepPinnedIDs() {
        let bookmark = WorkspaceFileBookmark(
            workspaceScopeFingerprint: "host|alice||~/.hermes",
            remotePath: "/srv/hermes/context.md",
            title: "  Shared Context  "
        )
        let reference = WorkspaceFileReference.bookmark(bookmark)

        #expect(bookmark.displayTitle == "Shared Context")
        #expect(reference.id == bookmark.fileID)
        #expect(reference.remotePath == "/srv/hermes/context.md")
        #expect(RemoteTrackedFile.memory.workspaceFileID == "canonical:memory")
    }

    @Test
    func bookmarkGroupsAreDerivedFromParentDirectoryAndSorted() {
        let scope = "host|alice||~/.hermes"
        let zeta = WorkspaceFileBookmark(
            id: UUID(uuidString: "94204522-EC09-49A1-A10A-EDC8FB47E60B")!,
            workspaceScopeFingerprint: scope,
            remotePath: "/srv/hermes/prompts/zeta.md"
        )
        let alpha = WorkspaceFileBookmark(
            id: UUID(uuidString: "362D301A-EF75-42CE-8908-B73D3F4095B5")!,
            workspaceScopeFingerprint: scope,
            remotePath: "/srv/hermes/prompts/alpha.md"
        )
        let memory = WorkspaceFileBookmark(
            id: UUID(uuidString: "6F49D7D8-69BF-43D6-B453-8FF15B8F8A99")!,
            workspaceScopeFingerprint: scope,
            remotePath: "/srv/hermes/memories/context.md"
        )
        let canonical = WorkspaceFileReference.canonical(
            .memory,
            remotePath: "~/.hermes/memory.md"
        )

        let groups = WorkspaceFileBookmarkGroup.groups(for: [
            WorkspaceFileReference.bookmark(zeta),
            canonical,
            WorkspaceFileReference.bookmark(memory),
            WorkspaceFileReference.bookmark(alpha)
        ])

        #expect(groups.map(\.directoryPath) == [
            "/srv/hermes/memories",
            "/srv/hermes/prompts"
        ])
        #expect(groups.map(\.title) == ["memories", "prompts"])
        #expect(groups[1].references.map(\.title) == ["alpha.md", "zeta.md"])
        #expect(groups.flatMap(\.references).allSatisfy { $0.bookmarkID != nil })
    }

    @Test
    func bookmarkParentDirectoryHandlesCommonRemotePathShapes() {
        #expect(WorkspaceFileBookmarkGroup.parentDirectoryPath(for: "~/.zshrc") == "~")
        #expect(WorkspaceFileBookmarkGroup.parentDirectoryPath(for: "~/notes/today.md") == "~/notes")
        #expect(WorkspaceFileBookmarkGroup.parentDirectoryPath(for: "/README.md") == "/")
        #expect(WorkspaceFileBookmarkGroup.parentDirectoryPath(for: "README.md") == ".")
        #expect(WorkspaceFileBookmarkGroup.parentDirectoryPath(for: " /srv/app/config.json/ ") == "/srv/app")

        #expect(WorkspaceFileBookmarkGroup.displayTitle(forDirectoryPath: "/srv/app") == "app")
        #expect(WorkspaceFileBookmarkGroup.displayTitle(forDirectoryPath: "~") == "~")
        #expect(WorkspaceFileBookmarkGroup.displayTitle(forDirectoryPath: "/") == "/")
    }

    @Test
    func directoryEntryBlocksBookmarksAboveEditableLimit() {
        let entry = RemoteDirectoryEntry(
            name: "large.log",
            path: "/tmp/large.log",
            displayPath: "~/large.log",
            kind: .file,
            size: WorkspaceFileLimits.maxEditableFileBytes + 1,
            modifiedAt: nil,
            isReadable: true,
            isWritable: true,
            isSymlink: false
        )

        #expect(entry.isTooLargeToEdit)
        #expect(!entry.canBookmark)
    }

    @Test
    func directoryEntryDoesNotTreatSymlinksAsPlainEditableFiles() {
        let entry = RemoteDirectoryEntry(
            name: "config-link.json",
            path: "/tmp/config-link.json",
            displayPath: "~/config-link.json",
            kind: .symlink,
            size: 128,
            modifiedAt: nil,
            isReadable: true,
            isWritable: true,
            isSymlink: true
        )

        #expect(!entry.canBookmark)
        #expect(!entry.canOpenDirectory)
    }
}
