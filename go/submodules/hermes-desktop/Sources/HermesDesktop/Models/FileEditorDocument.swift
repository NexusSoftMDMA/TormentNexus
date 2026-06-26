import Foundation

struct FileEditorDocument {
    let fileID: String
    var title: String
    var remotePath: String
    var content: String = ""
    var originalContent: String = ""
    var remoteContentHash: String?
    var isLoading = false
    var errorMessage: String?
    var lastSavedAt: Date?
    var hasLoaded = false

    var isDirty: Bool {
        content != originalContent
    }

    mutating func discardChanges() {
        content = originalContent
    }
}
