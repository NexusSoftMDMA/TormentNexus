import SwiftUI

/// Project root the dashboard widgets resolve relative `path` fields against.
/// Set by `ProjectsView` from the currently-selected project; nil when no
/// project is active. v2.7+ file-reading widgets (markdown_file, log_tail,
/// image-local) read this via the environment.
private struct SelectedProjectRootKey: EnvironmentKey {
    static let defaultValue: String? = nil
}

extension EnvironmentValues {
    var selectedProjectRoot: String? {
        get { self[SelectedProjectRootKey.self] }
        set { self[SelectedProjectRootKey.self] = newValue }
    }
}

/// Resolves a widget's `path` field against the project root. Rejects paths
/// that escape the project boundary via `..` segments after normalization,
/// rejects absolute paths, and rejects empty / nil inputs. The returned
/// path is suitable to hand to `transport.readFile`.
///
/// Returns nil + the reason if the path is invalid; widgets surface that
/// reason via `WidgetErrorCard`.
enum WidgetPathResolver {
    enum ResolveError: Error, Equatable {
        case noProject
        case missingPath
        case absolutePath
        case escapesProject
    }

    static func resolve(_ relativePath: String?, projectRoot: String?) -> Result<String, ResolveError> {
        guard let projectRoot, !projectRoot.isEmpty else { return .failure(.noProject) }
        guard let relativePath, !relativePath.isEmpty else { return .failure(.missingPath) }
        if relativePath.hasPrefix("/") { return .failure(.absolutePath) }
        // Strip a single leading "./" — common in template-authored paths.
        let trimmed = relativePath.hasPrefix("./") ? String(relativePath.dropFirst(2)) : relativePath
        // Walk the segments and reject any "..": the project root is the
        // trust boundary, anything reaching outside it is rejected. We do
        // this BEFORE join+standardize so symlink games can't smuggle a
        // ".." through path canonicalization.
        let segments = trimmed.split(separator: "/", omittingEmptySubsequences: true)
        for s in segments where s == ".." { return .failure(.escapesProject) }
        let joined = (projectRoot as NSString).appendingPathComponent(trimmed)
        let standardized = (joined as NSString).standardizingPath
        // Belt and suspenders: ensure the standardized path is still
        // beneath projectRoot. Standardizing resolves "./" and may follow
        // symlinks; this catch checks the final string prefix.
        let rootStd = (projectRoot as NSString).standardizingPath
        if !standardized.hasPrefix(rootStd) {
            return .failure(.escapesProject)
        }
        return .success(standardized)
    }
}

extension WidgetPathResolver.ResolveError {
    var userMessage: String {
        switch self {
        case .noProject:       return "No project selected."
        case .missingPath:     return "Missing required `path` field."
        case .absolutePath:    return "Path must be relative to the project root, not absolute."
        case .escapesProject:  return "Path escapes the project root (`..` segments are not allowed)."
        }
    }
}
